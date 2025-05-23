# Go Init Generator

A specialized microservice that generates Go code templates based on user specifications. This service consumes requests from Kafka, creates customized template archives, and streams them to the publisher service.

## Architecture Role

The Generator service is the code creation engine in the Go Init system:

1. Consumes template generation requests from Kafka
2. Optionally retrieves additional details from the Manager service via gRPC
3. Generates customized Go code templates based on user specifications
4. Streams the generated archive to the Publisher service via gRPC
5. Supports various template features (endpoints, databases, Docker, etc.)

## Features

- Template generation from configurable options
- Support for multiple protocol types (REST, GraphQL, gRPC)
- Database integration setup (PostgreSQL, MySQL, MongoDB)
- Docker container configuration generation
- Customizable authentication and documentation options
- Streaming archive generation to publisher service

## Prerequisites

- Go 1.16 or later
- Docker and Docker Compose
- Kafka for event consumption

## Local Setup

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd go-init-generator
   ```

2. **Install dependencies:**
   ```bash
   make install
   ```

3. **Configure environment variables:**
   Create a `.env` file in the root directory:
   ```
   HTTP_SERVER_PORT=8081
   GRPC_SERVER_PORT=50051
   KAFKA_BROKERS=localhost:9092
   ```

4. **Run the service:**
   ```bash
   make run
   ```

   For development with hot reload:
   ```bash
   make run-dev
   ```

## Configuration

### Environment Variables

The generator uses these environment variables:

```
# Static files to always include
STATIC_FILES_ARRAY=[
  "build/docker/Dockerfile",
  "README.md",
  "go.mod"
]

# Dynamic files that are conditionally included
DYNAMIC_FILES_ARRAY=[
  "api/graphql/schema.graphql",
  "api/grpc/service.proto"
]

# Save generated archives for debugging
GENERATOR_SAVE_ARCHIVE_LOCALLY=true
```

## Template Customization

### File Structure

The generator creates a template with the following structure:

```
├── api/                    # API definitions (REST, GraphQL, gRPC)
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── config/             # Configuration management
│   ├── domain/             # Business logic and domain models
│   ├── repository/         # Data access layer
│   └── service/            # Service implementations
├── pkg/                    # Public library code
├── build/                  # Build-related files
│   └── docker/             # Docker configuration
└── config/                 # Configuration files
```

### Modifying Templates

To add or modify templates:

1. Update the template files in the generator's internal template repository
2. Modify the template processor to include new options
3. Update the serialization and streaming logic if necessary

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