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
	"os"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	ttntypes "go.thethings.network/lorawan-stack/pkg/types"
)

var (
	selectEndDeviceListFlags = &pflag.FlagSet{}
	selectEndDeviceFlags     = &pflag.FlagSet{}
	setEndDeviceFlags        = &pflag.FlagSet{}
)

func endDeviceIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("application-id", "", "")
	flagSet.String("device-id", "", "")
	flagSet.String("join-eui", "", "(hex)")
	flagSet.String("dev-eui", "", "(hex)")
	return flagSet
}

var (
	errNoEndDeviceID            = errors.DefineInvalidArgument("no_end_device_id", "no end device ID set")
	errNoEndDeviceEUI           = errors.DefineInvalidArgument("no_end_device_eui", "no end device EUIs set")
	errInconsistentEndDeviceEUI = errors.DefineInvalidArgument("inconsistent_end_device_eui", "given end device EUIs do not match registered EUIs")
	errEndDeviceEUIUpdate       = errors.DefineInvalidArgument("end_device_eui_update", "end device EUIs can not be updated")
)

func getEndDeviceID(flagSet *pflag.FlagSet, args []string, requireID bool) (*ttnpb.EndDeviceIdentifiers, error) {
	applicationID, _ := flagSet.GetString("application-id")
	deviceID, _ := flagSet.GetString("device-id")
	switch len(args) {
	case 0:
	case 1:
		logger.Warn("Only single ID found in arguments, not considering arguments")
	case 2:
		applicationID = args[0]
		deviceID = args[1]
	default:
		logger.Warn("multiple IDs found in arguments, considering the first")
		applicationID = args[0]
		deviceID = args[1]
	}
	if applicationID == "" && requireID {
		return nil, errNoApplicationID
	}
	if deviceID == "" && requireID {
		return nil, errNoEndDeviceID
	}
	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: applicationID},
		DeviceID:               deviceID,
	}
	if joinEUIHex, _ := flagSet.GetString("join-eui"); joinEUIHex != "" {
		var joinEUI ttntypes.EUI64
		if err := joinEUI.UnmarshalText([]byte(joinEUIHex)); err != nil {
			return nil, err
		}
		ids.JoinEUI = &joinEUI
	}
	if devEUIHex, _ := flagSet.GetString("dev-eui"); devEUIHex != "" {
		var devEUI ttntypes.EUI64
		if err := devEUI.UnmarshalText([]byte(devEUIHex)); err != nil {
			return nil, err
		}
		ids.DevEUI = &devEUI
	}
	return ids, nil
}

func generateKey(length int) []byte {
	key := make([]byte, length)
	random.Read(key)
	return key
}

var (
	endDevicesCommand = &cobra.Command{
		Use:     "end-devices",
		Aliases: []string{"end-device", "devices", "device", "dev", "ed", "d"},
		Short:   "End Device commands",
	}
	endDevicesListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List end devices",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectEndDeviceListFlags)
			if len(paths) == 0 {
				logger.Warnf("No fields selected, selecting %v", defaultGetPaths)
				paths = append(paths, defaultGetPaths...)
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewEndDeviceRegistryClient(is).List(ctx, &ttnpb.ListEndDevicesRequest{
				ApplicationIdentifiers: *appID,
				FieldMask:              types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res.EndDevices)
		},
	}
	endDevicesGetCommand = &cobra.Command{
		Use:     "get",
		Aliases: []string{"info"},
		Short:   "Get an end device",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectEndDeviceFlags)
			if len(paths) == 0 {
				logger.Warnf("No fields selected, selecting %v", defaultGetPaths)
				paths = append(paths, defaultGetPaths...)
			}

			isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceGetPaths(paths...)

			if len(nsPaths) > 0 {
				isPaths = append(isPaths, "network_server_address")
			}
			if len(asPaths) > 0 {
				isPaths = append(isPaths, "application_server_address")
			}
			if len(jsPaths) > 0 {
				isPaths = append(isPaths, "join_server_address")
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: *devID,
				FieldMask:            types.FieldMask{Paths: isPaths},
			})
			if err != nil {
				return err
			}

			compareServerAddresses(device, config)

			res, err := getEndDevice(device.EndDeviceIdentifiers, nsPaths, asPaths, jsPaths, true)
			if err != nil {
				return err
			}

			device.SetFields(res, append(append(nsPaths, asPaths...), jsPaths...)...)

			return io.Write(os.Stdout, config.OutputFormat, device)
		},
	}
	endDevicesCreateCommand = &cobra.Command{
		Use:     "create",
		Aliases: []string{"add", "register"},
		Short:   "Create an end device",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, false)
			if err != nil {
				return err
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setEndDeviceFlags, attributesFlags())

			var device ttnpb.EndDevice
			if inputDecoder != nil {
				jsonPaths, err := inputDecoder.Decode(&device)
				if err != nil {
					return err
				}
				paths = append(paths, jsonPaths...)
			}

			setDefaults, _ := cmd.Flags().GetBool("defaults")
			if setDefaults {
				device.NetworkServerAddress = config.NetworkServerAddress
				device.ApplicationServerAddress = config.ApplicationServerAddress
				device.JoinServerAddress = config.JoinServerAddress
				device.LoRaWANVersion = ttnpb.MAC_V1_1
				device.LoRaWANPHYVersion = ttnpb.PHY_V1_1_REV_B
				device.Uses32BitFCnt = true
				device.MACSettings = &ttnpb.MACSettings{
					UseADR:    true,
					ADRMargin: 15,
				}
				paths = append(paths,
					"network_server_address", "application_server_address", "join_server_address",
					"lorawan_version", "lorawan_phy_version",
					"uses_32_bit_fcnt", "mac_settings",
				)
			}
			if otaa, _ := cmd.Flags().GetBool("otaa"); otaa {
				// TODO: Set JoinEUI and DevEUI (https://github.com/TheThingsIndustries/lorawan-stack/issues/1392).
				device.RootKeys = &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{Key: generateKey(16)},
					NwkKey: &ttnpb.KeyEnvelope{Key: generateKey(16)},
				}
				paths = append(paths,
					"root_keys.app_key",
					"root_keys.nwk_key",
				)
			}
			if abp, _ := cmd.Flags().GetBool("abp"); abp {
				device.Session = &ttnpb.Session{
					// TODO: Generate DevAddr (https://github.com/TheThingsIndustries/lorawan-stack/issues/1392).
					SessionKeys: ttnpb.SessionKeys{
						FNwkSIntKey: &ttnpb.KeyEnvelope{Key: generateKey(16)},
						SNwkSIntKey: &ttnpb.KeyEnvelope{Key: generateKey(16)},
						NwkSEncKey:  &ttnpb.KeyEnvelope{Key: generateKey(16)},
						AppSKey:     &ttnpb.KeyEnvelope{Key: generateKey(16)},
					},
				}
				device.DevAddr = &device.Session.DevAddr
				// TODO: Set device.NetID (https://github.com/TheThingsIndustries/lorawan-stack/issues/1392).
				paths = append(paths,
					"session.keys.f_nwk_s_int_key",
					"session.keys.s_nwk_s_int_key",
					"session.keys.nwk_s_enc_key",
					"session.keys.app_s_key",
				)
			}

			if err = util.SetFields(&device, setEndDeviceFlags); err != nil {
				return err
			}
			device.Attributes = mergeAttributes(device.Attributes, cmd.Flags())
			if devID != nil {
				if devID.DeviceID != "" {
					device.DeviceID = devID.DeviceID
				}
				if devID.ApplicationID != "" {
					device.ApplicationID = devID.ApplicationID
				}
				if devID.JoinEUI != nil {
					device.JoinEUI = devID.JoinEUI
				}
				if devID.DevEUI != nil {
					device.DevEUI = devID.DevEUI
				}
			}

			if device.ApplicationID == "" {
				return errNoApplicationID
			}
			if device.DeviceID == "" {
				return errNoEndDeviceID
			}

			isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceSetPaths(paths...)

			// Require EUIs for devices that need to be added to the Join Server.
			if len(jsPaths) > 0 && (device.JoinEUI == nil || device.DevEUI == nil) {
				return errNoEndDeviceEUI
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			isRes, err := ttnpb.NewEndDeviceRegistryClient(is).Create(ctx, &ttnpb.CreateEndDeviceRequest{
				EndDevice: device,
			})
			if err != nil {
				return err
			}

			device.SetFields(isRes, append(isPaths, "created_at", "updated_at")...)

			res, err := setEndDevice(&device, nil, nsPaths, asPaths, jsPaths)
			if err != nil {
				logger.WithError(err).Error("Could not create end device, rolling back...")
				return deleteEndDevice(&device.EndDeviceIdentifiers)
			}

			device.SetFields(res, append(append(nsPaths, asPaths...), jsPaths...)...)

			return io.Write(os.Stdout, config.OutputFormat, &device)
		},
	}
	endDevicesUpdateCommand = &cobra.Command{
		Use:     "update",
		Aliases: []string{"set"},
		Short:   "Update an end device",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setEndDeviceFlags, attributesFlags())
			if len(paths) == 0 {
				logger.Warn("No fields selected, won't update anything")
				return nil
			}
			var device ttnpb.EndDevice
			if err = util.SetFields(&device, setEndDeviceFlags); err != nil {
				return err
			}
			device.Attributes = mergeAttributes(device.Attributes, cmd.Flags())
			device.EndDeviceIdentifiers = *devID

			isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceSetPaths(paths...)

			if len(nsPaths) > 0 {
				isPaths = append(isPaths, "network_server_address")
			}
			if len(asPaths) > 0 {
				isPaths = append(isPaths, "application_server_address")
			}
			if len(jsPaths) > 0 {
				isPaths = append(isPaths, "join_server_address")
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			existingDevice, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: *devID,
				FieldMask:            types.FieldMask{Paths: isPaths},
			})
			if err != nil {
				return err
			}

			// EUIs can not be updated, so we only accept EUI flags if they are equal to the existing ones.
			if device.JoinEUI != nil {
				if existingDevice.JoinEUI != nil && *device.JoinEUI != *existingDevice.JoinEUI {
					return errEndDeviceEUIUpdate
				}
			} else {
				device.JoinEUI = existingDevice.JoinEUI
			}
			if device.DevEUI != nil {
				if existingDevice.DevEUI != nil && *device.DevEUI != *existingDevice.DevEUI {
					return errEndDeviceEUIUpdate
				}
			} else {
				device.DevEUI = existingDevice.DevEUI
			}

			// Require EUIs for devices that need to be updated in the Join Server.
			if len(jsPaths) > 0 && (device.JoinEUI == nil || device.DevEUI == nil) {
				return errNoEndDeviceEUI
			}

			compareServerAddresses(existingDevice, config)

			res, err := setEndDevice(&device, isPaths, nsPaths, asPaths, jsPaths)
			if err != nil {
				return err
			}

			res.SetFields(&device, "ids")
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	endDevicesDeleteCommand = &cobra.Command{
		Use:   "delete",
		Short: "Delete an end device",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerAddress)
			if err != nil {
				return err
			}
			existingDevice, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: *devID,
				FieldMask: types.FieldMask{Paths: []string{
					"network_server_address",
					"application_server_address",
					"join_server_address",
				}},
			})
			if err != nil {
				return err
			}

			// EUIs must match registered EUIs if set.
			if devID.JoinEUI != nil {
				if existingDevice.JoinEUI != nil && *devID.JoinEUI != *existingDevice.JoinEUI {
					return errInconsistentEndDeviceEUI
				}
			} else {
				devID.JoinEUI = existingDevice.JoinEUI
			}
			if devID.DevEUI != nil {
				if existingDevice.DevEUI != nil && *devID.DevEUI != *existingDevice.DevEUI {
					return errInconsistentEndDeviceEUI
				}
			} else {
				devID.DevEUI = existingDevice.DevEUI
			}

			compareServerAddresses(existingDevice, config)

			return deleteEndDevice(devID)
		},
	}
)

func init() {
	util.FieldMaskFlags(&ttnpb.EndDevice{}).VisitAll(func(flag *pflag.Flag) {
		path := strings.Split(flag.Name, ".")
		if getEndDevicePathFromIS(path...) {
			selectEndDeviceListFlags.AddFlag(flag)
			selectEndDeviceFlags.AddFlag(flag)
		}
		if getEndDevicePathFromNS(path...) ||
			getEndDevicePathFromAS(path...) ||
			getEndDevicePathFromJS(path...) {
			selectEndDeviceFlags.AddFlag(flag)
		}
	})

	util.FieldFlags(&ttnpb.EndDevice{}).VisitAll(func(flag *pflag.Flag) {
		path := strings.Split(flag.Name, ".")
		if setEndDevicePathToIS(path...) ||
			setEndDevicePathToNS(path...) ||
			setEndDevicePathToAS(path...) ||
			setEndDevicePathToJS(path...) {
			setEndDeviceFlags.AddFlag(flag)
		}
	})

	endDevicesListCommand.Flags().AddFlagSet(applicationIDFlags())
	endDevicesListCommand.Flags().AddFlagSet(selectEndDeviceListFlags)
	endDevicesCommand.AddCommand(endDevicesListCommand)
	endDevicesGetCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesGetCommand.Flags().AddFlagSet(selectEndDeviceFlags)
	endDevicesCommand.AddCommand(endDevicesGetCommand)
	endDevicesCreateCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesCreateCommand.Flags().AddFlagSet(setEndDeviceFlags)
	endDevicesCreateCommand.Flags().AddFlagSet(attributesFlags())
	endDevicesCreateCommand.Flags().Bool("defaults", true, "configure end device with defaults")
	endDevicesCreateCommand.Flags().Bool("otaa", true, "configure end device as OTAA")
	endDevicesCreateCommand.Flags().Bool("abp", false, "configure end device as ABP")
	endDevicesCommand.AddCommand(endDevicesCreateCommand)
	endDevicesUpdateCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesUpdateCommand.Flags().AddFlagSet(setEndDeviceFlags)
	endDevicesUpdateCommand.Flags().AddFlagSet(attributesFlags())
	endDevicesCommand.AddCommand(endDevicesUpdateCommand)
	endDevicesDeleteCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesCommand.AddCommand(endDevicesDeleteCommand)

	endDevicesCommand.AddCommand(applicationsDownlinkCommand)

	Root.AddCommand(endDevicesCommand)
}

func compareServerAddresses(device *ttnpb.EndDevice, config *Config) {
	if device.NetworkServerAddress != "" && device.NetworkServerAddress != config.NetworkServerAddress {
		logger.WithFields(log.Fields(
			"configured", config.NetworkServerAddress,
			"registered", device.NetworkServerAddress,
		)).Warn("Registered Network Server address does not match CLI configuration")
	}
	if device.ApplicationServerAddress != "" && device.ApplicationServerAddress != config.ApplicationServerAddress {
		logger.WithFields(log.Fields(
			"configured", config.ApplicationServerAddress,
			"registered", device.ApplicationServerAddress,
		)).Warn("Registered Application Server address does not match CLI configuration")
	}
	if device.JoinServerAddress != "" && device.JoinServerAddress != config.JoinServerAddress {
		logger.WithFields(log.Fields(
			"configured", config.JoinServerAddress,
			"registered", device.JoinServerAddress,
		)).Warn("Registered Join Server address does not match CLI configuration")
	}
}
