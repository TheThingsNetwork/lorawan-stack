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
	"fmt"
	"os"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func updateContactInfo(entityID *ttnpb.EntityIdentifiers, updater func([]*ttnpb.ContactInfo) ([]*ttnpb.ContactInfo, error)) ([]*ttnpb.ContactInfo, error) {
	is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
	if err != nil {
		return nil, err
	}
	fieldMask := types.FieldMask{Paths: []string{"contact_info"}}
	var res interface{}
	switch id := entityID.Identifiers().(type) {
	case *ttnpb.ApplicationIdentifiers:
		res, err = ttnpb.NewApplicationRegistryClient(is).Get(ctx, &ttnpb.GetApplicationRequest{ApplicationIdentifiers: *id, FieldMask: fieldMask})
	case *ttnpb.ClientIdentifiers:
		res, err = ttnpb.NewClientRegistryClient(is).Get(ctx, &ttnpb.GetClientRequest{ClientIdentifiers: *id, FieldMask: fieldMask})
	case *ttnpb.GatewayIdentifiers:
		res, err = ttnpb.NewGatewayRegistryClient(is).Get(ctx, &ttnpb.GetGatewayRequest{GatewayIdentifiers: *id, FieldMask: fieldMask})
	case *ttnpb.OrganizationIdentifiers:
		res, err = ttnpb.NewOrganizationRegistryClient(is).Get(ctx, &ttnpb.GetOrganizationRequest{OrganizationIdentifiers: *id, FieldMask: fieldMask})
	case *ttnpb.UserIdentifiers:
		res, err = ttnpb.NewUserRegistryClient(is).Get(ctx, &ttnpb.GetUserRequest{UserIdentifiers: *id, FieldMask: fieldMask})
	default:
		panic(fmt.Errorf("no contact info in %T", id))
	}
	if err != nil {
		return nil, err
	}
	switch res := res.(type) {
	case *ttnpb.Application:
		res.ContactInfo, err = updater(res.ContactInfo)
		if err != nil {
			return nil, err
		}
		res, err = ttnpb.NewApplicationRegistryClient(is).Update(ctx, &ttnpb.UpdateApplicationRequest{
			Application: *res,
			FieldMask:   fieldMask,
		})
	case *ttnpb.Client:
		res.ContactInfo, err = updater(res.ContactInfo)
		if err != nil {
			return nil, err
		}
		res, err = ttnpb.NewClientRegistryClient(is).Update(ctx, &ttnpb.UpdateClientRequest{
			Client:    *res,
			FieldMask: fieldMask,
		})
	case *ttnpb.Gateway:
		res.ContactInfo, err = updater(res.ContactInfo)
		if err != nil {
			return nil, err
		}
		res, err = ttnpb.NewGatewayRegistryClient(is).Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway:   *res,
			FieldMask: fieldMask,
		})
	case *ttnpb.Organization:
		res.ContactInfo, err = updater(res.ContactInfo)
		if err != nil {
			return nil, err
		}
		res, err = ttnpb.NewOrganizationRegistryClient(is).Update(ctx, &ttnpb.UpdateOrganizationRequest{
			Organization: *res,
			FieldMask:    fieldMask,
		})
	case *ttnpb.User:
		res.ContactInfo, err = updater(res.ContactInfo)
		if err != nil {
			return nil, err
		}
		res, err = ttnpb.NewUserRegistryClient(is).Update(ctx, &ttnpb.UpdateUserRequest{
			User:      *res,
			FieldMask: fieldMask,
		})
	}
	if err != nil {
		return nil, err
	}
	return res.(interface{ GetContactInfo() []*ttnpb.ContactInfo }).GetContactInfo(), nil
}

var contactInfoFlags = util.FieldFlags(&ttnpb.ContactInfo{})

var (
	errContactInfoExists           = errors.DefineAlreadyExists("contact_info_exists", "contact info already exists")
	errMatchingContactInfoNotFound = errors.DefineAlreadyExists("contact_info_not_found", "matching contact info not found")
)

func contactInfoCommands(entity string, getID func(cmd *cobra.Command, args []string) (*ttnpb.EntityIdentifiers, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contact-info",
		Short: fmt.Sprintf("Manage %s contact info", entity),
	}
	add := &cobra.Command{
		Use: fmt.Sprintf("add [%s-id]", entity),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getID(cmd, args)
			if err != nil {
				return err
			}
			var contactInfo ttnpb.ContactInfo
			if err = util.SetFields(&contactInfo, contactInfoFlags); err != nil {
				return err
			}
			updatedInfo, err := updateContactInfo(id, func(existing []*ttnpb.ContactInfo) ([]*ttnpb.ContactInfo, error) {
				for _, existing := range existing {
					if existing.ContactMethod == contactInfo.ContactMethod && existing.ContactType == contactInfo.ContactType && existing.Value == contactInfo.Value {
						return nil, errContactInfoExists
					}
				}
				return append(existing, &contactInfo), nil
			})
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, updatedInfo)
		},
	}
	remove := &cobra.Command{
		Use: fmt.Sprintf("remove [%s-id]", entity),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getID(cmd, args)
			if err != nil {
				return err
			}
			var contactInfo ttnpb.ContactInfo
			if err = util.SetFields(&contactInfo, contactInfoFlags); err != nil {
				return err
			}
			updatedInfo, err := updateContactInfo(id, func(existing []*ttnpb.ContactInfo) ([]*ttnpb.ContactInfo, error) {
				var updatedInfo []*ttnpb.ContactInfo
				var found bool
				for _, existing := range existing {
					if existing.ContactMethod != contactInfo.ContactMethod || existing.ContactType != contactInfo.ContactType || existing.Value != contactInfo.Value {
						updatedInfo = append(updatedInfo, existing)
					} else {
						found = true
					}
				}
				if !found {
					return nil, errMatchingContactInfoNotFound
				}
				return updatedInfo, nil
			})
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, updatedInfo)
		},
	}
	add.Flags().AddFlagSet(contactInfoFlags)
	remove.Flags().AddFlagSet(contactInfoFlags)
	cmd.AddCommand(add, remove)
	return cmd
}
