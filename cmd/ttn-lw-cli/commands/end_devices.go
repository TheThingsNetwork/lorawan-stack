// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"crypto/rand"
	"encoding/hex"
	stdio "io"
	"mime"
	"os"
	"path"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
)

var (
	selectEndDeviceListFlags   = &pflag.FlagSet{}
	selectEndDeviceFlags       = &pflag.FlagSet{}
	setEndDeviceFlags          = &pflag.FlagSet{}
	endDeviceFlattenPaths      = []string{"provisioning_data"}
	endDevicePictureFlags      = &pflag.FlagSet{}
	endDeviceLocationFlags     = util.FieldFlags(&ttnpb.Location{}, "location")
	getDefaultMACSettingsFlags = util.FieldFlags(&ttnpb.GetDefaultMACSettingsRequest{})

	selectAllEndDeviceFlags = util.SelectAllFlagSet("end devices")
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
	errConflictingPaths             = errors.DefineInvalidArgument("conflicting_paths", "conflicting set and unset field mask paths")
	errEndDeviceEUIUpdate           = errors.DefineInvalidArgument("end_device_eui_update", "end device EUIs can not be updated")
	errEndDeviceKeysWithProvisioner = errors.DefineInvalidArgument("end_device_keys_provisioner", "end device ABP or OTAA keys cannot be set when there is a provisioner")
	errInconsistentEndDeviceEUI     = errors.DefineInvalidArgument("inconsistent_end_device_eui", "given end device EUIs do not match registered EUIs")
	errInvalidDataRateIndex         = errors.DefineInvalidArgument("data_rate_index", "Data rate index is invalid")
	errInvalidMACVersion            = errors.DefineInvalidArgument("mac_version", "LoRaWAN MAC version is invalid")
	errInvalidPHYVersion            = errors.DefineInvalidArgument("phy_version", "LoRaWAN PHY version is invalid")
	errNoEndDeviceEUI               = errors.DefineInvalidArgument("no_end_device_eui", "no end device EUIs set")
	errInvalidJoinEUI               = errors.DefineInvalidArgument("invalid_join_eui", "invalid JoinEUI")
	errInvalidDevEUI                = errors.DefineInvalidArgument("invalid_dev_eui", "invalid DevEUI")
	errInvalidNetID                 = errors.DefineInvalidArgument("invalid_net_id", "invalid NetID")
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
		logger.Warn("Multiple IDs found in arguments, considering the first")
		applicationID = args[0]
		deviceID = args[1]
	}
	if applicationID == "" && requireID {
		return nil, errNoApplicationID.New()
	}
	if deviceID == "" && requireID {
		return nil, errNoEndDeviceID.New()
	}
	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: applicationID},
		DeviceId:       deviceID,
	}
	if joinEUIHex, _ := flagSet.GetString("join-eui"); joinEUIHex != "" {
		var joinEUI types.EUI64
		if err := joinEUI.UnmarshalText([]byte(joinEUIHex)); err != nil {
			return nil, errInvalidJoinEUI.WithCause(err)
		}
		ids.JoinEui = &joinEUI
	}
	if devEUIHex, _ := flagSet.GetString("dev-eui"); devEUIHex != "" {
		var devEUI types.EUI64
		if err := devEUI.UnmarshalText([]byte(devEUIHex)); err != nil {
			return nil, errInvalidDevEUI.WithCause(err)
		}
		ids.DevEui = &devEUI
	}
	return ids, nil
}

func generateKey() *types.AES128Key {
	var key types.AES128Key
	rand.Read(key[:])
	return &key
}

var (
	errJoinServerDisabled    = errors.DefineFailedPrecondition("join_server_disabled", "Join Server is disabled")
	errNetworkServerDisabled = errors.DefineFailedPrecondition("network_server_disabled", "Network Server is disabled")
)

var searchEndDevicesFlags = func() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.AddFlagSet(searchFlags)
	// NOTE: These flags need to be named with underscores, not dashes!
	flagSet.String("dev_eui_contains", "", "")
	flagSet.String("join_eui_contains", "", "")
	flagSet.String("dev_addr_contains", "", "")
	flagSet.Lookup("dev_addr_contains").Hidden = true // Part of the API but not actually supported.
	return flagSet
}()

var (
	endDevicesCommand = &cobra.Command{
		Use:     "end-devices",
		Aliases: []string{"end-device", "devices", "device", "dev", "ed", "d"},
		Short:   "End Device commands",
	}
	endDevicesListFrequencyPlans = &cobra.Command{
		Use:               "list-frequency-plans",
		Aliases:           []string{"get-frequency-plans", "frequency-plans", "fps"},
		Short:             "List available frequency plans for end devices",
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.NetworkServerEnabled {
				return errNetworkServerDisabled.New()
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
				return errNoApplicationID.New()
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectEndDeviceListFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.EndDeviceRegistry/List"].Allowed)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewEndDeviceRegistryClient(is).List(ctx, &ttnpb.ListEndDevicesRequest{
				ApplicationIds: appID,
				FieldMask:      &pbtypes.FieldMask{Paths: paths},
				Limit:          limit,
				Page:           page,
				Order:          getOrder(cmd.Flags()),
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.EndDevices)
		},
	}
	endDevicesSearchCommand = &cobra.Command{
		Use:   "search [application-id]",
		Short: "Search for end devices",
		RunE: func(cmd *cobra.Command, args []string) error {
			forwardDeprecatedDeviceFlags(cmd.Flags())

			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID.New()
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectEndDeviceListFlags)

			req := &ttnpb.SearchEndDevicesRequest{}
			if err := util.SetFields(req, searchEndDevicesFlags); err != nil {
				return err
			}
			var (
				opt      grpc.CallOption
				getTotal func() uint64
			)
			req.Limit, req.Page, opt, getTotal = withPagination(cmd.Flags())
			req.ApplicationIds = appID
			req.FieldMask = &pbtypes.FieldMask{Paths: paths}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewEndDeviceRegistrySearchClient(is).SearchEndDevices(ctx, req, opt)
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
				EndDeviceIds: devID,
				FieldMask:    &pbtypes.FieldMask{Paths: isPaths},
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

			res, err := getEndDevice(device.Ids, nsPaths, asPaths, jsPaths, true)
			if err != nil {
				return err
			}

			device.SetFields(res, "ids.dev_addr")
			device.SetFields(res, append(append(nsPaths, asPaths...), jsPaths...)...)
			if device.CreatedAt == nil || (res.CreatedAt != nil && ttnpb.StdTime(res.CreatedAt).Before(*ttnpb.StdTime(device.CreatedAt))) {
				device.CreatedAt = res.CreatedAt
			}
			if res.UpdatedAt != nil && ttnpb.StdTime(res.UpdatedAt).After(*ttnpb.StdTime(device.UpdatedAt)) {
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

			abp, _ := cmd.Flags().GetBool("abp")
			multicast, _ := cmd.Flags().GetBool("multicast")
			abp = abp || multicast
			device := &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{},
			}
			if inputDecoder != nil {
				decodedPaths, err := inputDecoder.Decode(device)
				if err != nil {
					return err
				}
				paths = append(paths, ttnpb.FlattenPaths(decodedPaths, endDeviceFlattenPaths)...)

				if abp && device.SupportsJoin {
					logger.Warn("Reading from standard input, ignoring --abp and --multicast flags")
				}
				abp = !device.SupportsJoin
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

			if picture, err := cmd.Flags().GetString("picture"); err == nil && picture != "" {
				device.Picture, err = readPicture(picture)
				if err != nil {
					return err
				}
			}

			if abp {
				device.SupportsJoin = false
				if config.NetworkServerEnabled {
					paths = append(paths, "supports_join")
				}
				if withSession, _ := cmd.Flags().GetBool("with-session"); withSession {
					if device.ProvisionerId != "" {
						return errEndDeviceKeysWithProvisioner.New()
					}
					ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
					if err != nil {
						return err
					}
					devAddrRes, err := ttnpb.NewNsClient(ns).GenerateDevAddr(ctx, ttnpb.Empty)
					if err != nil {
						return err
					}
					device.Ids.DevAddr = devAddrRes.DevAddr
					device.Session = &ttnpb.Session{
						DevAddr: *devAddrRes.DevAddr,
						Keys: &ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{Key: generateKey()},
							AppSKey:     &ttnpb.KeyEnvelope{Key: generateKey()},
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
						return errInvalidMACVersion.WithCause(err)
					}
					if err := macVersion.Validate(); err != nil {
						return errInvalidMACVersion.WithCause(err)
					}
					if macVersion.Compare(ttnpb.MACVersion_MAC_V1_1) >= 0 {
						device.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{Key: generateKey()}
						device.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{Key: generateKey()}
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
						if device.Ids.JoinEui == nil {
							// Get the default JoinEUI for JS.
							logger.WithField("join_server_address", config.JoinServerGRPCAddress).Info("JoinEUI empty but defaults flag is set, fetch default JoinEUI of the Join Server")
							js, err := api.Dial(ctx, config.JoinServerGRPCAddress)
							if err != nil {
								return err
							}
							defaultJoinEUI, err := ttnpb.NewJsClient(js).GetDefaultJoinEUI(ctx, ttnpb.Empty)
							if err != nil {
								return err
							}
							logger.WithField("default_join_eui", defaultJoinEUI.JoinEui.String()).
								Info("Successfully obtained Join Server's default JoinEUI")
							device.Ids.JoinEui = defaultJoinEUI.JoinEui
						}
					}
				}
				if withKeys, _ := cmd.Flags().GetBool("with-root-keys"); withKeys {
					if device.ProvisionerId != "" {
						return errEndDeviceKeysWithProvisioner.New()
					}
					// TODO: Set JoinEUI and DevEUI (https://github.com/TheThingsNetwork/lorawan-stack/issues/47).
					device.RootKeys = &ttnpb.RootKeys{
						RootKeyId: "ttn-lw-cli-generated",
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
			if hasUpdateDeviceLocationFlags(cmd.Flags()) {
				updateDeviceLocation(device, cmd.Flags())
				paths = append(paths, "locations")
			}

			if err = util.SetFields(device, setEndDeviceFlags); err != nil {
				return err
			}

			device.Attributes = mergeAttributes(device.Attributes, cmd.Flags())
			if devID != nil {
				if devID.DeviceId != "" {
					device.Ids.DeviceId = devID.DeviceId
				}
				if devID.ApplicationIds != nil {
					device.Ids.ApplicationIds = devID.ApplicationIds
				}
				if device.SupportsJoin && devID.JoinEui != nil {
					device.Ids.JoinEui = devID.JoinEui
				}
				if devID.DevEui != nil {
					device.Ids.DevEui = devID.DevEui
				}
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			requestDevEUI, _ := cmd.Flags().GetBool("request-dev-eui")
			if requestDevEUI {
				logger.Debug("request-dev-eui flag set, requesting a DevEUI")
				devEUIResponse, err := ttnpb.NewApplicationRegistryClient(is).IssueDevEUI(ctx, devID.ApplicationIds)
				if err != nil {
					return err
				}
				logger.WithField("dev_eui", devEUIResponse.DevEui.String()).
					Info("Successfully obtained DevEUI")
				device.Ids.DevEui = &devEUIResponse.DevEui
			}
			newPaths, err := parsePayloadFormatterParameterFlags("formatters", device.Formatters, cmd.Flags())
			if err != nil {
				return err
			}
			paths = append(paths, newPaths...)

			if device.GetIds().GetApplicationIds().GetApplicationId() == "" {
				return errNoApplicationID.New()
			}
			if device.Ids.DeviceId == "" {
				return errNoEndDeviceID.New()
			}

			isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceSetPaths(device.SupportsJoin, paths...)

			// Require EUIs for devices that need to be added to the Join Server.
			if len(jsPaths) > 0 && (device.Ids.JoinEui == nil || device.Ids.DevEui == nil) {
				return errNoEndDeviceEUI.New()
			}
			var isDevice ttnpb.EndDevice
			logger.WithField("paths", isPaths).Debug("Create end device on Identity Server")
			isDevice.SetFields(device, append(isPaths, "ids")...)
			isRes, err := ttnpb.NewEndDeviceRegistryClient(is).Create(ctx, &ttnpb.CreateEndDeviceRequest{
				EndDevice: isDevice,
			})
			if err != nil {
				return err
			}

			device.SetFields(isRes, append(isPaths, "created_at", "updated_at")...)

			res, err := setEndDevice(device, nil, nsPaths, asPaths, jsPaths, nil, true, false)
			if err != nil {
				logger.WithError(err).Error("Could not create end device, rolling back...")
				if err := deleteEndDevice(context.Background(), device.Ids); err != nil {
					logger.WithError(err).Error("Could not roll back end device creation")
				}
				return err
			}

			device.SetFields(res, append(append(nsPaths, asPaths...), jsPaths...)...)
			if device.CreatedAt == nil || (res.CreatedAt != nil && ttnpb.StdTime(res.CreatedAt).Before(*ttnpb.StdTime(device.CreatedAt))) {
				device.CreatedAt = res.CreatedAt
			}
			if res.UpdatedAt != nil && ttnpb.StdTime(res.UpdatedAt).After(*ttnpb.StdTime(device.UpdatedAt)) {
				device.UpdatedAt = res.UpdatedAt
			}

			return io.Write(os.Stdout, config.OutputFormat, device)
		}),
	}
	endDevicesSetCommand = &cobra.Command{
		Use:     "set [application-id] [device-id]",
		Aliases: []string{"update"},
		Short:   "Set properties of an end device",
		RunE: func(cmd *cobra.Command, args []string) error {
			forwardDeprecatedDeviceFlags(cmd.Flags())

			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setEndDeviceFlags, attributesFlags(), endDevicePictureFlags)
			rawUnsetPaths, _ := cmd.Flags().GetStringSlice("unset")
			unsetPaths := util.NormalizePaths(rawUnsetPaths)

			if hasUpdateDeviceLocationFlags(cmd.Flags()) {
				paths = append(paths, "locations")
			}

			if len(paths)+len(unsetPaths) == 0 {
				logger.Warn("No fields selected, won't update anything")
				return nil
			}
			if remainingPaths := ttnpb.ExcludeFields(paths, unsetPaths...); len(remainingPaths) != len(paths) {
				overlapPaths := ttnpb.ExcludeFields(paths, remainingPaths...)
				return errConflictingPaths.WithAttributes("field_mask_paths", overlapPaths)
			}
			device := &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{},
			}
			if ttnpb.HasAnyField(paths, setEndDeviceToJS...) || ttnpb.HasAnyField(unsetPaths, setEndDeviceToJS...) {
				device.SupportsJoin = true
			}
			if err = util.SetFields(device, setEndDeviceFlags); err != nil {
				return err
			}
			newPaths, err := parsePayloadFormatterParameterFlags("formatters", device.Formatters, cmd.Flags())
			if err != nil {
				return err
			}
			paths = append(paths, newPaths...)
			device.Attributes = mergeAttributes(device.Attributes, cmd.Flags())
			device.Ids = devID

			paths = append(paths, unsetPaths...)
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

			if picture, err := cmd.Flags().GetString("picture"); err == nil && picture != "" {
				device.Picture, err = readPicture(picture)
				if err != nil {
					return err
				}
				isPaths = append(paths, "picture")
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.WithField("paths", isPaths).Debug("Get end device from Identity Server")
			existingDevice, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
				FieldMask:    &pbtypes.FieldMask{Paths: ttnpb.ExcludeFields(isPaths, unsetPaths...)},
			})
			if err != nil {
				return err
			}

			// EUIs can not be updated, so we only accept EUI flags if they are equal to the existing ones.
			if device.Ids.JoinEui != nil {
				if existingDevice.Ids.JoinEui != nil && *device.Ids.JoinEui != *existingDevice.Ids.JoinEui {
					return errEndDeviceEUIUpdate.New()
				}
			} else {
				device.Ids.JoinEui = existingDevice.Ids.JoinEui
			}
			if device.Ids.DevEui != nil {
				if existingDevice.Ids.DevEui != nil && *device.Ids.DevEui != *existingDevice.Ids.DevEui {
					return errEndDeviceEUIUpdate.New()
				}
			} else {
				device.Ids.DevEui = existingDevice.Ids.DevEui
			}

			// Require EUIs for devices that need to be updated in the Join Server.
			if len(jsPaths) > 0 && (device.Ids.JoinEui == nil || device.Ids.DevEui == nil) {
				return errNoEndDeviceEUI.New()
			}

			if nsMismatch, asMismatch, jsMismatch := compareServerAddressesEndDevice(existingDevice, config); nsMismatch || asMismatch || jsMismatch {
				return errAddressMismatchEndDevice.New()
			}

			if hasUpdateDeviceLocationFlags(cmd.Flags()) {
				device.SetFields(existingDevice, "locations")
				updateDeviceLocation(device, cmd.Flags())
			}

			touch, _ := cmd.Flags().GetBool("touch")
			res, err := setEndDevice(device, isPaths, nsPaths, asPaths, jsPaths, unsetPaths, false, touch)
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
				return errNoApplicationID.New()
			}

			provisionerID, _ := cmd.Flags().GetString("provisioner-id")
			data, err := getDataBytes("", cmd.Flags())
			if err != nil {
				return err
			}

			req := &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIds:   appID,
				ProvisionerId:    provisionerID,
				ProvisioningData: data,
			}

			var joinEUI types.EUI64
			if joinEUIHex, _ := cmd.Flags().GetString("join-eui"); joinEUIHex != "" {
				if err := joinEUI.UnmarshalText([]byte(joinEUIHex)); err != nil {
					return errInvalidJoinEUI.WithCause(err)
				}
			}
			if inputDecoder != nil {
				list := &ttnpb.ProvisionEndDevicesRequest_IdentifiersList{
					JoinEui: &joinEUI,
				}
				for {
					var ids ttnpb.EndDeviceIdentifiers
					_, err := inputDecoder.Decode(&ids)
					if err == stdio.EOF {
						break
					}
					if err != nil {
						return err
					}
					ids.ApplicationIds = appID
					list.EndDeviceIds = append(list.EndDeviceIds, &ids)
				}
				req.EndDevices = &ttnpb.ProvisionEndDevicesRequest_List{
					List: list,
				}
			} else {
				if startDevEUIHex, _ := cmd.Flags().GetString("start-dev-eui"); startDevEUIHex != "" {
					var startDevEUI types.EUI64
					if err := startDevEUI.UnmarshalText([]byte(startDevEUIHex)); err != nil {
						return errInvalidDevEUI.WithCause(err)
					}
					req.EndDevices = &ttnpb.ProvisionEndDevicesRequest_Range{
						Range: &ttnpb.ProvisionEndDevicesRequest_IdentifiersRange{
							StartDevEui: startDevEUI,
							JoinEui:     &joinEUI,
						},
					}
				} else {
					req.EndDevices = &ttnpb.ProvisionEndDevicesRequest_FromData{
						FromData: &ttnpb.ProvisionEndDevicesRequest_IdentifiersFromData{
							JoinEui: &joinEUI,
						},
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
	endDevicesResetCommand = &cobra.Command{
		Use:   "reset [application-id] [device-id]",
		Short: "Reset state of an end device to factory defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			forwardDeprecatedDeviceFlags(cmd.Flags())

			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectEndDeviceFlags)

			isPaths, nsPaths, _, _ := splitEndDeviceGetPaths(paths...)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.WithField("paths", isPaths).Debug("Get end device from Identity Server")
			device, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
				FieldMask:    &pbtypes.FieldMask{Paths: isPaths},
			})
			if err != nil {
				return err
			}

			nsMismatch, _, _ := compareServerAddressesEndDevice(device, config)
			if nsMismatch {
				return errors.New("Network Server address does not match")
			}

			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			logger.WithField("paths", nsPaths).Debug("Reset end device to factory defaults on Network Server")
			nsDevice, err := ttnpb.NewNsEndDeviceRegistryClient(ns).ResetFactoryDefaults(ctx, &ttnpb.ResetAndGetEndDeviceRequest{
				EndDeviceIds: devID,
				FieldMask:    &pbtypes.FieldMask{Paths: nsPaths},
			})
			if err != nil {
				return err
			}
			device.SetFields(nsDevice, "ids.dev_addr")
			device.SetFields(nsDevice, ttnpb.AllowedBottomLevelFields(nsPaths, getEndDeviceFromNS)...)
			if device.CreatedAt == nil || (nsDevice.CreatedAt != nil && ttnpb.StdTime(nsDevice.CreatedAt).Before(*ttnpb.StdTime(device.CreatedAt))) {
				device.CreatedAt = nsDevice.CreatedAt
			}
			if nsDevice.UpdatedAt != nil && ttnpb.StdTime(nsDevice.UpdatedAt).After(*ttnpb.StdTime(device.UpdatedAt)) {
				device.UpdatedAt = nsDevice.UpdatedAt
			}
			return io.Write(os.Stdout, config.OutputFormat, device)
		},
	}
	endDevicesDeleteCommand = &cobra.Command{
		Use:     "delete [application-id] [device-id]",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete an end device",
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
				EndDeviceIds: devID,
				FieldMask: &pbtypes.FieldMask{Paths: []string{
					"network_server_address",
					"application_server_address",
					"join_server_address",
				}},
			})
			if err != nil {
				return err
			}

			// EUIs must match registered EUIs if set.
			if devID.JoinEui != nil {
				if existingDevice.Ids.JoinEui != nil && *devID.JoinEui != *existingDevice.Ids.JoinEui {
					return errInconsistentEndDeviceEUI.New()
				}
			} else {
				devID.JoinEui = existingDevice.Ids.JoinEui
			}
			if devID.DevEui != nil {
				if existingDevice.Ids.DevEui != nil && *devID.DevEui != *existingDevice.Ids.DevEui {
					return errInconsistentEndDeviceEUI.New()
				}
			} else {
				devID.DevEui = existingDevice.Ids.DevEui
			}

			if nsMismatch, asMismatch, jsMismatch := compareServerAddressesEndDevice(existingDevice, config); nsMismatch || asMismatch || jsMismatch {
				return errAddressMismatchEndDevice.New()
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
				return errNoApplicationID.New()
			}

			req := &ttnpb.ClaimEndDeviceRequest{
				TargetApplicationIds: targetAppID,
			}

			var joinEUI, devEUI *types.EUI64
			if joinEUIHex, _ := cmd.Flags().GetString("source-join-eui"); joinEUIHex != "" {
				joinEUI = new(types.EUI64)
				if err := joinEUI.UnmarshalText([]byte(joinEUIHex)); err != nil {
					return errInvalidJoinEUI.WithCause(err)
				}
			}
			if devEUIHex, _ := cmd.Flags().GetString("source-dev-eui"); devEUIHex != "" {
				devEUI = new(types.EUI64)
				if err := devEUI.UnmarshalText([]byte(devEUIHex)); err != nil {
					return errInvalidDevEUI.WithCause(err)
				}
			}
			if joinEUI != nil && devEUI != nil {
				authenticationCode, _ := cmd.Flags().GetString("source-authentication-code")
				req.SourceDevice = &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
						JoinEui:            *joinEUI,
						DevEui:             *devEUI,
						AuthenticationCode: authenticationCode,
					},
				}
			} else {
				if joinEUI != nil || devEUI != nil {
					logger.Warn("Either target JoinEUI or DevEUI specified but need both, not considering any and using scan mode")
				}
				rd, ok := io.BufferedPipe(os.Stdin)
				if !ok {
					logger.Info("Scan QR code")
					rd = bufio.NewReader(os.Stdin)
				}
				qrCode, err := rd.ReadBytes('\n')
				if err != nil {
					return err
				}
				qrCode = qrCode[:len(qrCode)-1]
				logger.WithField("code", string(qrCode)).Debug("Scanned QR code")
				req.SourceDevice = &ttnpb.ClaimEndDeviceRequest_QrCode{
					QrCode: qrCode,
				}
			}

			req.TargetDeviceId, _ = cmd.Flags().GetString("target-device-id")
			if netIDHex, _ := cmd.Flags().GetString("target-net-id"); netIDHex != "" {
				if err := req.TargetNetId.UnmarshalText([]byte(netIDHex)); err != nil {
					return errInvalidNetID.WithCause(err)
				}
			}
			if config.NetworkServerEnabled {
				req.TargetNetworkServerAddress = config.NetworkServerGRPCAddress
			}
			req.TargetNetworkServerKekLabel, _ = cmd.Flags().GetString("target-network-server-kek-label")
			if config.ApplicationServerEnabled {
				req.TargetApplicationServerAddress = config.ApplicationServerGRPCAddress
			}
			req.TargetApplicationServerKekLabel, _ = cmd.Flags().GetString("target-application-server-kek-label")
			req.TargetApplicationServerId, _ = cmd.Flags().GetString("target-application-server-id")
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
		Aliases: []string{"ls-qr-formats", "listqrformats", "lsqrformats", "lsqrfmts", "lsqrfmt", "qr-formats"},
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
		Example: `
  To generate a QR code for a single end device:
    $ ttn-lw-cli end-devices generate-qr app1 dev1

  To generate a QR code for multiple end devices:
    $ ttn-lw-cli end-devices list app1 \
      | ttn-lw-cli end-devices generate-qr`,
		RunE: asBulk(func(cmd *cobra.Command, args []string) error {
			var ids *ttnpb.EndDeviceIdentifiers
			if inputDecoder != nil {
				var dev ttnpb.EndDevice
				if _, err := inputDecoder.Decode(&dev); err != nil {
					return err
				}
				if dev.GetIds().GetApplicationIds().GetApplicationId() == "" {
					return errNoApplicationID.New()
				}
				if dev.Ids.DeviceId == "" {
					return errNoEndDeviceID.New()
				}
				ids = dev.Ids
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
				FormatId: formatID,
			})
			if err != nil {
				return err
			}

			isPaths, nsPaths, asPaths, jsPaths := splitEndDeviceGetPaths(format.FieldMask.GetPaths()...)

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
				EndDeviceIds: ids,
				FieldMask:    &pbtypes.FieldMask{Paths: isPaths},
			})
			if err != nil {
				return err
			}

			nsMismatch, asMismatch, jsMismatch := compareServerAddressesEndDevice(device, config)
			if len(nsPaths) > 0 && nsMismatch {
				return errAddressMismatchEndDevice.New()
			}
			if len(asPaths) > 0 && asMismatch {
				return errAddressMismatchEndDevice.New()
			}
			if len(jsPaths) > 0 && jsMismatch {
				return errAddressMismatchEndDevice.New()
			}

			dev, err := getEndDevice(device.Ids, nsPaths, asPaths, jsPaths, true)
			if err != nil {
				return err
			}
			device.SetFields(dev, append(append(nsPaths, asPaths...), jsPaths...)...)

			size, _ := cmd.Flags().GetUint32("size")
			res, err := client.Generate(ctx, &ttnpb.GenerateEndDeviceQRCodeRequest{
				FormatId:  formatID,
				EndDevice: device,
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
			filename := path.Join(folder, device.Ids.DeviceId+ext)
			if err := os.WriteFile(filename, res.Image.Embedded.Data, 0o644); err != nil {
				return err
			}

			logger.WithFields(log.Fields(
				"value", res.Text,
				"filename", filename,
			)).Info("Generated QR code")
			return nil
		}),
	}
	endDevicesExternalJSCommand = &cobra.Command{
		Use:     "use-external-join-server [application-id] [device-id]",
		Aliases: []string{"use-external-js", "use-ext-js"},
		Short:   "Disassociate and delete the device from Join Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			if !config.JoinServerEnabled {
				return errJoinServerDisabled.New()
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			dev, err := ttnpb.NewEndDeviceRegistryClient(is).Get(ctx, &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: devID,
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"join_server_address",
					},
				},
			})
			if err != nil {
				return err
			}
			if _, _, nok := compareServerAddressesEndDevice(dev, config); nok {
				return errAddressMismatchEndDevice.New()
			}

			js, err := api.Dial(ctx, config.JoinServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewJsEndDeviceRegistryClient(js).Delete(ctx, devID)
			if err != nil {
				return err
			}

			_, err = ttnpb.NewEndDeviceRegistryClient(is).Update(ctx, &ttnpb.UpdateEndDeviceRequest{
				EndDevice: ttnpb.EndDevice{
					Ids: devID,
				},
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"join_server_address",
					},
				},
			})
			return err
		},
	}
	endDevicesGetDefaultMACSettingsCommand = &cobra.Command{
		Use:               "get-default-mac-settings",
		Short:             "Get Network Server default MAC settings for frequency plan and LoRaWAN version",
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.NetworkServerEnabled {
				return errNetworkServerDisabled.New()
			}

			req := &ttnpb.GetDefaultMACSettingsRequest{}
			if err := util.SetFields(req, getDefaultMACSettingsFlags); err != nil {
				return err
			}
			ns, err := api.Dial(ctx, config.NetworkServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewNsClient(ns).GetDefaultMACSettings(ctx, req)
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, res)
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

	endDevicePictureFlags.String("picture", "", "upload the end device picture from this file")

	endDevicesListFrequencyPlans.Flags().Uint32("base-frequency", 0, "base frequency in MHz for hardware support (433, 470, 868 or 915)")
	endDevicesCommand.AddCommand(endDevicesListFrequencyPlans)
	endDevicesListCommand.Flags().AddFlagSet(applicationIDFlags())
	endDevicesListCommand.Flags().AddFlagSet(selectEndDeviceListFlags)
	endDevicesListCommand.Flags().AddFlagSet(selectAllEndDeviceFlags)
	endDevicesListCommand.Flags().AddFlagSet(paginationFlags())
	endDevicesListCommand.Flags().AddFlagSet(orderFlags())
	endDevicesCommand.AddCommand(endDevicesListCommand)
	endDevicesSearchCommand.Flags().AddFlagSet(applicationIDFlags())
	endDevicesSearchCommand.Flags().AddFlagSet(searchEndDevicesFlags)
	endDevicesSearchCommand.Flags().AddFlagSet(selectApplicationFlags)
	endDevicesSearchCommand.Flags().AddFlagSet(selectAllEndDeviceFlags)
	endDevicesCommand.AddCommand(endDevicesSearchCommand)
	endDevicesGetCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesGetCommand.Flags().AddFlagSet(selectEndDeviceFlags)
	endDevicesGetCommand.Flags().AddFlagSet(selectAllEndDeviceFlags)
	endDevicesCommand.AddCommand(endDevicesGetCommand)
	endDevicesCreateCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesCreateCommand.Flags().AddFlagSet(setEndDeviceFlags)
	endDevicesCreateCommand.Flags().AddFlagSet(attributesFlags())
	endDevicesCreateCommand.Flags().AddFlagSet(payloadFormatterParameterFlags("formatters"))
	endDevicesCreateCommand.Flags().Bool("defaults", true, "configure end device with defaults")
	endDevicesCreateCommand.Flags().Bool("with-root-keys", false, "generate OTAA root keys")
	endDevicesCreateCommand.Flags().Bool("abp", false, "configure end device as ABP")
	endDevicesCreateCommand.Flags().Bool("with-session", false, "generate ABP session DevAddr and keys")
	endDevicesCreateCommand.Flags().Bool("with-claim-authentication-code", false, "generate claim authentication code of 4 bytes")
	endDevicesCreateCommand.Flags().Bool("request-dev-eui", false, "request a new DevEUI")
	endDevicesCreateCommand.Flags().AddFlagSet(endDevicePictureFlags)
	endDevicesCreateCommand.Flags().AddFlagSet(endDeviceLocationFlags)
	endDevicesCommand.AddCommand(endDevicesCreateCommand)
	endDevicesSetCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesSetCommand.Flags().AddFlagSet(setEndDeviceFlags)
	endDevicesSetCommand.Flags().AddFlagSet(attributesFlags())
	endDevicesSetCommand.Flags().AddFlagSet(payloadFormatterParameterFlags("formatters"))
	endDevicesSetCommand.Flags().Bool("touch", false, "set in all registries even if no fields are specified")
	endDevicesSetCommand.Flags().AddFlagSet(endDevicePictureFlags)
	endDevicesSetCommand.Flags().AddFlagSet(endDeviceLocationFlags)
	endDevicesSetCommand.Flags().AddFlagSet(util.UnsetFlagSet())
	endDevicesCommand.AddCommand(endDevicesSetCommand)
	endDevicesProvisionCommand.Flags().AddFlagSet(applicationIDFlags())
	endDevicesProvisionCommand.Flags().AddFlagSet(dataFlags("", ""))
	endDevicesProvisionCommand.Flags().String("provisioner-id", "", "provisioner service")
	endDevicesProvisionCommand.Flags().String("join-eui", "", "(hex)")
	endDevicesProvisionCommand.Flags().String("start-dev-eui", "", "starting DevEUI to provision (hex)")
	endDevicesCommand.AddCommand(endDevicesProvisionCommand)
	endDevicesResetCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesResetCommand.Flags().AddFlagSet(selectEndDeviceFlags)
	endDevicesResetCommand.Flags().AddFlagSet(selectAllEndDeviceFlags)
	endDevicesCommand.AddCommand(endDevicesResetCommand)
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
	endDevicesExternalJSCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDevicesCommand.AddCommand(endDevicesExternalJSCommand)

	endDevicesCommand.AddCommand(applicationsDownlinkCommand)

	endDevicesGetDefaultMACSettingsCommand.Flags().AddFlagSet(getDefaultMACSettingsFlags)
	endDevicesCommand.AddCommand(endDevicesGetDefaultMACSettingsCommand)

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
