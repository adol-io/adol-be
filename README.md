# ADOL - Enterprise Point of Sale Backend System

ADOL is a modern, enterprise-grade Point of Sale (POS) backend system built with Go, PostgreSQL, and hexagonal architecture. It provides comprehensive features for product management, inventory tracking, sales processing, and invoice generation with support for multiple paper sizes.

## 🏗️ Architecture

This project follows **Hexagonal Architecture** (Ports and Adapters) principles:

- **Domain Layer**: Core business logic, entities, and domain services
- **Application Layer**: Use cases and application services
- **Infrastructure Layer**: External dependencies (database, HTTP, email, etc.)
- **Adapters**: Interface adapters for external systems

## 🚀 Features

### Core Features
- **User Management**: Role-based authentication and authorization (Admin, Manager, Cashier, Employee)
- **Product Management**: Complete CRUD operations with categorization and SKU management
- **Inventory Management**: Real-time stock tracking with automatic reorder alerts
- **Sales Processing**: Point-of-sale transactions with multiple payment methods
- **Invoice Generation**: Professional invoices with multiple paper size support (A4, A5, Letter, Legal, Receipt)
- **Reporting**: Comprehensive sales and inventory reports

### Technical Features
- **RESTful API**: Clean, well-documented REST endpoints
- **JWT Authentication**: Secure token-based authentication
- **Database Migrations**: Version-controlled database schema changes
- **Docker Support**: Containerized deployment with Docker Compose
- **Comprehensive Logging**: Structured logging with multiple levels
- **Error Handling**: Centralized error handling with custom error types
- **Input Validation**: Request validation with detailed error messages
- **Audit Trail**: Complete audit logging for all operations

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Docker and Docker Compose (for containerized deployment)
- Redis (optional, for caching and sessions)

## 🛠️ Installation

### Option 1: Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone https://github.com/nicklaros/adol.git
   cd adol
   ```

2. **Copy environment variables**
   ```bash
   cp .env.example .env
   ```

3. **Edit environment variables**
   ```bash
   nano .env  # Adjust configuration as needed
   ```

4. **Start the services**
   ```bash
   docker-compose up -d
   ```

5. **Run migrations** (after PostgreSQL is ready)
   ```bash
   # Wait for PostgreSQL to be ready, then run migrations
   make db-up
   ```
   
   **Note**: Even in Docker environments, migrations should be run separately to ensure proper sequencing and avoid race conditions.

6. **Check service status**
   ```bash
   docker-compose ps
   ```

The API will be available at `http://localhost:8080` and pgAdmin at `http://localhost:5050`.

### Option 2: Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/nicklaros/adol.git
   cd adol
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up PostgreSQL database**
   ```bash
   createdb adol_pos
   ```

4. **Run migrations**
   ```bash
   make db-up
   ```
   
   **Important**: Migrations are NOT run automatically on application startup to prevent race conditions in multi-instance deployments. Always run migrations manually using the provided commands.

5. **Copy and configure environment variables**
   ```bash
   cp .env.example .env
   # Edit .env file with your configuration
   ```

6. **Run the application**
   ```bash
   go run cmd/api/main.go
   ```

## 📚 API Documentation

### Authentication

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your_refresh_token"
}
```

#### Logout
```http
POST /api/v1/auth/logout
Authorization: Bearer your_access_token
```

### User Management

#### List Users
```http
GET /api/v1/users?page=1&limit=10&role=admin&status=active
Authorization: Bearer your_access_token
```

#### Create User
```http
POST /api/v1/users
Authorization: Bearer your_access_token
Content-Type: application/json

{
  "username": "newuser",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "password": "securepassword",
  "role": "cashier",
  "status": "active"
}
```

### Product Management

#### List Products
```http
GET /api/v1/products?page=1&limit=10&category=electronics&search=laptop
Authorization: Bearer your_access_token
```

#### Create Product
```http
POST /api/v1/products
Authorization: Bearer your_access_token
Content-Type: application/json

{
  "sku": "LAPTOP001",
  "name": "Gaming Laptop",
  "description": "High-performance gaming laptop",
  "category": "Electronics",
  "price": "1299.99",
  "cost": "999.99",
  "unit": "pcs",
  "min_stock": 5,
  "initial_stock": 10
}
```

### Sales Management

#### Create Sale
```http
POST /api/v1/sales
Authorization: Bearer your_access_token
Content-Type: application/json

{
  "customer_name": "John Customer",
  "customer_email": "john@example.com",
  "customer_phone": "+1234567890"
}
```

#### Add Item to Sale
```http
POST /api/v1/sales/{sale_id}/items
Authorization: Bearer your_access_token
Content-Type: application/json

{
  "product_id": "product_uuid",
  "quantity": 2
}
```

#### Complete Sale
```http
POST /api/v1/sales/{sale_id}/complete
Authorization: Bearer your_access_token
Content-Type: application/json

{
  "paid_amount": "149.98",
  "payment_method": "cash",
  "discount_amount": "10.00",
  "tax_percentage": "8.5",
  "notes": "Customer discount applied"
}
```

### Invoice Management

#### Generate Invoice PDF
```http
GET /api/v1/invoices/{invoice_id}/pdf?paper_size=a4
Authorization: Bearer your_access_token
```

#### Send Invoice Email
```http
POST /api/v1/invoices/{invoice_id}/email
Authorization: Bearer your_access_token
Content-Type: application/json

{
  "email_to": "customer@example.com",
  "subject": "Your Invoice",
  "paper_size": "a4"
}
```

## 🏃‍♂️ Usage Examples

### Default Admin Account
- **Username**: `admin`
- **Password**: `password`
- **Email**: `admin@adol.pos`

### Environment Variables

Key environment variables for configuration:

```bash
# Server
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=adol_pos
DB_USER=postgres
DB_PASSWORD=postgres

# JWT
JWT_SECRET_KEY=your-secret-key
JWT_EXPIRATION_TIME=24h

# Company Info (for invoices)
COMPANY_NAME=Your Company
COMPANY_ADDRESS=Your Address
COMPANY_PHONE=+1234567890
COMPANY_EMAIL=info@company.com
```

## 🗄️ Database Migrations

**⚠️ Important**: Migrations are NOT executed automatically on application startup to prevent race conditions in multi-instance deployments.

### Migration Commands

```bash
# Run all pending migrations
make db-up

# Rollback the last migration
make db-down

# Reset database (rollback all, then apply all)
make db-reset

# Create a new migration file
make db-create-migration
```

### Production Deployment Workflow

1. **Before deploying**: Run migrations on a single instance or dedicated migration job
2. **Deploy application**: Start application instances without migrations
3. **Verify**: Ensure all instances connect successfully

### Development Workflow

```bash
# Setup database and run migrations
make setup        # Creates .env file
make db-up        # Applies migrations
make run          # Starts the application
```

## 🧪 Testing

### Run Unit Tests
```bash
go test ./...
```

### Run Integration Tests
```bash
go test -tags=integration ./...
```

### API Health Check
```bash
curl http://localhost:8080/health
```

## 📊 Database Schema

The system uses PostgreSQL with the following main tables:
- `users` - User accounts and authentication
- `products` - Product catalog
- `stock` - Inventory levels
- `stock_movements` - Inventory transaction history
- `sales` - Sales transactions
- `sale_items` - Individual items in sales
- `invoices` - Generated invoices
- `invoice_items` - Invoice line items

## 🔧 Development

### Project Structure
```
adol/
├── cmd/api/                    # Application entry point
├── internal/
│   ├── domain/                 # Domain layer
│   │   ├── entities/          # Business entities
│   │   ├── repositories/      # Repository interfaces
│   │   └── services/         # Domain services
│   ├── application/           # Application layer
│   │   ├── usecases/         # Use cases
│   │   └── ports/            # Application ports
│   ├── infrastructure/        # Infrastructure layer
│   │   ├── database/         # Database connections
│   │   ├── http/            # HTTP handlers
│   │   └── config/          # Configuration
│   └── adapters/             # Adapters
├── pkg/                       # Shared packages
│   ├── errors/               # Error handling
│   ├── logger/               # Logging
│   └── utils/                # Utilities
├── migrations/                # Database migrations
├── docs/                     # Documentation
└── test/                     # Test files
```

### Adding New Features

1. **Domain Entity**: Add business logic in `internal/domain/entities/`
2. **Repository Interface**: Define data access in `internal/domain/repositories/`
3. **Use Case**: Implement business rules in `internal/application/usecases/`
4. **HTTP Handler**: Add API endpoints in `internal/infrastructure/http/`
5. **Migration**: Create database changes in `migrations/`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- Create an issue on GitHub
- Email: support@adol.pos
- Documentation: [Wiki](https://github.com/nicklaros/adol/wiki)

## 🚧 Roadmap

- [ ] Mobile API endpoints
- [ ] Real-time notifications
- [ ] Advanced reporting dashboard
- [ ] Multi-tenant support
- [ ] Integration with external payment gateways
- [ ] Barcode scanning support
- [ ] Multi-language support

## ⭐ Acknowledgments

Built with:
- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Golang Migrate](https://github.com/golang-migrate/migrate) - Database migrations
- [Logrus](https://github.com/sirupsen/logrus) - Structured logging
- [UUID](https://github.com/google/uuid) - UUID generation
- [Decimal](https://github.com/shopspring/decimal) - Decimal arithmetic