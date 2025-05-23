# Go Init Publisher

A specialized microservice responsible for storing and serving generated template archives. The publisher service receives archives from the generator service, stores them in MinIO, and publishes events back to the manager service.

## Architecture Role

The Publisher service is the storage and delivery component in the Go Init system:

1. Receives generated template archives from the Generator service via gRPC streaming
2. Stores archives in MinIO object storage with unique identifiers
3. Records archive metadata in PostgreSQL database
4. Publishes completion events to Kafka with download links
5. Provides public access to the generated templates

## Features

- gRPC streaming interface for receiving archives
- MinIO integration for durable object storage
- PostgreSQL database for archive metadata
- Kafka event publishing for system coordination
- Configurable retention policies for archives
- Direct download links for completed templates

## Prerequisites

- Go 1.16 or later
- Docker and Docker Compose
- PostgreSQL database
- MinIO object storage
- Kafka for event publishing

## Local Setup

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd go-init-publisher
   ```

2. **Install dependencies:**
   ```bash
   make install
   ```

3. **Configure environment variables:**
   Create a `.env` file in the root directory:
   ```
   HTTP_SERVER_PORT=8082
   GRPC_SERVER_PORT=50052
   POSTGRES_HOST=localhost
   POSTGRES_PORT=5432
   POSTGRES_USER=go_init_usr
   POSTGRES_PASSWORD=1234
   POSTGRES_DB=go_init_publisher_db
   KAFKA_BROKERS=localhost:9092
   MINIO_ENDPOINT=localhost:9000
   MINIO_ACCESS_KEY=B6NZaLCC7AmyqSkdB2Rr
   MINIO_SECRET_KEY=C8gZg0GSg31mEUs3BRVj7Dh5nIZ4HCvMPEHEamYq
   MINIO_BUCKET=go-init-archives
   ```

4. **Run the service:**
   ```bash
   make run
   ```

   For development with hot reload:
   ```bash
   make run-dev
   ```

## API

### gRPC Interface

The publisher exposes a gRPC interface for receiving archives from the generator service:

```proto
service ArchiverService {
  rpc UploadArchive(stream ArchiveChunk) returns (ArchiveResponse) {}
}

message ArchiveChunk {
  bytes content = 1;
  string template_id = 2;
  string file_name = 3;
}

message ArchiveResponse {
  bool success = 1;
  string message = 2;
  string download_url = 3;
}
```

### Kafka Events

The publisher sends events to Kafka when archives are successfully stored:

- Topic: `go-init-done`
- Message format: JSON containing template ID and download URL

## Database Schema

The publisher maintains a simple database schema for tracking archives:

```sql
CREATE TABLE archives (
  id SERIAL PRIMARY KEY,
  template_id VARCHAR(36) NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  object_path VARCHAR(255) NOT NULL,
  download_url VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Storage Configuration

The publisher uses MinIO for object storage. By default, all files are stored in the `go-init-archives` bucket with public read access.

## Testing

- Run tests using:
  ```bash
  make test
  ```

- Run tests with coverage:
  ```bash
  make test-coverage
  ```

## Building

- Build binary:
  ```bash
  make build
  ```

- Build Docker image:
  ```bash
  make docker-build
  ```

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request for any improvements.

## License

This project is licensed under the MIT License.
