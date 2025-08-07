# KYB Tool - Enterprise-Grade Know Your Business Platform

## Overview

KYB Tool is an enterprise-grade platform that provides comprehensive business verification, classification, risk assessment, and compliance monitoring capabilities. Built with Go 1.24+ and following Clean Architecture principles, the platform offers robust APIs for business intelligence and regulatory compliance.

## Features

- **Business Classification**: AI-powered business categorization with industry code mapping
- **Risk Assessment**: Comprehensive risk scoring and monitoring
- **Compliance Framework**: SOC 2, PCI DSS, GDPR, and regional compliance tracking
- **Authentication & Authorization**: JWT-based authentication with RBAC
- **Observability**: Distributed tracing, metrics, and structured logging
- **Cloud-Ready**: Designed for scalable cloud deployment

## Architecture

The project follows Clean Architecture principles with clear separation of concerns:

```
├── cmd/api/           # Application entry points
├── internal/          # Core application logic
│   ├── auth/         # Authentication and authorization
│   ├── classification/ # Business classification engine
│   ├── risk/         # Risk assessment engine
│   ├── compliance/   # Compliance framework
│   ├── database/     # Data access layer
│   ├── api/          # HTTP handlers and middleware
│   ├── config/       # Configuration management
│   └── observability/ # Logging, metrics, tracing
├── pkg/              # Shared utilities and packages
│   ├── validators/   # Input validation utilities
│   └── encryption/   # Encryption utilities
├── docs/api/         # API documentation
├── deployments/      # Docker and deployment configs
└── scripts/          # Build and deployment scripts
```

## Prerequisites

- Go 1.24+ (installed via `brew install go`)
- Git
- Docker (for containerized deployment)
- Make (for build automation)

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd kyb-tool
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Set Up Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 4. Run the Application

```bash
# Development mode
go run cmd/api/main.go

# Or using make
make run
```

### 5. Verify Installation

```bash
curl http://localhost:8080/health
```

## Development

### Project Structure

The project follows Go best practices and Clean Architecture:

- **cmd/**: Application entry points
- **internal/**: Private application code
- **pkg/**: Public libraries that can be imported by other projects
- **docs/**: Documentation
- **deployments/**: Infrastructure and deployment configurations

### Development Commands

```bash
# Run tests
make test

# Run linting
make lint

# Build application
make build

# Run with hot reload (requires air)
make dev
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/auth/...
```

## API Documentation

Once the application is running, visit:
- Swagger UI: `http://localhost:8080/docs`
- API Documentation: `http://localhost:8080/api/docs`

## Configuration

The application uses environment-based configuration. Key environment variables:

- `PORT`: Server port (default: 8080)
- `ENV`: Environment (dev/staging/prod)
- `DB_URL`: Database connection string
- `JWT_SECRET`: JWT signing secret
- `LOG_LEVEL`: Logging level (debug/info/warn/error)

## Deployment

### Docker

```bash
# Build image
docker build -t kyb-tool .

# Run container
docker run -p 8080:8080 kyb-tool
```

### Kubernetes

```bash
kubectl apply -f deployments/k8s/
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

[License information to be added]

## Support

For support and questions, please refer to the documentation or create an issue in the repository.
