# ADOL POS - Multi-Tenant Point of Sale System

ADOL is a comprehensive, multi-tenant Point of Sale (POS) system built with Go, designed to serve multiple independent organizations through a single SaaS platform. The system provides complete data isolation, subscription-based feature gating, and scalable architecture.

## 🚀 Features

### Core POS Functionality
- **Product Management**: Full inventory management with categories, pricing, and stock tracking
- **Sales Processing**: Complete sales workflow with multiple payment methods
- **User Management**: Role-based access control (Admin, Manager, Cashier, Employee)
- **Reporting**: Basic and advanced analytics with export capabilities
- **Invoice Generation**: Professional invoice creation and management

### Multi-Tenant Architecture
- **Complete Tenant Isolation**: Database-level and application-level data separation
- **Subscription Management**: Three-tier subscription system (Starter, Professional, Enterprise)
- **Feature Gating**: Subscription-based access control to advanced features
- **Usage Monitoring**: Real-time usage tracking with automatic limit enforcement
- **Trial Management**: 30-day free trial for new tenants
- **Custom Branding**: Tenant-specific configuration and branding options

### Enterprise Features
- **API Access**: RESTful API for integrations (Enterprise plan)
- **Multi-Location Support**: Manage multiple business locations (Professional+)
- **Advanced Reporting**: Detailed analytics and business intelligence (Professional+)
- **Webhook Support**: Real-time event notifications
- **Audit Logging**: Comprehensive audit trail for compliance
- **High Availability**: Scalable architecture with monitoring

## 🏗️ Architecture

### Technology Stack
- **Backend**: Go 1.21+ with Gin web framework
- **Database**: PostgreSQL 14+ with Row Level Security (RLS)
- **Authentication**: JWT with tenant-aware claims
- **Logging**: Structured logging with tenant context
- **Monitoring**: Real-time usage and performance tracking
- **Testing**: Comprehensive unit and integration tests

### Multi-Tenancy Pattern
- **Pattern**: Shared Database with Tenant Isolation
- **Isolation**: Row Level Security (RLS) + Application filtering
- **Identification**: UUID-based tenant identification
- **Resolution**: Multiple methods (headers, subdomains, domains)

## 📋 Subscription Plans

### 🎯 Starter Plan (Free)
- **Cost**: Free with 1% transaction fee
- **Users**: Up to 2 users
- **Products**: Unlimited
- **Sales**: Unlimited
- **Features**: Basic POS, Inventory, Basic Reporting
- **Support**: Community support

### 💼 Professional Plan (Rp300,000/month)
- **Cost**: Rp300,000 per month
- **Users**: Up to 10 users
- **Products**: Unlimited
- **Sales**: Unlimited
- **Features**: All Starter + Advanced Reporting, Multi-Location
- **Support**: Email support

### 🏢 Enterprise Plan (Rp1,500,000/month)
- **Cost**: Rp1,500,000 per month
- **Users**: Unlimited
- **Products**: Unlimited
- **Sales**: Unlimited
- **API Calls**: 10,000 per month
- **Features**: All Professional + API Access, Custom Integrations
- **Support**: Priority support with SLA

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 14 or higher
- Git

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/nicklaros/adol.git
   cd adol
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Setup environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Setup database**:
   ```bash
   createdb adol_pos
   go run cmd/migrate/main.go up
   ```

5. **Start the server**:
   ```bash
   go run cmd/api/main.go
   ```

The server will start on `http://localhost:8080`

### Register Your First Tenant

```bash
curl -X POST http://localhost:8080/api/v1/tenants/register \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_name": "My Store",
    "admin_username": "admin",
    "admin_email": "admin@mystore.com",
    "admin_password": "secure_password",
    "admin_first_name": "John",
    "admin_last_name": "Doe"
  }'
```

## 📚 Documentation

- [**Multi-Tenant Architecture**](docs/MULTI_TENANT_ARCHITECTURE.md) - Detailed architecture overview
- [**API Examples**](docs/API_EXAMPLES.md) - Practical API usage examples
- [**Deployment Guide**](docs/DEPLOYMENT.md) - Production deployment instructions
- [**Configuration Reference**](docs/CONFIGURATION.md) - Environment variables and settings

## 🔗 API Usage

### Authentication

All API requests require tenant context and authentication:

```bash
# Using tenant ID header
curl -X GET http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer your_jwt_token" \
  -H "X-Tenant-ID: tenant-uuid"

# Using tenant slug header
curl -X GET http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer your_jwt_token" \
  -H "X-Tenant-Slug: my-store"

# Using subdomain (in production)
curl -X GET https://my-store.yourdomain.com/api/v1/products \
  -H "Authorization: Bearer your_jwt_token"
```

### Key Endpoints

- `POST /api/v1/tenants/register` - Register new tenant
- `POST /api/v1/auth/login` - User authentication
- `GET /api/v1/subscription` - Get subscription details
- `POST /api/v1/subscription/upgrade` - Upgrade subscription
- `GET /api/v1/products` - List products (tenant-filtered)
- `POST /api/v1/sales` - Create sale
- `GET /api/v1/reports/advanced/*` - Advanced reporting (Pro+)

## 🧪 Testing

### Run Unit Tests
```bash
go test ./internal/domain/entities -v
go test ./internal/application/usecases -v
```

### Run Integration Tests
```bash
go test ./tests/integration -v
```

### Test Multi-Tenant Isolation
```bash
# Run the multi-tenant integration test suite
go test ./tests/integration -run TestMultiTenantIntegrationTestSuite -v
```

## 🌍 Environment Variables

Key configuration options:

```bash
# Server Configuration
SERVER_PORT=8080

# Database Configuration  
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=adol_pos

# JWT Configuration
JWT_SECRET_KEY=your-secret-key-min-32-chars
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h
JWT_INCLUDE_TENANT_CLAIMS=true

# Multi-Tenancy Features
FEATURE_ENABLE_MULTI_TENANCY=true
FEATURE_ENABLE_SUBSCRIPTIONS=true
FEATURE_ENABLE_USAGE_LIMITS=true
FEATURE_ENABLE_TRIAL_PERIODS=true

# Tenant Configuration
TENANT_DEFAULT_TRIAL_DAYS=30
TENANT_ALLOW_SUBDOMAINS=true
```

See [Configuration Reference](docs/CONFIGURATION.md) for complete options.

## 🔒 Security Features

### Data Isolation
- **Row Level Security (RLS)**: Database-level tenant isolation
- **Application Filtering**: Additional security layer in application code
- **Tenant Context Validation**: Automatic tenant boundary enforcement

### Authentication & Authorization
- **JWT Tokens**: Secure token-based authentication with tenant claims
- **Role-based Access**: Granular permissions per user role
- **Feature Gates**: Subscription-based feature access control

### Audit & Monitoring
- **Tenant-aware Logging**: All actions logged with tenant context
- **Usage Tracking**: Real-time monitoring of resource usage
- **Security Events**: Automatic logging of security-relevant events

## 📈 Monitoring & Observability

### Metrics
- Tenant registration and churn rates
- Usage patterns and limit violations
- Performance metrics per tenant
- Revenue and subscription analytics

### Logging
All logs include tenant context for easy filtering and debugging:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "message": "Product created",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_slug": "my-store",
  "user_id": "770g8400-e29b-41d4-a716-446655440000"
}
```

## 🚀 Deployment

### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up -d
```

### Production Deployment

See the [Deployment Guide](docs/DEPLOYMENT.md) for:
- Database setup and migrations
- Environment configuration
- Load balancer setup
- SSL certificate configuration
- Monitoring setup

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Write comprehensive tests for new features
- Ensure tenant isolation in all new functionality
- Update documentation for API changes
- Add proper error handling and logging

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Community Support (All Plans)
- GitHub Issues
- Documentation
- Community Forum

### Professional Support (Pro+ Plans)
- Email support: support@adol.pos
- Response time: 24-48 hours

### Enterprise Support (Enterprise Plan)
- Priority support with SLA
- Dedicated support channel
- Custom integration assistance

## 🗺️ Roadmap

### Q1 2024
- [ ] Multi-currency support
- [ ] Advanced inventory management
- [ ] Mobile app for POS operations

### Q2 2024
- [ ] E-commerce integration
- [ ] Customer loyalty programs
- [ ] Advanced analytics dashboard

### Q3 2024
- [ ] AI-powered sales forecasting
- [ ] Third-party integrations marketplace
- [ ] White-label solutions

---

## 📞 Contact

- **Website**: [adol.pos](https://adol.pos)
- **Email**: contact@adol.pos
- **Support**: support@adol.pos
- **Documentation**: [docs.adol.pos](https://docs.adol.pos)

---

Built with ❤️ in Indonesia 🇮🇩