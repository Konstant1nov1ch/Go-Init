package work

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"go-init-gen/internal/api/grpc"
	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine"

	"gitlab.com/go-init/go-init-common/default/kafka"
	"gitlab.com/go-init/go-init-common/default/logger"
)

type Worker struct {
	ctx             context.Context
	cancel          context.CancelFunc
	log             *logger.Logger
	kafkaConsumer   *kafka.ClientConfig
	publisherClient *grpc.PublisherClient
	messageChan     chan []byte
	wg              sync.WaitGroup
	workerCount     int
	isRunning       bool
	mu              sync.Mutex
}

// Add a new struct for CloudEvent format
type CloudEvent struct {
	SpecVersion     string          `json:"specversion"`
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	DataContentType string          `json:"datacontenttype"`
	Time            string          `json:"time"`
	Data            json.RawMessage `json:"data"`
}

func NewWorker(ctx context.Context, log *logger.Logger, kafkaConsumer *kafka.ClientConfig, publisherClient *grpc.PublisherClient) *Worker {
	workerCtx, cancel := context.WithCancel(ctx)
	return &Worker{
		ctx:             workerCtx,
		cancel:          cancel,
		log:             log,
		kafkaConsumer:   kafkaConsumer,
		publisherClient: publisherClient,
		messageChan:     make(chan []byte, 100),
		workerCount:     5, // Configurable worker count
	}
}

// Start initializes and starts the worker pool to process messages
func (w *Worker) Start() error {
	w.mu.Lock()
	if w.isRunning {
		w.mu.Unlock()
		return nil
	}
	w.isRunning = true
	w.mu.Unlock()

	w.log.Info("Starting worker pool")
	w.startWorkerPool()

	// Keep the Start method running until context is canceled
	<-w.ctx.Done()
	w.log.Info("Stopping worker pool due to context cancellation")

	// Close the message channel to signal workers to stop
	close(w.messageChan)

	// Wait for all workers to complete
	w.wg.Wait()
	w.log.Info("Worker pool stopped")
	return nil
}

func (w *Worker) Stop() error {
	w.log.Info("Stopping worker...")
	w.cancel()
	return nil
}

func (w *Worker) startWorkerPool() {
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}
}

func (w *Worker) worker(id int) {
	defer w.wg.Done()
	w.log.Info(fmt.Sprintf("Starting worker %d", id))

	for message := range w.messageChan {
		// Check if context is canceled before processing
		select {
		case <-w.ctx.Done():
			w.log.Info(fmt.Sprintf("Worker %d stopping due to context cancellation", id))
			return
		default:
			// Debug log the raw message
			w.log.Debug(fmt.Sprintf("Worker %d received message: %s", id, string(message)))

			// First unmarshal the CloudEvent structure
			var cloudEvent CloudEvent
			if err := json.Unmarshal(message, &cloudEvent); err != nil {
				w.log.Error(fmt.Sprintf("Worker %d failed to unmarshal CloudEvent: %v", id, err))
				continue
			}

			// Debug log the extracted data
			w.log.Debug(fmt.Sprintf("Worker %d extracted CloudEvent data: %s", id, string(cloudEvent.Data)))

			// Then unmarshal the data field into ProcessTemplate
			var template eventdata.ProcessTemplate
			if err := json.Unmarshal(cloudEvent.Data, &template); err != nil {
				w.log.Error(fmt.Sprintf("Worker %d failed to unmarshal ProcessTemplate from CloudEvent data: %v", id, err))
				continue
			}

			// Debug log the parsed template
			w.log.Debug(fmt.Sprintf("Worker %d parsed template - ID: %s, Status: %s, Name: %s",
				id, template.ID, template.Status, template.Data.Name))

			// Process the message
			w.log.Info(fmt.Sprintf("Worker %d processing message with template ID: %s", id, template.ID))
			archive, err := w.generateArchive(template)
			if err != nil {
				w.log.Error(fmt.Sprintf("Worker %d failed to generate archive: %v", id, err))
				continue
			}

			err = w.streamArchive(archive, template.ID)
			if err != nil {
				w.log.Error(fmt.Sprintf("Worker %d failed to stream archive: %v", id, err))
			}
		}
	}

	w.log.Info(fmt.Sprintf("Worker %d stopped", id))
}

func (w *Worker) generateArchive(template eventdata.ProcessTemplate) ([]byte, error) {
	// Log template data for debugging
	w.log.Info(fmt.Sprintf("Generating archive for template ID: %s", template.ID))
	w.log.Debug(fmt.Sprintf("Template details: Status=%s, Name=%s", template.Status, template.Data.Name))

	// Log endpoints
	if len(template.Data.Endpoints) > 0 {
		for i, endpoint := range template.Data.Endpoints {
			w.log.Debug(fmt.Sprintf("Endpoint %d: Protocol=%s, Role=%s",
				i, endpoint.Protocol, endpoint.Role))
		}
	}

	// Log database info
	w.log.Debug(fmt.Sprintf("Database: Type=%s, DDL=%s",
		template.Data.Database.Type, template.Data.Database.DDL))

	// Log docker info
	w.log.Debug(fmt.Sprintf("Docker: Registry=%s, ImageName=%s",
		template.Data.Docker.Registry, template.Data.Docker.ImageName))

	// Log advanced settings if present
	if template.Data.Advanced != nil {
		w.log.Debug(fmt.Sprintf("Advanced: EnableAuth=%t, GenSwagger=%t",
			template.Data.Advanced.EnableAuthentication, template.Data.Advanced.GenerateSwaggerDocs))
	}

	// Create generator and generate template
	gen := engine.New()
	archive, err := gen.Generate(w.ctx, &template)
	if err != nil {
		return nil, fmt.Errorf("failed to generate template: %w", err)
	}

	w.log.Info(fmt.Sprintf("Successfully generated archive for template ID: %s", template.ID))
	return archive, nil
}

// streamArchive отправляет сгенерированный архив через gRPC в сервис publisher.
func (w *Worker) streamArchive(archive []byte, templateID string) error {
	w.log.Info("Отправка архива в publisher сервис", "templateID", templateID)

	// Используем метод потоковой передачи для отправки архива
	err := w.publisherClient.StreamArchive(context.Background(), templateID, archive)
	if err != nil {
		return fmt.Errorf("ошибка при отправке архива: %w", err)
	}

	w.log.Info("Архив успешно отправлен в publisher сервис через стриминг", "templateID", templateID)
	return nil
}

// Work implements the ConsumerWorker interface.
// It is called by the Kafka consumer when a new message is received.
func (w *Worker) Work(ctx context.Context, value []byte) error {
	// Quick validation of the message
	if len(value) == 0 {
		return fmt.Errorf("received empty message")
	}

	// Ensure the worker pool is started
	w.mu.Lock()
	isRunning := w.isRunning
	w.mu.Unlock()

	if !isRunning {
		// Auto-start the worker pool if it's not running yet
		go func() {
			if err := w.Start(); err != nil {
				w.log.Error(fmt.Sprintf("Failed to start worker pool: %v", err))
			}
		}()
	}

	// Try to send the message to the worker pool
	select {
	case w.messageChan <- value:
		// Message successfully sent to the worker pool
		return nil
	case <-ctx.Done():
		// Context was canceled while trying to send the message
		return ctx.Err()
	default:
		// Channel is full, process the message directly
		w.log.Warn("Worker pool queue is full, processing message directly")

		// First unmarshal the CloudEvent structure
		var cloudEvent CloudEvent
		if err := json.Unmarshal(value, &cloudEvent); err != nil {
			w.log.Error(fmt.Sprintf("Failed to unmarshal CloudEvent: %v", err))
			return err
		}

		// Then unmarshal the data field into ProcessTemplate
		var template eventdata.ProcessTemplate
		if err := json.Unmarshal(cloudEvent.Data, &template); err != nil {
			w.log.Error(fmt.Sprintf("Failed to unmarshal ProcessTemplate from CloudEvent data: %v", err))
			return err
		}

		w.log.Info(fmt.Sprintf("Processing template with ID: %s", template.ID))

		archive, err := w.generateArchive(template)
		if err != nil {
			w.log.Error(fmt.Sprintf("Failed to generate archive: %v", err))
			return err
		}

		// Передаем ID шаблона (который соответствует RequestUUID в Manager)
		if err := w.streamArchive(archive, template.ID); err != nil {
			w.log.Error(fmt.Sprintf("Failed to stream archive: %v", err))
			return err
		}

		return nil
	}
}
