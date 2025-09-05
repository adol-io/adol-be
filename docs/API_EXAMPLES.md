# ADOL POS Multi-Tenant API Examples

This document provides practical examples of using the ADOL POS multi-tenant API.

## Authentication & Tenant Context

### 1. Tenant Registration

Register a new tenant with admin user and subscription:

```bash
curl -X POST https://api.adol.pos/api/v1/tenants/register \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_name": "Acme Corporation",
    "admin_username": "admin",
    "admin_email": "admin@acme.com",
    "admin_password": "SecurePass123!",
    "admin_first_name": "John",
    "admin_last_name": "Doe",
    "domain": "acme.com"
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "tenant": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Acme Corporation", 
      "slug": "acme-corporation",
      "domain": "acme.com",
      "status": "trial",
      "trial_start": "2024-01-01T00:00:00Z",
      "trial_end": "2024-01-31T23:59:59Z",
      "configuration": {
        "business_info": {
          "name": "Acme Corporation",
          "currency": "USD",
          "tax_rate": 0.0
        },
        "feature_flags": {
          "pos": true,
          "inventory": true,
          "reporting": true,
          "advanced_reporting": false
        }
      }
    },
    "subscription": {
      "id": "660f8400-e29b-41d4-a716-446655440000",
      "plan_type": "starter",
      "status": "trial",
      "features": {
        "pos": true,
        "inventory": true,
        "reporting": true,
        "advanced_reporting": false,
        "multi_location": false,
        "api_access": false
      },
      "usage_limits": {
        "users": 2,
        "products": -1,
        "sales_per_month": -1,
        "api_calls_per_month": 0
      }
    },
    "admin_user": {
      "id": "770g8400-e29b-41d4-a716-446655440000",
      "username": "admin",
      "email": "admin@acme.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "admin"
    }
  }
}
```

### 2. User Authentication

Login with tenant context:

```bash
curl -X POST https://api.adol.pos/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Slug: acme-corporation" \
  -d '{
    "username": "admin",
    "password": "SecurePass123!"
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "user": {
      "id": "770g8400-e29b-41d4-a716-446655440000",
      "username": "admin",
      "email": "admin@acme.com",
      "role": "admin",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000"
    },
    "access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
    "refresh_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
    "expires_at": "2024-01-01T12:15:00Z",
    "token_type": "Bearer"
  }
}
```

## Product Management

### 3. Create Product

Create a product with tenant isolation:

```bash
curl -X POST https://api.adol.pos/api/v1/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "sku": "LAPTOP-001",
    "name": "Gaming Laptop",
    "description": "High-performance gaming laptop",
    "category": "Electronics",
    "price": 1299.99,
    "cost": 899.99,
    "unit": "pcs",
    "min_stock": 5
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "880h8400-e29b-41d4-a716-446655440000",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "sku": "LAPTOP-001",
    "name": "Gaming Laptop",
    "description": "High-performance gaming laptop",
    "category": "Electronics",
    "price": 1299.99,
    "cost": 899.99,
    "unit": "pcs",
    "min_stock": 5,
    "status": "active",
    "created_at": "2024-01-01T12:00:00Z",
    "created_by": "770g8400-e29b-41d4-a716-446655440000"
  }
}
```

### 4. List Products (Tenant-Filtered)

List products for current tenant:

```bash
curl -X GET https://api.adol.pos/api/v1/products \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." \
  -H "X-Tenant-Slug: acme-corporation"
```

**Response:**
```json
{
  "status": "success",
  "data": [
    {
      "id": "880h8400-e29b-41d4-a716-446655440000",
      "sku": "LAPTOP-001",
      "name": "Gaming Laptop",
      "category": "Electronics",
      "price": 1299.99,
      "status": "active",
      "current_stock": 10
    }
  ],
  "pagination": {
    "total": 1,
    "page": 1,
    "per_page": 50,
    "total_pages": 1
  }
}
```

## User Management

### 5. Create User (With Usage Limit Check)

Create a new user within subscription limits:

```bash
curl -X POST https://api.adol.pos/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "username": "cashier1",
    "email": "cashier@acme.com",
    "first_name": "Jane",
    "last_name": "Smith",
    "password": "CashierPass123!",
    "role": "cashier"
  }'
```

**Success Response:**
```json
{
  "status": "success",
  "data": {
    "id": "990i8400-e29b-41d4-a716-446655440000",
    "username": "cashier1",
    "email": "cashier@acme.com",
    "first_name": "Jane",
    "last_name": "Smith",
    "role": "cashier",
    "status": "active",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2024-01-01T12:30:00Z"
  }
}
```

**Usage Limit Exceeded Response:**
```json
{
  "status": "error",
  "error": {
    "code": "USAGE_LIMIT_EXCEEDED",
    "message": "User limit exceeded for current subscription plan",
    "details": {
      "current_usage": 2,
      "limit": 2,
      "plan_type": "starter",
      "upgrade_required": "professional"
    }
  }
}
```

## Subscription Management

### 6. Get Current Subscription

Check subscription details and usage:

```bash
curl -X GET https://api.adol.pos/api/v1/subscription \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "subscription": {
      "id": "660f8400-e29b-41d4-a716-446655440000",
      "plan_type": "starter",
      "status": "trial",
      "trial_end": "2024-01-31T23:59:59Z",
      "features": {
        "pos": true,
        "inventory": true,
        "reporting": true,
        "advanced_reporting": false,
        "multi_location": false,
        "api_access": false
      }
    },
    "usage": {
      "users": {
        "current": 2,
        "limit": 2,
        "percentage": 100
      },
      "products": {
        "current": 1,
        "limit": -1,
        "percentage": 0
      },
      "sales_this_month": {
        "current": 0,
        "limit": -1,
        "percentage": 0
      }
    },
    "billing": {
      "current_period_start": "2024-01-01T00:00:00Z",
      "current_period_end": "2024-01-31T23:59:59Z",
      "next_billing_date": "2024-02-01T00:00:00Z"
    }
  }
}
```

### 7. Upgrade Subscription

Upgrade to a higher plan:

```bash
curl -X POST https://api.adol.pos/api/v1/subscription/upgrade \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "plan_type": "professional"
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "subscription": {
      "id": "660f8400-e29b-41d4-a716-446655440000",
      "plan_type": "professional",
      "status": "active",
      "monthly_fee": 300000,
      "billing_start": "2024-01-01T12:45:00Z",
      "billing_end": "2024-02-01T12:45:00Z",
      "features": {
        "pos": true,
        "inventory": true,
        "reporting": true,
        "advanced_reporting": true,
        "multi_location": true,
        "api_access": false
      },
      "usage_limits": {
        "users": 10,
        "products": -1,
        "sales_per_month": -1,
        "api_calls_per_month": 0
      }
    },
    "message": "Subscription upgraded successfully. New features are now available."
  }
}
```

## Feature Access Control

### 8. Access Advanced Reporting (Feature-Gated)

Try to access advanced reporting:

```bash
curl -X GET https://api.adol.pos/api/v1/reports/advanced/sales-analytics \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Success Response (Professional/Enterprise plan):**
```json
{
  "status": "success",
  "data": {
    "report_type": "sales_analytics",
    "period": "last_30_days",
    "metrics": {
      "total_sales": 15750.00,
      "transaction_count": 127,
      "average_transaction": 124.02,
      "top_products": [
        {
          "sku": "LAPTOP-001",
          "name": "Gaming Laptop",
          "sales": 5199.96,
          "quantity": 4
        }
      ]
    }
  }
}
```

**Feature Access Denied Response (Starter plan):**
```json
{
  "status": "error",
  "error": {
    "code": "FEATURE_ACCESS_DENIED",
    "message": "Feature access denied: advanced_reporting",
    "details": {
      "required_plan": "professional",
      "current_plan": "starter",
      "upgrade_url": "/api/v1/subscription/upgrade"
    }
  }
}
```

## Sales Processing

### 9. Create Sale

Process a sale transaction:

```bash
curl -X POST https://api.adol.pos/api/v1/sales \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "customer_name": "Alice Johnson",
    "customer_email": "alice@example.com",
    "items": [
      {
        "product_id": "880h8400-e29b-41d4-a716-446655440000",
        "quantity": 1,
        "unit_price": 1299.99
      }
    ],
    "payment_method": "card",
    "tax_amount": 104.00,
    "total_amount": 1403.99,
    "paid_amount": 1403.99
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "aa0j8400-e29b-41d4-a716-446655440000",
    "sale_number": "SALE-2024-001",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "customer_name": "Alice Johnson",
    "subtotal": 1299.99,
    "tax_amount": 104.00,
    "total_amount": 1403.99,
    "status": "completed",
    "items": [
      {
        "product_id": "880h8400-e29b-41d4-a716-446655440000",
        "product_sku": "LAPTOP-001",
        "product_name": "Gaming Laptop",
        "quantity": 1,
        "unit_price": 1299.99,
        "total_price": 1299.99
      }
    ],
    "created_at": "2024-01-01T14:30:00Z",
    "created_by": "990i8400-e29b-41d4-a716-446655440000"
  }
}
```

## Tenant Configuration

### 10. Update Business Information

Update tenant business configuration:

```bash
curl -X PUT https://api.adol.pos/api/v1/tenants/configuration/business \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "name": "Acme Corporation Ltd",
    "address": "123 Business St, City, State 12345",
    "phone": "+1-555-123-4567",
    "email": "info@acme.com",
    "tax_id": "TAX123456789",
    "currency": "USD",
    "tax_rate": 8.5
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "business_info": {
      "name": "Acme Corporation Ltd",
      "address": "123 Business St, City, State 12345",
      "phone": "+1-555-123-4567",
      "email": "info@acme.com",
      "tax_id": "TAX123456789",
      "currency": "USD",
      "tax_rate": 8.5
    },
    "updated_at": "2024-01-01T15:00:00Z"
  }
}
```

## Error Handling Examples

### Common Error Responses

#### 1. Missing Tenant Context
```json
{
  "status": "error",
  "error": {
    "code": "TENANT_CONTEXT_REQUIRED",
    "message": "Tenant context not found in request",
    "details": {
      "supported_methods": [
        "X-Tenant-ID header",
        "X-Tenant-Slug header", 
        "subdomain",
        "URL parameter"
      ]
    }
  }
}
```

#### 2. Invalid Tenant
```json
{
  "status": "error",
  "error": {
    "code": "TENANT_NOT_FOUND",
    "message": "Tenant not found or inactive",
    "details": {
      "tenant_id": "invalid-id"
    }
  }
}
```

#### 3. Subscription Suspended
```json
{
  "status": "error",
  "error": {
    "code": "SUBSCRIPTION_SUSPENDED",
    "message": "Tenant subscription is suspended",
    "details": {
      "status": "suspended",
      "reason": "Payment overdue",
      "contact_support": "support@adol.pos"
    }
  }
}
```

## Rate Limiting & Usage Tracking

### API Rate Limits by Plan

#### Starter Plan
- **Rate Limit**: 100 requests/minute
- **Burst**: 200 requests
- **API Access**: No external API access

#### Professional Plan  
- **Rate Limit**: 500 requests/minute
- **Burst**: 1000 requests
- **API Access**: No external API access

#### Enterprise Plan
- **Rate Limit**: 2000 requests/minute
- **Burst**: 5000 requests
- **API Access**: 10,000 calls/month to external endpoints

### Rate Limit Headers

All API responses include rate limit information:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 85
X-RateLimit-Reset: 1640995200
X-RateLimit-Retry-After: 45
```

## Webhook Examples

### Subscription Events

Configure webhooks to receive subscription events:

```bash
curl -X POST https://api.adol.pos/api/v1/webhooks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "url": "https://acme.com/webhooks/adol",
    "events": [
      "subscription.upgraded",
      "subscription.cancelled",
      "usage.limit_reached",
      "trial.ending"
    ],
    "secret": "webhook_secret_key"
  }'
```

**Webhook Payload Example:**
```json
{
  "event": "subscription.upgraded",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-01-01T12:45:00Z",
  "data": {
    "subscription_id": "660f8400-e29b-41d4-a716-446655440000",
    "old_plan": "starter",
    "new_plan": "professional",
    "effective_date": "2024-01-01T12:45:00Z"
  }
}
```

## Testing with Different Tenant Contexts

### Using Subdomain Resolution

```bash
# Request with subdomain
curl -X GET https://acme-corporation.adol.pos/api/v1/products \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."
```

### Using Custom Domain

```bash
# Request with custom domain (if configured)
curl -X GET https://pos.acme.com/api/v1/products \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."
```

### Multiple Tenant Operations

```bash
# Create products for different tenants
curl -X POST https://api.adol.pos/api/v1/products \
  -H "X-Tenant-Slug: acme-corporation" \
  -H "Authorization: Bearer token_for_acme" \
  -d '{"sku": "ACME-001", "name": "Acme Product"}'

curl -X POST https://api.adol.pos/api/v1/products \
  -H "X-Tenant-Slug: other-company" \
  -H "Authorization: Bearer token_for_other" \
  -d '{"sku": "OTHER-001", "name": "Other Product"}'

# Verify isolation - each tenant only sees their products
curl -X GET https://api.adol.pos/api/v1/products \
  -H "X-Tenant-Slug: acme-corporation" \
  -H "Authorization: Bearer token_for_acme"
```

---

For more examples and detailed API reference, see the [OpenAPI specification](./api-spec.yaml).