// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package templates

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var writeGolden = flag.Bool("write-golden", false, "Write golden files")

var testTemplateData = email.NewTemplateData(
	&email.NetworkConfig{
		Name:              "The Things Network",
		IdentityServerURL: "https://eu1.cloud.thethings.network/oauth",
		ConsoleURL:        "https://console.cloud.thethings.network",
		AssetsBaseURL:     "https://assets.cloud.thethings.network",
		BrandingBaseURL:   "https://assets.cloud.thethings.network/branding",
	}, &ttnpb.User{
		Ids: &ttnpb.UserIdentifiers{
			UserId: "john-doe",
		},
		Name:                "John Doe",
		PrimaryEmailAddress: "john.doe@example.com",
		Admin:               true,
	},
)

func TestEmailTemplates(t *testing.T) {
	t.Parallel()
	defer func() {
		if t.Failed() {
			t.Log("NOTE: If you encounter a diff, you may have to run this test with the -write-golden flag.")
		}
	}()

	usrIDs := &ttnpb.UserIdentifiers{
		UserId: "foo-usr",
	}

	for _, tc := range []struct {
		TemplateName ttnpb.NotificationType
		TemplateData email.TemplateData
	}{
		{
			TemplateName: ttnpb.NotificationType_INVITATION,
			TemplateData: &InvitationData{
				TemplateData:    testTemplateData,
				SenderIds:       usrIDs,
				InvitationToken: "TOKEN",
				TTL:             time.Hour,
			},
		},

		{
			TemplateName: ttnpb.NotificationType_LOGIN_TOKEN,
			TemplateData: &LoginTokenData{
				TemplateData: testTemplateData,
				LoginToken:   "TOKEN",
				TTL:          time.Hour,
			},
		},

		{
			TemplateName: ttnpb.NotificationType_TEMPORARY_PASSWORD,
			TemplateData: &TemporaryPasswordData{
				TemplateData:      testTemplateData,
				TemporaryPassword: "TEMPORARY",
				TTL:               time.Hour,
			},
		},

		{
			TemplateName: ttnpb.NotificationType_VALIDATE,
			TemplateData: &ValidateData{
				TemplateData:      testTemplateData,
				EntityIdentifiers: usrIDs.GetEntityIdentifiers(),
				ID:                "ID",
				Token:             "TOKEN",
				TTL:               time.Hour,
			},
		},
	} {
		tc := tc
		t.Run(tc.TemplateName.String(), func(t *testing.T) {
			t.Parallel()
			a, ctx := test.New(t)

			emailTemplate := email.GetTemplate(ctx, tc.TemplateName)
			message, err := emailTemplate.Execute(tc.TemplateData)
			if a.So(err, should.BeNil) && a.So(message, should.NotBeNil) {
				if err = compareMessageToGolden(message); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestNotificationEmailTemplates(t *testing.T) {
	t.Parallel()
	defer func() {
		if t.Failed() {
			t.Log("NOTE: If you encounter a diff, you may have to run this test with the -write-golden flag.")
		}
	}()

	appIDs := &ttnpb.ApplicationIdentifiers{
		ApplicationId: "foo-app",
	}
	cliIDs := &ttnpb.ClientIdentifiers{
		ClientId: "foo-cli",
	}
	usrIDs := &ttnpb.UserIdentifiers{
		UserId: "foo-usr",
	}
	now := timestamppb.Now()

	for _, notification := range []*ttnpb.Notification{
		{
			EntityIds:        appIDs.GetEntityIdentifiers(),
			NotificationType: ttnpb.NotificationType_API_KEY_CHANGED,
			Data: ttnpb.MustMarshalAny(&ttnpb.APIKey{
				Id:   "TEST",
				Name: "API Key Name",
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_INFO,
					ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
				},
				CreatedAt: now,
				UpdatedAt: now,
			}),
			SenderIds: usrIDs,
			Receivers: []ttnpb.NotificationReceiver{
				ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
			},
		},

		{
			Id:               "with_sender_ids",
			EntityIds:        appIDs.GetEntityIdentifiers(),
			NotificationType: ttnpb.NotificationType_API_KEY_CREATED,
			Data: ttnpb.MustMarshalAny(&ttnpb.APIKey{
				Id:   "TEST",
				Name: "API Key Name",
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_INFO,
					ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
				},
				CreatedAt: now,
				UpdatedAt: now,
			}),
			SenderIds: usrIDs,
			Receivers: []ttnpb.NotificationReceiver{
				ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
			},
		},

		{
			Id:               "with_api_key_name",
			EntityIds:        cliIDs.GetEntityIdentifiers(),
			NotificationType: ttnpb.NotificationType_CLIENT_REQUESTED,
			Data: ttnpb.MustMarshalAny(&ttnpb.CreateClientEmailMessage{
				ApiKey: &ttnpb.APIKey{
					Name: "My API key Name",
				},
				CreateClientRequest: &ttnpb.CreateClientRequest{
					Client: &ttnpb.Client{
						Ids:                   cliIDs,
						CreatedAt:             now,
						UpdatedAt:             now,
						Name:                  "Foo Client",
						Description:           "Foo Client Description",
						AdministrativeContact: usrIDs.GetOrganizationOrUserIdentifiers(),
						TechnicalContact:      usrIDs.GetOrganizationOrUserIdentifiers(),
						RedirectUris:          []string{"https://example.com/oauth/callback"},
						LogoutRedirectUris:    []string{"https://example.com/logout/success"},
						State:                 ttnpb.State_STATE_REQUESTED,
						Grants:                []ttnpb.GrantType{ttnpb.GrantType_GRANT_AUTHORIZATION_CODE},
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_USER_INFO,
							ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST,
							ttnpb.Right_RIGHT_USER_ORGANIZATIONS_LIST,
							ttnpb.Right_RIGHT_ORGANIZATION_INFO,
							ttnpb.Right_RIGHT_ORGANIZATION_APPLICATIONS_LIST,
							ttnpb.Right_RIGHT_APPLICATION_INFO,
							ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						},
					},
					Collaborator: usrIDs.GetOrganizationOrUserIdentifiers(),
				},
			}),
			SenderIds: nil,
		},

		{
			Id:               "with_api_key_id",
			EntityIds:        cliIDs.GetEntityIdentifiers(),
			NotificationType: ttnpb.NotificationType_CLIENT_REQUESTED,
			Data: ttnpb.MustMarshalAny(&ttnpb.CreateClientEmailMessage{
				ApiKey: &ttnpb.APIKey{
					Id: "My API key ID",
				},
				CreateClientRequest: &ttnpb.CreateClientRequest{
					Client: &ttnpb.Client{
						Ids:                   cliIDs,
						CreatedAt:             now,
						UpdatedAt:             now,
						Name:                  "Foo Client",
						Description:           "Foo Client Description",
						AdministrativeContact: usrIDs.GetOrganizationOrUserIdentifiers(),
						TechnicalContact:      usrIDs.GetOrganizationOrUserIdentifiers(),
						RedirectUris:          []string{"https://example.com/oauth/callback"},
						LogoutRedirectUris:    []string{"https://example.com/logout/success"},
						State:                 ttnpb.State_STATE_REQUESTED,
						Grants:                []ttnpb.GrantType{ttnpb.GrantType_GRANT_AUTHORIZATION_CODE},
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_USER_INFO,
							ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST,
							ttnpb.Right_RIGHT_USER_ORGANIZATIONS_LIST,
							ttnpb.Right_RIGHT_ORGANIZATION_INFO,
							ttnpb.Right_RIGHT_ORGANIZATION_APPLICATIONS_LIST,
							ttnpb.Right_RIGHT_APPLICATION_INFO,
							ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						},
					},
					Collaborator: usrIDs.GetOrganizationOrUserIdentifiers(),
				},
			}),
			SenderIds: nil,
		},
		{
			EntityIds:        cliIDs.GetEntityIdentifiers(),
			NotificationType: ttnpb.NotificationType_CLIENT_REQUESTED,
			Data: ttnpb.MustMarshalAny(&ttnpb.CreateClientEmailMessage{
				CreateClientRequest: &ttnpb.CreateClientRequest{
					Client: &ttnpb.Client{
						Ids:                   cliIDs,
						CreatedAt:             now,
						UpdatedAt:             now,
						Name:                  "Foo Client",
						Description:           "Foo Client Description",
						AdministrativeContact: usrIDs.GetOrganizationOrUserIdentifiers(),
						TechnicalContact:      usrIDs.GetOrganizationOrUserIdentifiers(),
						RedirectUris:          []string{"https://example.com/oauth/callback"},
						LogoutRedirectUris:    []string{"https://example.com/logout/success"},
						State:                 ttnpb.State_STATE_REQUESTED,
						Grants:                []ttnpb.GrantType{ttnpb.GrantType_GRANT_AUTHORIZATION_CODE},
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_USER_INFO,
							ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST,
							ttnpb.Right_RIGHT_USER_ORGANIZATIONS_LIST,
							ttnpb.Right_RIGHT_ORGANIZATION_INFO,
							ttnpb.Right_RIGHT_ORGANIZATION_APPLICATIONS_LIST,
							ttnpb.Right_RIGHT_APPLICATION_INFO,
							ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						},
					},
					Collaborator: usrIDs.GetOrganizationOrUserIdentifiers(),
				},
			}),
			SenderIds: usrIDs,
		},

		{
			EntityIds:        appIDs.GetEntityIdentifiers(),
			NotificationType: ttnpb.NotificationType_COLLABORATOR_CHANGED,
			Data: ttnpb.MustMarshalAny(&ttnpb.Collaborator{
				Ids: usrIDs.GetOrganizationOrUserIdentifiers(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_INFO,
					ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
				},
			}),
			SenderIds: usrIDs,
		},

		{
			EntityIds:        cliIDs.GetEntityIdentifiers(),
			NotificationType: ttnpb.NotificationType_ENTITY_STATE_CHANGED,
			Data: ttnpb.MustMarshalAny(&ttnpb.EntityStateChangedNotification{
				State:            ttnpb.State_STATE_FLAGGED,
				StateDescription: "This OAuch client has been sending a large number of invalid requests.",
			}),
			SenderIds: usrIDs,
		},

		{
			EntityIds:        usrIDs.GetEntityIdentifiers(),
			NotificationType: ttnpb.NotificationType_PASSWORD_CHANGED,
		},

		{
			EntityIds:        usrIDs.GetEntityIdentifiers(),
			NotificationType: ttnpb.NotificationType_USER_REQUESTED,
			Data: ttnpb.MustMarshalAny(&ttnpb.CreateUserRequest{
				User: &ttnpb.User{
					Ids:                 usrIDs,
					CreatedAt:           now,
					UpdatedAt:           now,
					Name:                "Foo User",
					Description:         "Foo User Description",
					PrimaryEmailAddress: "foo@example.com",
					State:               ttnpb.State_STATE_REQUESTED,
				},
			}),
		},
	} {
		notification := ttnpb.Clone(notification)
		notification.CreatedAt = now
		notification.Email = true
		t.Run(notification.NotificationType.String(), func(t *testing.T) {
			t.Parallel()
			a, ctx := test.New(t)

			emailNotification := email.GetNotification(ctx, notification.GetNotificationType())
			emailTemplate := email.GetTemplate(ctx, emailNotification.EmailTemplateName)
			templateData, err := emailNotification.DataBuilder(
				ctx,
				email.NewNotificationTemplateData(testTemplateData, notification),
			)
			a.So(err, should.BeNil)
			message, err := emailTemplate.Execute(templateData)
			if a.So(err, should.BeNil) && a.So(message, should.NotBeNil) {
				if err = compareMessageToGolden(message, appendSuffix(notification.Id)); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

type pathTransformation func(string) string

func appendSuffix(suffix string) pathTransformation {
	return func(base string) string {
		if suffix == "" {
			return base
		}
		return fmt.Sprintf("%s.%s", base, suffix)
	}
}

func compareMessageToGolden(message *email.Message, ops ...pathTransformation) error {
	goldenFile := func(part, ext string) string {
		for _, op := range ops {
			part = op(part)
		}
		return filepath.Join("testdata", fmt.Sprintf("%s.%s.golden.%s", message.TemplateName, part, ext))
	}

	if *writeGolden {
		if err := os.WriteFile(goldenFile("subject", "txt"), []byte(message.Subject), 0o644); err != nil {
			return fmt.Errorf("failed to write subject golden file: %w", err)
		}
		if err := os.WriteFile(goldenFile("body", "html"), []byte(message.HTMLBody), 0o644); err != nil {
			return fmt.Errorf("failed to write HTML body golden file: %w", err)
		}
		if err := os.WriteFile(goldenFile("body", "txt"), []byte(message.TextBody), 0o644); err != nil {
			return fmt.Errorf("failed to write text body golden file: %w", err)
		}
		return nil
	}

	expectedSubject, err := os.ReadFile(goldenFile("subject", "txt"))
	if err != nil {
		return fmt.Errorf("failed to read subject golden file: %w", err)
	}
	if diff := cmp.Diff(message.Subject, string(expectedSubject)); diff != "" {
		return fmt.Errorf("unexpected diff in Subject: %s", diff)
	}

	expectedHTMLBody, err := os.ReadFile(goldenFile("body", "html"))
	if err != nil {
		return fmt.Errorf("failed to read HTML body golden file: %w", err)
	}
	if diff := cmp.Diff(message.HTMLBody, string(expectedHTMLBody)); diff != "" {
		return fmt.Errorf("unexpected diff in HTMLBody: %s", diff)
	}

	expectedTextBody, err := os.ReadFile(goldenFile("body", "txt"))
	if err != nil {
		return fmt.Errorf("failed to read text body golden file: %w", err)
	}
	if diff := cmp.Diff(message.TextBody, string(expectedTextBody)); diff != "" {
		return fmt.Errorf("unexpected diff in TextBody: %s", diff)
	}

	return nil
}
