-- Modify "tenants" table
ALTER TABLE "tenants" ADD COLUMN "type" character varying NOT NULL DEFAULT 'PERSONAL';
-- Create index "tenant_id_owner_id_type" to table: "tenants"
CREATE INDEX "tenant_id_owner_id_type" ON "tenants" ("id", "owner_id", "type");
