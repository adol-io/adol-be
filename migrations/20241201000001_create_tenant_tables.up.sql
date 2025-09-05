-- Multi-Tenant Management Schema
-- This migration creates the core tenant management tables

-- Tenants table - Central tenant management
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    domain VARCHAR(255) UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'trial')),
    configuration JSONB DEFAULT '{}',
    trial_start TIMESTAMP WITH TIME ZONE,
    trial_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID
);

-- Create indexes for tenants table
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_domain ON tenants(domain) WHERE domain IS NOT NULL;
CREATE INDEX idx_tenants_status ON tenants(status);

-- Tenant subscriptions table
CREATE TABLE tenant_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_type VARCHAR(50) NOT NULL CHECK (plan_type IN ('starter', 'professional', 'enterprise')),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'cancelled', 'trial')),
    billing_start TIMESTAMP WITH TIME ZONE,
    billing_end TIMESTAMP WITH TIME ZONE,
    monthly_fee DECIMAL(15,2) NOT NULL DEFAULT 0 CHECK (monthly_fee >= 0),
    features JSONB DEFAULT '{}',
    usage_limits JSONB DEFAULT '{}',
    current_usage JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create unique constraint and indexes for tenant subscriptions
ALTER TABLE tenant_subscriptions ADD CONSTRAINT uk_tenant_subscriptions_tenant_id UNIQUE (tenant_id);
CREATE INDEX idx_tenant_subscriptions_status ON tenant_subscriptions(status);
CREATE INDEX idx_tenant_subscriptions_plan_type ON tenant_subscriptions(plan_type);

-- Tenant settings table for flexible configuration
CREATE TABLE tenant_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    setting_key VARCHAR(255) NOT NULL,
    setting_value JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create unique constraint and indexes for tenant settings
ALTER TABLE tenant_settings ADD CONSTRAINT uk_tenant_settings_tenant_key UNIQUE (tenant_id, setting_key);
CREATE INDEX idx_tenant_settings_tenant_id ON tenant_settings(tenant_id);

-- Add tenant_id to existing tables for multi-tenancy

-- Update users table
ALTER TABLE users ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
-- Create index for tenant_id on users
CREATE INDEX idx_users_tenant_id ON users(tenant_id);

-- Update products table
ALTER TABLE products ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
-- Create index for tenant_id on products
CREATE INDEX idx_products_tenant_id ON products(tenant_id);
-- Create composite index for tenant-scoped SKU uniqueness
ALTER TABLE products DROP CONSTRAINT products_sku_key;
CREATE UNIQUE INDEX uk_products_tenant_sku ON products(tenant_id, sku) WHERE deleted_at IS NULL;

-- Update sales table
ALTER TABLE sales ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
-- Create index for tenant_id on sales
CREATE INDEX idx_sales_tenant_id ON sales(tenant_id);
-- Create composite index for tenant-scoped sale number uniqueness
ALTER TABLE sales DROP CONSTRAINT sales_sale_number_key;
CREATE UNIQUE INDEX uk_sales_tenant_sale_number ON sales(tenant_id, sale_number) WHERE deleted_at IS NULL;

-- Update invoices table
ALTER TABLE invoices ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
-- Create index for tenant_id on invoices
CREATE INDEX idx_invoices_tenant_id ON invoices(tenant_id);
-- Create composite index for tenant-scoped invoice number uniqueness
ALTER TABLE invoices DROP CONSTRAINT invoices_invoice_number_key;
CREATE UNIQUE INDEX uk_invoices_tenant_invoice_number ON invoices(tenant_id, invoice_number) WHERE deleted_at IS NULL;

-- Update stock table (inherits tenant_id from product relationship)
-- No direct tenant_id needed since it's linked to product

-- Update stock_movements table (inherits tenant_id from product relationship)
-- No direct tenant_id needed since it's linked to product

-- Update sale_items table (inherits tenant_id from sale relationship)
-- No direct tenant_id needed since it's linked to sale

-- Update invoice_items table (inherits tenant_id from invoice relationship)
-- No direct tenant_id needed since it's linked to invoice

-- Row Level Security Policies for tenant isolation
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;

-- RLS Policies for tenants table (system-level access only)
CREATE POLICY tenant_isolation_tenants ON tenants
    USING (true); -- Will be controlled at application level

-- RLS Policies for tenant_subscriptions
CREATE POLICY tenant_isolation_subscriptions ON tenant_subscriptions
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- RLS Policies for tenant_settings
CREATE POLICY tenant_isolation_settings ON tenant_settings
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- RLS Policies for users
CREATE POLICY tenant_isolation_users ON users
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- RLS Policies for products
CREATE POLICY tenant_isolation_products ON products
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- RLS Policies for sales
CREATE POLICY tenant_isolation_sales ON sales
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- RLS Policies for invoices
CREATE POLICY tenant_isolation_invoices ON invoices
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Create updated_at triggers for new tables
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenant_subscriptions_updated_at BEFORE UPDATE ON tenant_subscriptions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenant_settings_updated_at BEFORE UPDATE ON tenant_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create default tenant for existing data (if any)
INSERT INTO tenants (name, slug, domain, status, configuration, created_at, updated_at)
VALUES (
    'Default Tenant', 
    'default',
    'default.adol.pos',
    'active',
    '{"business_info": {"name": "Default Store", "currency": "USD", "tax_rate": 0}, "pos_settings": {"auto_print": true, "receipt_template": "standard"}, "feature_flags": {"advanced_reporting": true, "multi_location": false, "api_access": true}}',
    NOW(),
    NOW()
);

-- Create default subscription for the default tenant
INSERT INTO tenant_subscriptions (tenant_id, plan_type, status, monthly_fee, features, usage_limits, created_at, updated_at)
SELECT 
    id,
    'enterprise',
    'active',
    0.00,
    '{"pos": true, "inventory": true, "reporting": true, "multi_location": false, "api_access": true, "advanced_reporting": true}',
    '{"users": -1, "products": -1, "sales_per_month": -1, "api_calls_per_month": -1}',
    NOW(),
    NOW()
FROM tenants WHERE slug = 'default';

-- Update existing users to belong to default tenant
UPDATE users 
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default')
WHERE tenant_id IS NULL;

-- Update existing products to belong to default tenant
UPDATE products 
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default')
WHERE tenant_id IS NULL;

-- Update existing sales to belong to default tenant
UPDATE sales 
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default')
WHERE tenant_id IS NULL;

-- Update existing invoices to belong to default tenant
UPDATE invoices 
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default')
WHERE tenant_id IS NULL;

-- Make tenant_id NOT NULL after data migration
ALTER TABLE users ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE products ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE sales ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE invoices ALTER COLUMN tenant_id SET NOT NULL;

-- Add composite indexes for better query performance
CREATE INDEX idx_users_tenant_email ON users(tenant_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_tenant_username ON users(tenant_id, username) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_tenant_category ON products(tenant_id, category) WHERE deleted_at IS NULL;
CREATE INDEX idx_sales_tenant_status ON sales(tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_tenant_status ON invoices(tenant_id, status) WHERE deleted_at IS NULL;