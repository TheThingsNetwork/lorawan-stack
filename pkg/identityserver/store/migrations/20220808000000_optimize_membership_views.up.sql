DROP VIEW IF EXISTS direct_entity_memberships CASCADE;
DROP VIEW IF EXISTS indirect_entity_memberships CASCADE;
DROP VIEW IF EXISTS entity_friendly_ids CASCADE;

--bun:split

CREATE VIEW direct_entity_memberships AS
SELECT
  acc.account_type AS account_type,
  acc.id AS account_id,
  acc.uid AS account_friendly_id,
  mem.rights AS rights,
  mem.entity_type AS entity_type,
  mem.entity_id AS entity_id,
  CASE
    WHEN mem.entity_type = 'application' THEN (SELECT application_id FROM applications WHERE id = mem.entity_id)
    WHEN mem.entity_type = 'client' THEN (SELECT client_id FROM clients WHERE id = mem.entity_id)
    WHEN mem.entity_type = 'gateway' THEN (SELECT gateway_id FROM gateways WHERE id = mem.entity_id)
    WHEN mem.entity_type = 'organization' THEN (SELECT uid FROM accounts WHERE account_type = 'organization' AND account_id = mem.entity_id)
  END AS entity_friendly_id
FROM
  accounts AS acc
  JOIN memberships AS mem ON mem.account_id = acc.id
WHERE
  acc.deleted_at IS NULL;

--bun:split

CREATE VIEW indirect_entity_memberships AS
SELECT
  usr_acc.id AS user_account_id,
  usr_acc.uid AS user_account_friendly_id,
  dmem.rights AS user_rights,
  org_acc.id AS organization_account_id,
  org_acc.uid AS organization_account_friendly_id,
  imem.rights AS entity_rights,
  imem.entity_type AS entity_type,
  imem.entity_id AS entity_id,
  CASE
    WHEN imem.entity_type = 'application' THEN (SELECT application_id FROM applications WHERE id = imem.entity_id)
    WHEN imem.entity_type = 'client' THEN (SELECT client_id FROM clients WHERE id = imem.entity_id)
    WHEN imem.entity_type = 'gateway' THEN (SELECT gateway_id FROM gateways WHERE id = imem.entity_id)
  END AS entity_friendly_id
FROM
  accounts AS usr_acc
  JOIN memberships AS dmem ON dmem.account_id = usr_acc.id
  JOIN accounts org_acc ON dmem.entity_type = org_acc.account_type
  AND dmem.entity_id = org_acc.account_id
  JOIN memberships AS imem ON imem.account_id = org_acc.id
WHERE
  usr_acc.deleted_at IS NULL
  AND usr_acc.account_type = 'user'
  AND dmem.entity_type = 'organization'
  AND org_acc.deleted_at IS NULL;
