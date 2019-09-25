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
	"encoding/binary"
	"fmt"
	stdio "io"
	"os"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var (
	endDeviceTemplateFlattenPaths = []string{"end_device.provisioning_data"}
)

func templateFormatIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("format-id", "", "")
	return flagSet
}

var (
	errNoTemplateFormatID             = errors.DefineInvalidArgument("no_template_format_id", "no template format ID set")
	errEndDeviceMappingNotFound       = errors.DefineNotFound("mapped_end_device_not_found", "end device mapping not found")
	errNoEndDeviceTemplateJoinEUI     = errors.DefineInvalidArgument("no_end_device_template_join_eui", "no end device template JoinEUI set")
	errNoEndDeviceTemplateStartDevEUI = errors.DefineInvalidArgument("no_end_device_template_start_dev_eui", "no end device template start DevEUI set")
)

func getTemplateFormatID(flagSet *pflag.FlagSet, args []string) string {
	var formatID string
	if len(args) > 0 {
		if len(args) > 1 {
			logger.Warn("Multiple IDs found in arguments, considering only the first")
		}
		formatID = args[0]
	} else {
		formatID, _ = flagSet.GetString("format-id")
	}
	if formatID == "" {
		return ""
	}
	return formatID
}

var (
	endDeviceTemplatesCommand = &cobra.Command{
		Use:     "templates",
		Aliases: []string{"template", "tmpl"},
		Short:   "End Device template commands",
	}
	endDeviceTemplatesExtendCommand = &cobra.Command{
		Use:               "extend [flags]",
		Short:             "Extend an end device template (EXPERIMENTAL)",
		PersistentPreRunE: preRun(),
		RunE: asBulk(func(cmd *cobra.Command, args []string) error {
			forwardDeprecatedDeviceFlags(cmd.Flags())
			paths := util.UpdateFieldMask(cmd.Flags(), setEndDeviceFlags, attributesFlags())

			var res ttnpb.EndDeviceTemplate
			if inputDecoder != nil {
				_, err := inputDecoder.Decode(&res)
				if err != nil {
					return err
				}
				paths = append(paths, res.FieldMask.Paths...)
			}

			if mappingKey, _ := cmd.Flags().GetString("mapping-key"); mappingKey != "" {
				res.MappingKey = mappingKey
			}
			if err := util.SetFields(&res.EndDevice, setEndDeviceFlags); err != nil {
				return err
			}
			res.EndDevice.Attributes = mergeAttributes(res.EndDevice.Attributes, cmd.Flags())
			res.FieldMask.Paths = ttnpb.BottomLevelFields(paths)

			return io.Write(os.Stdout, config.OutputFormat, &res)
		}),
	}
	endDeviceTemplatesCreateCommand = &cobra.Command{
		Use:   "create [flags]",
		Short: "Create an end device template from an existing device (EXPERIMENTAL)",
		Long: `Create an end device template from an existing device (EXPERIMENTAL)

By default, this command strips the device's application ID, device ID, JoinEUI,
DevEUI and server addresses to create a generic template. You can include the
end device identifiers by passing the concerning flags: --application-id,
--device-id, --join-eui and --dev-eui.

This command takes end devices from stdin.`,
		PersistentPreRunE: preRun(),
		RunE: asBulk(func(cmd *cobra.Command, args []string) error {
			if inputDecoder == nil {
				return nil
			}

			forwardDeprecatedDeviceFlags(cmd.Flags())
			paths := util.UpdateFieldMask(cmd.Flags(), selectEndDeviceFlags)

			excludePaths := []string{
				"ids.dev_addr",
				"created_at",
				"updated_at",
				"network_server_address",
				"application_server_address",
				"join_server_address",
			}
			if appID, _ := cmd.Flags().GetBool("application-id"); !appID {
				excludePaths = append(excludePaths, "ids.application_ids.application_id")
			}
			if devID, _ := cmd.Flags().GetBool("device-id"); !devID {
				excludePaths = append(excludePaths, "ids.device_id")
			}
			if joinEUI, _ := cmd.Flags().GetBool("join-eui"); !joinEUI {
				excludePaths = append(excludePaths, "ids.join_eui")
			}
			if devEUI, _ := cmd.Flags().GetBool("dev-eui"); !devEUI {
				excludePaths = append(excludePaths, "ids.dev_eui")
			}

			var input ttnpb.EndDevice
			decodedPaths, err := inputDecoder.Decode(&input)
			if err != nil {
				return err
			}
			decodedPaths = ttnpb.FlattenPaths(decodedPaths, endDeviceFlattenPaths)
			decodedPaths = ttnpb.ExcludeFields(decodedPaths, excludePaths...)
			paths = append(paths, decodedPaths...)

			mappingKey, _ := cmd.Flags().GetString("mapping-key")
			res := &ttnpb.EndDeviceTemplate{
				FieldMask: pbtypes.FieldMask{
					Paths: paths,
				},
				MappingKey: mappingKey,
			}
			res.EndDevice.SetFields(&input, paths...)

			return io.Write(os.Stdout, config.OutputFormat, res)
		}),
	}
	endDeviceTemplatesExecuteCommand = &cobra.Command{
		Use:     "execute [flags]",
		Aliases: []string{"exec"},
		Short:   "Execute the template to an end device (EXPERIMENTAL)",
		Long: `Execute the template to an end device (EXPERIMENTAL)

This command takes end device templates from stdin.`,
		PersistentPreRunE: preRun(),
		RunE: asBulk(func(cmd *cobra.Command, args []string) error {
			if inputDecoder == nil {
				return nil
			}

			forwardDeprecatedDeviceFlags(cmd.Flags())

			var input ttnpb.EndDeviceTemplate
			_, err := inputDecoder.Decode(&input)
			if err != nil {
				return err
			}

			var device ttnpb.EndDevice
			device.SetFields(&input.EndDevice, input.FieldMask.Paths...)
			if err := util.SetFields(&device, setEndDeviceFlags); err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, &device)
		}),
	}
	endDeviceTemplatesAssignEUIsCommand = &cobra.Command{
		Use:     "assign-euis [join-eui] [start-dev-eui] [flags]",
		Aliases: []string{"euis", "eui"},
		Short:   "Assign JoinEUI and DevEUIs to end device templates (EXPERIMENTAL)",
		Long: `Assign JoinEUI and DevEUIs to end device templates (EXPERIMENTAL)

Pass --count=N to assign N number of DevEUIs to each input end device template.

This command takes end device templates from stdin.`,
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputDecoder == nil {
				return nil
			}

			forwardDeprecatedDeviceFlags(cmd.Flags())
			joinEUIHex, _ := cmd.Flags().GetString("join-eui")
			startDevEUIHex, _ := cmd.Flags().GetString("start-dev-eui")
			switch len(args) {
			case 0:
			case 1:
				logger.Warn("only single EUI found in arguments, not considering arguments")
			case 2:
				joinEUIHex = args[0]
				startDevEUIHex = args[1]
			default:
				logger.Warn("Multiple EUIs found in arguments, considering the first")
				joinEUIHex = args[0]
				startDevEUIHex = args[1]
			}
			if joinEUIHex == "" {
				return errNoEndDeviceTemplateJoinEUI
			}
			if startDevEUIHex == "" {
				return errNoEndDeviceTemplateStartDevEUI
			}

			var joinEUI types.EUI64
			if err := joinEUI.UnmarshalText([]byte(joinEUIHex)); err != nil {
				return err
			}
			var startDevEUI types.EUI64
			if err := startDevEUI.UnmarshalText([]byte(startDevEUIHex)); err != nil {
				return err
			}
			devEUIInt := binary.BigEndian.Uint64(startDevEUI[:])

			count, _ := cmd.Flags().GetInt("count")
			for {
				var template ttnpb.EndDeviceTemplate
				if _, err := inputDecoder.Decode(&template); err != nil {
					if err == stdio.EOF {
						return nil
					}
					return err
				}

				for i := 0; i < count; i++ {
					res := template

					var devEUI types.EUI64
					binary.BigEndian.PutUint64(devEUI[:], devEUIInt)
					devEUIInt++

					res.EndDevice.DeviceID = fmt.Sprintf("eui-%s", strings.ToLower(devEUI.String()))
					res.EndDevice.JoinEUI = &joinEUI
					res.EndDevice.DevEUI = &devEUI
					res.FieldMask.Paths = ttnpb.BottomLevelFields(append(res.FieldMask.Paths,
						"ids.device_id",
						"ids.join_eui",
						"ids.dev_eui",
					))

					if err := io.Write(os.Stdout, config.OutputFormat, &res); err != nil {
						return err
					}
				}
			}
		},
	}
	endDeviceTemplatesListFormats = &cobra.Command{
		Use:               "list-formats",
		Aliases:           []string{"ls-formats", "listformats", "lsformats"},
		Short:             "List available end device template formats (EXPERIMENTAL)",
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			dtc, err := api.Dial(ctx, config.DeviceTemplateConverterGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewEndDeviceTemplateConverterClient(dtc).ListFormats(ctx, ttnpb.Empty)
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	endDeviceTemplatesFromDataCommand = &cobra.Command{
		Use:               "from-data [format-id]",
		Aliases:           []string{"fromdata"},
		Short:             "Convert data to an end device template (EXPERIMENTAL)",
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			formatID := getTemplateFormatID(cmd.Flags(), args)
			if formatID == "" {
				return errNoTemplateFormatID
			}
			data, err := getDataBytes("", cmd.Flags())
			if err != nil {
				return err
			}

			dtc, err := api.Dial(ctx, config.DeviceTemplateConverterGRPCAddress)
			if err != nil {
				return err
			}
			stream, err := ttnpb.NewEndDeviceTemplateConverterClient(dtc).Convert(ctx, &ttnpb.ConvertEndDeviceTemplateRequest{
				FormatID: formatID,
				Data:     data,
			})
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
	endDeviceTemplatesMapCommand = &cobra.Command{
		Use:   "map [flags]",
		Short: "Map end device templates (EXPERIMENTAL)",
		Long: `Map end device templates (EXPERIMENTAL)

This command matches the input templates with the mapping file to create new
templates. The mapping file contains end device templates in the same format
as input.

The matching from input to a mapping template is, in order, by mapping key, end
device identifiers and DevEUI. If you don't specify a mapping key, end device
identifiers nor DevEUI, the mapping entry always matches. This is useful for
mapping many end device templates with a generic template.

Typical use cases are:

1. Assigning identifiers from a mapping file to device templates matching on
   mapping key.
2. Mapping a device profile (i.e. MAC and PHY versions, frequency plan and class
   B/C support) from a mapping file to many end device templates.

Use the create command to create a mapping file and (optionally) the assign-euis
command to assign EUIs to map to end device templates.`,
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputDecoder := inputDecoder
			if inputDecoder == nil {
				reader, err := getDataReader("input", cmd.Flags())
				if err != nil {
					return err
				}
				inputDecoder, err = getInputDecoder(reader)
				if err != nil {
					return err
				}
			}
			var input []ttnpb.EndDeviceTemplate
			for {
				var entry ttnpb.EndDeviceTemplate
				_, err := inputDecoder.Decode(&entry)
				if err != nil {
					if err == stdio.EOF {
						break
					}
					return err
				}
				input = append(input, entry)
			}

			var mapping []ttnpb.EndDeviceTemplate
			reader, err := getDataReader("mapping", cmd.Flags())
			if err != nil {
				return err
			}
			mappingDecoder, err := getInputDecoder(reader)
			if err != nil {
				return err
			}
			for {
				var entry ttnpb.EndDeviceTemplate
				if _, err := mappingDecoder.Decode(&entry); err != nil {
					if err == stdio.EOF {
						break
					}
					return err
				}
				mapping = append(mapping, entry)
			}

			for _, inputEntry := range input {
				var mappedEntry *ttnpb.EndDeviceTemplate
				for _, e := range mapping {
					switch {
					case e.MappingKey != "" && e.MappingKey == inputEntry.MappingKey:
					case e.EndDevice.ApplicationID != "" && e.EndDevice.ApplicationID == inputEntry.EndDevice.ApplicationID &&
						e.EndDevice.DeviceID != "" && e.EndDevice.DeviceID == inputEntry.EndDevice.DeviceID:
					case e.EndDevice.DevEUI != nil && inputEntry.EndDevice.DevEUI != nil && e.EndDevice.DevEUI.Equal(*inputEntry.EndDevice.DevEUI):
					case e.EndDevice.EndDeviceIdentifiers.IsZero():
					default:
						continue
					}
					mappedEntry = &e
					break
				}
				if mappedEntry == nil {
					if fail, _ := cmd.Flags().GetBool("fail-not-found"); fail {
						return errEndDeviceMappingNotFound
					}
					continue
				}

				var res ttnpb.EndDeviceTemplate
				res.EndDevice.SetFields(&inputEntry.EndDevice, inputEntry.FieldMask.Paths...)
				res.EndDevice.SetFields(&mappedEntry.EndDevice, mappedEntry.FieldMask.Paths...)
				res.FieldMask.Paths = ttnpb.BottomLevelFields(append(inputEntry.FieldMask.Paths, mappedEntry.FieldMask.Paths...))

				if err := io.Write(os.Stdout, config.OutputFormat, &res); err != nil {
					return err
				}
			}
			return nil
		},
	}
)

func init() {
	endDeviceTemplatesExtendCommand.Flags().AddFlagSet(attributesFlags())
	endDeviceTemplatesExtendCommand.Flags().String("mapping-key", "", "")
	endDeviceTemplatesCommand.AddCommand(endDeviceTemplatesExtendCommand)
	endDeviceTemplatesCreateCommand.Flags().AddFlagSet(selectEndDeviceIDFlags())
	endDeviceTemplatesCreateCommand.Flags().String("mapping-key", "", "")
	endDeviceTemplatesCommand.AddCommand(endDeviceTemplatesCreateCommand)
	endDeviceTemplatesCommand.AddCommand(endDeviceTemplatesExecuteCommand)
	endDeviceTemplatesAssignEUIsCommand.Flags().String("join-eui", "", "(hex)")
	endDeviceTemplatesAssignEUIsCommand.Flags().String("start-dev-eui", "", "(hex)")
	endDeviceTemplatesAssignEUIsCommand.Flags().Int("count", 1, "")
	endDeviceTemplatesCommand.AddCommand(endDeviceTemplatesAssignEUIsCommand)
	endDeviceTemplatesCommand.AddCommand(endDeviceTemplatesListFormats)
	endDeviceTemplatesFromDataCommand.Flags().AddFlagSet(templateFormatIDFlags())
	endDeviceTemplatesFromDataCommand.Flags().AddFlagSet(dataFlags("", ""))
	endDeviceTemplatesCommand.AddCommand(endDeviceTemplatesFromDataCommand)
	endDeviceTemplatesMapCommand.Flags().AddFlagSet(dataFlags("input", "input file"))
	endDeviceTemplatesMapCommand.Flags().AddFlagSet(dataFlags("mapping", "mapping file"))
	endDeviceTemplatesMapCommand.Flags().Bool("fail-not-found", false, "fail if no matching mapping is found")
	endDeviceTemplatesCommand.AddCommand(endDeviceTemplatesMapCommand)
	endDevicesCommand.AddCommand(endDeviceTemplatesCommand)

	Root.AddCommand(endDeviceTemplatesCommand)
}
