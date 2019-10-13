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

package config

// ACME represents ACME configuration.
type ACME struct {
	// TODO: Remove Enable (https://github.com/TheThingsNetwork/lorawan-stack/issues/1450)
	Enable      bool     `name:"enable" description:"Enable automated certificate management (ACME). This setting is deprecated; set the TLS config source to acme instead"`
	Endpoint    string   `name:"endpoint" description:"ACME endpoint"`
	Dir         string   `name:"dir" description:"Location of ACME storage directory"`
	Email       string   `name:"email" description:"Email address to register with the ACME account"`
	Hosts       []string `name:"hosts" description:"Hosts to enable automatic certificates for"`
	DefaultHost string   `name:"default-host" description:"Default host to assume for clients without SNI"`
}

// IsZero returns whether the ACME configuration is empty.
func (a ACME) IsZero() bool {
	return !a.Enable &&
		a.Endpoint == "" &&
		a.Dir == "" &&
		a.Email == "" &&
		len(a.Hosts) == 0
}

// TLSKeyVault defines configuration for loading a certificate from the key vault.
type TLSKeyVault struct {
	ID string `name:"id" description:"ID of the certificate"`
}

// IsZero returns whether the TLSKeyVault is empty.
func (t TLSKeyVault) IsZero() bool {
	return t.ID == ""
}

// TLS represents TLS configuration.
type TLS struct {
	RootCA             string `name:"root-ca" description:"Location of TLS root CA certificate (optional)"`
	InsecureSkipVerify bool   `name:"insecure-skip-verify" description:"Skip verification of certificate chains (insecure)"`

	Source string `name:"source" description:"Source of the TLS certificate (file, acme, key-vault)"`

	Certificate string `name:"certificate" description:"Location of TLS certificate"`
	Key         string `name:"key" description:"Location of TLS private key"`

	ACME ACME `name:"acme"`

	KeyVault TLSKeyVault `name:"key-vault"`
}

// IsZero returns whether the TLS configuration is empty.
func (t TLS) IsZero() bool {
	return t.Source == "" &&
		t.RootCA == "" &&
		t.Certificate == "" &&
		t.Key == "" &&
		t.ACME.IsZero() &&
		t.KeyVault.IsZero()
}
