-- The primary email address of a user should be unique only amount the active accounts, otherwise
-- this will prevent the registration of new users with the same email address as the deleted account.

DROP INDEX uix_users_primary_email_address;
CREATE UNIQUE INDEX uix_users_primary_email_address ON users USING btree ((LOWER(primary_email_address))) WHERE deleted_at IS NULL;
