# KYB Platform - Business Verification & Risk Assessment

A comprehensive Know Your Business (KYB) platform built with Go, providing business verification, risk assessment, and compliance monitoring capabilities.

## 🚀 Quick Start with Supabase

The KYB Platform now uses **Supabase** for cost-effective MVP deployment during the product discovery phase.

### Prerequisites

- Go 1.22 or later
- Docker and Docker Compose
- Supabase account and project
- PostgreSQL client (optional, for database testing)

### 1. Setup Supabase

1. **Create a Supabase project:**
   - Go to [supabase.com](https://supabase.com)
   - Create a new project
   - Note your project URL and API keys

2. **Run the setup script:**
   ```bash
   ./scripts/setup_supabase.sh
   ```
   
   This script will:
   - Prompt for your Supabase credentials
   - Create a `.env` file with proper configuration
   - Test database connectivity
   - Set up Supabase configuration files

### 2. Start the Application

**Option A: Using Docker (Recommended)**
```bash
# Development
docker-compose -f docker-compose.dev.yml up

# Production
docker-compose up
```

**Option B: Running Locally**
```bash
# Install dependencies
go mod download

# Run the application
go run ./cmd/api
```

### 3. Access the Platform

- **Application:** http://localhost:8080
- **API Documentation:** http://localhost:8080/docs
- **Health Check:** http://localhost:8080/health
- **Metrics:** http://localhost:8080/metrics

### 4. Monitoring & Observability

- **Grafana Dashboard:** http://localhost:3000 (admin/admin)
- **Prometheus Metrics:** http://localhost:9090
- **Jaeger Tracing:** http://localhost:16686 (development only)

## 🏗️ Architecture

### Provider Abstraction

The platform uses a **provider abstraction layer** that allows easy switching between different cloud providers:

- **Current:** Supabase (MVP phase)
- **Future:** AWS, GCP, Azure (enterprise phase)

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │  Authentication │    │   Database      │
│   (Go/HTTP)     │◄──►│   (Supabase)    │◄──►│   (PostgreSQL)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Classification │    │  Risk Assessment│    │   Compliance    │
│     Service     │    │     Service     │    │     Service     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Cache       │    │   Monitoring    │    │   Storage       │
│   (Supabase)    │    │  (Prometheus)   │    │   (Supabase)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📊 Features

### Core KYB Features
- **Business Classification:** Industry code mapping and categorization
- **Risk Assessment:** Multi-factor risk scoring and analysis
- **Compliance Monitoring:** Regulatory framework tracking
- **Audit Trail:** Complete activity logging and traceability

### Technical Features
- **Provider Abstraction:** Easy migration between cloud providers
- **Row Level Security:** Supabase RLS for data protection
- **Real-time Updates:** Supabase real-time subscriptions
- **API-First Design:** RESTful API with OpenAPI documentation
- **Monitoring:** Prometheus metrics and Grafana dashboards
- **Tracing:** Distributed tracing with Jaeger

## 🔧 Configuration

### Environment Variables

The platform uses environment variables for configuration. Key variables:

```bash
# Provider Selection
PROVIDER_DATABASE=supabase
PROVIDER_AUTH=supabase
PROVIDER_CACHE=supabase
PROVIDER_STORAGE=supabase

# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_API_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
SUPABASE_JWT_SECRET=your_jwt_secret

# Database Configuration
DB_HOST=db.your-project.supabase.co
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=your_password
DB_DATABASE=postgres
DB_SSL_MODE=require
```

### Configuration Files

- **Development:** `configs/development.env`
- **Production:** `configs/production.env`
- **Example:** `env.example`

## 🗄️ Database Schema

The platform uses PostgreSQL with the following key tables:

- **users:** User accounts and profiles
- **businesses:** Business entities and information
- **classifications:** Industry classifications and codes
- **risk_assessments:** Risk analysis results
- **compliance_status:** Compliance tracking
- **audit_logs:** Activity audit trail
- **cache:** Application caching layer

### Row Level Security (RLS)

All tables have RLS policies enabled for data protection:
- Users can only access their own data
- Admins have broader access for management
- Automatic field population (created_by, updated_at)

## 🔄 Migration Strategy

### Current: Supabase (MVP Phase)
- **Cost:** ~$25/month
- **Features:** PostgreSQL, Auth, Real-time, Storage
- **Benefits:** Fast setup, low cost, managed services

### Future: AWS (Enterprise Phase)
- **Cost:** ~$200-500/month
- **Features:** Advanced analytics, ML, global distribution
- **Benefits:** Enterprise features, scalability, compliance

### Provider Switching
The platform supports easy migration between providers:
```bash
# Switch to AWS
PROVIDER_DATABASE=aws
PROVIDER_AUTH=aws
PROVIDER_CACHE=aws
PROVIDER_STORAGE=aws
```

## 🧪 Testing

### Run Tests
```bash
# Unit tests
go test ./...

# Integration tests
go test ./test/integration/...

# Performance tests
go test ./test/performance/...
```

### Test Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📚 API Documentation

### OpenAPI Specification
- **Swagger UI:** http://localhost:8080/docs
- **OpenAPI JSON:** http://localhost:8080/docs/swagger.json

### Key Endpoints
- `POST /v1/auth/register` - User registration
- `POST /v1/auth/login` - User authentication
- `POST /v1/businesses/classify` - Business classification
- `POST /v1/businesses/assess-risk` - Risk assessment
- `GET /v1/businesses/{id}` - Get business details
- `GET /v1/health` - Health check

## 🚀 Deployment

### Development
```bash
# Using Docker Compose
docker-compose -f docker-compose.dev.yml up

# Using local Go
go run ./cmd/api
```

### Production
```bash
# Using Docker Compose
docker-compose up -d

# Using Kubernetes
kubectl apply -f deployments/kubernetes/
```

### Environment-Specific Configurations
- **Development:** Local development with hot reload
- **Staging:** Pre-production testing environment
- **Production:** Live production environment

## 📊 Monitoring & Observability

### Metrics (Prometheus)
- Request rates and latencies
- Error rates and types
- Database connection pool stats
- Cache hit/miss ratios

### Logging
- Structured JSON logging
- Request tracing with correlation IDs
- Error tracking and alerting

### Dashboards (Grafana)
- **KYB Business Dashboard:** Business metrics and trends
- **KYB Performance Dashboard:** System performance metrics

## 🔒 Security

### Authentication & Authorization
- JWT-based authentication
- Role-based access control (RBAC)
- API key management
- Session management

### Data Protection
- Row Level Security (RLS)
- Encrypted data transmission (TLS)
- Secure password handling
- Audit logging

### Rate Limiting
- Request rate limiting
- Authentication-specific limits
- IP-based blocking

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Development Guidelines
- Follow Go best practices
- Write comprehensive tests
- Update documentation
- Use conventional commits

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Documentation
- [API Documentation](./docs/api/)
- [User Guides](./docs/user-guides/)
- [Architecture Documentation](./docs/architecture.md)

### Troubleshooting
- Check application logs: `docker-compose logs kyb-platform`
- Test database connection: `PGPASSWORD='password' psql -h host -U user -d database`
- Verify environment variables: `cat .env`

### Getting Help
- Create an issue on GitHub
- Check the documentation
- Review the troubleshooting guide

---

**Built with ❤️ using Go and Supabase**
