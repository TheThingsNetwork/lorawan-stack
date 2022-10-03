DROP INDEX gateway_eui_index;
CREATE UNIQUE INDEX gateway_eui_index ON gateways USING btree (gateway_eui);
