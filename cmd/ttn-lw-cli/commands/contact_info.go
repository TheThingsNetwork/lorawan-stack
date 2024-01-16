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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func listContactInfo(entityID *ttnpb.EntityIdentifiers) ([]*ttnpb.ContactInfo, error) {
	is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
	if err != nil {
		return nil, err
	}
	fieldMask := ttnpb.FieldMask("contact_info")
	var res any
	switch id := entityID.GetIds().(type) {
	case *ttnpb.EntityIdentifiers_ApplicationIds:
		res, err = ttnpb.NewApplicationRegistryClient(is).Get(ctx, &ttnpb.GetApplicationRequest{ApplicationIds: id.ApplicationIds, FieldMask: fieldMask})
	case *ttnpb.EntityIdentifiers_ClientIds:
		res, err = ttnpb.NewClientRegistryClient(is).Get(ctx, &ttnpb.GetClientRequest{ClientIds: id.ClientIds, FieldMask: fieldMask})
	case *ttnpb.EntityIdentifiers_GatewayIds:
		res, err = ttnpb.NewGatewayRegistryClient(is).Get(ctx, &ttnpb.GetGatewayRequest{GatewayIds: id.GatewayIds, FieldMask: fieldMask})
	case *ttnpb.EntityIdentifiers_OrganizationIds:
		res, err = ttnpb.NewOrganizationRegistryClient(is).Get(ctx, &ttnpb.GetOrganizationRequest{OrganizationIds: id.OrganizationIds, FieldMask: fieldMask})
	case *ttnpb.EntityIdentifiers_UserIds:
		res, err = ttnpb.NewUserRegistryClient(is).Get(ctx, &ttnpb.GetUserRequest{UserIds: id.UserIds, FieldMask: fieldMask})
	default:
		panic(fmt.Errorf("no contact info in %T", id))
	}
	if err != nil {
		return nil, err
	}
	return res.(interface{ GetContactInfo() []*ttnpb.ContactInfo }).GetContactInfo(), nil
}

func updateContactInfo(entityID *ttnpb.EntityIdentifiers, updater func([]*ttnpb.ContactInfo) ([]*ttnpb.ContactInfo, error)) ([]*ttnpb.ContactInfo, error) {
	is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
	if err != nil {
		return nil, err
	}
	fieldMask := ttnpb.FieldMask("contact_info")
	var res any
	switch id := entityID.GetIds().(type) {
	case *ttnpb.EntityIdentifiers_ApplicationIds:
		res, err = ttnpb.NewApplicationRegistryClient(is).Get(ctx, &ttnpb.GetApplicationRequest{ApplicationIds: id.ApplicationIds, FieldMask: fieldMask})
	case *ttnpb.EntityIdentifiers_ClientIds:
		res, err = ttnpb.NewClientRegistryClient(is).Get(ctx, &ttnpb.GetClientRequest{ClientIds: id.ClientIds, FieldMask: fieldMask})
	case *ttnpb.EntityIdentifiers_GatewayIds:
		res, err = ttnpb.NewGatewayRegistryClient(is).Get(ctx, &ttnpb.GetGatewayRequest{GatewayIds: id.GatewayIds, FieldMask: fieldMask})
	case *ttnpb.EntityIdentifiers_OrganizationIds:
		res, err = ttnpb.NewOrganizationRegistryClient(is).Get(ctx, &ttnpb.GetOrganizationRequest{OrganizationIds: id.OrganizationIds, FieldMask: fieldMask})
	case *ttnpb.EntityIdentifiers_UserIds:
		res, err = ttnpb.NewUserRegistryClient(is).Get(ctx, &ttnpb.GetUserRequest{UserIds: id.UserIds, FieldMask: fieldMask})
	default:
		panic(fmt.Errorf("no contact info in %T", id))
	}
	if err != nil {
		return nil, err
	}

	var contactInfoer interface {
		GetContactInfo() []*ttnpb.ContactInfo
	}
	switch res := res.(type) {
	case *ttnpb.Application:
		res.ContactInfo, err = updater(res.ContactInfo)
		if err != nil {
			return nil, err
		}
		contactInfoer, err = ttnpb.NewApplicationRegistryClient(is).Update(ctx, &ttnpb.UpdateApplicationRequest{
			Application: res,
			FieldMask:   fieldMask,
		})
	case *ttnpb.Client:
		res.ContactInfo, err = updater(res.ContactInfo)
		if err != nil {
			return nil, err
		}
		contactInfoer, err = ttnpb.NewClientRegistryClient(is).Update(ctx, &ttnpb.UpdateClientRequest{
			Client:    res,
			FieldMask: fieldMask,
		})
	case *ttnpb.Gateway:
		res.ContactInfo, err = updater(res.ContactInfo)
		if err != nil {
			return nil, err
		}
		contactInfoer, err = ttnpb.NewGatewayRegistryClient(is).Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway:   res,
			FieldMask: fieldMask,
		})
	case *ttnpb.Organization:
		res.ContactInfo, err = updater(res.ContactInfo)
		if err != nil {
			return nil, err
		}
		contactInfoer, err = ttnpb.NewOrganizationRegistryClient(is).Update(ctx, &ttnpb.UpdateOrganizationRequest{
			Organization: res,
			FieldMask:    fieldMask,
		})
	case *ttnpb.User:
		res.ContactInfo, err = updater(res.ContactInfo)
		if err != nil {
			return nil, err
		}
		contactInfoer, err = ttnpb.NewUserRegistryClient(is).Update(ctx, &ttnpb.UpdateUserRequest{
			User:      res,
			FieldMask: fieldMask,
		})
	}
	if err != nil {
		return nil, err
	}
	return contactInfoer.GetContactInfo(), nil
}

var (
	errContactInfoExists           = errors.DefineAlreadyExists("contact_info_exists", "contact info already exists")
	errMatchingContactInfoNotFound = errors.DefineAlreadyExists("contact_info_not_found", "matching contact info not found")
)

func contactInfoCommands(entity string, getID func(cmd *cobra.Command, args []string) (*ttnpb.EntityIdentifiers, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contact-info",
		Short: fmt.Sprintf("Manage %[1]s contact info (DEPRECATED. Instead, use administrative_contact and technical_contact fields of %[1]s)", entity),
	}
	list := &cobra.Command{
		Use:     fmt.Sprintf("list [%s-id]", entity),
		Aliases: []string{"ls", "get"},
		Short:   fmt.Sprintf("List %[1]s contact info (DEPRECATED. Instead, select administrative_contact and technical_contact fields of %[1]s)", entity),
		Hidden:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getID(cmd, args)
			if err != nil {
				return err
			}
			contactInfo, err := listContactInfo(id)
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, contactInfo)
		},
	}
	add := &cobra.Command{
		Use:     fmt.Sprintf("create [%s-id]", entity),
		Aliases: []string{"add", "register"},
		Short:   fmt.Sprintf("Add %[1]s contact info (DEPRECATED. Instead, set administrative_contact and technical_contact fields of %[1]s)", entity),
		Hidden:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getID(cmd, args)
			if err != nil {
				return err
			}
			var contactInfo ttnpb.ContactInfo
			_, err = contactInfo.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			updatedInfo, err := updateContactInfo(id, func(existing []*ttnpb.ContactInfo) ([]*ttnpb.ContactInfo, error) {
				for _, existing := range existing {
					if existing.ContactMethod == contactInfo.ContactMethod && existing.ContactType == contactInfo.ContactType && existing.Value == contactInfo.Value {
						return nil, errContactInfoExists.New()
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
		Use:     fmt.Sprintf("delete [%s-id]", entity),
		Aliases: []string{"del", "remove", "rm"},
		Short:   fmt.Sprintf("Remove %[1]s contact info (DEPRECATED. Instead, unset administrative_contact and technical_contact fields of %[1]s)", entity),
		Hidden:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getID(cmd, args)
			if err != nil {
				return err
			}
			var contactInfo ttnpb.ContactInfo
			_, err = contactInfo.SetFromFlags(cmd.Flags(), "")
			if err != nil {
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
					return nil, errMatchingContactInfoNotFound.New()
				}
				return updatedInfo, nil
			})
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, updatedInfo)
		},
	}
	requestValidation := &cobra.Command{
		Use:    fmt.Sprintf("request-validation [%s-id]", entity),
		Short:  "Request validation for entity contact info (DEPRECATED. Use `user email-validation request` instead.",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getID(cmd, args)
			if err != nil {
				return err
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewContactInfoRegistryClient(is).RequestValidation(ctx, id)
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	validate := &cobra.Command{
		Use:    "validate [reference] [token]",
		Short:  "Validate contact info (DEPRECATED. Use `user email-validation validate` instead.",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			reference, _ := cmd.Flags().GetString("reference")
			token, _ := cmd.Flags().GetString("token")
			switch len(args) {
			case 1:
				reference = args[0]
			case 2:
				reference = args[0]
				token = args[1]
			default:
			}
			if reference == "" {
				return errNoValidationReference.New()
			}
			if token == "" {
				return errNoValidationToken.New()
			}
			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewContactInfoRegistryClient(is).Validate(ctx, &ttnpb.ContactInfoValidation{
				Id:    reference,
				Token: token,
			})
			if err != nil {
				return err
			}
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	ttnpb.AddSetFlagsForContactInfo(add.Flags(), "", false)
	cmd.AddCommand(add)
	ttnpb.AddSetFlagsForContactInfo(list.Flags(), "", false)
	cmd.AddCommand(list)
	ttnpb.AddSetFlagsForContactInfo(remove.Flags(), "", false)
	cmd.AddCommand(remove)
	ttnpb.AddSetFlagsForContactInfo(requestValidation.Flags(), "", false)
	cmd.AddCommand(requestValidation)
	validate.Flags().String("reference", "", "Reference of the requested validation")
	validate.Flags().String("token", "", "Token that you received")
	cmd.AddCommand(validate)
	return cmd
}
