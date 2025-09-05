-- Rollback Multi-Tenant Management Schema
-- This migration rolls back the tenant management tables and related changes

-- Drop RLS policies first
DROP POLICY IF EXISTS tenant_isolation_invoices ON invoices;
DROP POLICY IF EXISTS tenant_isolation_sales ON sales;
DROP POLICY IF EXISTS tenant_isolation_products ON products;
DROP POLICY IF EXISTS tenant_isolation_users ON users;
DROP POLICY IF EXISTS tenant_isolation_settings ON tenant_settings;
DROP POLICY IF EXISTS tenant_isolation_subscriptions ON tenant_subscriptions;
DROP POLICY IF EXISTS tenant_isolation_tenants ON tenants;

-- Disable RLS
ALTER TABLE invoices DISABLE ROW LEVEL SECURITY;
ALTER TABLE sales DISABLE ROW LEVEL SECURITY;
ALTER TABLE products DISABLE ROW LEVEL SECURITY;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_settings DISABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_subscriptions DISABLE ROW LEVEL SECURITY;
ALTER TABLE tenants DISABLE ROW LEVEL SECURITY;

-- Drop composite indexes
DROP INDEX IF EXISTS idx_invoices_tenant_status;
DROP INDEX IF EXISTS idx_sales_tenant_status;
DROP INDEX IF EXISTS idx_products_tenant_category;
DROP INDEX IF EXISTS idx_users_tenant_username;
DROP INDEX IF EXISTS idx_users_tenant_email;

-- Drop unique constraints for tenant-scoped uniqueness
DROP INDEX IF EXISTS uk_invoices_tenant_invoice_number;
DROP INDEX IF EXISTS uk_sales_tenant_sale_number;
DROP INDEX IF EXISTS uk_products_tenant_sku;

-- Restore original unique constraints
CREATE UNIQUE CONSTRAINT invoices_invoice_number_key ON invoices(invoice_number);
CREATE UNIQUE CONSTRAINT sales_sale_number_key ON sales(sale_number);
CREATE UNIQUE CONSTRAINT products_sku_key ON products(sku);

-- Drop tenant_id indexes
DROP INDEX IF EXISTS idx_invoices_tenant_id;
DROP INDEX IF EXISTS idx_sales_tenant_id;
DROP INDEX IF EXISTS idx_products_tenant_id;
DROP INDEX IF EXISTS idx_users_tenant_id;

-- Remove tenant_id columns from existing tables
ALTER TABLE invoices DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE sales DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE products DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;

-- Drop tenant management tables
DROP TABLE IF EXISTS tenant_settings;
DROP TABLE IF EXISTS tenant_subscriptions;
DROP TABLE IF EXISTS tenants;

-- Drop indexes for tenant tables
DROP INDEX IF EXISTS idx_tenant_settings_tenant_id;
DROP INDEX IF EXISTS idx_tenant_subscriptions_plan_type;
DROP INDEX IF EXISTS idx_tenant_subscriptions_status;
DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_domain;
DROP INDEX IF EXISTS idx_tenants_slug;