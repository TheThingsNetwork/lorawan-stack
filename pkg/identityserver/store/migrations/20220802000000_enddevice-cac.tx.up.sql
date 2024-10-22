ALTER TABLE end_devices ADD claim_authentication_code_secret bytea;
ALTER TABLE end_devices ADD claim_authentication_code_valid_from timestamp with time zone;
ALTER TABLE end_devices ADD claim_authentication_code_valid_to timestamp with time zone;
