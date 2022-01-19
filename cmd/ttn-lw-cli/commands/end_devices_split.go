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
	"context"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	getEndDeviceFromIS = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.EndDeviceRegistry/Get"].Allowed
	getEndDeviceFromNS = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.NsEndDeviceRegistry/Get"].Allowed
	getEndDeviceFromAS = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.AsEndDeviceRegistry/Get"].Allowed
	getEndDeviceFromJS = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.JsEndDeviceRegistry/Get"].Allowed
	setEndDeviceToIS   = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.EndDeviceRegistry/Update"].Allowed
	setEndDeviceToNS   = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.NsEndDeviceRegistry/Set"].Allowed
	setEndDeviceToAS   = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.AsEndDeviceRegistry/Set"].Allowed
	setEndDeviceToJS   = ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.JsEndDeviceRegistry/Set"].Allowed
)

func nonImplicitPaths(paths ...string) []string {
	nonImplicitPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "ids" || strings.HasPrefix(path, "ids.") {
			continue
		}
		if path == "created_at" || path == "updated_at" {
			continue
		}
		nonImplicitPaths = append(nonImplicitPaths, path)
	}
	return nonImplicitPaths
}

func splitEndDeviceGetPaths(paths ...string) (is, ns, as, js []string) {
	nonImplicitPaths := nonImplicitPaths(paths...)
	is = ttnpb.AllowedFields(nonImplicitPaths, getEndDeviceFromIS)
	ns = ttnpb.AllowedFields(nonImplicitPaths, getEndDeviceFromNS)
	as = ttnpb.AllowedFields(nonImplicitPaths, getEndDeviceFromAS)
	js = ttnpb.AllowedFields(nonImplicitPaths, getEndDeviceFromJS)
	return
}

func splitEndDeviceSetPaths(supportsJoin bool, paths ...string) (is, ns, as, js []string) {
	nonImplicitPaths := nonImplicitPaths(paths...)
	is = ttnpb.AllowedFields(nonImplicitPaths, setEndDeviceToIS)
	ns = ttnpb.AllowedFields(nonImplicitPaths, setEndDeviceToNS)
	as = ttnpb.AllowedFields(nonImplicitPaths, setEndDeviceToAS)
	if supportsJoin {
		js = ttnpb.AllowedFields(nonImplicitPaths, setEndDeviceToJS)
	}
	return
}

func getEndDevice(ids *ttnpb.EndDeviceIdentifiers, nsPaths, asPaths, jsPaths []string, continueOnError bool) (*ttnpb.EndDevice, error) {
	var res ttnpb.EndDevice
	if len(jsPaths) > 0 {
		if !config.JoinServerEnabled {
			logger.WithField("paths", jsPaths).Warn("Join Server disabled but fields specified to get")
		} else {
			js, err := api.Dial(ctx, config.JoinServerGRPCAddress)
			if err != nil {
				if !continueOnError {
					return nil, err
				}
				logger.WithError(err).Error("Could not connect to Join Server")
			} else {
				logger.WithField("paths", jsPaths).Debug("Get end device from Join Server")
				jsRes, err := ttnpb.NewJsEndDeviceRegistryClient(js).Get(ctx, &ttnpb.GetEndDeviceRequest{
					EndDeviceIds: ids,
					FieldMask:    &pbtypes.FieldMask{Paths: jsPaths},
				})
				if err != nil {
					if !continueOnError {
						return nil, err
					}
					logger.WithError(err).Error("Could not get end device from Join Server")
				} else {
					res.SetFields(jsRes, ttnpb.AllowedBottomLevelFields(jsPaths, getEndDeviceFromJS)...)
					if res.CreatedAt == nil || (jsRes.CreatedAt != nil && ttnpb.StdTime(jsRes.CreatedAt).Before(*ttnpb.StdTime(res.CreatedAt))) {
						res.CreatedAt = jsRes.CreatedAt
					}
					if res.UpdatedAt == nil || (jsRes.UpdatedAt != nil && ttnpb.StdTime(jsRes.UpdatedAt).After(*ttnpb.StdTime(res.UpdatedAt))) {
						res.UpdatedAt = jsRes.UpdatedAt
					}
				}
			}
		}
	}

	if len(asPaths) > 0 {
		if !config.ApplicationServerEnabled {
			logger.WithField("paths", asPaths).Warn("Application Server disabled but fields specified to get")
		} else {
			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				if !continueOnError {
					return nil, err
				}
				logger.WithError(err).Error("Could not connect to Application Server")
			} else {
				logger.WithField("paths", asPaths).Debug("Get end device from Application Server")
				asRes, err := ttnpb.NewAsEndDeviceRegistryClient(as).Get(ctx, &ttnpb.GetEndDeviceRequest{
					EndDeviceIds: ids,
					FieldMask:    &pbtypes.FieldMask{Paths: asPaths},
				})
				if err != nil {
					if !continueOnError {
						return nil, err
					}
					logger.WithError(err).Error("Could not get end device from Application Server")
				} else {
					res.SetFields(asRes, ttnpb.AllowedBottomLevelFields(asPaths, getEndDeviceFromAS)...)
					if res.CreatedAt == nil || (asRes.CreatedAt != nil && ttnpb.StdTime(asRes.CreatedAt).Before(*ttnpb.StdTime(res.CreatedAt))) {
						res.CreatedAt = asRes.CreatedAt
					}
					if res.UpdatedAt == nil || (asRes.UpdatedAt != nil && ttnpb.StdTime(asRes.UpdatedAt).After(*ttnpb.StdTime(res.UpdatedAt))) {
						res.UpdatedAt = asRes.UpdatedAt
					}
				}
			}
		}
	}

	if len(nsPaths) > 0 {
		if !config.NetworkServerEnabled {
			logger.WithField("paths", nsPaths).Warn("Network Server disabled but fields specified to get")
		} else {
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				if !continueOnError {
					return nil, err
				}
				logger.WithError(err).Error("Could not connect to Network Server")
			} else {
				logger.WithField("paths", nsPaths).Debug("Get end device from Network Server")
				nsRes, err := ttnpb.NewNsEndDeviceRegistryClient(ns).Get(ctx, &ttnpb.GetEndDeviceRequest{
					EndDeviceIds: ids,
					FieldMask:    &pbtypes.FieldMask{Paths: nsPaths},
				})
				if err != nil {
					if !continueOnError {
						return nil, err
					}
					logger.WithError(err).Error("Could not get end device from Network Server")
				} else {
					res.SetFields(nsRes, "ids.dev_addr")
					res.SetFields(nsRes, ttnpb.AllowedBottomLevelFields(nsPaths, getEndDeviceFromNS)...)
					if res.CreatedAt == nil || (nsRes.CreatedAt != nil && ttnpb.StdTime(nsRes.CreatedAt).Before(*ttnpb.StdTime(res.CreatedAt))) {
						res.CreatedAt = nsRes.CreatedAt
					}
					if res.UpdatedAt == nil || (nsRes.UpdatedAt != nil && ttnpb.StdTime(nsRes.UpdatedAt).After(*ttnpb.StdTime(res.UpdatedAt))) {
						res.UpdatedAt = nsRes.UpdatedAt
					}
				}
			}
		}
	}

	return &res, nil
}

func setEndDevice(device *ttnpb.EndDevice, isPaths, nsPaths, asPaths, jsPaths, unsetPaths []string, isCreate, touch bool) (*ttnpb.EndDevice, error) {
	var res ttnpb.EndDevice
	res.SetFields(device, "ids", "created_at", "updated_at")

	if len(isPaths) > 0 && !isCreate {
		is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
		if err != nil {
			return nil, err
		}
		var isDevice ttnpb.EndDevice
		logger.WithField("paths", isPaths).Debug("Set end device on Identity Server")
		isDevice.SetFields(device, append(ttnpb.ExcludeFields(isPaths, unsetPaths...), "ids")...)
		isRes, err := ttnpb.NewEndDeviceRegistryClient(is).Update(ctx, &ttnpb.UpdateEndDeviceRequest{
			EndDevice: isDevice,
			FieldMask: &pbtypes.FieldMask{Paths: isPaths},
		})
		if err != nil {
			return nil, err
		}
		res.SetFields(isRes, isPaths...)
		if res.CreatedAt == nil || (isRes.CreatedAt != nil && ttnpb.StdTime(isRes.CreatedAt).Before(*ttnpb.StdTime(res.CreatedAt))) {
			res.CreatedAt = isRes.CreatedAt
		}
		if res.UpdatedAt == nil || ttnpb.StdTime(isRes.UpdatedAt).After(*ttnpb.StdTime(res.UpdatedAt)) {
			res.UpdatedAt = isRes.UpdatedAt
		}
	}

	if len(jsPaths) > 0 && !config.JoinServerEnabled {
		logger.WithField("paths", jsPaths).Warn("Join Server disabled but fields specified to set")
	} else if (len(jsPaths) > 0 || touch && device.SupportsJoin) && config.JoinServerEnabled {
		js, err := api.Dial(ctx, config.JoinServerGRPCAddress)
		if err != nil {
			return nil, err
		}
		var jsDevice ttnpb.EndDevice
		logger.WithField("paths", jsPaths).Debug("Set end device on Join Server")
		jsDevice.SetFields(device, append(ttnpb.ExcludeFields(jsPaths, unsetPaths...), "ids")...)
		jsRes, err := ttnpb.NewJsEndDeviceRegistryClient(js).Set(ctx, &ttnpb.SetEndDeviceRequest{
			EndDevice: jsDevice,
			FieldMask: &pbtypes.FieldMask{Paths: jsPaths},
		})
		if err != nil {
			return nil, err
		}
		res.SetFields(jsRes, jsPaths...)
		if res.CreatedAt == nil || (jsRes.CreatedAt != nil && ttnpb.StdTime(jsRes.CreatedAt).Before(*ttnpb.StdTime(res.CreatedAt))) {
			res.CreatedAt = jsRes.CreatedAt
		}
		if res.UpdatedAt == nil || (jsRes.UpdatedAt != nil && ttnpb.StdTime(jsRes.UpdatedAt).After(*ttnpb.StdTime(res.UpdatedAt))) {
			res.UpdatedAt = jsRes.UpdatedAt
		}
	}

	if len(nsPaths) > 0 && !config.NetworkServerEnabled {
		logger.WithField("paths", nsPaths).Warn("Network Server disabled but fields specified to set")
	} else if (len(nsPaths) > 0 || isCreate || touch) && config.NetworkServerEnabled {
		ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
		if err != nil {
			return nil, err
		}
		var nsDevice ttnpb.EndDevice
		logger.WithField("paths", nsPaths).Debug("Set end device on Network Server")
		nsDevice.SetFields(device, append(ttnpb.ExcludeFields(nsPaths, unsetPaths...), "ids")...)
		nsRes, err := ttnpb.NewNsEndDeviceRegistryClient(ns).Set(ctx, &ttnpb.SetEndDeviceRequest{
			EndDevice: nsDevice,
			FieldMask: &pbtypes.FieldMask{Paths: nsPaths},
		})
		if err != nil {
			return nil, err
		}
		res.SetFields(nsRes, nsPaths...)
		if res.CreatedAt == nil || (nsRes.CreatedAt != nil && ttnpb.StdTime(nsRes.CreatedAt).Before(*ttnpb.StdTime(res.CreatedAt))) {
			res.CreatedAt = nsRes.CreatedAt
		}
		if res.UpdatedAt == nil || (nsRes.UpdatedAt != nil && ttnpb.StdTime(nsRes.UpdatedAt).After(*ttnpb.StdTime(res.UpdatedAt))) {
			res.UpdatedAt = nsRes.UpdatedAt
		}
	}

	if len(asPaths) > 0 && !config.ApplicationServerEnabled {
		logger.WithField("paths", asPaths).Warn("Application Server disabled but fields specified to set")
	} else if (len(asPaths) > 0 || isCreate || touch) && config.ApplicationServerEnabled {
		as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
		if err != nil {
			return nil, err
		}
		var asDevice ttnpb.EndDevice
		logger.WithField("paths", asPaths).Debug("Set end device on Application Server")
		asDevice.SetFields(device, append(ttnpb.ExcludeFields(asPaths, unsetPaths...), "ids")...)
		asRes, err := ttnpb.NewAsEndDeviceRegistryClient(as).Set(ctx, &ttnpb.SetEndDeviceRequest{
			EndDevice: asDevice,
			FieldMask: &pbtypes.FieldMask{Paths: asPaths},
		})
		if err != nil {
			return nil, err
		}
		res.SetFields(asRes, asPaths...)
		if res.CreatedAt == nil || (asRes.CreatedAt != nil && ttnpb.StdTime(asRes.CreatedAt).Before(*ttnpb.StdTime(res.CreatedAt))) {
			res.CreatedAt = asRes.CreatedAt
		}
		if res.UpdatedAt == nil || (asRes.UpdatedAt != nil && ttnpb.StdTime(asRes.UpdatedAt).After(*ttnpb.StdTime(res.UpdatedAt))) {
			res.UpdatedAt = asRes.UpdatedAt
		}
	}

	return &res, ctx.Err()
}

func deleteEndDevice(ctx context.Context, devID *ttnpb.EndDeviceIdentifiers) error {
	if config.ApplicationServerEnabled {
		as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
		if err != nil {
			return err
		}
		_, err = ttnpb.NewAsEndDeviceRegistryClient(as).Delete(ctx, devID)
		if errors.IsNotFound(err) {
			logger.WithError(err).Error("Could not delete end device from Application Server")
		} else if err != nil {
			return err
		}
	}

	if config.NetworkServerEnabled {
		ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
		if err != nil {
			return err
		}
		_, err = ttnpb.NewNsEndDeviceRegistryClient(ns).Delete(ctx, devID)
		if errors.IsNotFound(err) {
			logger.WithError(err).Error("Could not delete end device from Network Server")
		} else if err != nil {
			return err
		}
	}

	if config.JoinServerEnabled {
		if devID.JoinEui != nil && devID.DevEui != nil {
			js, err := api.Dial(ctx, config.JoinServerGRPCAddress)
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
	}

	is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
	if err != nil {
		return err
	}
	_, err = ttnpb.NewEndDeviceRegistryClient(is).Delete(ctx, devID)
	if err != nil {
		return err
	}

	return nil
}

func hasUpdateDeviceLocationFlags(flags *pflag.FlagSet) bool {
	return flags.Changed("location.latitude") ||
		flags.Changed("location.longitude") ||
		flags.Changed("location.altitude") ||
		flags.Changed("location.accuracy")
}

func updateDeviceLocation(device *ttnpb.EndDevice, flags *pflag.FlagSet) {
	if device.Locations == nil {
		device.Locations = make(map[string]*ttnpb.Location)
	}
	loc, ok := device.Locations["user"]
	if !ok {
		loc = &ttnpb.Location{}
	}
	loc.Source = ttnpb.LocationSource_SOURCE_REGISTRY
	if flags.Changed("location.longitude") {
		longitude, _ := flags.GetFloat64("location.longitude")
		loc.Longitude = longitude
	}
	if flags.Changed("location.latitude") {
		latitude, _ := flags.GetFloat64("location.latitude")
		loc.Latitude = latitude
	}
	if flags.Changed("location.altitude") {
		altitude, _ := flags.GetInt32("location.altitude")
		loc.Altitude = altitude
	}
	if flags.Changed("location.accuracy") {
		accuracy, _ := flags.GetInt32("location.accuracy")
		loc.Accuracy = accuracy
	}
	device.Locations["user"] = loc
}
