ALTER TABLE organizations
DROP COLUMN IF EXISTS fanout_notifications CASCADE;

--bun:split
CREATE OR REPLACE VIEW organization_accounts AS
SELECT
  acc.id AS account_id,
  acc.created_at AS account_created_at,
  acc.updated_at AS account_updated_at,
  acc.deleted_at AS account_deleted_at,
  acc.uid AS account_uid,
  org.*
FROM
  accounts acc
  JOIN organizations org ON org.id = acc.account_id
  AND acc.account_type = 'organization';
