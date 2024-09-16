-- Drop index "invite_tenant_id_email_status" from table: "invites"
DROP INDEX "invite_tenant_id_email_status";
-- Drop index "invite_tenant_id_user_id_status" from table: "invites"
DROP INDEX "invite_tenant_id_user_id_status";
-- Create index "invite_tenant_id_email" to table: "invites"
CREATE UNIQUE INDEX "invite_tenant_id_email" ON "invites" ("tenant_id", "email");
-- Create index "invite_tenant_id_user_id" to table: "invites"
CREATE UNIQUE INDEX "invite_tenant_id_user_id" ON "invites" ("tenant_id", "user_id");
