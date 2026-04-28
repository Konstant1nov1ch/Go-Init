package closer

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.com/go-init/go-init-common/default/logger"
)

var (
	// Дефолтные сигналы для вызова CloseAll у объекта Closer
	defaultCloseSignals = []os.Signal{
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}
	// приватный инстанс Closer
	instance *Closer
)

// InitCloser инициализирует singleton-инстанс Closer.
func InitCloser(log *logger.Logger) {
	instance = New(log, defaultCloseSignals...)
}

// Add добавляет функции завершения.
func Add(f ...func() error) {
	instance.Add(f...)
}

// Wait ожидает завершения всех функций завершения.
func Wait() {
	instance.Wait()
}

// CloseAll выполняет все функции завершения.
func CloseAll() {
	instance.CloseAll()
}

// Closer агрегирует функции завершения, выполняемые при остановке сервиса.
type Closer struct {
	logger *logger.Logger
	mutex  sync.Mutex
	once   sync.Once
	done   chan struct{}
	funcs  []func() error
}

// New создает новый инстанс Closer.
func New(log *logger.Logger, signals ...os.Signal) *Closer {
	c := &Closer{
		logger: log,
		done:   make(chan struct{}),
	}

	c.logger.Info("Closer initialized")

	if len(signals) > 0 {
		go c.listenForSignals(signals...)
	}
	return c
}

// Add добавляет функции завершения.
func (c *Closer) Add(f ...func() error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.funcs = append(c.funcs, f...)
}

// Wait блокирует выполнение до завершения всех функций завершения.
func (c *Closer) Wait() {
	<-c.done
}

// CloseAll выполняет все функции завершения один раз.
func (c *Closer) CloseAll() {
	c.once.Do(func() {
		c.logger.Info("Starting service shutdown")
		defer close(c.done)

		c.mutex.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mutex.Unlock()

		errCh := make(chan error, len(funcs))
		var wg sync.WaitGroup

		for _, f := range funcs {
			wg.Add(1)
			go func(f func() error) {
				defer wg.Done()
				if err := f(); err != nil {
					errCh <- err
				}
			}(f)
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			c.logger.Error("Error during shutdown", logger.Error(err))
		}

		c.logger.Info("Service shutdown complete")
	})
}

// listenForSignals ждет системные сигналы и вызывает CloseAll.
func (c *Closer) listenForSignals(signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	<-ch
	signal.Stop(ch)
	c.CloseAll()
}
