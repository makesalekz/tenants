-- Modify "invites" table
ALTER TABLE "invites" ADD COLUMN "role_id" bigint NULL, ADD COLUMN "resource" character varying NULL, ADD COLUMN "resource_id" bigint NULL;
