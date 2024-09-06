DROP INDEX IF EXISTS "invite_tenant_id_email_status";
DROP INDEX IF EXISTS "invite_tenant_id_user_id_status";

CREATE UNIQUE INDEX "invite_tenant_id_email" ON "invites" ("tenant_id", "email");
CREATE UNIQUE INDEX "invite_tenant_id_user_id" ON "invites" ("tenant_id", "user_id");
