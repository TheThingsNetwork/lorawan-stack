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

func templateFormatIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("format-id", "", "")
	return flagSet
}

var errNoTemplateFormatID = errors.DefineInvalidArgument("no_template_format_id", "no template format ID set")

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
			data, err := getData(cmd.Flags())
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
)

func init() {
	endDeviceTemplatesCommand.AddCommand(endDeviceTemplatesListFormats)
	endDeviceTemplatesFromDataCommand.Flags().AddFlagSet(templateFormatIDFlags())
	endDeviceTemplatesFromDataCommand.Flags().AddFlagSet(dataFlags("", ""))
	endDeviceTemplatesCommand.AddCommand(endDeviceTemplatesFromDataCommand)
	endDevicesCommand.AddCommand(endDeviceTemplatesCommand)

	Root.AddCommand(endDeviceTemplatesCommand)
}
