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

package identityserver

import (
	"context"
	"os"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/email/sendgrid"
	"go.thethings.network/lorawan-stack/v3/pkg/email/smtp"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/oauth"
	telemetry "go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	ttntypes "go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Config for the Identity Server.
type Config struct {
	DatabaseURI      string `name:"database-uri" description:"Database connection URI"`
	UserRegistration struct {
		Enabled    bool `name:"enabled" description:"Enable user registration"`
		Invitation struct {
			Required bool          `name:"required" description:"Require invitations for new users"`
			TokenTTL time.Duration `name:"token-ttl" description:"TTL of user invitation tokens"`
		} `name:"invitation"`
		ContactInfoValidation struct {
			Required      bool          `name:"required" description:"Require contact info validation for new users"`
			TokenTTL      time.Duration `name:"token-ttl" description:"TTL of contact info validation tokens"`
			RetryInterval time.Duration `name:"retry-interval" description:"Minimum interval for resending contact info validation emails"` // nolint:lll
		} `name:"contact-info-validation"`
		AdminApproval struct {
			Required bool `name:"required" description:"Require admin approval for new users"`
		} `name:"admin-approval"`
		PasswordRequirements struct {
			MinLength    int  `name:"min-length" description:"Minimum password length"`
			MaxLength    int  `name:"max-length" description:"Maximum password length"`
			MinUppercase int  `name:"min-uppercase" description:"Minimum number of uppercase letters"`
			MinDigits    int  `name:"min-digits" description:"Minimum number of digits"`
			MinSpecial   int  `name:"min-special" description:"Minimum number of special characters"`
			RejectUserID bool `name:"reject-user-id" description:"Reject passwords that contain user ID"`
			RejectCommon bool `name:"reject-common" description:"Reject common passwords"`
		} `name:"password-requirements"`
	} `name:"user-registration"`
	AuthCache struct {
		MembershipTTL time.Duration `name:"membership-ttl" description:"TTL of membership caches"`
	} `name:"auth-cache"`
	OAuth          oauth.Config `name:"oauth"`
	ProfilePicture struct {
		DisableUpload bool   `name:"disable-upload" description:"Disable uploading profile pictures"`
		UseGravatar   bool   `name:"use-gravatar" description:"Use Gravatar fallback for users without profile picture"`
		Bucket        string `name:"bucket" description:"Bucket used for storing profile pictures"`
		BucketURL     string `name:"bucket-url" description:"Base URL for public bucket access"`
	} `name:"profile-picture"`
	EndDevicePicture struct {
		DisableUpload bool   `name:"disable-upload" description:"Disable uploading end device pictures"`
		Bucket        string `name:"bucket" description:"Bucket used for storing end device pictures"`
		BucketURL     string `name:"bucket-url" description:"Base URL for public bucket access"`
	} `name:"end-device-picture"`
	UserRights struct {
		CreateApplications  bool `name:"create-applications" description:"Allow non-admin users to create applications in their user account"`
		CreateClients       bool `name:"create-clients" description:"Allow non-admin users to create OAuth clients in their user account"`
		CreateGateways      bool `name:"create-gateways" description:"Allow non-admin users to create gateways in their user account"`
		CreateOrganizations bool `name:"create-organizations" description:"Allow non-admin users to create organizations in their user account"`
	} `name:"user-rights"`
	AdminRights struct {
		All bool `name:"all" description:"Grant all rights to admins, including _KEYS and _ALL"`
	} `name:"admin-rights"`
	CollaboratorRights struct {
		SetOthersAsContacts bool `name:"set-others-as-contacts" description:"Allow users to set other users as entity contacts"` // nolint:lll
	} `name:"collaborator-rights"`
	LoginTokens struct {
		Enabled  bool          `name:"enabled" description:"enable users requesting login tokens"`
		TokenTTL time.Duration `name:"token-ttl" description:"TTL of login tokens"`
	} `name:"login-tokens"`
	Email struct {
		email.Config `name:",squash"`
		Provider     string               `name:"provider" description:"Email provider to use"`
		Dir          string               `name:"dir" description:"Directory to write emails to if the dir provider is used (development only)"` // nolint:lll
		SendGrid     sendgrid.Config      `name:"sendgrid"`
		SMTP         smtp.Config          `name:"smtp"`
		Templates    emailTemplatesConfig `name:"templates"`
	} `name:"email"`
	EndDevices struct {
		EncryptionKeyID string `name:"encryption-key-id" description:"ID of the key used to encrypt end device secrets at rest"` //nolint:lll
	} `name:"end-devices"`
	Gateways struct {
		EncryptionKeyID string        `name:"encryption-key-id" description:"ID of the key used to encrypt gateway secrets at rest"`
		TokenValidity   time.Duration `name:"token-validity" description:"Time in seconds after creation when a gateway token is valid"` //nolint:lll
	} `name:"gateways"`
	Delete struct {
		Restore time.Duration `name:"restore" description:"How long after soft-deletion an entity can be restored"`
	} `name:"delete"`
	DevEUIBlock struct {
		Enabled          bool                 `name:"enabled" description:"Enable DevEUI address issuing from IEEE MAC block"`
		ApplicationLimit int                  `name:"application-limit" description:"Maximum DevEUI addresses to be issued per application"`
		Prefix           ttntypes.EUI64Prefix `name:"prefix" description:"DevEUI block prefix"`
		InitCounter      int64                `name:"init-counter" description:"Initial counter value for the addresses to be issued (default 0)"`
	} `name:"dev-eui-block" description:"IEEE MAC block used to issue DevEUIs to devices that are not yet programmed"`
	Network struct {
		NetID    ttntypes.NetID  `name:"net-id" description:"NetID of this network"`
		NSID     *ttntypes.EUI64 `name:"ns-id" description:"NSID of this network (EUI)"`
		TenantID string          `name:"tenant-id" description:"Tenant ID"`
	} `name:"network"`
	TelemetryQueue telemetry.TaskQueue `name:"-"`
	Pagination     struct {
		DefaultLimit uint32 `name:"default-limit" description:"The default limit applied to paginated requests if not specified"` // nolint:lll
	} `name:"pagination" description:"Pagination settings"`
}

type emailTemplatesConfig struct {
	Source    string                `name:"source" description:"Source of the email template files (directory, url, blob)"` // nolint:lll
	Static    map[string][]byte     `name:"-"`
	Directory string                `name:"directory" description:"Retrieve the email templates from the filesystem"`
	URL       string                `name:"url" description:"Retrieve the email templates from a web server"`
	Blob      config.BlobPathConfig `name:"blob"`

	Includes []string `name:"includes" description:"The email templates that will be preloaded on startup"`
}

// Fetcher returns a fetch.Interface based on the configuration.
// If no configuration source is set, this method returns nil, nil.
func (c emailTemplatesConfig) Fetcher(ctx context.Context, blobConf config.BlobConfig, httpClientProvider httpclient.Provider) (fetch.Interface, error) {
	// TODO: Remove detection mechanism (https://github.com/TheThingsNetwork/lorawan-stack/issues/1450)
	if c.Source == "" {
		switch {
		case c.Static != nil:
			c.Source = "static"
		case c.Directory != "":
			if stat, err := os.Stat(c.Directory); err == nil && stat.IsDir() {
				c.Source = "directory"
				break
			}
			fallthrough
		case c.URL != "":
			c.Source = "url"
		case !c.Blob.IsZero():
			c.Source = "blob"
		}
	}
	switch c.Source {
	case "static":
		return fetch.NewMemFetcher(c.Static), nil
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		httpClient, err := httpClientProvider.HTTPClient(ctx, httpclient.WithCache(true))
		if err != nil {
			return nil, err
		}
		return fetch.FromHTTP(httpClient, c.URL)
	case "blob":
		b, err := blobConf.Bucket(ctx, c.Blob.Bucket, httpClientProvider)
		if err != nil {
			return nil, err
		}
		return fetch.FromBucket(ctx, b, c.Blob.Path), nil
	default:
		return nil, nil
	}
}

func (c Config) toProto() *ttnpb.IsConfiguration {
	return &ttnpb.IsConfiguration{
		UserRegistration: &ttnpb.IsConfiguration_UserRegistration{
			Enabled: c.UserRegistration.Enabled,
			Invitation: &ttnpb.IsConfiguration_UserRegistration_Invitation{
				Required: &wrapperspb.BoolValue{Value: c.UserRegistration.Invitation.Required},
				TokenTtl: durationpb.New(c.UserRegistration.Invitation.TokenTTL),
			},
			ContactInfoValidation: &ttnpb.IsConfiguration_UserRegistration_ContactInfoValidation{
				Required:      &wrapperspb.BoolValue{Value: c.UserRegistration.ContactInfoValidation.Required},
				TokenTtl:      durationpb.New(c.UserRegistration.ContactInfoValidation.TokenTTL),
				RetryInterval: durationpb.New(c.UserRegistration.ContactInfoValidation.RetryInterval),
			},
			AdminApproval: &ttnpb.IsConfiguration_UserRegistration_AdminApproval{
				Required: &wrapperspb.BoolValue{Value: c.UserRegistration.AdminApproval.Required},
			},
			PasswordRequirements: &ttnpb.IsConfiguration_UserRegistration_PasswordRequirements{
				MinLength:    &wrapperspb.UInt32Value{Value: uint32(c.UserRegistration.PasswordRequirements.MinLength)},
				MaxLength:    &wrapperspb.UInt32Value{Value: uint32(c.UserRegistration.PasswordRequirements.MaxLength)},
				MinUppercase: &wrapperspb.UInt32Value{Value: uint32(c.UserRegistration.PasswordRequirements.MinUppercase)},
				MinDigits:    &wrapperspb.UInt32Value{Value: uint32(c.UserRegistration.PasswordRequirements.MinDigits)},
				MinSpecial:   &wrapperspb.UInt32Value{Value: uint32(c.UserRegistration.PasswordRequirements.MinSpecial)},
			},
		},
		ProfilePicture: &ttnpb.IsConfiguration_ProfilePicture{
			DisableUpload: &wrapperspb.BoolValue{Value: c.ProfilePicture.DisableUpload},
			UseGravatar:   &wrapperspb.BoolValue{Value: c.ProfilePicture.UseGravatar},
		},
		EndDevicePicture: &ttnpb.IsConfiguration_EndDevicePicture{
			DisableUpload: &wrapperspb.BoolValue{Value: c.ProfilePicture.DisableUpload},
		},
		UserRights: &ttnpb.IsConfiguration_UserRights{
			CreateApplications:  &wrapperspb.BoolValue{Value: c.UserRights.CreateApplications},
			CreateClients:       &wrapperspb.BoolValue{Value: c.UserRights.CreateClients},
			CreateGateways:      &wrapperspb.BoolValue{Value: c.UserRights.CreateGateways},
			CreateOrganizations: &wrapperspb.BoolValue{Value: c.UserRights.CreateOrganizations},
		},
		AdminRights: &ttnpb.IsConfiguration_AdminRights{
			All: &wrapperspb.BoolValue{Value: c.AdminRights.All},
		},
		CollaboratorRights: &ttnpb.IsConfiguration_CollaboratorRights{
			SetOthersAsContacts: &wrapperspb.BoolValue{Value: c.CollaboratorRights.SetOthersAsContacts},
		},
	}
}

// GetConfiguration implements the RPC that returns the configuration of the Identity Server.
func (is *IdentityServer) GetConfiguration(ctx context.Context, _ *ttnpb.GetIsConfigurationRequest) (*ttnpb.GetIsConfigurationResponse, error) {
	return &ttnpb.GetIsConfigurationResponse{
		Configuration: is.configFromContext(ctx).toProto(),
	}, nil
}
