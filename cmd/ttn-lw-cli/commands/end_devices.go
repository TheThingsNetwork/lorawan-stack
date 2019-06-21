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
	stdio "io"
	"os"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var (
	selectEndDeviceListFlags = &pflag.FlagSet{}
	selectEndDeviceFlags     = &pflag.FlagSet{}
	setEndDeviceFlags        = &pflag.FlagSet{}
	endDeviceFlattenPaths    = []string{"provisioning_data"}
)

func endDeviceIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("application-id", "", "")
	flagSet.String("device-id", "", "")
	flagSet.String("join-eui", "", "(hex)")
	flagSet.String("dev-eui", "", "(hex)")
	addDeprecatedDeviceFlags(flagSet)
	return flagSet
}

func addDeprecatedDeviceFlags(flagSet *pflag.FlagSet) {
	util.DeprecateFlag(flagSet, "app-eui", "join-eui")
	util.DeprecateFlag(flagSet, "session.keys.nwk_s_key", "session.keys.f_nwk_s_int_key")
	util.DeprecateFlag(flagSet, "pending_session.keys.nwk_s_key", "pending_session.keys.f_nwk_s_int_key")
	util.DeprecateFlag(flagSet, "session.keys.nwk_s_key.key", "session.keys.f_nwk_s_int_key.key")
	util.DeprecateFlag(flagSet, "pending_session.keys.nwk_s_key.key", "pending_session.keys.f_nwk_s_int_key.key")
}

func forwardDeprecatedDeviceFlags(flagSet *pflag.FlagSet) {
	util.ForwardFlag(flagSet, "app-eui", "join-eui")
	util.ForwardFlag(flagSet, "session.keys.nwk_s_key", "session.keys.f_nwk_s_int_key")
	util.ForwardFlag(flagSet, "pending_session.keys.nwk_s_key", "pending_session.keys.f_nwk_s_int_key")
	util.ForwardFlag(flagSet, "session.keys.nwk_s_key.key", "session.keys.f_nwk_s_int_key.key")
	util.ForwardFlag(flagSet, "pending_session.keys.nwk_s_key.key", "pending_session.keys.f_nwk_s_int_key.key")
}

var (
	errEndDeviceEUIUpdate           = errors.DefineInvalidArgument("end_device_eui_update", "end device EUIs can not be updated")
	errEndDeviceKeysWithProvisioner = errors.DefineInvalidArgument("end_device_keys_provisioner", "end device ABP or OTAA keys cannot be set when there is a provisioner")
	errInconsistentEndDeviceEUI     = errors.DefineInvalidArgument("inconsistent_end_device_eui", "given end device EUIs do not match registered EUIs")
	errInvalidMACVerson             = errors.DefineInvalidArgument("mac_version", "LoRaWAN MAC version is invalid")
	errInvalidPHYVerson             = errors.DefineInvalidArgument("phy_version", "LoRaWAN PHY version is invalid")
	errNoEndDeviceEUI               = errors.DefineInvalidArgument("no_end_device_eui", "no end device EUIs set")
	errNoEndDeviceID                = errors.DefineInvalidArgument("no_end_device_id", "no end device ID set")
)

func getEndDeviceID(flagSet *pflag.FlagSet, args []string, requireID bool) (*ttnpb.EndDeviceIdentifiers, error) {
	forwardDeprecatedDeviceFlags(flagSet)
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
		var joinEUI types.EUI64
		if err := joinEUI.UnmarshalText([]byte(joinEUIHex)); err != nil {
			return nil, err
		}
		ids.JoinEUI = &joinEUI
	}
	if devEUIHex, _ := flagSet.GetString("dev-eui"); devEUIHex != "" {
		var devEUI types.EUI64
		if err := devEUI.UnmarshalText([]byte(devEUIHex)); err != nil {
			return nil, err
		}
		ids.DevEUI = &devEUI
	}
	return ids, nil
}

func generateBytes(length int) []byte {
	b := make([]byte, length)
	random.Read(b)
	return b
}

func generateKey() *types.AES128Key {
	var key types.AES128Key
	random.Read(key[:])
	return &key
}

func generateDevAddr(netID types.NetID) (types.DevAddr, error) {
	nwkAddr := make([]byte, types.NwkAddrLength(netID))
	random.Read(nwkAddr)
	nwkAddr[0] &= 0xff >> (8 - types.NwkAddrBits(netID)%8)
	devAddr, err := types.NewDevAddr(netID, nwkAddr)
	if err != nil {
		return types.DevAddr{}, err
	}
	return devAddr, nil
}

var (
	endDevicesCommand = &cobra.Command{
		Use:     "end-devices",
		Aliases: []string{"end-device", "devices", "device", "dev", "ed", "d"},
		Short:   "End Device commands",
	}
	endDevicesListFrequencyPlans = &cobra.Command{
		Use:               "list-frequency-plans",
		Short:             "List available frequency plans for end devices",
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseFrequency, _ := cmd.Flags().GetUint32("base-frequency")
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewConfigurationClient(ns).ListFrequencyPlans(ctx, &ttnpb.ListFrequencyPlansRequest{
				BaseFrequency: baseFrequency,
			})
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, res.FrequencyPlans)
		},
	}
	endDevicesListCommand = &cobra.Command{
		Use:     "list [application-id]",
		Aliases: []string{"ls"},
		Short:   "List end devices",
		RunE: func(cmd *cobra.Command, args []string) error {
			forwardDeprecatedDeviceFlags(cmd.Flags())

			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectEndDeviceListFlags)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewEndDeviceRegistryClient(is).List(ctx, &ttnpb.ListEndDevicesRequest{
				ApplicationIdentifiers: *appID,
				FieldMask:              pbtypes.FieldMask{Paths: paths},
				Limit:                  limit,
				Page:                   page,
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.EndDevices)
		},
	}
	endDevicesGetCommand = &cobra.Command{
		Use:     "get [application-id] [device-id]",
		Aliases: []string{"info"},
		Short:   "Get an end device",
		RunE: func(cmd *cobra.Command, args []string) error {
			forwardDeprecatedDeviceFlags(cmd.Flags())

			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectEndDeviceFlags)

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

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.WithField("paths", isPaths).Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: *devID,
				FieldMask:            pbtypes.FieldMask{Paths: isPaths},
			})
			if err != nil {
				return err
			}

			if len(jsPaths) > 0 && device.JoinServerAddress == "" {
				logger.WithField("paths", jsPaths).Debug("No registered Join Server address, deselecting Join Server paths")
				jsPaths = nil
			}

			nsMismatch, asMismatch, jsMismatch := compareServerAddressesEndDevice(device, config)
			if len(nsPaths) > 0 && nsMismatch {
				logger.WithField("paths", nsPaths).Warn("Deselecting Network Server paths")
				nsPaths = nil
			}
			if len(asPaths) > 0 && asMismatch {
				logger.WithField("paths", asPaths).Warn("Deselecting Application Server paths")
				asPaths = nil
			}
			if len(jsPaths) > 0 && jsMismatch {
				logger.WithField("paths", jsPaths).Warn("Deselecting Join Server paths")
				jsPaths = nil
			}

			res, err := getEndDevice(device.EndDeviceIdentifiers, nsPaths, asPaths, jsPaths, true)
			if err != nil {
				return err
			}

			device.SetFields(res, "ids.dev_addr")
			device.SetFields(res, append(append(nsPaths, asPaths...), jsPaths...)...)
			if device.CreatedAt.IsZero() || (!res.CreatedAt.IsZero() && res.CreatedAt.Before(res.CreatedAt)) {
				device.CreatedAt = res.CreatedAt
			}
			if res.UpdatedAt.After(device.UpdatedAt) {
				device.UpdatedAt = res.UpdatedAt
			}

			return io.Write(os.Stdout, config.OutputFormat, device)
		},
	}
	endDevicesCreateCommand = &cobra.Command{
		Use:     "create [application-id] [device-id]",
		Aliases: []string{"add", "register"},
		Short:   "Create an end device",
		RunE: asBulk(func(cmd *cobra.Command, args []string) (err error) {
			forwardDeprecatedDeviceFlags(cmd.Flags())

			devID, err := getEndDeviceID(cmd.Flags(), args, false)
			if err != nil {
				return err
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setEndDeviceFlags, attributesFlags())

			var device ttnpb.EndDevice
			if inputDecoder != nil {
				decodedPaths, err := inputDecoder.Decode(&device)
				if err != nil {
					return err
				}
				paths = append(paths, ttnpb.FlattenPaths(decodedPaths, endDeviceFlattenPaths)...)
			}

			var macVersion ttnpb.MACVersion
			s, err := setEndDeviceFlags.GetString("lorawan_version")
			if err != nil {
				return err
			}

			if err := macVersion.UnmarshalText([]byte(s)); err != nil {
				return err
			}
			if err := macVersion.Validate(); err != nil {
				return errInvalidMACVerson
			}

			setDefaults, _ := cmd.Flags().GetBool("defaults")
			if setDefaults {
				device.NetworkServerAddress = getHost(config.NetworkServerGRPCAddress)
				device.ApplicationServerAddress = getHost(config.ApplicationServerGRPCAddress)
				paths = append(paths,
					"application_server_address",
					"network_server_address",
				)
			}

			if abp, _ := cmd.Flags().GetBool("abp"); abp {
				device.SupportsJoin = false
				paths = append(paths, "supports_join")
				if withSession, _ := cmd.Flags().GetBool("with-session"); withSession {
					if device.ProvisionerID != "" {
						return errEndDeviceKeysWithProvisioner
					}
					// TODO: Generate DevAddr in cluster NetID (https://github.com/TheThingsNetwork/lorawan-stack/issues/47).
					devAddr, err := generateDevAddr(types.NetID{})
					if err != nil {
						return err
					}
					device.DevAddr = &devAddr
					device.Session = &ttnpb.Session{
						DevAddr: devAddr,
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: generateBytes(16),
							FNwkSIntKey:  &ttnpb.KeyEnvelope{Key: generateKey()},
							AppSKey:      &ttnpb.KeyEnvelope{Key: generateKey()},
						},
					}
					paths = append(paths,
						"session.keys.session_key_id",
						"session.keys.f_nwk_s_int_key.key",
						"session.keys.app_s_key.key",
						"session.dev_addr",
					)
					if macVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
						device.Session.SessionKeys.SNwkSIntKey = &ttnpb.KeyEnvelope{Key: generateKey()}
						device.Session.SessionKeys.NwkSEncKey = &ttnpb.KeyEnvelope{Key: generateKey()}
						paths = append(paths,
							"session.keys.s_nwk_s_int_key.key",
							"session.keys.nwk_s_enc_key.key",
						)
					}
				}
			} else {
				device.SupportsJoin = true
				paths = append(paths, "supports_join")
				if setDefaults {
					device.JoinServerAddress = getHost(config.JoinServerGRPCAddress)
					paths = append(paths,
						"join_server_address",
					)
				}
				if withKeys, _ := cmd.Flags().GetBool("with-root-keys"); withKeys {
					if device.ProvisionerID != "" {
						return errEndDeviceKeysWithProvisioner
					}
					// TODO: Set JoinEUI and DevEUI (https://github.com/TheThingsNetwork/lorawan-stack/issues/47).
					device.RootKeys = &ttnpb.RootKeys{
						RootKeyID: "ttn-lw-cli-generated",
						AppKey:    &ttnpb.KeyEnvelope{Key: generateKey()},
						NwkKey:    &ttnpb.KeyEnvelope{Key: generateKey()},
					}
					paths = append(paths,
						"root_keys.root_key_id",
						"root_keys.app_key.key",
						"root_keys.nwk_key.key",
					)
				}
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
				if device.SupportsJoin {
					if devID.JoinEUI != nil {
						device.JoinEUI = devID.JoinEUI
					}
					if devID.DevEUI != nil {
						device.DevEUI = devID.DevEUI
					}
				}
			}

			if device.ApplicationID == "" {
				return errNoApplicationID
			}
			if device.DeviceID == "" {
				return errNoEndDeviceID
			}

			isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceSetPaths(device.SupportsJoin, paths...)

			// Require EUIs for devices that need to be added to the Join Server.
			if len(jsPaths) > 0 && (device.JoinEUI == nil || device.DevEUI == nil) {
				return errNoEndDeviceEUI
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
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

			res, err := setEndDevice(&device, nil, nsPaths, asPaths, jsPaths, true)
			if err != nil {
				logger.WithError(err).Error("Could not create end device, rolling back...")
				if err := deleteEndDevice(context.Background(), &device.EndDeviceIdentifiers); err != nil {
					logger.WithError(err).Error("Could not roll back end device creation")
				}
				return err
			}

			device.SetFields(res, append(append(nsPaths, asPaths...), jsPaths...)...)
			if device.CreatedAt.IsZero() || (!res.CreatedAt.IsZero() && res.CreatedAt.Before(res.CreatedAt)) {
				device.CreatedAt = res.CreatedAt
			}
			if res.UpdatedAt.After(device.UpdatedAt) {
				device.UpdatedAt = res.UpdatedAt
			}

			return io.Write(os.Stdout, config.OutputFormat, &device)
		}),
	}
	endDevicesUpdateCommand = &cobra.Command{
		Use:     "update [application-id] [device-id]",
		Aliases: []string{"set"},
		Short:   "Update an end device",
		RunE: func(cmd *cobra.Command, args []string) error {
			forwardDeprecatedDeviceFlags(cmd.Flags())

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
			if ttnpb.HasAnyField(ttnpb.TopLevelFields(paths), "root_keys") {
				device.SupportsJoin = true
				paths = append(paths, "supports_join")
			}
			if err = util.SetFields(&device, setEndDeviceFlags); err != nil {
				return err
			}
			device.Attributes = mergeAttributes(device.Attributes, cmd.Flags())
			device.EndDeviceIdentifiers = *devID

			isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceSetPaths(device.SupportsJoin, paths...)

			if len(nsPaths) > 0 {
				isPaths = append(isPaths, "network_server_address")
			}
			if len(asPaths) > 0 {
				isPaths = append(isPaths, "application_server_address")
			}
			if len(jsPaths) > 0 {
				isPaths = append(isPaths, "join_server_address")
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.WithField("paths", isPaths).Debug("Get end device from Identity Server")
			existingDevice, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: *devID,
				FieldMask:            pbtypes.FieldMask{Paths: isPaths},
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

			if nsMismatch, asMismatch, jsMismatch := compareServerAddressesEndDevice(existingDevice, config); nsMismatch || asMismatch || jsMismatch {
				return errAddressMismatchEndDevice
			}

			res, err := setEndDevice(&device, isPaths, nsPaths, asPaths, jsPaths, false)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	endDevicesProvisionCommand = &cobra.Command{
		Use:   "provision",
		Short: "Provision end devices using vendor-specific data",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), nil)
			if appID == nil {
				return errNoApplicationID
			}

			provisionerID, _ := cmd.Flags().GetString("provisioner-id")
			data, err := getData(cmd.Flags())
			if err != nil {
				return err
			}

			req := &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: *appID,
				ProvisionerID:          provisionerID,
				ProvisioningData:       data,
			}

			var joinEUI types.EUI64
			if joinEUIHex, _ := cmd.Flags().GetString("join-eui"); joinEUIHex != "" {
				if err := joinEUI.UnmarshalText([]byte(joinEUIHex)); err != nil {
					return err
				}
			}
			if inputDecoder != nil {
				list := &ttnpb.ProvisionEndDevicesRequest_IdentifiersList{}
				for {
					var ids ttnpb.EndDeviceIdentifiers
					_, err := inputDecoder.Decode(&ids)
					if err == stdio.EOF {
						break
					}
					if err != nil {
						return err
					}
					ids.ApplicationIdentifiers = *appID
					if !joinEUI.IsZero() {
						list.JoinEUI = &joinEUI
					}
					list.EndDeviceIDs = append(list.EndDeviceIDs, ids)
				}
				req.EndDevices = &ttnpb.ProvisionEndDevicesRequest_List{
					List: list,
				}
			} else {
				if startDevEUIHex, _ := cmd.Flags().GetString("start-dev-eui"); startDevEUIHex != "" {
					var startDevEUI types.EUI64
					if err := startDevEUI.UnmarshalText([]byte(startDevEUIHex)); err != nil {
						return err
					}
					r := &ttnpb.ProvisionEndDevicesRequest_IdentifiersRange{
						StartDevEUI: startDevEUI,
					}
					if !joinEUI.IsZero() {
						r.JoinEUI = &joinEUI
					}
					req.EndDevices = &ttnpb.ProvisionEndDevicesRequest_Range{
						Range: r,
					}
				} else {
					fromData := &ttnpb.ProvisionEndDevicesRequest_IdentifiersFromData{}
					if !joinEUI.IsZero() {
						fromData.JoinEUI = &joinEUI
					}
					req.EndDevices = &ttnpb.ProvisionEndDevicesRequest_FromData{
						FromData: fromData,
					}
				}
			}

			js, err := api.Dial(ctx, config.JoinServerGRPCAddress)
			if err != nil {
				return err
			}
			stream, err := ttnpb.NewJsEndDeviceRegistryClient(js).Provision(ctx, req)
			if err != nil {
				return err
			}
			for {
				dev, err := stream.Recv()
				if err == stdio.EOF {
					return nil
				}
				if err != nil {
					return err
				}
				if err := io.Write(os.Stdout, config.OutputFormat, dev); err != nil {
					return err
				}
			}
		},
	}
	endDevicesDeleteCommand = &cobra.Command{
		Use:   "delete [application-id] [device-id]",
		Short: "Delete an end device",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			existingDevice, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: *devID,
				FieldMask: pbtypes.FieldMask{Paths: []string{
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

			if nsMismatch, asMismatch, jsMismatch := compareServerAddressesEndDevice(existingDevice, config); nsMismatch || asMismatch || jsMismatch {
				return errAddressMismatchEndDevice
			}

			return deleteEndDevice(ctx, devID)
		},
	}
)

func init() {
	util.FieldMaskFlags(&ttnpb.EndDevice{}).VisitAll(func(flag *pflag.Flag) {
		if ttnpb.ContainsField(flag.Name, getEndDeviceFromIS) {
			selectEndDeviceListFlags.AddFlag(flag)
			selectEndDeviceFlags.AddFlag(flag)
		} else if ttnpb.ContainsField(flag.Name, getEndDeviceFromNS) ||
			ttnpb.ContainsField(flag.Name, getEndDeviceFromAS) ||
			ttnpb.ContainsField(flag.Name, getEndDeviceFromJS) {
			selectEndDeviceFlags.AddFlag(flag)
		}
	})

	addDeprecatedDeviceFlags(selectEndDeviceListFlags)
	addDeprecatedDeviceFlags(selectEndDeviceFlags)

	util.FieldFlags(&ttnpb.EndDevice{}).VisitAll(func(flag *pflag.Flag) {
		if ttnpb.ContainsField(flag.Name, setEndDeviceToIS) ||
			ttnpb.ContainsField(flag.Name, setEndDeviceToNS) ||
			ttnpb.ContainsField(flag.Name, setEndDeviceToAS) ||
			ttnpb.ContainsField(flag.Name, setEndDeviceToJS) {
			setEndDeviceFlags.AddFlag(flag)
		}
	})

	addDeprecatedDeviceFlags(setEndDeviceFlags)

	endDevicesListFrequencyPlans.Flags().Uint32("base-frequency", 0, "Base frequency in MHz for hardware support (433, 470, 868 or 915)")
	endDevicesCommand.AddCommand(endDevicesListFrequencyPlans)
	endDevicesListCommand.Flags().AddFlagSet(applicationIDFlags())
	endDevicesListCommand.Flags().AddFlagSet(selectEndDeviceListFlags)
	endDevicesListCommand.Flags().AddFlagSet(paginationFlags())
	endDevicesCommand.AddCommand(endDevicesListCommand)
	endDevicesGetCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesGetCommand.Flags().AddFlagSet(selectEndDeviceFlags)
	endDevicesCommand.AddCommand(endDevicesGetCommand)
	endDevicesCreateCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesCreateCommand.Flags().AddFlagSet(setEndDeviceFlags)
	endDevicesCreateCommand.Flags().AddFlagSet(attributesFlags())
	endDevicesCreateCommand.Flags().Bool("defaults", true, "configure end device with defaults")
	endDevicesCreateCommand.Flags().Bool("with-root-keys", false, "generate OTAA root keys")
	endDevicesCreateCommand.Flags().Bool("abp", false, "configure end device as ABP")
	endDevicesCreateCommand.Flags().Bool("with-session", false, "generate ABP session DevAddr and keys")
	endDevicesCommand.AddCommand(endDevicesCreateCommand)
	endDevicesUpdateCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesUpdateCommand.Flags().AddFlagSet(setEndDeviceFlags)
	endDevicesUpdateCommand.Flags().AddFlagSet(attributesFlags())
	endDevicesCommand.AddCommand(endDevicesUpdateCommand)
	endDevicesProvisionCommand.Flags().AddFlagSet(applicationIDFlags())
	endDevicesProvisionCommand.Flags().AddFlagSet(dataFlags())
	endDevicesProvisionCommand.Flags().String("provisioner-id", "", "provisioner service")
	endDevicesProvisionCommand.Flags().String("join-eui", "", "(hex)")
	endDevicesProvisionCommand.Flags().String("start-dev-eui", "", "starting DevEUI to provision (hex)")
	endDevicesCommand.AddCommand(endDevicesProvisionCommand)
	endDevicesDeleteCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesCommand.AddCommand(endDevicesDeleteCommand)

	endDevicesCommand.AddCommand(applicationsDownlinkCommand)

	Root.AddCommand(endDevicesCommand)
}

var errAddressMismatchEndDevice = errors.DefineAborted("end_device_server_address_mismatch", "network/application/join server address mismatch")

func compareServerAddressesEndDevice(device *ttnpb.EndDevice, config *Config) (nsMismatch, asMismatch, jsMismatch bool) {
	nsHost, asHost, jsHost := getHost(config.NetworkServerGRPCAddress), getHost(config.ApplicationServerGRPCAddress), getHost(config.JoinServerGRPCAddress)
	if host := getHost(device.NetworkServerAddress); host != "" && host != nsHost {
		nsMismatch = true
		logger.WithFields(log.Fields(
			"configured", nsHost,
			"registered", host,
		)).Warn("Registered Network Server address does not match CLI configuration")
	}
	if host := getHost(device.ApplicationServerAddress); host != "" && host != asHost {
		asMismatch = true
		logger.WithFields(log.Fields(
			"configured", asHost,
			"registered", host,
		)).Warn("Registered Application Server address does not match CLI configuration")
	}
	if host := getHost(device.JoinServerAddress); host != "" && host != jsHost {
		jsMismatch = true
		logger.WithFields(log.Fields(
			"configured", jsHost,
			"registered", host,
		)).Warn("Registered Join Server address does not match CLI configuration")
	}
	return
}
