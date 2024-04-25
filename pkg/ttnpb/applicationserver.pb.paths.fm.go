// Code generated by protoc-gen-fieldmask. DO NOT EDIT.

package ttnpb

var ApplicationLinkFieldPathsNested = []string{
	"default_formatters",
	"default_formatters.down_formatter",
	"default_formatters.down_formatter_parameter",
	"default_formatters.up_formatter",
	"default_formatters.up_formatter_parameter",
	"skip_payload_crypto",
}

var ApplicationLinkFieldPathsTopLevel = []string{
	"default_formatters",
	"skip_payload_crypto",
}
var GetApplicationLinkRequestFieldPathsNested = []string{
	"application_ids",
	"application_ids.application_id",
	"field_mask",
}

var GetApplicationLinkRequestFieldPathsTopLevel = []string{
	"application_ids",
	"field_mask",
}
var SetApplicationLinkRequestFieldPathsNested = []string{
	"application_ids",
	"application_ids.application_id",
	"field_mask",
	"link",
	"link.default_formatters",
	"link.default_formatters.down_formatter",
	"link.default_formatters.down_formatter_parameter",
	"link.default_formatters.up_formatter",
	"link.default_formatters.up_formatter_parameter",
	"link.skip_payload_crypto",
}

var SetApplicationLinkRequestFieldPathsTopLevel = []string{
	"application_ids",
	"field_mask",
	"link",
}
var ApplicationLinkStatsFieldPathsNested = []string{
	"downlink_count",
	"last_downlink_forwarded_at",
	"last_up_received_at",
	"linked_at",
	"network_server_address",
	"up_count",
}

var ApplicationLinkStatsFieldPathsTopLevel = []string{
	"downlink_count",
	"last_downlink_forwarded_at",
	"last_up_received_at",
	"linked_at",
	"network_server_address",
	"up_count",
}
var AsConfigurationFieldPathsNested = []string{
	"pubsub",
	"pubsub.providers",
	"pubsub.providers.mqtt",
	"pubsub.providers.nats",
	"webhooks",
	"webhooks.unhealthy_attempts_threshold",
	"webhooks.unhealthy_retry_interval",
}

var AsConfigurationFieldPathsTopLevel = []string{
	"pubsub",
	"webhooks",
}
var GetAsConfigurationRequestFieldPathsNested []string
var GetAsConfigurationRequestFieldPathsTopLevel []string
var GetAsConfigurationResponseFieldPathsNested = []string{
	"configuration",
	"configuration.pubsub",
	"configuration.pubsub.providers",
	"configuration.pubsub.providers.mqtt",
	"configuration.pubsub.providers.nats",
	"configuration.webhooks",
	"configuration.webhooks.unhealthy_attempts_threshold",
	"configuration.webhooks.unhealthy_retry_interval",
}

var GetAsConfigurationResponseFieldPathsTopLevel = []string{
	"configuration",
}
var NsAsHandleUplinkRequestFieldPathsNested = []string{
	"application_ups",
}

var NsAsHandleUplinkRequestFieldPathsTopLevel = []string{
	"application_ups",
}
var EncodeDownlinkRequestFieldPathsNested = []string{
	"downlink",
	"downlink.class_b_c",
	"downlink.class_b_c.absolute_time",
	"downlink.class_b_c.gateways",
	"downlink.confirmed",
	"downlink.confirmed_retry",
	"downlink.confirmed_retry.attempt",
	"downlink.confirmed_retry.max_attempts",
	"downlink.correlation_ids",
	"downlink.decoded_payload",
	"downlink.decoded_payload_warnings",
	"downlink.f_cnt",
	"downlink.f_port",
	"downlink.frm_payload",
	"downlink.priority",
	"downlink.session_key_id",
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"formatter",
	"parameter",
	"version_ids",
	"version_ids.band_id",
	"version_ids.brand_id",
	"version_ids.firmware_version",
	"version_ids.hardware_version",
	"version_ids.model_id",
}

var EncodeDownlinkRequestFieldPathsTopLevel = []string{
	"downlink",
	"end_device_ids",
	"formatter",
	"parameter",
	"version_ids",
}
var EncodeDownlinkResponseFieldPathsNested = []string{
	"downlink",
	"downlink.class_b_c",
	"downlink.class_b_c.absolute_time",
	"downlink.class_b_c.gateways",
	"downlink.confirmed",
	"downlink.confirmed_retry",
	"downlink.confirmed_retry.attempt",
	"downlink.confirmed_retry.max_attempts",
	"downlink.correlation_ids",
	"downlink.decoded_payload",
	"downlink.decoded_payload_warnings",
	"downlink.f_cnt",
	"downlink.f_port",
	"downlink.frm_payload",
	"downlink.priority",
	"downlink.session_key_id",
}

var EncodeDownlinkResponseFieldPathsTopLevel = []string{
	"downlink",
}
var DecodeUplinkRequestFieldPathsNested = []string{
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"formatter",
	"parameter",
	"uplink",
	"uplink.app_s_key",
	"uplink.app_s_key.encrypted_key",
	"uplink.app_s_key.kek_label",
	"uplink.app_s_key.key",
	"uplink.confirmed",
	"uplink.consumed_airtime",
	"uplink.decoded_payload",
	"uplink.decoded_payload_warnings",
	"uplink.f_cnt",
	"uplink.f_port",
	"uplink.frm_payload",
	"uplink.last_a_f_cnt_down",
	"uplink.locations",
	"uplink.network_ids",
	"uplink.network_ids.cluster_address",
	"uplink.network_ids.cluster_id",
	"uplink.network_ids.net_id",
	"uplink.network_ids.ns_id",
	"uplink.network_ids.tenant_address",
	"uplink.network_ids.tenant_id",
	"uplink.normalized_payload",
	"uplink.normalized_payload_warnings",
	"uplink.packet_error_rate",
	"uplink.received_at",
	"uplink.rx_metadata",
	"uplink.session_key_id",
	"uplink.settings",
	"uplink.settings.concentrator_timestamp",
	"uplink.settings.data_rate",
	"uplink.settings.data_rate.modulation",
	"uplink.settings.data_rate.modulation.fsk",
	"uplink.settings.data_rate.modulation.fsk.bit_rate",
	"uplink.settings.data_rate.modulation.lora",
	"uplink.settings.data_rate.modulation.lora.bandwidth",
	"uplink.settings.data_rate.modulation.lora.coding_rate",
	"uplink.settings.data_rate.modulation.lora.spreading_factor",
	"uplink.settings.data_rate.modulation.lrfhss",
	"uplink.settings.data_rate.modulation.lrfhss.coding_rate",
	"uplink.settings.data_rate.modulation.lrfhss.modulation_type",
	"uplink.settings.data_rate.modulation.lrfhss.operating_channel_width",
	"uplink.settings.downlink",
	"uplink.settings.downlink.antenna_index",
	"uplink.settings.downlink.invert_polarization",
	"uplink.settings.downlink.tx_power",
	"uplink.settings.enable_crc",
	"uplink.settings.frequency",
	"uplink.settings.time",
	"uplink.settings.timestamp",
	"uplink.version_ids",
	"uplink.version_ids.band_id",
	"uplink.version_ids.brand_id",
	"uplink.version_ids.firmware_version",
	"uplink.version_ids.hardware_version",
	"uplink.version_ids.model_id",
	"version_ids",
	"version_ids.band_id",
	"version_ids.brand_id",
	"version_ids.firmware_version",
	"version_ids.hardware_version",
	"version_ids.model_id",
}

var DecodeUplinkRequestFieldPathsTopLevel = []string{
	"end_device_ids",
	"formatter",
	"parameter",
	"uplink",
	"version_ids",
}
var DecodeUplinkResponseFieldPathsNested = []string{
	"uplink",
	"uplink.app_s_key",
	"uplink.app_s_key.encrypted_key",
	"uplink.app_s_key.kek_label",
	"uplink.app_s_key.key",
	"uplink.confirmed",
	"uplink.consumed_airtime",
	"uplink.decoded_payload",
	"uplink.decoded_payload_warnings",
	"uplink.f_cnt",
	"uplink.f_port",
	"uplink.frm_payload",
	"uplink.last_a_f_cnt_down",
	"uplink.locations",
	"uplink.network_ids",
	"uplink.network_ids.cluster_address",
	"uplink.network_ids.cluster_id",
	"uplink.network_ids.net_id",
	"uplink.network_ids.ns_id",
	"uplink.network_ids.tenant_address",
	"uplink.network_ids.tenant_id",
	"uplink.normalized_payload",
	"uplink.normalized_payload_warnings",
	"uplink.packet_error_rate",
	"uplink.received_at",
	"uplink.rx_metadata",
	"uplink.session_key_id",
	"uplink.settings",
	"uplink.settings.concentrator_timestamp",
	"uplink.settings.data_rate",
	"uplink.settings.data_rate.modulation",
	"uplink.settings.data_rate.modulation.fsk",
	"uplink.settings.data_rate.modulation.fsk.bit_rate",
	"uplink.settings.data_rate.modulation.lora",
	"uplink.settings.data_rate.modulation.lora.bandwidth",
	"uplink.settings.data_rate.modulation.lora.coding_rate",
	"uplink.settings.data_rate.modulation.lora.spreading_factor",
	"uplink.settings.data_rate.modulation.lrfhss",
	"uplink.settings.data_rate.modulation.lrfhss.coding_rate",
	"uplink.settings.data_rate.modulation.lrfhss.modulation_type",
	"uplink.settings.data_rate.modulation.lrfhss.operating_channel_width",
	"uplink.settings.downlink",
	"uplink.settings.downlink.antenna_index",
	"uplink.settings.downlink.invert_polarization",
	"uplink.settings.downlink.tx_power",
	"uplink.settings.enable_crc",
	"uplink.settings.frequency",
	"uplink.settings.time",
	"uplink.settings.timestamp",
	"uplink.version_ids",
	"uplink.version_ids.band_id",
	"uplink.version_ids.brand_id",
	"uplink.version_ids.firmware_version",
	"uplink.version_ids.hardware_version",
	"uplink.version_ids.model_id",
}

var DecodeUplinkResponseFieldPathsTopLevel = []string{
	"uplink",
}
var DecodeDownlinkRequestFieldPathsNested = []string{
	"downlink",
	"downlink.class_b_c",
	"downlink.class_b_c.absolute_time",
	"downlink.class_b_c.gateways",
	"downlink.confirmed",
	"downlink.confirmed_retry",
	"downlink.confirmed_retry.attempt",
	"downlink.confirmed_retry.max_attempts",
	"downlink.correlation_ids",
	"downlink.decoded_payload",
	"downlink.decoded_payload_warnings",
	"downlink.f_cnt",
	"downlink.f_port",
	"downlink.frm_payload",
	"downlink.priority",
	"downlink.session_key_id",
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"formatter",
	"parameter",
	"version_ids",
	"version_ids.band_id",
	"version_ids.brand_id",
	"version_ids.firmware_version",
	"version_ids.hardware_version",
	"version_ids.model_id",
}

var DecodeDownlinkRequestFieldPathsTopLevel = []string{
	"downlink",
	"end_device_ids",
	"formatter",
	"parameter",
	"version_ids",
}
var DecodeDownlinkResponseFieldPathsNested = []string{
	"downlink",
	"downlink.class_b_c",
	"downlink.class_b_c.absolute_time",
	"downlink.class_b_c.gateways",
	"downlink.confirmed",
	"downlink.confirmed_retry",
	"downlink.confirmed_retry.attempt",
	"downlink.confirmed_retry.max_attempts",
	"downlink.correlation_ids",
	"downlink.decoded_payload",
	"downlink.decoded_payload_warnings",
	"downlink.f_cnt",
	"downlink.f_port",
	"downlink.frm_payload",
	"downlink.priority",
	"downlink.session_key_id",
}

var DecodeDownlinkResponseFieldPathsTopLevel = []string{
	"downlink",
}
var AsConfiguration_PubSubFieldPathsNested = []string{
	"providers",
	"providers.mqtt",
	"providers.nats",
}

var AsConfiguration_PubSubFieldPathsTopLevel = []string{
	"providers",
}
var AsConfiguration_WebhooksFieldPathsNested = []string{
	"unhealthy_attempts_threshold",
	"unhealthy_retry_interval",
}

var AsConfiguration_WebhooksFieldPathsTopLevel = []string{
	"unhealthy_attempts_threshold",
	"unhealthy_retry_interval",
}
var AsConfiguration_PubSub_ProvidersFieldPathsNested = []string{
	"mqtt",
	"nats",
}

var AsConfiguration_PubSub_ProvidersFieldPathsTopLevel = []string{
	"mqtt",
	"nats",
}
