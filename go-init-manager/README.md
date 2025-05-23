# Go Init Manager

The Go Init Manager is a service designed to manage the initialization of service templates, handling database configurations, Docker settings, and advanced options. It provides a GraphQL API for creating and managing service templates and coordinates with other microservices via Kafka.

## Features

- Create and manage service templates
- Configure endpoints, databases, and Docker settings
- Advanced options for authentication and documentation generation
- Integration with Kafka for event processing
- GraphQL API for client interactions
- PostgreSQL storage for template metadata

## Architecture Role

The Manager service is the entry point for client requests in the Go Init system:

1. Receives template generation requests from clients
2. Stores template parameters in PostgreSQL
3. Publishes generation events to Kafka
4. Exposes gRPC endpoints for other services to fetch details 
5. Consumes completion events from Kafka to update template status
6. Provides clients with download links for completed templates

## Prerequisites

- Go 1.16 or later
- Docker and Docker Compose
- PostgreSQL
- Kafka

## Local Setup

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd go-init-manager
   ```

2. **Install dependencies:**
   ```bash
   make install
   ```

3. **Configure environment variables:**
   Create a `.env` file in the root directory:
   ```
   HTTP_SERVER_PORT=8080
   POSTGRES_HOST=localhost
   POSTGRES_PORT=5432
   POSTGRES_USER=go_init_usr
   POSTGRES_PASSWORD=1234
   POSTGRES_DB=go_init_manager_db
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

## API Usage

The Manager service exposes a GraphQL API for client interactions.

### GraphQL Endpoint

Access the GraphQL API at `http://localhost:8080/graphql`

### Example Queries

#### CreateTemplate Mutation

```graphql
mutation CreateTemplate {
    createTemplate(
        input: {
            name: "MyServiceTemplate"
            endpoints: [{ protocol: GRPC, role: SERVER }, { protocol: REST, role: CLIENT }]
            database: {
                type: POSTGRESQL
                ddl: "CREATE TABLE example (id SERIAL PRIMARY KEY, name VARCHAR(100));"
            }
            docker: { registry: "docker.io/myuser", imageName: "my-service-image" }
            advanced: { enableAuthentication: true, generateSwaggerDocs: true }
        }
    ) {
        success
        message
        template {
            id
            name
            endpoints {
                protocol
                role
            }
            database {
                type
                ddl
            }
            docker {
                registry
                imageName
            }
            advanced {
                enableAuthentication
                generateSwaggerDocs
            }
            createdAt
            zipUrl
        }
    }
}
```

#### Query Template Status

```graphql
query GetTemplate {
    template(id: "template-id") {
        id
        name
        status
        zipUrl
    }
}
```

## Configuration

The service uses a YAML-based configuration system. Core settings are managed in `config/config.go`.

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
