INSERT INTO contact_info_validations
(created_at, updated_at, reference, token, entity_id, entity_type, contact_method, value, used, expires_at)
SELECT created_at, updated_at, reference, token, user_uuid, 'user', 1, email_address, false, expires_at
FROM email_validations
where (expires_at > now() or expires_at is null) and used = false;

--bun:split
DROP TABLE IF EXISTS email_validations;
