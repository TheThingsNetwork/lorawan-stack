CREATE TABLE email_validations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  expires_at timestamp with time zone,

  user_uuid uuid NOT NULL,
  email_address character varying NOT NULL,
  reference character varying NOT NULL,
  token character varying NOT NULL,
  used boolean DEFAULT FALSE NOT NULL
);

CREATE UNIQUE INDEX email_validation_index ON email_validations USING btree (reference, token);

--bun:split
INSERT INTO email_validations (created_at, updated_at, expires_at, user_uuid, email_address, reference, token, used)
SELECT created_at, updated_at, expires_at, entity_id, value, reference, token, false
FROM contact_info_validations
WHERE contact_method=1
  AND entity_type='user'
  AND (used = false OR used IS NULL)
  AND (expires_at > now() OR expires_at IS null);

--bun:split
delete from contact_info_validations
where contact_method = 1 and entity_type = 'user'
;
