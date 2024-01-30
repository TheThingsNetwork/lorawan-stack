// Code generated by protoc-gen-fieldmask. DO NOT EDIT.

package ttnpb

var CreateRelayRequestFieldPathsNested = []string{
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"settings",
	"settings.mode",
	"settings.mode.served",
	"settings.mode.served.backoff",
	"settings.mode.served.mode",
	"settings.mode.served.mode.always",
	"settings.mode.served.mode.dynamic",
	"settings.mode.served.mode.dynamic.smart_enable_level",
	"settings.mode.served.mode.end_device_controlled",
	"settings.mode.served.second_channel",
	"settings.mode.served.second_channel.ack_offset",
	"settings.mode.served.second_channel.data_rate_index",
	"settings.mode.served.second_channel.frequency",
	"settings.mode.served.serving_device_id",
	"settings.mode.serving",
	"settings.mode.serving.cad_periodicity",
	"settings.mode.serving.default_channel_index",
	"settings.mode.serving.limits",
	"settings.mode.serving.limits.join_requests",
	"settings.mode.serving.limits.join_requests.bucket_size",
	"settings.mode.serving.limits.join_requests.reload_rate",
	"settings.mode.serving.limits.notifications",
	"settings.mode.serving.limits.notifications.bucket_size",
	"settings.mode.serving.limits.notifications.reload_rate",
	"settings.mode.serving.limits.overall",
	"settings.mode.serving.limits.overall.bucket_size",
	"settings.mode.serving.limits.overall.reload_rate",
	"settings.mode.serving.limits.reset_behavior",
	"settings.mode.serving.limits.uplink_messages",
	"settings.mode.serving.limits.uplink_messages.bucket_size",
	"settings.mode.serving.limits.uplink_messages.reload_rate",
	"settings.mode.serving.second_channel",
	"settings.mode.serving.second_channel.ack_offset",
	"settings.mode.serving.second_channel.data_rate_index",
	"settings.mode.serving.second_channel.frequency",
	"settings.mode.serving.uplink_forwarding_rules",
}

var CreateRelayRequestFieldPathsTopLevel = []string{
	"end_device_ids",
	"settings",
}
var CreateRelayResponseFieldPathsNested = []string{
	"settings",
	"settings.mode",
	"settings.mode.served",
	"settings.mode.served.backoff",
	"settings.mode.served.mode",
	"settings.mode.served.mode.always",
	"settings.mode.served.mode.dynamic",
	"settings.mode.served.mode.dynamic.smart_enable_level",
	"settings.mode.served.mode.end_device_controlled",
	"settings.mode.served.second_channel",
	"settings.mode.served.second_channel.ack_offset",
	"settings.mode.served.second_channel.data_rate_index",
	"settings.mode.served.second_channel.frequency",
	"settings.mode.served.serving_device_id",
	"settings.mode.serving",
	"settings.mode.serving.cad_periodicity",
	"settings.mode.serving.default_channel_index",
	"settings.mode.serving.limits",
	"settings.mode.serving.limits.join_requests",
	"settings.mode.serving.limits.join_requests.bucket_size",
	"settings.mode.serving.limits.join_requests.reload_rate",
	"settings.mode.serving.limits.notifications",
	"settings.mode.serving.limits.notifications.bucket_size",
	"settings.mode.serving.limits.notifications.reload_rate",
	"settings.mode.serving.limits.overall",
	"settings.mode.serving.limits.overall.bucket_size",
	"settings.mode.serving.limits.overall.reload_rate",
	"settings.mode.serving.limits.reset_behavior",
	"settings.mode.serving.limits.uplink_messages",
	"settings.mode.serving.limits.uplink_messages.bucket_size",
	"settings.mode.serving.limits.uplink_messages.reload_rate",
	"settings.mode.serving.second_channel",
	"settings.mode.serving.second_channel.ack_offset",
	"settings.mode.serving.second_channel.data_rate_index",
	"settings.mode.serving.second_channel.frequency",
	"settings.mode.serving.uplink_forwarding_rules",
}

var CreateRelayResponseFieldPathsTopLevel = []string{
	"settings",
}
var GetRelayRequestFieldPathsNested = []string{
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"field_mask",
}

var GetRelayRequestFieldPathsTopLevel = []string{
	"end_device_ids",
	"field_mask",
}
var GetRelayResponseFieldPathsNested = []string{
	"settings",
	"settings.mode",
	"settings.mode.served",
	"settings.mode.served.backoff",
	"settings.mode.served.mode",
	"settings.mode.served.mode.always",
	"settings.mode.served.mode.dynamic",
	"settings.mode.served.mode.dynamic.smart_enable_level",
	"settings.mode.served.mode.end_device_controlled",
	"settings.mode.served.second_channel",
	"settings.mode.served.second_channel.ack_offset",
	"settings.mode.served.second_channel.data_rate_index",
	"settings.mode.served.second_channel.frequency",
	"settings.mode.served.serving_device_id",
	"settings.mode.serving",
	"settings.mode.serving.cad_periodicity",
	"settings.mode.serving.default_channel_index",
	"settings.mode.serving.limits",
	"settings.mode.serving.limits.join_requests",
	"settings.mode.serving.limits.join_requests.bucket_size",
	"settings.mode.serving.limits.join_requests.reload_rate",
	"settings.mode.serving.limits.notifications",
	"settings.mode.serving.limits.notifications.bucket_size",
	"settings.mode.serving.limits.notifications.reload_rate",
	"settings.mode.serving.limits.overall",
	"settings.mode.serving.limits.overall.bucket_size",
	"settings.mode.serving.limits.overall.reload_rate",
	"settings.mode.serving.limits.reset_behavior",
	"settings.mode.serving.limits.uplink_messages",
	"settings.mode.serving.limits.uplink_messages.bucket_size",
	"settings.mode.serving.limits.uplink_messages.reload_rate",
	"settings.mode.serving.second_channel",
	"settings.mode.serving.second_channel.ack_offset",
	"settings.mode.serving.second_channel.data_rate_index",
	"settings.mode.serving.second_channel.frequency",
	"settings.mode.serving.uplink_forwarding_rules",
}

var GetRelayResponseFieldPathsTopLevel = []string{
	"settings",
}
var UpdateRelayRequestFieldPathsNested = []string{
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"field_mask",
	"settings",
	"settings.mode",
	"settings.mode.served",
	"settings.mode.served.backoff",
	"settings.mode.served.mode",
	"settings.mode.served.mode.always",
	"settings.mode.served.mode.dynamic",
	"settings.mode.served.mode.dynamic.smart_enable_level",
	"settings.mode.served.mode.end_device_controlled",
	"settings.mode.served.second_channel",
	"settings.mode.served.second_channel.ack_offset",
	"settings.mode.served.second_channel.data_rate_index",
	"settings.mode.served.second_channel.frequency",
	"settings.mode.served.serving_device_id",
	"settings.mode.serving",
	"settings.mode.serving.cad_periodicity",
	"settings.mode.serving.default_channel_index",
	"settings.mode.serving.limits",
	"settings.mode.serving.limits.join_requests",
	"settings.mode.serving.limits.join_requests.bucket_size",
	"settings.mode.serving.limits.join_requests.reload_rate",
	"settings.mode.serving.limits.notifications",
	"settings.mode.serving.limits.notifications.bucket_size",
	"settings.mode.serving.limits.notifications.reload_rate",
	"settings.mode.serving.limits.overall",
	"settings.mode.serving.limits.overall.bucket_size",
	"settings.mode.serving.limits.overall.reload_rate",
	"settings.mode.serving.limits.reset_behavior",
	"settings.mode.serving.limits.uplink_messages",
	"settings.mode.serving.limits.uplink_messages.bucket_size",
	"settings.mode.serving.limits.uplink_messages.reload_rate",
	"settings.mode.serving.second_channel",
	"settings.mode.serving.second_channel.ack_offset",
	"settings.mode.serving.second_channel.data_rate_index",
	"settings.mode.serving.second_channel.frequency",
	"settings.mode.serving.uplink_forwarding_rules",
}

var UpdateRelayRequestFieldPathsTopLevel = []string{
	"end_device_ids",
	"field_mask",
	"settings",
}
var UpdateRelayResponseFieldPathsNested = []string{
	"settings",
	"settings.mode",
	"settings.mode.served",
	"settings.mode.served.backoff",
	"settings.mode.served.mode",
	"settings.mode.served.mode.always",
	"settings.mode.served.mode.dynamic",
	"settings.mode.served.mode.dynamic.smart_enable_level",
	"settings.mode.served.mode.end_device_controlled",
	"settings.mode.served.second_channel",
	"settings.mode.served.second_channel.ack_offset",
	"settings.mode.served.second_channel.data_rate_index",
	"settings.mode.served.second_channel.frequency",
	"settings.mode.served.serving_device_id",
	"settings.mode.serving",
	"settings.mode.serving.cad_periodicity",
	"settings.mode.serving.default_channel_index",
	"settings.mode.serving.limits",
	"settings.mode.serving.limits.join_requests",
	"settings.mode.serving.limits.join_requests.bucket_size",
	"settings.mode.serving.limits.join_requests.reload_rate",
	"settings.mode.serving.limits.notifications",
	"settings.mode.serving.limits.notifications.bucket_size",
	"settings.mode.serving.limits.notifications.reload_rate",
	"settings.mode.serving.limits.overall",
	"settings.mode.serving.limits.overall.bucket_size",
	"settings.mode.serving.limits.overall.reload_rate",
	"settings.mode.serving.limits.reset_behavior",
	"settings.mode.serving.limits.uplink_messages",
	"settings.mode.serving.limits.uplink_messages.bucket_size",
	"settings.mode.serving.limits.uplink_messages.reload_rate",
	"settings.mode.serving.second_channel",
	"settings.mode.serving.second_channel.ack_offset",
	"settings.mode.serving.second_channel.data_rate_index",
	"settings.mode.serving.second_channel.frequency",
	"settings.mode.serving.uplink_forwarding_rules",
}

var UpdateRelayResponseFieldPathsTopLevel = []string{
	"settings",
}
var DeleteRelayRequestFieldPathsNested = []string{
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
}

var DeleteRelayRequestFieldPathsTopLevel = []string{
	"end_device_ids",
}
var DeleteRelayResponseFieldPathsNested []string
var DeleteRelayResponseFieldPathsTopLevel []string
var CreateRelayUplinkForwardingRuleRequestFieldPathsNested = []string{
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"index",
	"rule",
	"rule.device_id",
	"rule.last_w_f_cnt",
	"rule.limits",
	"rule.limits.bucket_size",
	"rule.limits.reload_rate",
	"rule.session_key_id",
}

var CreateRelayUplinkForwardingRuleRequestFieldPathsTopLevel = []string{
	"end_device_ids",
	"index",
	"rule",
}
var CreateRelayUplinkForwardingRuleResponseFieldPathsNested = []string{
	"rule",
	"rule.device_id",
	"rule.last_w_f_cnt",
	"rule.limits",
	"rule.limits.bucket_size",
	"rule.limits.reload_rate",
	"rule.session_key_id",
}

var CreateRelayUplinkForwardingRuleResponseFieldPathsTopLevel = []string{
	"rule",
}
var GetRelayUplinkForwardingRuleRequestFieldPathsNested = []string{
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"field_mask",
	"index",
}

var GetRelayUplinkForwardingRuleRequestFieldPathsTopLevel = []string{
	"end_device_ids",
	"field_mask",
	"index",
}
var GetRelayUplinkForwardingRuleResponseFieldPathsNested = []string{
	"rule",
	"rule.device_id",
	"rule.last_w_f_cnt",
	"rule.limits",
	"rule.limits.bucket_size",
	"rule.limits.reload_rate",
	"rule.session_key_id",
}

var GetRelayUplinkForwardingRuleResponseFieldPathsTopLevel = []string{
	"rule",
}
var ListRelayUplinkForwardingRulesRequestFieldPathsNested = []string{
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"field_mask",
}

var ListRelayUplinkForwardingRulesRequestFieldPathsTopLevel = []string{
	"end_device_ids",
	"field_mask",
}
var ListRelayUplinkForwardingRulesResponseFieldPathsNested = []string{
	"rules",
}

var ListRelayUplinkForwardingRulesResponseFieldPathsTopLevel = []string{
	"rules",
}
var UpdateRelayUplinkForwardingRuleRequestFieldPathsNested = []string{
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"field_mask",
	"index",
	"rule",
	"rule.device_id",
	"rule.last_w_f_cnt",
	"rule.limits",
	"rule.limits.bucket_size",
	"rule.limits.reload_rate",
	"rule.session_key_id",
}

var UpdateRelayUplinkForwardingRuleRequestFieldPathsTopLevel = []string{
	"end_device_ids",
	"field_mask",
	"index",
	"rule",
}
var UpdateRelayUplinkForwardingRuleResponseFieldPathsNested = []string{
	"rule",
	"rule.device_id",
	"rule.last_w_f_cnt",
	"rule.limits",
	"rule.limits.bucket_size",
	"rule.limits.reload_rate",
	"rule.session_key_id",
}

var UpdateRelayUplinkForwardingRuleResponseFieldPathsTopLevel = []string{
	"rule",
}
var DeleteRelayUplinkForwardingRuleRequestFieldPathsNested = []string{
	"end_device_ids",
	"end_device_ids.application_ids",
	"end_device_ids.application_ids.application_id",
	"end_device_ids.dev_addr",
	"end_device_ids.dev_eui",
	"end_device_ids.device_id",
	"end_device_ids.join_eui",
	"index",
}

var DeleteRelayUplinkForwardingRuleRequestFieldPathsTopLevel = []string{
	"end_device_ids",
	"index",
}
var DeleteRelayUplinkForwardingRuleResponseFieldPathsNested []string
var DeleteRelayUplinkForwardingRuleResponseFieldPathsTopLevel []string