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
	"strconv"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	selectApplicationPackageAssociationsFlags = util.FieldMaskFlags(&ttnpb.ApplicationPackageAssociation{})
	setApplicationPackageAssociationsFlags    = util.FieldFlags(&ttnpb.ApplicationPackageAssociation{})
)

func applicationPackageAssociationIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("application-id", "", "")
	flagSet.String("device-id", "", "")
	flagSet.Uint8("f-port", 1, "")
	return flagSet
}

var errNoFPort = errors.DefineInvalidArgument("no_f_port", "no FPort set")

func getApplicationPackageAssociationID(flagSet *pflag.FlagSet, args []string) (*ttnpb.ApplicationPackageAssociationIdentifiers, error) {
	applicationID, _ := flagSet.GetString("application-id")
	deviceID, _ := flagSet.GetString("device-id")
	fport, _ := flagSet.GetUint8("f-port")
	switch len(args) {
	case 0:
	case 1:
	case 2:
		logger.Warn("Only single ID found in arguments, not considering arguments")
	case 3:
		applicationID = args[0]
		deviceID = args[1]
		fport64, err := strconv.ParseUint(args[2], 10, 8)
		if err != nil {
			return nil, err
		}
		fport = uint8(fport64)
	default:
		logger.Warn("multiple IDs found in arguments, considering the first")
		applicationID = args[0]
		deviceID = args[1]
		fport64, err := strconv.ParseUint(args[2], 10, 8)
		if err != nil {
			return nil, err
		}
		fport = uint8(fport64)
	}
	if applicationID == "" {
		return nil, errNoApplicationID
	}
	if deviceID == "" {
		return nil, errNoEndDeviceID
	}
	if fport == 0 {
		return nil, errNoFPort
	}
	return &ttnpb.ApplicationPackageAssociationIdentifiers{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: applicationID},
			DeviceID:               deviceID,
		},
		FPort: uint32(fport),
	}, nil
}

var (
	applicationsPackagesCommand = &cobra.Command{
		Use:     "packages",
		Aliases: []string{"package", "pkg", "pkgs"},
		Short:   "Application packages commands",
	}
	applicationsPackagesListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the available application packages for the device",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationPackageRegistryClient(as).List(ctx, devID)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsPackagesAssociationsCommand = &cobra.Command{
		Use:     "associations",
		Aliases: []string{"assoc", "assocs"},
		Short:   "Application packages associations commands",
	}
	applicationsPackageAssociationGetCommand = &cobra.Command{
		Use:     "get [application-id] [device-id] [f-port]",
		Aliases: []string{"info-association"},
		Short:   "Get the properties of an application package association",
		RunE: func(cmd *cobra.Command, args []string) error {
			assocID, err := getApplicationPackageAssociationID(cmd.Flags(), args)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationPackageAssociationsFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectApplicationPackageAssociationsFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, strings.Replace(flag.Name, "-", "_", -1))
				})
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationPackageRegistryClient(as).GetAssociation(ctx, &ttnpb.GetApplicationPackageAssociationRequest{
				ApplicationPackageAssociationIdentifiers: *assocID,
				FieldMask:                                types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsPackageAssociationsListCommand = &cobra.Command{
		Use:     "list [application-id] [device-id]",
		Aliases: []string{"ls"},
		Short:   "List application package associations",
		RunE: func(cmd *cobra.Command, args []string) error {
			devID, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationPackageAssociationsFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectApplicationPackageAssociationsFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, strings.Replace(flag.Name, "-", "_", -1))
				})
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewApplicationPackageRegistryClient(as).ListAssociations(ctx, &ttnpb.ListApplicationPackageAssociationRequest{
				EndDeviceIdentifiers: *devID,
				Limit:                limit,
				Page:                 page,
				FieldMask:            types.FieldMask{Paths: paths},
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsPackageAssociationSetCommand = &cobra.Command{
		Use:     "set [application-id] [device-id] [f-port]",
		Aliases: []string{"update"},
		Short:   "Set the properties of an application package association",
		RunE: func(cmd *cobra.Command, args []string) error {
			assocID, err := getApplicationPackageAssociationID(cmd.Flags(), args)
			if err != nil {
				return err
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setApplicationPackageAssociationsFlags)

			var association ttnpb.ApplicationPackageAssociation
			if err = util.SetFields(&association, setApplicationPackageAssociationsFlags); err != nil {
				return err
			}
			association.ApplicationPackageAssociationIdentifiers = *assocID

			reader, err := getDataReader("data", cmd.Flags())
			if err != nil {
				logger.WithError(err).Warn("Package data not available")
			} else {
				var st types.Struct
				err := jsonpb.TTN().NewDecoder(reader).Decode(&st)
				if err != nil {
					return err
				}

				association.Data = &st
				paths = append(paths, "data")
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationPackageRegistryClient(as).SetAssociation(ctx, &ttnpb.SetApplicationPackageAssociationRequest{
				ApplicationPackageAssociation: association,
				FieldMask:                     types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsPackageAssociationDeleteCommand = &cobra.Command{
		Use:     "delete [application-id] [device-id] [f-port]",
		Aliases: []string{"del", "rm"},
		Short:   "Delete an application package association",
		RunE: func(cmd *cobra.Command, args []string) error {
			assocID, err := getApplicationPackageAssociationID(cmd.Flags(), args)
			if err != nil {
				return err
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationPackageRegistryClient(as).DeleteAssociation(ctx, assocID)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	applicationsPackagesCommand.AddCommand(applicationsPackagesListCommand)
	applicationsPackageAssociationGetCommand.Flags().AddFlagSet(applicationPackageAssociationIDFlags())
	applicationsPackageAssociationGetCommand.Flags().AddFlagSet(selectApplicationPackageAssociationsFlags)
	applicationsPackagesCommand.AddCommand(applicationsPackagesAssociationsCommand)
	applicationsPackagesAssociationsCommand.AddCommand(applicationsPackageAssociationGetCommand)
	applicationsPackageAssociationsListCommand.Flags().AddFlagSet(endDeviceIDFlags())
	applicationsPackageAssociationsListCommand.Flags().AddFlagSet(selectApplicationPackageAssociationsFlags)
	applicationsPackageAssociationsListCommand.Flags().AddFlagSet(paginationFlags())
	applicationsPackagesAssociationsCommand.AddCommand(applicationsPackageAssociationsListCommand)
	applicationsPackageAssociationSetCommand.Flags().AddFlagSet(applicationPackageAssociationIDFlags())
	applicationsPackageAssociationSetCommand.Flags().AddFlagSet(setApplicationPackageAssociationsFlags)
	applicationsPackageAssociationSetCommand.Flags().AddFlagSet(dataFlags("data", "package data"))
	applicationsPackagesAssociationsCommand.AddCommand(applicationsPackageAssociationSetCommand)
	applicationsPackageAssociationDeleteCommand.Flags().AddFlagSet(applicationPackageAssociationIDFlags())
	applicationsPackagesAssociationsCommand.AddCommand(applicationsPackageAssociationDeleteCommand)
	applicationsCommand.AddCommand(applicationsPackagesCommand)
}
