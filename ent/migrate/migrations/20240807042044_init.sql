-- Drop index "invite_tenant_id_email" from table: "invites"
DROP INDEX "invite_tenant_id_email";
-- Modify "invites" table
ALTER TABLE "invites" ADD COLUMN "role_id" bigint NULL, ADD COLUMN "resource" character varying NULL, ADD COLUMN "resource_id" bigint NULL;
-- Create index "invite_tenant_id_email_status" to table: "invites"
CREATE UNIQUE INDEX "invite_tenant_id_email_status" ON "invites" ("tenant_id", "email", "status") WHERE (((status)::text = 'accepted'::text) AND (email IS NOT NULL));
-- Create index "invite_tenant_id_user_id_status" to table: "invites"
CREATE UNIQUE INDEX "invite_tenant_id_user_id_status" ON "invites" ("tenant_id", "user_id", "status") WHERE (((status)::text = 'accepted'::text) AND (user_id IS NOT NULL));
