DROP INDEX uix_users_primary_email_address;
CREATE UNIQUE INDEX uix_users_primary_email_address ON users USING btree (primary_email_address);
