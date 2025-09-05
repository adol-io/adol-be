# ADOL POS Multi-Tenant Architecture

This document provides comprehensive information about the multi-tenant architecture implementation in the ADOL Point of Sale (POS) system.

## Overview

The ADOL POS system has been transformed from a single-tenant application into a multi-tenant SaaS platform, enabling multiple independent organizations to use the system while maintaining complete data isolation and customization capabilities.

## Architecture Design

### Multi-Tenancy Pattern
- **Pattern**: Shared Database with Tenant Isolation
- **Isolation Method**: Row Level Security (RLS) + Application-level filtering
- **Identification**: Tenant ID (UUID) in every table

### Key Components

#### 1. Tenant Management
- **Entity**: `Tenant`
- **Features**: Trial management, feature flags, business configuration
- **Slug Generation**: Automatic URL-friendly slug creation
- **Status Management**: Active, Inactive, Suspended, Trial states

#### 2. Subscription System
- **Plans**: Starter (Free), Professional ($300k), Enterprise ($1.5M)
- **Features**: Tiered feature access with usage limits
- **Trial Period**: 30-day free trial for new tenants
- **Usage Tracking**: Real-time usage monitoring and limit enforcement

#### 3. Tenant Context Management
- **Resolution Methods**: 
  - HTTP Headers (`X-Tenant-ID`, `X-Tenant-Slug`, `X-Tenant-Domain`)
  - Subdomain extraction
  - URL parameters
- **Context Storage**: Request-scoped tenant information
- **Middleware**: Automatic tenant resolution and context injection

#### 4. Enhanced Authentication
- **JWT Enhancement**: Tenant-aware token claims
- **User Isolation**: Users belong to specific tenants
- **Role-based Access**: Admin, Manager, Cashier, Employee roles per tenant

## Database Schema

### Core Tables

#### Tenants Table
```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    domain VARCHAR(255) UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'trial',
    configuration JSONB NOT NULL DEFAULT '{}',
    trial_start TIMESTAMP WITH TIME ZONE,
    trial_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);
```

#### Tenant Subscriptions Table
```sql
CREATE TABLE tenant_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'trial',
    billing_start TIMESTAMP WITH TIME ZONE,
    billing_end TIMESTAMP WITH TIME ZONE,
    monthly_fee DECIMAL(15,2) NOT NULL DEFAULT 0,
    features JSONB NOT NULL DEFAULT '{}',
    usage_limits JSONB NOT NULL DEFAULT '{}',
    current_usage JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### Row Level Security (RLS)

All existing tables have been enhanced with `tenant_id` columns and RLS policies:

```sql
-- Example RLS Policy for Products
CREATE POLICY tenant_isolation_policy ON products
    FOR ALL
    TO authenticated_users
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

## API Documentation

### Tenant Registration

**Endpoint**: `POST /api/v1/tenants/register`

**Request**:
```json
{
  "tenant_name": "My Company",
  "admin_username": "admin",
  "admin_email": "admin@mycompany.com",
  "admin_password": "secure_password",
  "admin_first_name": "John",
  "admin_last_name": "Doe",
  "domain": "mycompany.com" // optional
}
```

**Response**:
```json
{
  "tenant": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "My Company",
    "slug": "my-company",
    "status": "trial",
    "trial_start": "2024-01-01T00:00:00Z",
    "trial_end": "2024-01-31T23:59:59Z"
  },
  "subscription": {
    "id": "660f8400-e29b-41d4-a716-446655440000",
    "plan_type": "starter",
    "status": "trial"
  },
  "admin_user": {
    "id": "770g8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "email": "admin@mycompany.com",
    "role": "admin"
  }
}
```

### Tenant Context Headers

All API requests must include tenant identification:

**Option 1 - Tenant ID Header**:
```
X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Option 2 - Tenant Slug Header**:
```
X-Tenant-Slug: my-company
```

**Option 3 - Subdomain**:
```
Host: my-company.adol.pos
```

### Subscription Management

**Get Subscription Details**: `GET /api/v1/subscription`

**Upgrade Plan**: `POST /api/v1/subscription/upgrade`
```json
{
  "plan_type": "professional"
}
```

**Check Usage**: `GET /api/v1/subscription/usage`

## Configuration

### Environment Variables

```bash
# Multi-Tenancy Features
FEATURE_ENABLE_MULTI_TENANCY=true
FEATURE_ENABLE_SUBSCRIPTIONS=true
FEATURE_ENABLE_USAGE_LIMITS=true
FEATURE_ENABLE_TRIAL_PERIODS=true
FEATURE_ENABLE_SUBDOMAINS=true
FEATURE_ENABLE_FEATURE_GATING=true

# Tenant Configuration
TENANT_DEFAULT_TRIAL_DAYS=30
TENANT_MAX_PER_DOMAIN=1
TENANT_SLUG_MIN_LENGTH=3
TENANT_SLUG_MAX_LENGTH=63
TENANT_ALLOW_SUBDOMAINS=true
TENANT_REQUIRE_SSL=false

# JWT Configuration (Enhanced for Multi-Tenancy)
JWT_INCLUDE_TENANT_CLAIMS=true
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h
```

### Feature Flags by Plan

#### Starter Plan (Free with 1% transaction fee)
- ✅ Basic POS functionality
- ✅ Inventory management
- ✅ Basic reporting
- ❌ Advanced reporting
- ❌ Multi-location support
- ❌ API access
- **User Limit**: 2 users
- **Products**: Unlimited
- **Sales**: Unlimited

#### Professional Plan (Rp300,000/month)
- ✅ All Starter features
- ✅ Advanced reporting
- ✅ Multi-location support
- ❌ API access
- ❌ Custom integrations
- **User Limit**: 10 users
- **Products**: Unlimited
- **Sales**: Unlimited

#### Enterprise Plan (Rp1,500,000/month)
- ✅ All Professional features
- ✅ Full API access
- ✅ Custom integrations
- ✅ Priority support
- **User Limit**: Unlimited
- **Products**: Unlimited
- **Sales**: Unlimited
- **API Calls**: 10,000/month

## Usage Examples

### Creating Tenant-Aware Products

```go
// In your handler
tenantContext := GetTenantContext(c)
if tenantContext == nil {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "No tenant context"})
    return
}

product, err := entities.NewProduct(
    tenantContext.TenantID, // Tenant ID is automatically included
    "PROD-001",
    "Product Name",
    "Description",
    "Category",
    "pcs",
    decimal.NewFromFloat(99.99),
    decimal.NewFromFloat(50.00),
    10,
    userID,
)
```

### Checking Feature Access

```go
// Check if tenant has access to a feature
err := tenantContext.ValidateFeatureAccess("advanced_reporting")
if err != nil {
    c.JSON(http.StatusForbidden, gin.H{
        "error": "Feature access denied",
        "required_plan": "professional",
    })
    return
}
```

### Tracking Usage

```go
// Track API usage
err := monitor.TrackUsage(ctx, tenantID, "api_requests", 1)
if err != nil {
    logger.Error("Failed to track usage", "error", err)
}

// Check usage limits
limitStatus, err := monitor.CheckSubscriptionLimits(ctx, tenantID)
if err == nil && limitStatus.IsOverLimit {
    // Handle over-limit scenario
}
```

## Monitoring and Logging

### Tenant-Specific Logging

All log entries automatically include tenant context:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "message": "Product created successfully",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_slug": "my-company",
  "user_id": "770g8400-e29b-41d4-a716-446655440000",
  "request_id": "req-123",
  "method": "POST",
  "path": "/api/v1/products"
}
```

### Usage Monitoring

- **Real-time Usage Tracking**: Automatic tracking of users, products, sales, API calls
- **Limit Enforcement**: Proactive blocking when limits are exceeded
- **Usage Alerts**: Automatic alerts at 80% and 100% usage thresholds
- **Health Monitoring**: Per-tenant health status tracking

## Security Considerations

### Data Isolation
- **Database Level**: Row Level Security (RLS) policies
- **Application Level**: Tenant-aware repositories and services  
- **API Level**: Tenant context validation middleware

### Access Control
- **Tenant Boundaries**: Users cannot access other tenants' data
- **Feature Gating**: Subscription-based feature access control
- **Usage Limits**: Automatic enforcement of plan limits

### Audit Logging
- **Tenant-Specific Audits**: All actions logged with tenant context
- **Security Events**: Failed access attempts and suspicious activity
- **Compliance**: Full audit trail for regulatory requirements

## Deployment Guide

### Database Migration

1. **Run Tenant Tables Migration**:
   ```bash
   migrate -path ./migrations -database postgres://... up
   ```

2. **Verify RLS Policies**:
   ```sql
   SELECT schemaname, tablename, policyname, cmd, roles 
   FROM pg_policies 
   WHERE schemaname = 'public';
   ```

### Application Deployment

1. **Environment Configuration**: Set all required environment variables
2. **Feature Flags**: Enable multi-tenancy features gradually
3. **Monitoring Setup**: Configure tenant-specific dashboards
4. **Load Testing**: Test with multiple concurrent tenants

### Migration from Single-Tenant

For existing single-tenant installations:

1. **Backup Data**: Full database backup before migration
2. **Create Default Tenant**: Migrate existing data to first tenant
3. **Update Existing Records**: Add tenant_id to all existing data
4. **Enable Multi-Tenancy**: Activate multi-tenant features
5. **Test Isolation**: Verify data isolation works correctly

## Troubleshooting

### Common Issues

#### 1. Tenant Context Not Found
**Symptom**: 401 Unauthorized errors
**Solution**: Ensure proper tenant headers are set in requests

#### 2. Feature Access Denied
**Symptom**: 403 Forbidden for valid features
**Solution**: Verify subscription plan and feature flags

#### 3. Usage Limit Exceeded
**Symptom**: 403 Forbidden with usage limit message
**Solution**: Upgrade plan or reduce usage

#### 4. RLS Policy Blocking Access
**Symptom**: Empty results or access denied
**Solution**: Check `current_setting('app.current_tenant_id')` is set

### Debug Commands

```bash
# Check tenant context in database
SELECT current_setting('app.current_tenant_id', true);

# Verify RLS policies
\d+ products  -- Should show RLS is enabled

# Test tenant isolation
SELECT COUNT(*) FROM products WHERE tenant_id = 'tenant-id';
```

## Performance Considerations

### Database Optimization
- **Indexing**: All tenant_id columns are indexed
- **Partitioning**: Consider table partitioning for large datasets
- **Query Optimization**: Ensure tenant_id is always in WHERE clauses

### Application Performance
- **Caching**: Tenant-aware caching strategies
- **Connection Pooling**: Per-tenant connection pools if needed
- **Resource Isolation**: CPU and memory usage monitoring per tenant

### Scaling Strategies
- **Horizontal Scaling**: Multiple application instances
- **Database Scaling**: Read replicas and sharding if needed
- **CDN Integration**: Tenant-specific static asset delivery

## Support and Maintenance

### Monitoring Dashboards
- Tenant registration and churn metrics
- Usage patterns and limit violations
- Performance metrics per tenant
- Revenue and subscription analytics

### Backup and Recovery
- **Per-Tenant Backups**: Isolated backup and restore capabilities
- **Point-in-Time Recovery**: Tenant-specific recovery procedures
- **Data Export**: Self-service data export for customers

### Support Procedures
- **Tenant Identification**: Quick tenant lookup by domain/slug
- **Usage Analysis**: Understanding tenant usage patterns
- **Issue Escalation**: Tenant-aware support ticket routing

---

For more detailed information, see the individual component documentation in the `/docs` directory.