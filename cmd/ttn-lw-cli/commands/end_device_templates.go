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
	stdio "io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
	errNoTemplateFormatID       = errors.DefineInvalidArgument("no_template_format_id", "no template format ID set")
	errEndDeviceMappingNotFound = errors.DefineNotFound("mapped_end_device_not_found", "end device mapping not found")
)

func getTemplateFormatID(flagSet *pflag.FlagSet, args []string) string {
	var formatID string
	if len(args) > 0 {
		if len(args) > 1 {
			logger.Warn("multiple IDs found in arguments, considering only the first")
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
	endDeviceTemplatesListFormats = &cobra.Command{
		Use:               "list-formats",
		Short:             "List available end device template formats",
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
		Short:             "Convert data to an end device template",
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
		Short: "Map end device templates",
		Long: `Map end device templates

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
			data, err := getData("mapping", cmd.Flags())
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
