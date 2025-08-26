# ADOL POS API Documentation

This document provides comprehensive API documentation for the ADOL Point of Sale system.

## Table of Contents

- [Authentication](#authentication)
- [Error Handling](#error-handling)
- [Pagination](#pagination)
- [User Management API](#user-management-api)
- [Product Management API](#product-management-api)
- [Stock Management API](#stock-management-api)
- [Sales Management API](#sales-management-api)
- [Invoice Management API](#invoice-management-api)
- [Reports API](#reports-api)
- [System API](#system-api)
- [Response Examples](#response-examples)
- [Error Codes](#error-codes)

## Authentication

The ADOL API uses JWT (JSON Web Token) based authentication. Include the access token in the Authorization header for all protected endpoints.

### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

**Response:**
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "username": "admin",
      "email": "admin@adol.pos",
      "role": "admin",
      "status": "active"
    }
  },
  "message": "Login successful",
  "request_id": "req_123456789"
}
```

### Using the Token

Include the access token in all subsequent requests:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Refresh Token

When the access token expires, use the refresh token to get a new one:

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Error Handling

All API endpoints return consistent error responses with the following structure:

```json
{
  "error": {
    "type": "validation_error",
    "message": "Invalid request data",
    "details": "Field 'email' is required"
  },
  "request_id": "req_123456789"
}
```

### Error Types

- `validation_error`: Invalid input data
- `authentication_error`: Authentication required or invalid token
- `authorization_error`: Insufficient permissions
- `not_found_error`: Resource not found
- `conflict_error`: Resource already exists
- `internal_error`: Server-side error

## Pagination

List endpoints support pagination with the following parameters:

- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 100)

**Example:**
```http
GET /api/v1/products?page=2&limit=20
```

**Response includes pagination info:**
```json
{
  "data": {
    "products": [...],
    "pagination": {
      "page": 2,
      "limit": 20,
      "total_count": 150,
      "total_pages": 8,
      "has_next": true,
      "has_prev": true
    }
  }
}
```

## User Management API

### List Users

```http
GET /api/v1/users?page=1&limit=10&role=admin&status=active&search=john
Authorization: Bearer <token>
```

**Query Parameters:**
- `page`: Page number
- `limit`: Items per page
- `role`: Filter by role (admin, manager, cashier, employee)
- `status`: Filter by status (active, inactive, suspended)
- `search`: Search in username, email, first_name, last_name

### Create User

```http
POST /api/v1/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "password": "securepassword123",
  "role": "cashier",
  "status": "active"
}
```

### Get User

```http
GET /api/v1/users/123e4567-e89b-12d3-a456-426614174000
Authorization: Bearer <token>
```

### Update User

```http
PUT /api/v1/users/123e4567-e89b-12d3-a456-426614174000
Authorization: Bearer <token>
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Smith",
  "email": "john.smith@example.com",
  "role": "manager"
}
```

### User Status Management

```http
PUT /api/v1/users/123e4567-e89b-12d3-a456-426614174000/activate
PUT /api/v1/users/123e4567-e89b-12d3-a456-426614174000/deactivate
PUT /api/v1/users/123e4567-e89b-12d3-a456-426614174000/suspend
Authorization: Bearer <token>
```

## Product Management API

### List Products

```http
GET /api/v1/products?category=electronics&status=active&search=laptop&min_price=100&max_price=2000
Authorization: Bearer <token>
```

**Query Parameters:**
- `category`: Filter by category
- `status`: Filter by status (active, inactive, discontinued)
- `search`: Search in name, description, SKU
- `min_price`: Minimum price filter
- `max_price`: Maximum price filter

### Create Product

```http
POST /api/v1/products
Authorization: Bearer <token>
Content-Type: application/json

{
  "sku": "LAPTOP001",
  "name": "Gaming Laptop",
  "description": "High-performance gaming laptop with RTX graphics",
  "category": "Electronics",
  "price": "1299.99",
  "cost": "999.99",
  "unit": "pcs",
  "min_stock": 5,
  "initial_stock": 10
}
```

### Get Product

```http
GET /api/v1/products/123e4567-e89b-12d3-a456-426614174000
Authorization: Bearer <token>
```

### Get Product by SKU

```http
GET /api/v1/products/sku/LAPTOP001
Authorization: Bearer <token>
```

### Update Product

```http
PUT /api/v1/products/123e4567-e89b-12d3-a456-426614174000
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Gaming Laptop Pro",
  "price": "1399.99",
  "cost": "1099.99",
  "min_stock": 3
}
```

### Get Categories

```http
GET /api/v1/products/categories
Authorization: Bearer <token>
```

### Get Low Stock Products

```http
GET /api/v1/products/low-stock?page=1&limit=10
Authorization: Bearer <token>
```

## Stock Management API

### Get Stock for Product

```http
GET /api/v1/stock/123e4567-e89b-12d3-a456-426614174000
Authorization: Bearer <token>
```

### Adjust Stock

```http
POST /api/v1/stock/adjust
Authorization: Bearer <token>
Content-Type: application/json

{
  "product_id": "123e4567-e89b-12d3-a456-426614174000",
  "quantity": 50,
  "movement_type": "adjustment",
  "reason": "Physical inventory count",
  "reference": "INV-2024-001"
}
```

**Movement Types:**
- `sale`: Stock sold
- `purchase`: Stock purchased
- `adjustment`: Manual adjustment
- `return`: Return from customer
- `damage`: Damaged goods
- `transfer`: Transfer between locations

### Reserve Stock

```http
POST /api/v1/stock/reserve
Authorization: Bearer <token>
Content-Type: application/json

{
  "product_id": "123e4567-e89b-12d3-a456-426614174000",
  "quantity": 2,
  "reference": "SALE-2024-001",
  "expires_at": "2024-01-15T10:00:00Z"
}
```

### Release Reserved Stock

```http
POST /api/v1/stock/release
Authorization: Bearer <token>
Content-Type: application/json

{
  "product_id": "123e4567-e89b-12d3-a456-426614174000",
  "quantity": 2,
  "reference": "SALE-2024-001"
}
```

### Get Stock Movements

```http
GET /api/v1/stock/movements?product_id=123e4567-e89b-12d3-a456-426614174000&from_date=2024-01-01&to_date=2024-01-31
Authorization: Bearer <token>
```

## Sales Management API

### Create Sale

```http
POST /api/v1/sales
Authorization: Bearer <token>
Content-Type: application/json

{
  "customer_name": "John Customer",
  "customer_email": "john@customer.com",
  "customer_phone": "+1234567890"
}
```

### Add Item to Sale

```http
POST /api/v1/sales/123e4567-e89b-12d3-a456-426614174000/items
Authorization: Bearer <token>
Content-Type: application/json

{
  "product_id": "456e7890-e89b-12d3-a456-426614174111",
  "quantity": 2
}
```

### Update Sale Item

```http
PUT /api/v1/sales/123e4567-e89b-12d3-a456-426614174000/items
Authorization: Bearer <token>
Content-Type: application/json

{
  "product_id": "456e7890-e89b-12d3-a456-426614174111",
  "quantity": 3
}
```

### Remove Sale Item

```http
DELETE /api/v1/sales/123e4567-e89b-12d3-a456-426614174000/items/456e7890-e89b-12d3-a456-426614174111
Authorization: Bearer <token>
```

### Complete Sale

```http
POST /api/v1/sales/123e4567-e89b-12d3-a456-426614174000/complete
Authorization: Bearer <token>
Content-Type: application/json

{
  "paid_amount": "149.98",
  "payment_method": "cash",
  "discount_amount": "10.00",
  "tax_percentage": "8.5",
  "notes": "Customer discount applied"
}
```

**Payment Methods:**
- `cash`: Cash payment
- `card`: Credit/debit card
- `bank_transfer`: Bank transfer
- `digital_wallet`: Digital wallet payment

### List Sales

```http
GET /api/v1/sales?status=completed&from_date=2024-01-01&to_date=2024-01-31&customer_name=john
Authorization: Bearer <token>
```

### Get Sale

```http
GET /api/v1/sales/123e4567-e89b-12d3-a456-426614174000
Authorization: Bearer <token>
```

### Get Sale by Number

```http
GET /api/v1/sales/number/SALE-2024-001
Authorization: Bearer <token>
```

## Invoice Management API

### List Invoices

```http
GET /api/v1/invoices?status=pending&from_date=2024-01-01&to_date=2024-01-31
Authorization: Bearer <token>
```

### Create Invoice

```http
POST /api/v1/invoices
Authorization: Bearer <token>
Content-Type: application/json

{
  "sale_id": "123e4567-e89b-12d3-a456-426614174000",
  "due_date": "2024-01-31T23:59:59Z",
  "paper_size": "a4",
  "notes": "Payment due within 30 days"
}
```

### Get Invoice

```http
GET /api/v1/invoices/123e4567-e89b-12d3-a456-426614174000
Authorization: Bearer <token>
```

### Generate Invoice PDF

```http
GET /api/v1/invoices/123e4567-e89b-12d3-a456-426614174000/pdf?paper_size=a4
Authorization: Bearer <token>
```

**Paper Sizes:**
- `a4`: A4 (210mm x 297mm)
- `a5`: A5 (148mm x 210mm)
- `letter`: US Letter (8.5" x 11")
- `legal`: US Legal (8.5" x 14")
- `receipt`: Thermal receipt (80mm)

### Send Invoice Email

```http
POST /api/v1/invoices/123e4567-e89b-12d3-a456-426614174000/email
Authorization: Bearer <token>
Content-Type: application/json

{
  "email_to": "customer@example.com",
  "subject": "Your Invoice #INV-2024-001",
  "message": "Please find your invoice attached.",
  "paper_size": "a4"
}
```

### Print Invoice

```http
POST /api/v1/invoices/123e4567-e89b-12d3-a456-426614174000/print
Authorization: Bearer <token>
Content-Type: application/json

{
  "printer_name": "Office Printer",
  "paper_size": "a4"
}
```

### Mark Invoice as Paid

```http
PUT /api/v1/invoices/123e4567-e89b-12d3-a456-426614174000/paid
Authorization: Bearer <token>
Content-Type: application/json

{
  "payment_method": "bank_transfer",
  "payment_reference": "TXN123456",
  "paid_at": "2024-01-15T10:30:00Z"
}
```

### Get Overdue Invoices

```http
GET /api/v1/invoices/overdue?page=1&limit=10
Authorization: Bearer <token>
```

### Get Invoice Templates

```http
GET /api/v1/invoices/templates?paper_size=a4
Authorization: Bearer <token>
```

### Get Available Paper Sizes

```http
GET /api/v1/invoices/paper-sizes
Authorization: Bearer <token>
```

### Get Available Printers

```http
GET /api/v1/invoices/printers
Authorization: Bearer <token>
```

## Reports API

### Sales Report

```http
GET /api/v1/reports/sales?from_date=2024-01-01&to_date=2024-01-31&group_by=day
Authorization: Bearer <token>
```

**Group By Options:**
- `day`: Daily breakdown
- `week`: Weekly breakdown
- `month`: Monthly breakdown
- `product`: By product
- `category`: By category

### Daily Sales Report

```http
GET /api/v1/reports/sales/daily?date=2024-01-15
Authorization: Bearer <token>
```

### Invoice Report

```http
GET /api/v1/reports/invoices?from_date=2024-01-01&to_date=2024-01-31&status=pending
Authorization: Bearer <token>
```

### Top Selling Products

```http
GET /api/v1/reports/products/top-selling?from_date=2024-01-01&to_date=2024-01-31&limit=10
Authorization: Bearer <token>
```

## System API

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "adol-pos-api",
  "version": "1.0.0"
}
```

### Detailed Health Check

```http
GET /health/detailed
```

### Metrics

```http
GET /metrics
Authorization: Bearer <token>
```

## Response Examples

### Success Response

```json
{
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "username": "john_doe",
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "cashier",
      "status": "active"
    }
  },
  "message": "User created successfully",
  "request_id": "req_123456789"
}
```

### List Response with Pagination

```json
{
  "data": {
    "products": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "sku": "LAPTOP001",
        "name": "Gaming Laptop",
        "price": "1299.99",
        "cost": "999.99",
        "category": "Electronics",
        "status": "active"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total_count": 25,
      "total_pages": 3,
      "has_next": true,
      "has_prev": false
    }
  },
  "message": "Products retrieved successfully",
  "request_id": "req_123456789"
}
```

### Error Response

```json
{
  "error": {
    "type": "validation_error",
    "message": "Invalid request data",
    "details": "Field 'email' must be a valid email address"
  },
  "request_id": "req_123456789"
}
```

## Error Codes

| HTTP Status | Error Type | Description |
|-------------|------------|-------------|
| 400 | validation_error | Invalid request data |
| 401 | authentication_error | Authentication required |
| 403 | authorization_error | Insufficient permissions |
| 404 | not_found_error | Resource not found |
| 409 | conflict_error | Resource conflict |
| 429 | rate_limit_error | Too many requests |
| 500 | internal_error | Server error |

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Authenticated requests**: 1000 requests per hour per user
- **Unauthenticated requests**: 100 requests per hour per IP

Rate limit headers are included in all responses:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1705318800
```

## SDK and Libraries

### cURL Examples

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# Create product
curl -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"sku":"TEST001","name":"Test Product","price":"99.99","cost":"49.99","unit":"pcs"}'

# Get products
curl -X GET "http://localhost:8080/api/v1/products?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### JavaScript/Node.js

```javascript
const ADOL_API_BASE = 'http://localhost:8080/api/v1';
let accessToken = '';

// Login
async function login(username, password) {
  const response = await fetch(`${ADOL_API_BASE}/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password }),
  });
  const data = await response.json();
  accessToken = data.data.access_token;
  return data;
}

// Create product
async function createProduct(product) {
  const response = await fetch(`${ADOL_API_BASE}/products`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(product),
  });
  return response.json();
}
```

### Python

```python
import requests

class ADOLClient:
    def __init__(self, base_url='http://localhost:8080/api/v1'):
        self.base_url = base_url
        self.access_token = None
    
    def login(self, username, password):
        response = requests.post(f'{self.base_url}/auth/login', json={
            'username': username,
            'password': password
        })
        data = response.json()
        self.access_token = data['data']['access_token']
        return data
    
    def get_headers(self):
        return {'Authorization': f'Bearer {self.access_token}'}
    
    def create_product(self, product):
        response = requests.post(
            f'{self.base_url}/products',
            json=product,
            headers=self.get_headers()
        )
        return response.json()
```

## Postman Collection

A Postman collection with all API endpoints and examples is available at:
`docs/postman/ADOL-POS-API.postman_collection.json`

Import this collection into Postman for easy testing and development.

---

For more information, visit our [GitHub repository](https://github.com/nicklaros/adol) or contact support at support@adol.pos.