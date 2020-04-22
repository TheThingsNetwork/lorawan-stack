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
	"time"

	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/email/sendgrid"
	"go.thethings.network/lorawan-stack/pkg/email/smtp"
	"go.thethings.network/lorawan-stack/pkg/oauth"
)

// Config for the Identity Server
type Config struct {
	DatabaseURI      string `name:"database-uri" description:"Database connection URI"`
	UserRegistration struct {
		Invitation struct {
			Required bool          `name:"required" description:"Require invitations for new users"`
			TokenTTL time.Duration `name:"token-ttl" description:"TTL of user invitation tokens"`
		} `name:"invitation"`
		ContactInfoValidation struct {
			Required bool `name:"required" description:"Require contact info validation for new users"`
		} `name:"contact-info-validation"`
		AdminApproval struct {
			Required bool `name:"required" description:"Require admin approval for new users"`
		} `name:"admin-approval"`
		PasswordRequirements struct {
			MinLength    int `name:"min-length" description:"Minimum password length"`
			MaxLength    int `name:"max-length" description:"Maximum password length"`
			MinUppercase int `name:"min-uppercase" description:"Minimum number of uppercase letters"`
			MinDigits    int `name:"min-digits" description:"Minimum number of digits"`
			MinSpecial   int `name:"min-special" description:"Minimum number of special characters"`
		} `name:"password-requirements"`
	} `name:"user-registration"`
	AuthCache struct {
		MembershipTTL time.Duration `name:"membership-ttl" description:"TTL of membership caches"`
	} `name:"auth-cache"`
	OAuth          oauth.Config `name:"oauth"`
	ProfilePicture struct {
		UseGravatar bool   `name:"use-gravatar" description:"Use Gravatar fallback for users without profile picture"`
		Bucket      string `name:"bucket" description:"Bucket used for storing profile pictures"`
		BucketURL   string `name:"bucket-url" description:"Base URL for public bucket access"`
	} `name:"profile-picture"`
	EndDevicePicture struct {
		Bucket    string `name:"bucket" description:"Bucket used for storing end device pictures"`
		BucketURL string `name:"bucket-url" description:"Base URL for public bucket access"`
	} `name:"end-device-picture"`
	Email struct {
		email.Config `name:",squash"`
		SendGrid     sendgrid.Config      `name:"sendgrid"`
		SMTP         smtp.Config          `name:"smtp"`
		Templates    emailTemplatesConfig `name:"templates"`
	} `name:"email"`
}
