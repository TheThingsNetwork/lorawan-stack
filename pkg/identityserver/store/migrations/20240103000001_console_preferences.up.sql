ALTER TABLE users
ADD COLUMN IF NOT EXISTS console_preferences bytea;

--bun:split
CREATE TABLE user_bookmarks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,

  user_id character varying(36) NOT NULL,
  entity_id character varying(36) NOT NULL,
  entity_type character varying(32) NOT NULL
);

CREATE UNIQUE INDEX user_bookmarks_user_id_entity_id_entity_type_idx
  ON user_bookmarks (user_id, entity_id, entity_type);

--bun:split
CREATE OR REPLACE VIEW user_accounts AS
SELECT
  acc.id AS account_id,
  acc.created_at AS account_created_at,
  acc.updated_at AS account_updated_at,
  acc.deleted_at AS account_deleted_at,
  acc.uid AS account_uid,
  usr.*
FROM
  accounts acc
  JOIN users usr ON usr.id = acc.account_id
  AND acc.account_type = 'user';
