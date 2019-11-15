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
	"bufio"
	"context"
	"encoding/hex"
	stdio "io"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"strings"

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

func selectEndDeviceIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Bool("application-id", false, "")
	flagSet.Bool("device-id", false, "")
	flagSet.Bool("join-eui", false, "")
	flagSet.Bool("dev-eui", false, "")
	addDeprecatedDeviceFlags(flagSet)
	return flagSet
}

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
	errQRCodeFormat                 = errors.DefineInvalidArgument("qr_code_format", "invalid QR code format")
	errNoQRCodeTarget               = errors.DefineInvalidArgument("no_qr_code_target", "no QR code target specified")
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
		logger.Warn("Multiple IDs found in arguments, considering the first")
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

var errGatewayServerDisabled = errors.DefineFailedPrecondition("gateway_server_disabled", "Gateway Server is disabled")

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
			if !config.GatewayServerEnabled {
				return errGatewayServerDisabled
			}

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

			setDefaults, _ := cmd.Flags().GetBool("defaults")
			if setDefaults {
				if config.NetworkServerEnabled {
					device.NetworkServerAddress = getHost(config.NetworkServerGRPCAddress)
					paths = append(paths, "network_server_address")
				}
				if config.ApplicationServerEnabled {
					device.ApplicationServerAddress = getHost(config.ApplicationServerGRPCAddress)
					paths = append(paths, "application_server_address")
				}
			}

			abp, _ := cmd.Flags().GetBool("abp")
			multicast, _ := cmd.Flags().GetBool("multicast")
			if abp || multicast {
				device.SupportsJoin = false
				if config.NetworkServerEnabled {
					paths = append(paths, "supports_join")
				}
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
				if config.NetworkServerEnabled {
					paths = append(paths, "supports_join")
				}
				if setDefaults {
					if config.JoinServerEnabled {
						device.JoinServerAddress = getHost(config.JoinServerGRPCAddress)
						paths = append(paths,
							"join_server_address",
						)
					}
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
			if withClaimAuthenticationCode, _ := cmd.Flags().GetBool("with-claim-authentication-code"); withClaimAuthenticationCode {
				device.ClaimAuthenticationCode = &ttnpb.EndDeviceAuthenticationCode{
					Value: strings.ToUpper(hex.EncodeToString(random.Bytes(4))),
				}
				paths = append(paths, "claim_authentication_code")
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

			res, err := setEndDevice(&device, nil, nsPaths, asPaths, jsPaths, true, false)
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
			if ttnpb.HasAnyField(paths, setEndDeviceToJS...) {
				device.SupportsJoin = true
			}
			if err = util.SetFields(&device, setEndDeviceFlags); err != nil {
				return err
			}
			device.Attributes = mergeAttributes(device.Attributes, cmd.Flags())
			device.EndDeviceIdentifiers = *devID

			isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceSetPaths(device.SupportsJoin, paths...)

			if len(nsPaths) > 0 && config.NetworkServerEnabled {
				if device.NetworkServerAddress == "" {
					device.NetworkServerAddress = getHost(config.NetworkServerGRPCAddress)
				}
				isPaths = append(isPaths, "network_server_address")
			}
			if len(asPaths) > 0 && config.ApplicationServerEnabled {
				if device.ApplicationServerAddress == "" {
					device.ApplicationServerAddress = getHost(config.ApplicationServerGRPCAddress)
				}
				isPaths = append(isPaths, "application_server_address")
			}
			if len(jsPaths) > 0 && config.JoinServerEnabled {
				if device.JoinServerAddress == "" {
					device.JoinServerAddress = getHost(config.JoinServerGRPCAddress)
				}
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

			touch, _ := cmd.Flags().GetBool("touch")
			res, err := setEndDevice(&device, isPaths, nsPaths, asPaths, jsPaths, false, touch)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/999)
	endDevicesProvisionCommand = &cobra.Command{
		Use:   "provision",
		Short: "Provision end devices using vendor-specific data",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Warn("This command is deprecated. Please use `device template from-data` instead")

			appID := getApplicationID(cmd.Flags(), nil)
			if appID == nil {
				return errNoApplicationID
			}

			provisionerID, _ := cmd.Flags().GetString("provisioner-id")
			data, err := getDataBytes("", cmd.Flags())
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
	endDevicesClaimCommand = &cobra.Command{
		Use:   "claim [application-id]",
		Short: "Claim an end device (EXPERIMENTAL)",
		Long: `Claim an end device (EXPERIMENTAL)

The claiming procedure transfers devices from the source application to the
target application using the Device Claiming Server, thereby transferring
ownership of the device.

Authentication of device claiming is by the device's JoinEUI, DevEUI and claim
authentication code as stored in the Join Server. This information is typically
encoded in a QR code. This command supports claiming by QR code (via stdin), as
well as providing the claim information through the flags --source-join-eui,
--source-dev-eui, --source-authentication-code.

Claim authentication code validity is controlled by the owner of the device by
setting the value and optionally a time window when the code is valid. As part
of the claiming, the claim authentication code is invalidated by default to
block subsequent claiming attempts. You can keep the claim authentication code
valid by specifying --invalidate-authentication-code=false.

As part of claiming, you can optionally provide the target NetID, Network Server
KEK label and Application Server ID and KEK label. The Network Server and
Application Server addresses will be taken from the CLI configuration. These
values will be stored in the Join Server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetAppID := getApplicationID(cmd.Flags(), args)
			if targetAppID == nil {
				return errNoApplicationID
			}

			req := &ttnpb.ClaimEndDeviceRequest{
				TargetApplicationIDs: *targetAppID,
			}

			var joinEUI, devEUI *types.EUI64
			if joinEUIHex, _ := cmd.Flags().GetString("source-join-eui"); joinEUIHex != "" {
				joinEUI = new(types.EUI64)
				if err := joinEUI.UnmarshalText([]byte(joinEUIHex)); err != nil {
					return err
				}
			}
			if devEUIHex, _ := cmd.Flags().GetString("source-dev-eui"); devEUIHex != "" {
				devEUI = new(types.EUI64)
				if err := devEUI.UnmarshalText([]byte(devEUIHex)); err != nil {
					return err
				}
			}
			if joinEUI != nil && devEUI != nil {
				authenticationCode, _ := cmd.Flags().GetString("source-authentication-code")
				req.SourceDevice = &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
						JoinEUI:            *joinEUI,
						DevEUI:             *devEUI,
						AuthenticationCode: authenticationCode,
					},
				}
			} else {
				if joinEUI != nil || devEUI != nil {
					logger.Warn("Either target JoinEUI or DevEUI specified but need both, not considering any and using scan mode")
				}
				if !io.IsPipe(os.Stdin) {
					logger.Info("Scan QR code")
				}
				qrCode, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
				if err != nil {
					return err
				}
				qrCode = qrCode[:len(qrCode)-1]
				logger.WithField("code", string(qrCode)).Debug("Scanned QR code")
				req.SourceDevice = &ttnpb.ClaimEndDeviceRequest_QRCode{
					QRCode: qrCode,
				}
			}

			req.TargetDeviceID, _ = cmd.Flags().GetString("target-device-id")
			if netIDHex, _ := cmd.Flags().GetString("target-net-id"); netIDHex != "" {
				if err := req.TargetNetID.UnmarshalText([]byte(netIDHex)); err != nil {
					return err
				}
			}
			if config.NetworkServerEnabled {
				req.TargetNetworkServerAddress = config.NetworkServerGRPCAddress
			}
			req.TargetNetworkServerKEKLabel, _ = cmd.Flags().GetString("target-network-server-kek-label")
			if config.ApplicationServerEnabled {
				req.TargetApplicationServerAddress = config.ApplicationServerGRPCAddress
			}
			req.TargetApplicationServerKEKLabel, _ = cmd.Flags().GetString("target-application-server-kek-label")
			req.TargetApplicationServerID, _ = cmd.Flags().GetString("target-application-server-id")
			req.InvalidateAuthenticationCode, _ = cmd.Flags().GetBool("invalidate-authentication-code")

			dcs, err := api.Dial(ctx, config.DeviceClaimingServerGRPCAddress)
			if err != nil {
				return err
			}
			ids, err := ttnpb.NewEndDeviceClaimingServerClient(dcs).Claim(ctx, req)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, ids)
		},
	}
	endDevicesListQRCodeFormatsCommand = &cobra.Command{
		Use:     "list-qr-formats",
		Aliases: []string{"ls-qr-formats", "listqrformats", "lsqrformats", "lsqrfmts", "lsqrfmt"},
		Short:   "List QR code formats (EXPERIMENTAL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			qrg, err := api.Dial(ctx, config.QRCodeGeneratorGRPCAddress)
			if err != nil {
				return err
			}

			res, err := ttnpb.NewEndDeviceQRCodeGeneratorClient(qrg).ListFormats(ctx, ttnpb.Empty)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	endDevicesGenerateQRCommand = &cobra.Command{
		Use:     "generate-qr [application-id] [device-id]",
		Aliases: []string{"genqr"},
		Short:   "Generate an end device QR code (EXPERIMENTAL)",
		Long: `Generate an end device QR code (EXPERIMENTAL)

This command saves a QR code in PNG format in the given folder. The filename is
the device ID.

This command may take end device identifiers from stdin.`,
		Example: `To generate a QR code for a single end device:
  ttn-lw-cli end-devices generate-qr app1 dev1

To generate a QR code for multiple end devices:
  ttn-lw-cli end-devices list app1 \
    | ttn-lw-cli end-devices generate-qr`,
		RunE: asBulk(func(cmd *cobra.Command, args []string) error {
			var ids *ttnpb.EndDeviceIdentifiers
			if inputDecoder != nil {
				var dev ttnpb.EndDevice
				if _, err := inputDecoder.Decode(&dev); err != nil {
					return err
				}
				if dev.ApplicationID == "" {
					return errNoApplicationID
				}
				if dev.DeviceID == "" {
					return errNoEndDeviceID
				}
				ids = &dev.EndDeviceIdentifiers
			} else {
				var err error
				ids, err = getEndDeviceID(cmd.Flags(), args, true)
				if err != nil {
					return err
				}
			}

			formatID, _ := cmd.Flags().GetString("format-id")

			qrg, err := api.Dial(ctx, config.QRCodeGeneratorGRPCAddress)
			if err != nil {
				return err
			}
			client := ttnpb.NewEndDeviceQRCodeGeneratorClient(qrg)
			format, err := client.GetFormat(ctx, &ttnpb.GetQRCodeFormatRequest{
				FormatID: formatID,
			})
			if err != nil {
				return err
			}

			isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceGetPaths(format.FieldMask.Paths...)

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
				EndDeviceIdentifiers: *ids,
				FieldMask:            pbtypes.FieldMask{Paths: isPaths},
			})
			if err != nil {
				return err
			}

			nsMismatch, asMismatch, jsMismatch := compareServerAddressesEndDevice(device, config)
			if len(nsPaths) > 0 && nsMismatch {
				return errAddressMismatchEndDevice
			}
			if len(asPaths) > 0 && asMismatch {
				return errAddressMismatchEndDevice
			}
			if len(jsPaths) > 0 && jsMismatch {
				return errAddressMismatchEndDevice
			}

			dev, err := getEndDevice(device.EndDeviceIdentifiers, nsPaths, asPaths, jsPaths, true)
			if err != nil {
				return err
			}
			device.SetFields(dev, append(append(nsPaths, asPaths...), jsPaths...)...)

			size, _ := cmd.Flags().GetUint32("size")
			res, err := client.Generate(ctx, &ttnpb.GenerateEndDeviceQRCodeRequest{
				FormatID:  formatID,
				EndDevice: *device,
				Image: &ttnpb.GenerateEndDeviceQRCodeRequest_Image{
					ImageSize: size,
				},
			})
			if err != nil {
				return err
			}

			folder, _ := cmd.Flags().GetString("folder")
			if folder == "" {
				folder, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			var ext string
			if exts, err := mime.ExtensionsByType(res.Image.Embedded.MimeType); err == nil && len(exts) > 0 {
				ext = exts[0]
			}
			filename := path.Join(folder, device.DeviceID+ext)
			if err := ioutil.WriteFile(filename, res.Image.Embedded.Data, 0644); err != nil {
				return err
			}

			logger.WithFields(log.Fields(
				"value", res.Text,
				"filename", filename,
			)).Info("Generated QR code")
			return nil
		}),
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
	endDevicesCreateCommand.Flags().Bool("with-claim-authentication-code", false, "generate claim authentication code of 4 bytes")
	endDevicesCommand.AddCommand(endDevicesCreateCommand)
	endDevicesUpdateCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesUpdateCommand.Flags().AddFlagSet(setEndDeviceFlags)
	endDevicesUpdateCommand.Flags().AddFlagSet(attributesFlags())
	endDevicesUpdateCommand.Flags().Bool("touch", false, "set in all registries even if no fields are specified")
	endDevicesCommand.AddCommand(endDevicesUpdateCommand)
	endDevicesProvisionCommand.Flags().AddFlagSet(applicationIDFlags())
	endDevicesProvisionCommand.Flags().AddFlagSet(dataFlags("", ""))
	endDevicesProvisionCommand.Flags().String("provisioner-id", "", "provisioner service")
	endDevicesProvisionCommand.Flags().String("join-eui", "", "(hex)")
	endDevicesProvisionCommand.Flags().String("start-dev-eui", "", "starting DevEUI to provision (hex)")
	endDevicesCommand.AddCommand(endDevicesProvisionCommand)
	endDevicesDeleteCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesCommand.AddCommand(endDevicesDeleteCommand)
	endDevicesClaimCommand.Flags().AddFlagSet(applicationIDFlags())
	endDevicesClaimCommand.Flags().String("source-join-eui", "", "(hex)")
	endDevicesClaimCommand.Flags().String("source-dev-eui", "", "(hex)")
	endDevicesClaimCommand.Flags().String("source-authentication-code", "", "(hex)")
	endDevicesClaimCommand.Flags().String("target-device-id", "", "")
	endDevicesClaimCommand.Flags().String("target-net-id", "", "(hex)")
	endDevicesClaimCommand.Flags().String("target-network-server-kek-label", "", "")
	endDevicesClaimCommand.Flags().String("target-application-server-kek-label", "", "")
	endDevicesClaimCommand.Flags().String("target-application-server-id", "", "")
	endDevicesClaimCommand.Flags().Bool("invalidate-authentication-code", true, "invalidate the claim authentication code to block subsequent claiming attempts")
	endDevicesCommand.AddCommand(endDevicesClaimCommand)
	endDevicesCommand.AddCommand(endDevicesListQRCodeFormatsCommand)
	endDevicesGenerateQRCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesGenerateQRCommand.Flags().String("format-id", "", "")
	endDevicesGenerateQRCommand.Flags().Uint32("size", 300, "size of the image in pixels")
	endDevicesGenerateQRCommand.Flags().String("folder", "", "folder to write the QR code image to")
	endDevicesCommand.AddCommand(endDevicesGenerateQRCommand)

	endDevicesCommand.AddCommand(applicationsDownlinkCommand)

	Root.AddCommand(endDevicesCommand)

	endDeviceTemplatesExtendCommand.Flags().AddFlagSet(setEndDeviceFlags)
	endDeviceTemplatesCreateCommand.Flags().AddFlagSet(selectEndDeviceFlags)
	endDeviceTemplatesExecuteCommand.Flags().AddFlagSet(setEndDeviceFlags)
}

var errAddressMismatchEndDevice = errors.DefineAborted("end_device_server_address_mismatch", "Network/Application/Join Server address mismatch")

func compareServerAddressesEndDevice(device *ttnpb.EndDevice, config *Config) (nsMismatch, asMismatch, jsMismatch bool) {
	nsHost, asHost, jsHost := getHost(config.NetworkServerGRPCAddress), getHost(config.ApplicationServerGRPCAddress), getHost(config.JoinServerGRPCAddress)
	if host := getHost(device.NetworkServerAddress); config.NetworkServerEnabled && host != "" && host != nsHost {
		nsMismatch = true
		logger.WithFields(log.Fields(
			"configured", nsHost,
			"registered", host,
		)).Warn("Registered Network Server address does not match CLI configuration")
	}
	if host := getHost(device.ApplicationServerAddress); config.ApplicationServerEnabled && host != "" && host != asHost {
		asMismatch = true
		logger.WithFields(log.Fields(
			"configured", asHost,
			"registered", host,
		)).Warn("Registered Application Server address does not match CLI configuration")
	}
	if host := getHost(device.JoinServerAddress); config.JoinServerEnabled && host != "" && host != jsHost {
		jsMismatch = true
		logger.WithFields(log.Fields(
			"configured", jsHost,
			"registered", host,
		)).Warn("Registered Join Server address does not match CLI configuration")
	}
	return
}
