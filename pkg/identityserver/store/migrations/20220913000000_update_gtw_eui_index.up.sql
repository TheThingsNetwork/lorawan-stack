-- The gateway_eui_index should be unique only amongst the non deleted gateways, otherwise it will prevent the creation
-- of gateways with the same eui as deleted entities.
DROP INDEX gateway_eui_index;
CREATE UNIQUE INDEX gateway_eui_index ON gateways USING btree (gateway_eui) WHERE deleted_at IS NULL;
