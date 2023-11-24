// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networkserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/proto"
)

var (
	emptyRelayUplinkForwardingRule = &ttnpb.RelayUplinkForwardingRule{}

	errRelayAlreadyExists = errors.DefineAlreadyExists("relay_already_exists", "relay already exists")
	errRelayNotFound      = errors.DefineNotFound("relay_not_found", "relay not found")
	errRelayNotServed     = errors.DefineUnavailable("relay_not_served", "relay not served")
	errRelayNotServing    = errors.DefineUnavailable("relay_not_serving", "relay not serving")

	errRelayUplinkForwardingRuleAlreadyExists = errors.DefineAlreadyExists(
		"relay_uplink_forwarding_rule_already_exists", "relay uplink forwarding rule already exists",
	)
	errRelayUplinkForwardingRuleNotFound = errors.DefineNotFound(
		"relay_uplink_forwarding_rule_not_found", "relay uplink forwarding rule not found",
	)
)

type nsRelayConfigurationService struct {
	ttnpb.UnimplementedNsRelayConfigurationServiceServer

	devices        DeviceRegistry
	frequencyPlans func(context.Context) (*frequencyplans.Store, error)
}

// CreateRelay implements ttnpb.NsRelayConfigurationServiceServer.
func (s *nsRelayConfigurationService) CreateRelay(
	ctx context.Context, req *ttnpb.CreateRelayRequest,
) (*ttnpb.CreateRelayResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	fps, err := s.frequencyPlans(ctx)
	if err != nil {
		return nil, err
	}
	if _, ctx, err := s.devices.SetByID(
		ctx,
		req.EndDeviceIds.ApplicationIds,
		req.EndDeviceIds.DeviceId,
		[]string{
			"frequency_plan_id",
			"lorawan_phy_version",
			"mac_settings.desired_relay",
			"mac_state.desired_parameters.relay",
			"pending_mac_state.desired_parameters.relay",
		},
		func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.New()
			}
			if dev.MacSettings.GetDesiredRelay() != nil {
				return nil, nil, errRelayAlreadyExists.New()
			}
			phy, err := DeviceBand(dev, fps)
			if err != nil {
				return nil, nil, err
			}
			parameters := relayParametersFromConfiguration(req.Configuration)
			if err := validateRelayConfiguration(parameters, phy, "configuration"); err != nil {
				return nil, nil, err
			}
			dev.MacSettings = &ttnpb.MACSettings{DesiredRelay: parameters}
			paths := []string{"mac_settings.desired_relay"}
			for path, desiredParameters := range map[string]*ttnpb.MACParameters{
				"mac_state.desired_parameters.relay":         dev.MacState.GetDesiredParameters(),
				"pending_mac_state.desired_parameters.relay": dev.PendingMacState.GetDesiredParameters(),
			} {
				if desiredParameters == nil {
					continue
				}
				desiredParameters.Relay = parameters
				paths = ttnpb.AddFields(paths, path)
			}
			return dev, paths, nil
		},
	); err != nil {
		logRegistryRPCError(ctx, err, "Failed to create relay")
		return nil, err
	}
	return &ttnpb.CreateRelayResponse{
		Configuration: req.Configuration,
	}, nil
}

// GetRelay implements ttnpb.NsRelayConfigurationServiceServer.
func (s *nsRelayConfigurationService) GetRelay(
	ctx context.Context, req *ttnpb.GetRelayRequest,
) (*ttnpb.GetRelayResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
	); err != nil {
		return nil, err
	}
	dev, ctx, err := s.devices.GetByID(
		ctx,
		req.EndDeviceIds.ApplicationIds,
		req.EndDeviceIds.DeviceId,
		ttnpb.FieldsWithPrefix("mac_settings.desired_relay", req.FieldMask.GetPaths()...),
	)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to get relay")
		return nil, err
	}
	if dev.MacSettings.GetDesiredRelay() == nil {
		return nil, errRelayNotFound.New()
	}
	return &ttnpb.GetRelayResponse{
		Configuration: relayConfigurationFromParameters(dev.MacSettings.DesiredRelay),
	}, nil
}

// UpdateRelay implements ttnpb.NsRelayConfigurationServiceServer.
func (s *nsRelayConfigurationService) UpdateRelay(
	ctx context.Context, req *ttnpb.UpdateRelayRequest,
) (*ttnpb.UpdateRelayResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	fps, err := s.frequencyPlans(ctx)
	if err != nil {
		return nil, err
	}
	if _, ctx, err := s.devices.SetByID(
		ctx,
		req.EndDeviceIds.ApplicationIds,
		req.EndDeviceIds.DeviceId,
		[]string{
			"frequency_plan_id",
			"lorawan_phy_version",
			"mac_settings.desired_relay",
			"mac_state.desired_parameters.relay",
			"pending_mac_state.desired_parameters.relay",
		},
		func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.New()
			}
			if dev.MacSettings.GetDesiredRelay() == nil {
				return nil, nil, errRelayNotFound.New()
			}
			phy, err := DeviceBand(dev, fps)
			if err != nil {
				return nil, nil, err
			}
			parameters := dev.MacSettings.DesiredRelay
			if err := parameters.SetFields(
				relayParametersFromConfiguration(req.Configuration), req.FieldMask.GetPaths()...,
			); err != nil {
				return nil, nil, err
			}
			if err := validateRelayConfiguration(parameters, phy, "configuration"); err != nil {
				return nil, nil, err
			}
			dev.MacSettings.DesiredRelay = parameters
			paths := []string{"mac_settings.desired_relay"}
			for path, desiredParameters := range map[string]*ttnpb.MACParameters{
				"mac_state.desired_parameters.relay":         dev.MacState.GetDesiredParameters(),
				"pending_mac_state.desired_parameters.relay": dev.PendingMacState.GetDesiredParameters(),
			} {
				if desiredParameters == nil {
					continue
				}
				desiredParameters.Relay = parameters
				paths = ttnpb.AddFields(paths, path)
			}
			return dev, paths, nil
		},
	); err != nil {
		logRegistryRPCError(ctx, err, "Failed to update relay")
		return nil, err
	}
	return &ttnpb.UpdateRelayResponse{
		Configuration: req.Configuration,
	}, nil
}

// DeleteRelay implements ttnpb.NsRelayConfigurationServiceServer.
func (s *nsRelayConfigurationService) DeleteRelay(
	ctx context.Context, req *ttnpb.DeleteRelayRequest,
) (*ttnpb.DeleteRelayResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	if _, ctx, err := s.devices.SetByID(
		ctx,
		req.EndDeviceIds.ApplicationIds,
		req.EndDeviceIds.DeviceId,
		[]string{
			"mac_settings.desired_relay",
			"mac_state.desired_parameters.relay",
			"pending_mac_state.desired_parameters.relay",
		},
		func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.New()
			}
			if dev.MacSettings.GetDesiredRelay() == nil {
				return nil, nil, errRelayNotFound.New()
			}
			dev.MacSettings.DesiredRelay = nil
			paths := []string{"mac_settings.desired_relay"}
			for path, desiredParameters := range map[string]*ttnpb.MACParameters{
				"mac_state.desired_parameters.relay":         dev.MacState.GetDesiredParameters(),
				"pending_mac_state.desired_parameters.relay": dev.PendingMacState.GetDesiredParameters(),
			} {
				if desiredParameters == nil {
					continue
				}
				desiredParameters.Relay = nil
				paths = ttnpb.AddFields(paths, path)
			}
			return dev, paths, nil
		},
	); err != nil {
		logRegistryRPCError(ctx, err, "Failed to delete relay")
		return nil, err
	}
	return &ttnpb.DeleteRelayResponse{}, nil
}

// CreateRelayUplinkForwardingRule implements ttnpb.NsRelayConfigurationServiceServer.
func (s *nsRelayConfigurationService) CreateRelayUplinkForwardingRule( // nolint:gocyclo
	ctx context.Context, req *ttnpb.CreateRelayUplinkForwardingRuleRequest,
) (*ttnpb.CreateRelayUplinkForwardingRuleResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	_, ctx, err := s.devices.SetByID(
		ctx,
		req.EndDeviceIds.ApplicationIds,
		req.UplinkForwardingRule.DeviceId,
		[]string{
			"mac_settings.desired_relay",
			"mac_state.desired_parameters.relay",
			"pending_mac_state.desired_parameters.relay",
			"pending_session.keys.session_key_id",
			"session.keys.session_key_id",
		},
		func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.New()
			}
			if dev.MacSettings.GetDesiredRelay() == nil {
				return nil, nil, errRelayNotFound.New()
			}
			servedSettings := dev.MacSettings.DesiredRelay.GetServed()
			if servedSettings == nil {
				return nil, nil, errRelayNotServed.New()
			}
			servedSettings.ServingDeviceId = req.EndDeviceIds.DeviceId
			paths := []string{"mac_settings.desired_relay.mode.served.serving_device_id"}
			for path, served := range map[string]*ttnpb.ServedRelayParameters{
				"mac_state.desired_parameters.relay.mode.served.serving_device_id":         dev.MacState.GetDesiredParameters().GetRelay().GetServed(),        // nolint: lll
				"pending_mac_state.desired_parameters.relay.mode.served.serving_device_id": dev.PendingMacState.GetDesiredParameters().GetRelay().GetServed(), // nolint: lll
			} {
				if served == nil {
					continue
				}
				served.ServingDeviceId = req.EndDeviceIds.DeviceId
				paths = ttnpb.AddFields(paths, path)
			}
			var sessionKeyID []byte
			for _, session := range []*ttnpb.Session{
				dev.Session,
				dev.PendingSession,
			} {
				if session.GetKeys() == nil {
					continue
				}
				if len(session.Keys.SessionKeyId) == 0 {
					continue
				}
				sessionKeyID = session.Keys.SessionKeyId
			}
			if _, _, err := s.devices.SetByID(
				ctx,
				req.EndDeviceIds.ApplicationIds,
				req.EndDeviceIds.DeviceId,
				[]string{
					"mac_settings.desired_relay",
					"mac_state.desired_parameters.relay",
					"pending_mac_state.desired_parameters.relay",
				},
				func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
					if dev == nil {
						return nil, nil, errDeviceNotFound.New()
					}
					if dev.MacSettings.GetDesiredRelay() == nil {
						return nil, nil, errRelayNotFound.New()
					}
					servingSettings := dev.MacSettings.DesiredRelay.GetServing()
					if servingSettings == nil {
						return nil, nil, errRelayNotServing.New()
					}
					if uint(req.Index) < uint(len(servingSettings.UplinkForwardingRules)) &&
						!proto.Equal(servingSettings.UplinkForwardingRules[req.Index], emptyRelayUplinkForwardingRule) {
						return nil, nil, errRelayUplinkForwardingRuleAlreadyExists.New()
					}
					if n := len(servingSettings.UplinkForwardingRules); uint(req.Index) >= uint(n) {
						servingSettings.UplinkForwardingRules = append(
							servingSettings.UplinkForwardingRules,
							make(
								[]*ttnpb.RelayUplinkForwardingRule,
								1+int(req.Index-uint32(n)),
							)...,
						)
						for i := n; i < len(servingSettings.UplinkForwardingRules); i++ {
							servingSettings.UplinkForwardingRules[i] = &ttnpb.RelayUplinkForwardingRule{}
						}
					}
					rule := relayUplinkForwardingRuleFromConfiguration(req.UplinkForwardingRule)
					rule.SessionKeyId = sessionKeyID
					servingSettings.UplinkForwardingRules[req.Index] = rule
					paths := []string{"mac_settings.desired_relay.mode.serving.uplink_forwarding_rules"}
					for path, serving := range map[string]*ttnpb.ServingRelayParameters{
						"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules":         dev.MacState.GetDesiredParameters().GetRelay().GetServing(),        // nolint: lll
						"pending_mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules": dev.PendingMacState.GetDesiredParameters().GetRelay().GetServing(), // nolint: lll
					} {
						if serving == nil {
							continue
						}
						serving.UplinkForwardingRules = servingSettings.UplinkForwardingRules
						paths = ttnpb.AddFields(paths, path)
					}
					return dev, paths, nil
				},
			); err != nil {
				return nil, nil, err
			}
			return dev, paths, nil
		},
	)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to create relay uplink forwarding rule")
		return nil, err
	}
	return &ttnpb.CreateRelayUplinkForwardingRuleResponse{
		UplinkForwardingRule: req.UplinkForwardingRule,
	}, nil
}

// GetRelayUplinkForwardingRule implements ttnpb.NsRelayConfigurationServiceServer.
func (s *nsRelayConfigurationService) GetRelayUplinkForwardingRule(
	ctx context.Context, req *ttnpb.GetRelayUplinkForwardingRuleRequest,
) (*ttnpb.GetRelayUplinkForwardingRuleResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
	); err != nil {
		return nil, err
	}
	dev, ctx, err := s.devices.GetByID(
		ctx,
		req.EndDeviceIds.ApplicationIds,
		req.EndDeviceIds.DeviceId,
		[]string{"mac_settings.desired_relay"},
	)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to get relay uplink forwarding rule")
		return nil, err
	}
	if dev.MacSettings.GetDesiredRelay() == nil {
		return nil, errRelayNotFound.New()
	}
	servingSettings := dev.MacSettings.DesiredRelay.GetServing()
	if servingSettings == nil {
		return nil, errRelayNotServing.New()
	}
	if uint(req.Index) >= uint(len(servingSettings.UplinkForwardingRules)) ||
		proto.Equal(servingSettings.UplinkForwardingRules[req.Index], emptyRelayUplinkForwardingRule) {
		return nil, errRelayUplinkForwardingRuleNotFound.New()
	}
	rule := &ttnpb.RelayUplinkForwardingRule{}
	if err := rule.SetFields(servingSettings.UplinkForwardingRules[req.Index], req.FieldMask.GetPaths()...); err != nil {
		return nil, err
	}
	return &ttnpb.GetRelayUplinkForwardingRuleResponse{
		UplinkForwardingRule: relayConfigurationUplinkForwardingRuleFromUplinkForwardingRule(rule),
	}, nil
}

// ListRelayUplinkForwardingRules implements ttnpb.NsRelayConfigurationServiceServer.
func (s *nsRelayConfigurationService) ListRelayUplinkForwardingRules(
	ctx context.Context, req *ttnpb.ListRelayUplinkForwardingRulesRequest,
) (*ttnpb.ListRelayUplinkForwardingRulesResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
	); err != nil {
		return nil, err
	}
	dev, ctx, err := s.devices.GetByID(
		ctx,
		req.EndDeviceIds.ApplicationIds,
		req.EndDeviceIds.DeviceId,
		[]string{"mac_settings.desired_relay"},
	)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to list relay uplink forwarding rules")
		return nil, err
	}
	if dev.MacSettings.GetDesiredRelay() == nil {
		return nil, errRelayNotFound.New()
	}
	servingSettings := dev.MacSettings.DesiredRelay.GetServing()
	if servingSettings == nil {
		return nil, errRelayNotServing.New()
	}
	rules := make([]*ttnpb.RelayConfigurationUplinkForwardingRule, 0, len(servingSettings.UplinkForwardingRules))
	for _, rule := range servingSettings.UplinkForwardingRules {
		rules = append(rules, relayConfigurationUplinkForwardingRuleFromUplinkForwardingRule(rule))
	}
	return &ttnpb.ListRelayUplinkForwardingRulesResponse{
		UplinkForwardingRules: rules,
	}, nil
}

// UpdateRelayUplinkForwardingRule implements ttnpb.NsRelayConfigurationServiceServer.
func (s *nsRelayConfigurationService) UpdateRelayUplinkForwardingRule( // nolint:gocyclo
	ctx context.Context, req *ttnpb.UpdateRelayUplinkForwardingRuleRequest,
) (*ttnpb.UpdateRelayUplinkForwardingRuleResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	updateServingDeviceID := ttnpb.HasAnyField(req.FieldMask.GetPaths(), "serving_device_id")
	var servedSessionKeyID []byte
	updateServingDevice := func(_ context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev == nil {
			return nil, nil, errDeviceNotFound.New()
		}
		if dev.MacSettings.GetDesiredRelay() == nil {
			return nil, nil, errRelayNotFound.New()
		}
		servingSettings := dev.MacSettings.DesiredRelay.GetServing()
		if servingSettings == nil {
			return nil, nil, errRelayNotServing.New()
		}
		if uint(req.Index) >= uint(len(servingSettings.UplinkForwardingRules)) ||
			proto.Equal(servingSettings.UplinkForwardingRules[req.Index], emptyRelayUplinkForwardingRule) {
			return nil, nil, errRelayUplinkForwardingRuleNotFound.New()
		}
		rule := servingSettings.UplinkForwardingRules[req.Index]
		patch := relayUplinkForwardingRuleFromConfiguration(req.UplinkForwardingRule)
		patchPaths := req.FieldMask.GetPaths()
		if updateServingDeviceID {
			patch.SessionKeyId = servedSessionKeyID
			patchPaths = ttnpb.AddFields(patchPaths, "session_key_id")
		}
		if err := rule.SetFields(patch, patchPaths...); err != nil {
			return nil, nil, err
		}
		servingSettings.UplinkForwardingRules[req.Index] = rule
		paths := []string{"mac_settings.desired_relay.mode.serving.uplink_forwarding_rules"}
		for path, serving := range map[string]*ttnpb.ServingRelayParameters{
			"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules":         dev.MacState.GetDesiredParameters().GetRelay().GetServing(),        // nolint: lll
			"pending_mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules": dev.PendingMacState.GetDesiredParameters().GetRelay().GetServing(), // nolint: lll
		} {
			if serving == nil {
				continue
			}
			serving.UplinkForwardingRules = servingSettings.UplinkForwardingRules
			paths = ttnpb.AddFields(paths, path)
		}
		return dev, paths, nil
	}
	ctx, err := ctx, error(nil)
	if updateServingDeviceID {
		_, ctx, err = s.devices.SetByID(
			ctx,
			req.EndDeviceIds.ApplicationIds,
			req.UplinkForwardingRule.DeviceId,
			[]string{
				"mac_settings.desired_relay",
				"mac_state.desired_parameters.relay",
				"pending_mac_state.desired_parameters.relay",
				"pending_session.keys.session_key_id",
				"session.keys.session_key_id",
			},
			func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				if dev == nil {
					return nil, nil, errDeviceNotFound.New()
				}
				if dev.MacSettings.GetDesiredRelay() == nil {
					return nil, nil, errRelayNotFound.New()
				}
				servedSettings := dev.MacSettings.DesiredRelay.GetServed()
				if servedSettings == nil {
					return nil, nil, errRelayNotServed.New()
				}
				servedSettings.ServingDeviceId = req.EndDeviceIds.DeviceId
				paths := []string{"mac_settings.desired_relay.mode.served.serving_device_id"}
				for path, served := range map[string]*ttnpb.ServedRelayParameters{
					"mac_state.desired_parameters.relay.mode.served.serving_device_id":         dev.MacState.GetDesiredParameters().GetRelay().GetServed(),        // nolint: lll
					"pending_mac_state.desired_parameters.relay.mode.served.serving_device_id": dev.PendingMacState.GetDesiredParameters().GetRelay().GetServed(), // nolint: lll
				} {
					if served == nil {
						continue
					}
					served.ServingDeviceId = req.EndDeviceIds.DeviceId
					paths = ttnpb.AddFields(paths, path)
				}
				for _, session := range []*ttnpb.Session{
					dev.Session,
					dev.PendingSession,
				} {
					if session.GetKeys() == nil {
						continue
					}
					if len(session.Keys.SessionKeyId) == 0 {
						continue
					}
					servedSessionKeyID = session.Keys.SessionKeyId
				}
				if _, _, err := s.devices.SetByID(
					ctx,
					req.EndDeviceIds.ApplicationIds,
					req.EndDeviceIds.DeviceId,
					[]string{
						"mac_settings.desired_relay",
						"mac_state.desired_parameters.relay",
						"pending_mac_state.desired_parameters.relay",
					},
					updateServingDevice,
				); err != nil {
					return nil, nil, err
				}
				return dev, paths, nil
			},
		)
	} else {
		_, ctx, err = s.devices.SetByID(
			ctx,
			req.EndDeviceIds.ApplicationIds,
			req.EndDeviceIds.DeviceId,
			[]string{
				"mac_settings.desired_relay",
				"mac_state.desired_parameters.relay",
				"pending_mac_state.desired_parameters.relay",
			},
			updateServingDevice,
		)
	}
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to update relay uplink forwarding rule")
		return nil, err
	}
	return &ttnpb.UpdateRelayUplinkForwardingRuleResponse{
		UplinkForwardingRule: req.UplinkForwardingRule,
	}, nil
}

// DeleteRelayUplinkForwardingRule implements ttnpb.NsRelayConfigurationServiceServer.
func (s *nsRelayConfigurationService) DeleteRelayUplinkForwardingRule(
	ctx context.Context, req *ttnpb.DeleteRelayUplinkForwardingRuleRequest,
) (*ttnpb.DeleteRelayUplinkForwardingRuleResponse, error) {
	if err := rights.RequireApplication(
		ctx, req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	_, ctx, err := s.devices.SetByID(
		ctx,
		req.EndDeviceIds.ApplicationIds,
		req.EndDeviceIds.DeviceId,
		[]string{
			"mac_settings.desired_relay",
			"mac_state.desired_parameters.relay",
			"pending_mac_state.desired_parameters.relay",
		},
		func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.New()
			}
			if dev.MacSettings.GetDesiredRelay() == nil {
				return nil, nil, errRelayNotFound.New()
			}
			servingSettings := dev.MacSettings.DesiredRelay.GetServing()
			if servingSettings == nil {
				return nil, nil, errRelayNotServing.New()
			}
			if uint(req.Index) >= uint(len(servingSettings.UplinkForwardingRules)) ||
				proto.Equal(servingSettings.UplinkForwardingRules[req.Index], emptyRelayUplinkForwardingRule) {
				return nil, nil, errRelayUplinkForwardingRuleNotFound.New()
			}
			servingSettings.UplinkForwardingRules[req.Index] = emptyRelayUplinkForwardingRule
			paths := []string{"mac_settings.desired_relay.mode.serving.uplink_forwarding_rules"}
			for path, serving := range map[string]*ttnpb.ServingRelayParameters{
				"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules":         dev.MacState.GetDesiredParameters().GetRelay().GetServing(),        // nolint: lll
				"pending_mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules": dev.PendingMacState.GetDesiredParameters().GetRelay().GetServing(), // nolint: lll
			} {
				if serving == nil {
					continue
				}
				serving.UplinkForwardingRules = servingSettings.UplinkForwardingRules
				paths = ttnpb.AddFields(paths, path)
			}
			return dev, paths, nil
		},
	)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to delete relay uplink forwarding rule")
		return nil, err
	}
	return &ttnpb.DeleteRelayUplinkForwardingRuleResponse{}, nil
}
