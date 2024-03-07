CREATE TABLE user_bookmarks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,

  user_id character varying(36) NOT NULL,
  entity_id character varying(36) NOT NULL,
  entity_type character varying(32) NOT NULL
);

CREATE UNIQUE INDEX user_bookmarks_user_id_entity_id_entity_type_idx
  ON user_bookmarks (user_id, entity_id, entity_type);
