-- Modify "tenants" table
ALTER TABLE "tenants" ADD COLUMN "type" character varying NOT NULL DEFAULT 'PERSONAL';
-- Create index "tenant_owner_id_type" to table: "tenants"
CREATE INDEX "tenant_owner_id_type" ON "tenants" ("owner_id", "type");
