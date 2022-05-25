CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE migrations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  name character varying
);

CREATE UNIQUE INDEX migration_name_index ON migrations USING btree (name);

CREATE TABLE accounts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  uid character varying(36),
  account_id uuid NOT NULL,
  account_type character varying(32) NOT NULL
);

CREATE UNIQUE INDEX account_uid_index ON accounts USING btree (uid);

CREATE INDEX idx_accounts_deleted_at ON accounts USING btree (deleted_at);

CREATE INDEX account_id_index ON accounts USING btree (account_id, account_type);

CREATE TABLE attributes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  entity_id uuid NOT NULL,
  entity_type character varying(32) NOT NULL,
  key character varying,
  value character varying
);

CREATE INDEX attribute_entity_index ON attributes USING btree (entity_id, entity_type);

CREATE TABLE contact_infos (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  contact_type integer NOT NULL,
  contact_method integer NOT NULL,
  value character varying,
  public boolean,
  validated_at timestamp with time zone,
  entity_id uuid NOT NULL,
  entity_type character varying(32) NOT NULL
);

CREATE INDEX contact_info_entity_index ON contact_infos USING btree (entity_id, entity_type);

CREATE TABLE contact_info_validations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  reference character varying,
  token character varying,
  entity_id uuid NOT NULL,
  entity_type character varying(32) NOT NULL,
  contact_method integer NOT NULL,
  value character varying,
  used boolean,
  expires_at timestamp with time zone
);

CREATE INDEX contact_info_validation_id_index ON contact_info_validations USING btree (reference, token);

CREATE INDEX contact_info_validation_entity_index ON contact_info_validations USING btree (entity_id, entity_type);

CREATE TABLE pictures (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  data bytea
);

CREATE INDEX idx_pictures_deleted_at ON pictures USING btree (deleted_at);

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  name character varying,
  description text,
  primary_email_address character varying NOT NULL,
  primary_email_address_validated_at timestamp with time zone,
  password character varying NOT NULL,
  password_updated_at timestamp with time zone NOT NULL,
  require_password_update boolean NOT NULL,
  state integer NOT NULL,
  state_description character varying,
  admin boolean NOT NULL,
  temporary_password character varying,
  temporary_password_created_at timestamp with time zone,
  temporary_password_expires_at timestamp with time zone,
  profile_picture_id uuid
);

CREATE INDEX idx_users_deleted_at ON users USING btree (deleted_at);

CREATE UNIQUE INDEX uix_users_primary_email_address ON users USING btree (primary_email_address);

CREATE INDEX user_profile_picture_index ON users USING btree (profile_picture_id);

CREATE TABLE user_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  user_id uuid NOT NULL,
  session_secret character varying,
  expires_at timestamp with time zone
);

CREATE INDEX user_session_user_index ON user_sessions USING btree (user_id);

CREATE TABLE login_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  user_id uuid,
  token character varying NOT NULL,
  expires_at timestamp with time zone,
  used boolean
);

CREATE UNIQUE INDEX login_token_index ON login_tokens USING btree (token);

CREATE TABLE invitations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  email character varying NOT NULL,
  token character varying NOT NULL,
  expires_at timestamp with time zone,
  accepted_by_id uuid,
  accepted_at timestamp with time zone
);

CREATE UNIQUE INDEX invitation_email_index ON invitations USING btree (email);

CREATE UNIQUE INDEX invitation_token_index ON invitations USING btree (token);

CREATE TABLE memberships (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  account_id uuid NOT NULL,
  rights integer [],
  entity_id uuid NOT NULL,
  entity_type character varying(32) NOT NULL
);

CREATE INDEX membership_account_index ON memberships USING btree (account_id);

CREATE INDEX membership_entity_index ON memberships USING btree (entity_id, entity_type);

CREATE TABLE organizations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  name character varying,
  description text,
  administrative_contact_id uuid,
  technical_contact_id uuid
);

CREATE INDEX idx_organizations_deleted_at ON organizations USING btree (deleted_at);

CREATE INDEX idx_organizations_administrative_contact_id ON organizations USING btree (administrative_contact_id);

CREATE INDEX idx_organizations_technical_contact_id ON organizations USING btree (technical_contact_id);

CREATE TABLE gateways (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  gateway_eui character varying(16),
  gateway_id character varying(36) NOT NULL,
  name character varying,
  description text,
  administrative_contact_id uuid,
  technical_contact_id uuid,
  brand_id character varying,
  model_id character varying,
  hardware_version character varying,
  firmware_version character varying,
  gateway_server_address character varying,
  auto_update boolean NOT NULL,
  update_channel character varying,
  frequency_plan_id character varying,
  status_public boolean NOT NULL,
  location_public boolean NOT NULL,
  schedule_downlink_late boolean NOT NULL,
  enforce_duty_cycle boolean NOT NULL,
  schedule_anytime_delay bigint DEFAULT 0 NOT NULL,
  downlink_path_constraint integer,
  update_location_from_status boolean DEFAULT false NOT NULL,
  lbs_lns_secret bytea,
  claim_authentication_code_secret bytea,
  claim_authentication_code_valid_from timestamp with time zone,
  claim_authentication_code_valid_to timestamp with time zone,
  target_cups_uri character varying,
  target_cups_key bytea,
  require_authenticated_connection boolean,
  supports_lrfhss boolean DEFAULT false NOT NULL,
  disable_packet_broker_forwarding boolean DEFAULT false NOT NULL
);

CREATE INDEX idx_gateways_deleted_at ON gateways USING btree (deleted_at);

CREATE UNIQUE INDEX gateway_eui_index ON gateways USING btree (gateway_eui);

CREATE UNIQUE INDEX gateway_id_index ON gateways USING btree (gateway_id);

CREATE INDEX idx_gateways_administrative_contact_id ON gateways USING btree (administrative_contact_id);

CREATE INDEX idx_gateways_technical_contact_id ON gateways USING btree (technical_contact_id);

CREATE TABLE gateway_antennas (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  gateway_id uuid NOT NULL,
  index integer NOT NULL,
  gain numeric,
  latitude numeric,
  longitude numeric,
  altitude integer,
  accuracy integer,
  placement integer
);

CREATE INDEX gateway_antenna_gateway_index ON gateway_antennas USING btree (gateway_id);

CREATE UNIQUE INDEX gateway_antenna_id_index ON gateway_antennas USING btree (gateway_id, index);

CREATE TABLE applications (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  application_id character varying(36) NOT NULL,
  name character varying,
  description text,
  administrative_contact_id uuid,
  technical_contact_id uuid,
  network_server_address character varying,
  application_server_address character varying,
  join_server_address character varying,
  dev_eui_counter integer DEFAULT 0
);

CREATE INDEX idx_applications_deleted_at ON applications USING btree (deleted_at);

CREATE UNIQUE INDEX application_id_index ON applications USING btree (application_id);

CREATE INDEX idx_applications_administrative_contact_id ON applications USING btree (administrative_contact_id);

CREATE INDEX idx_applications_technical_contact_id ON applications USING btree (technical_contact_id);

CREATE TABLE eui_blocks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  type character varying(10),
  start_eui character varying(16),
  end_counter bigint,
  current_counter bigint
);

CREATE TABLE end_devices (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  application_id character varying(36) NOT NULL,
  device_id character varying(36) NOT NULL,
  name character varying,
  description text,
  join_eui character varying(16),
  dev_eui character varying(16),
  brand_id character varying,
  model_id character varying,
  hardware_version character varying,
  firmware_version character varying,
  band_id character varying,
  network_server_address character varying,
  application_server_address character varying,
  join_server_address character varying,
  service_profile_id character varying,
  picture_id uuid,
  activated_at timestamp with time zone,
  last_seen_at timestamp with time zone
);

CREATE INDEX end_device_application_index ON end_devices USING btree (application_id);

CREATE UNIQUE INDEX end_device_id_index ON end_devices USING btree (application_id, device_id);

CREATE INDEX end_device_join_eui_index ON end_devices USING btree (join_eui);

CREATE INDEX end_device_dev_eui_index ON end_devices USING btree (dev_eui);

CREATE UNIQUE INDEX end_device_eui_index ON end_devices USING btree (join_eui, dev_eui);

CREATE INDEX end_device_picture_index ON end_devices USING btree (picture_id);

CREATE TABLE end_device_locations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  end_device_id uuid NOT NULL,
  service text,
  latitude numeric,
  longitude numeric,
  altitude integer,
  accuracy integer,
  source integer NOT NULL
);

CREATE INDEX end_device_device_index ON end_device_locations USING btree (end_device_id);

CREATE UNIQUE INDEX end_device_location_id_index ON end_device_locations USING btree (end_device_id, service);

CREATE TABLE clients (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  client_id character varying(36) NOT NULL,
  name character varying,
  description text,
  administrative_contact_id uuid,
  technical_contact_id uuid,
  client_secret character varying,
  redirect_uris character varying [],
  logout_redirect_uris character varying [],
  state integer NOT NULL,
  state_description character varying,
  skip_authorization boolean NOT NULL,
  endorsed boolean NOT NULL,
  grants integer [],
  rights integer []
);

CREATE INDEX idx_clients_deleted_at ON clients USING btree (deleted_at);

CREATE UNIQUE INDEX client_id_index ON clients USING btree (client_id);

CREATE INDEX idx_clients_administrative_contact_id ON clients USING btree (administrative_contact_id);

CREATE INDEX idx_clients_technical_contact_id ON clients USING btree (technical_contact_id);

CREATE TABLE client_authorizations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  client_id uuid NOT NULL,
  user_id uuid NOT NULL,
  rights integer []
);

CREATE INDEX idx_client_authorizations_client_id ON client_authorizations USING btree (client_id);

CREATE INDEX idx_client_authorizations_user_id ON client_authorizations USING btree (user_id);

CREATE TABLE authorization_codes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  client_id uuid NOT NULL,
  user_id uuid NOT NULL,
  user_session_id uuid,
  rights integer [],
  code character varying NOT NULL,
  redirect_uri character varying,
  state character varying,
  expires_at timestamp with time zone
);

CREATE INDEX idx_authorization_codes_client_id ON authorization_codes USING btree (client_id);

CREATE INDEX idx_authorization_codes_user_id ON authorization_codes USING btree (user_id);

CREATE INDEX idx_authorization_codes_user_session_id ON authorization_codes USING btree (user_session_id);

CREATE UNIQUE INDEX authorization_code_code_index ON authorization_codes USING btree (code);

CREATE TABLE access_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  client_id uuid NOT NULL,
  user_id uuid NOT NULL,
  user_session_id uuid,
  rights integer [],
  token_id character varying NOT NULL,
  previous_id character varying,
  access_token character varying NOT NULL,
  refresh_token character varying NOT NULL,
  expires_at timestamp with time zone
);

CREATE INDEX idx_access_tokens_client_id ON access_tokens USING btree (client_id);

CREATE INDEX idx_access_tokens_user_id ON access_tokens USING btree (user_id);

CREATE INDEX idx_access_tokens_user_session_id ON access_tokens USING btree (user_session_id);

CREATE UNIQUE INDEX access_token_id_index ON access_tokens USING btree (token_id);

CREATE INDEX access_token_previous_index ON access_tokens USING btree (previous_id);

CREATE TABLE api_keys (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  api_key_id character varying,
  key character varying,
  rights integer [],
  name character varying,
  entity_id uuid NOT NULL,
  entity_type character varying(32) NOT NULL,
  expires_at timestamp with time zone
);

CREATE UNIQUE INDEX api_key_id_index ON api_keys USING btree (api_key_id);

CREATE INDEX api_key_entity_index ON api_keys USING btree (entity_id, entity_type);

CREATE TABLE notifications (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  entity_id uuid NOT NULL,
  entity_type character varying(32) NOT NULL,
  entity_uid character varying(36) NOT NULL,
  notification_type text NOT NULL,
  data jsonb,
  sender_id uuid,
  sender_uid character varying(36) NOT NULL,
  receivers integer [],
  email boolean NOT NULL
);

CREATE INDEX notification_entity_index ON notifications USING btree (entity_id, entity_type);

CREATE INDEX notification_sender_index ON notifications USING btree (sender_id);

CREATE TABLE notification_receivers (
  notification_id uuid NOT NULL,
  receiver_id uuid NOT NULL,
  status integer NOT NULL,
  status_updated_at timestamp with time zone NOT NULL
);

CREATE INDEX notification_receiver_notification_id_index ON notification_receivers USING btree (notification_id);

CREATE INDEX notification_receiver_user_index ON notification_receivers USING btree (receiver_id);

CREATE UNIQUE INDEX notification_receiver_index ON notification_receivers USING btree (notification_id, receiver_id);
