// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package commands

import (
	"strings"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// getEndDevicePathFromIS returns whether the field with given path should be
// retrieved from the Identity Server.
func getEndDevicePathFromIS(pathParts ...string) bool {
	switch pathParts[0] {
	case
		"application_server_address",
		"attributes",
		"description",
		"join_server_address",
		"locations",
		"name",
		"network_server_address",
		"service_profile_id",
		"version_ids":
		return true
	}
	return false
}

// setEndDevicePathToIS returns whether the field with given path should be
// set in the Identity Server.
func setEndDevicePathToIS(pathParts ...string) bool {
	switch pathParts[0] {
	case
		"application_server_address",
		"attributes",
		"description",
		"join_server_address",
		"locations",
		"name",
		"network_server_address",
		"service_profile_id",
		"version_ids":
		return true
	}
	return false
}

// getEndDevicePathFromNS returns whether the field with given path should be
// retrieved from the Network Server.
func getEndDevicePathFromNS(pathParts ...string) bool {
	switch pathParts[0] {
	case
		"battery_percentage",
		"default_class",
		"default_mac_parameters",
		"downlink_margin",
		"frequency_plan_id",
		"last_dev_status_received_at",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"mac_state",
		"max_frequency",
		"min_frequency",
		"power_state",
		"recent_downlinks",
		"recent_uplinks",
		"resets_f_cnt",
		"resets_join_nonces",
		"supports_class_b",
		"supports_class_c",
		"supports_join",
		"uses_32_bit_f_cnt":
		return true
	case "session":
		if len(pathParts) == 1 {
			return true
		}
		switch pathParts[1] {
		case
			"dev_addr",
			"last_f_cnt_up",
			"last_n_f_cnt_down",
			"last_conf_f_cnt_down",
			"started_at":
			return true
		case "keys":
			if len(pathParts) == 2 {
				return true
			}
			switch pathParts[2] {
			case
				"f_nwk_s_int_key",
				"s_nwk_s_int_key",
				"nwk_s_enc_key":
				return true
			}
		}
	}
	return false
}

// setEndDevicePathToNS returns whether the field with given path should be
// set in the Network Server.
func setEndDevicePathToNS(pathParts ...string) bool {
	switch pathParts[0] {
	case
		"default_class",
		"default_mac_parameters",
		"frequency_plan_id",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"max_frequency",
		"min_frequency",
		"resets_f_cnt",
		"resets_join_nonces",
		"supports_class_b",
		"supports_class_c",
		"supports_join",
		"uses_32_bit_f_cnt":
		return true
	case "session":
		if len(pathParts) == 1 {
			return false // Only set specific fields.
		}
		switch pathParts[1] {
		case
			"dev_addr",
			"last_conf_f_cnt_down",
			"last_f_cnt_up",
			"last_n_f_cnt_down",
			"started_at":
			return true
		case "keys":
			if len(pathParts) == 2 {
				return false // Only set specific fields.
			}
			switch pathParts[2] {
			case
				"f_nwk_s_int_key",
				"nwk_s_enc_key",
				"s_nwk_s_int_key":
				return true
			}
		}
	}
	return false
}

// getEndDevicePathFromAS returns whether the field with given path should be
// retrieved from the Application Server.
func getEndDevicePathFromAS(pathParts ...string) bool {
	switch pathParts[0] {
	case
		"formatters",
		"queued_application_downlinks":
		return true
	case "session":
		if len(pathParts) == 1 {
			return true
		}
		switch pathParts[1] {
		case
			"last_a_f_cnt_down":
			return true
		case "keys":
			if len(pathParts) == 2 {
				return true
			}
			switch pathParts[2] {
			case
				"app_s_key":
				return true
			}
		}
	}
	return false
}

// setEndDevicePathToAS returns whether the field with given path should be
// set in the Application Server.
func setEndDevicePathToAS(pathParts ...string) bool {
	switch pathParts[0] {
	case "formatters":
		return true
	case "session":
		if len(pathParts) == 1 {
			return false // Only set specific fields.
		}
		switch pathParts[1] {
		case
			"dev_addr",
			"last_a_f_cnt_down":
			return true
		case "keys":
			if len(pathParts) == 2 {
				return false // Only set specific fields.
			}
			switch pathParts[2] {
			case
				"app_s_key":
				return true
			}
		}
	}
	return false
}

// getEndDevicePathFromJS returns whether the field with given path should be
// retrieved from the Join Server.
func getEndDevicePathFromJS(pathParts ...string) bool {
	switch pathParts[0] {
	case
		"last_dev_nonce",
		"last_join_nonce",
		"last_rj_count_0",
		"last_rj_count_1",
		"net_id",
		"provisioner_id",
		"provisioning_data",
		"resets_join_nonces",
		"root_keys",
		"used_dev_nonces":
		return true
	}
	return false
}

// setEndDevicePathToJS returns whether the field with given path should be
// set in the Join Server.
func setEndDevicePathToJS(pathParts ...string) bool {
	switch pathParts[0] {
	case
		"application_server_address",
		"last_dev_nonce",
		"last_join_nonce",
		"last_rj_count_0",
		"last_rj_count_1",
		"net_id",
		"network_server_address",
		"provisioner_id",
		"provisioning_data",
		"resets_join_nonces",
		"root_keys",
		"used_dev_nonces":
		return true
	}
	return false
}

func splitEndDeviceGetPaths(paths ...string) (is, ns, as, js []string) {
	var unassigned []string
	for _, path := range paths {
		switch path {
		case "session.keys.session_key_id": // Ignore.
			continue
		}
		parts := strings.Split(path, ".")
		switch parts[0] {
		case "ids", "created_at", "updated_at":
			continue
		case "pending_session": // Ignore.
			continue
		}
		var assigned bool
		if getEndDevicePathFromIS(parts...) {
			is = append(is, path)
			assigned = true
		}
		if getEndDevicePathFromNS(parts...) {
			ns = append(ns, path)
			assigned = true
		}
		if getEndDevicePathFromAS(parts...) {
			as = append(as, path)
			assigned = true
		}
		if getEndDevicePathFromJS(parts...) {
			js = append(js, path)
			assigned = true
		}
		if !assigned {
			switch path {
			// Here we can ignore intentionally unassigned paths/
			default:
				unassigned = append(unassigned, path)
			}
		}
	}
	if len(unassigned) > 0 {
		logger.WithField("fields", unassigned).Warn("Some fields could not be assigned to a server")
	}
	return
}

func splitEndDeviceSetPaths(supportsJoin bool, paths ...string) (is, ns, as, js []string) {
	var unassigned []string
	for _, path := range paths {
		switch path {
		case "session.keys.session_key_id": // Ignore.
			continue
		}
		parts := strings.Split(path, ".")
		switch parts[0] {
		case "ids", "created_at", "updated_at":
			continue
		case "pending_session": // Ignore.
			continue
		}
		var assigned bool
		if setEndDevicePathToIS(parts...) {
			is = append(is, path)
			assigned = true
		}
		if setEndDevicePathToNS(parts...) {
			ns = append(ns, path)
			assigned = true
		}
		if setEndDevicePathToAS(parts...) {
			as = append(as, path)
			assigned = true
		}
		if supportsJoin && setEndDevicePathToJS(parts...) {
			js = append(js, path)
			assigned = true
		}
		if !assigned {
			switch path {
			// Here we can ignore intentionally unassigned paths.
			default:
				unassigned = append(unassigned, path)
			}
		}
	}
	if len(unassigned) > 0 {
		logger.WithField("fields", unassigned).Warn("Some fields could not be assigned to a server")
	}
	return
}

func getEndDevice(ids ttnpb.EndDeviceIdentifiers, nsPaths, asPaths, jsPaths []string, continueOnError bool) (*ttnpb.EndDevice, error) {
	var res ttnpb.EndDevice

	if len(jsPaths) > 0 {
		js, err := api.Dial(ctx, config.JoinServerAddress)
		if err != nil {
			if !continueOnError {
				return nil, err
			}
			logger.WithError(err).Error("Could not connect to Join Server")
		} else {
			jsRes, err := ttnpb.NewJsEndDeviceRegistryClient(js).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ids,
				FieldMask:            types.FieldMask{Paths: jsPaths},
			})
			if err != nil {
				if !continueOnError {
					return nil, err
				}
				logger.WithError(err).Error("Could not get end device from Join Server")
			} else {
				res.SetFields(jsRes, jsPaths...)
			}
		}
	}

	if len(asPaths) > 0 {
		as, err := api.Dial(ctx, config.ApplicationServerAddress)
		if err != nil {
			if !continueOnError {
				return nil, err
			}
			logger.WithError(err).Error("Could not connect to Application Server")
		} else {
			asRes, err := ttnpb.NewAsEndDeviceRegistryClient(as).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ids,
				FieldMask:            types.FieldMask{Paths: asPaths},
			})
			if err != nil {
				if !continueOnError {
					return nil, err
				}
				logger.WithError(err).Error("Could not get end device from Application Server")
			} else {
				res.SetFields(asRes, asPaths...)
			}
		}
	}

	if len(nsPaths) > 0 {
		ns, err := api.Dial(ctx, config.NetworkServerAddress)
		if err != nil {
			if !continueOnError {
				return nil, err
			}
			logger.WithError(err).Error("Could not connect to Network Server")
		} else {
			nsRes, err := ttnpb.NewNsEndDeviceRegistryClient(ns).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ids,
				FieldMask:            types.FieldMask{Paths: nsPaths},
			})
			if err != nil {
				if !continueOnError {
					return nil, err
				}
				logger.WithError(err).Error("Could not get end device from Network Server")
			} else {
				res.SetFields(nsRes, nsPaths...)
			}
		}
	}

	return &res, nil
}

func setEndDevice(device *ttnpb.EndDevice, isPaths, nsPaths, asPaths, jsPaths []string, isCreate bool) (*ttnpb.EndDevice, error) {
	var res ttnpb.EndDevice

	if len(isPaths) > 0 || isCreate {
		is, err := api.Dial(ctx, config.IdentityServerAddress)
		if err != nil {
			return nil, err
		}
		var isDevice ttnpb.EndDevice
		isDevice.SetFields(device, append(isPaths, "ids")...)
		isRes, err := ttnpb.NewEndDeviceRegistryClient(is).Update(ctx, &ttnpb.UpdateEndDeviceRequest{
			EndDevice: isDevice,
			FieldMask: types.FieldMask{Paths: isPaths},
		})
		if err != nil {
			return nil, err
		}
		res.SetFields(isRes, isPaths...)
	}

	if len(jsPaths) > 0 {
		js, err := api.Dial(ctx, config.JoinServerAddress)
		if err != nil {
			return nil, err
		}
		var jsDevice ttnpb.EndDevice
		jsDevice.SetFields(device, append(jsPaths, "ids")...)
		jsRes, err := ttnpb.NewJsEndDeviceRegistryClient(js).Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device:    jsDevice,
			FieldMask: types.FieldMask{Paths: jsPaths},
		})
		if err != nil {
			return nil, err
		}
		res.SetFields(jsRes, jsPaths...)
	}

	if len(nsPaths) > 0 || isCreate {
		ns, err := api.Dial(ctx, config.NetworkServerAddress)
		if err != nil {
			return nil, err
		}
		var nsDevice ttnpb.EndDevice
		nsDevice.SetFields(device, append(nsPaths, "ids")...)
		nsRes, err := ttnpb.NewNsEndDeviceRegistryClient(ns).Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device:    nsDevice,
			FieldMask: types.FieldMask{Paths: nsPaths},
		})
		if err != nil {
			return nil, err
		}
		res.SetFields(nsRes, nsPaths...)
	}

	if len(asPaths) > 0 || isCreate {
		as, err := api.Dial(ctx, config.ApplicationServerAddress)
		if err != nil {
			return nil, err
		}
		var asDevice ttnpb.EndDevice
		asDevice.SetFields(device, append(asPaths, "ids")...)
		asRes, err := ttnpb.NewAsEndDeviceRegistryClient(as).Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device:    asDevice,
			FieldMask: types.FieldMask{Paths: asPaths},
		})
		if err != nil {
			return nil, err
		}
		res.SetFields(asRes, asPaths...)
	}

	return &res, nil
}

func deleteEndDevice(devID *ttnpb.EndDeviceIdentifiers) error {
	as, err := api.Dial(ctx, config.ApplicationServerAddress)
	if err != nil {
		return err
	}
	_, err = ttnpb.NewAsEndDeviceRegistryClient(as).Delete(ctx, devID)
	if errors.IsNotFound(err) {
		logger.WithError(err).Error("Could not delete end device from Application Server")
	} else if err != nil {
		return err
	}

	ns, err := api.Dial(ctx, config.NetworkServerAddress)
	if err != nil {
		return err
	}
	_, err = ttnpb.NewNsEndDeviceRegistryClient(ns).Delete(ctx, devID)
	if errors.IsNotFound(err) {
		logger.WithError(err).Error("Could not delete end device from Network Server")
	} else if err != nil {
		return err
	}

	if devID.JoinEUI != nil && devID.DevEUI != nil {
		js, err := api.Dial(ctx, config.JoinServerAddress)
		if err != nil {
			return err
		}
		_, err = ttnpb.NewJsEndDeviceRegistryClient(js).Delete(ctx, devID)
		if errors.IsNotFound(err) {
			logger.WithError(err).Error("Could not delete end device from Join Server")
		} else if err != nil {
			return err
		}
	}

	is, err := api.Dial(ctx, config.IdentityServerAddress)
	if err != nil {
		return err
	}
	_, err = ttnpb.NewEndDeviceRegistryClient(is).Delete(ctx, devID)
	if err != nil {
		return err
	}

	return nil
}
