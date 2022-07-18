ALTER TABLE end_devices ADD end_device_cac_secret bytea;
ALTER TABLE end_devices ADD end_device_cac_valid_from timestamp with time zone;
ALTER TABLE end_devices ADD end_device_cac_valid_to timestamp with time zone;
