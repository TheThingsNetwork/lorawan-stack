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

// Package tlsconfig provides configuration for TLS clients and servers.
package tlsconfig

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"sync"
	"sync/atomic"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// ACME represents ACME configuration.
type ACME struct {
	manager *autocert.Manager

	// TODO: Remove Enable (https://github.com/TheThingsNetwork/lorawan-stack/issues/1450)
	Enable      bool     `name:"enable" description:"Enable automated certificate management (ACME). This setting is deprecated; set the TLS config source to acme instead"` //nolint:lll
	Endpoint    string   `name:"endpoint" description:"ACME endpoint"`
	Dir         string   `name:"dir" description:"Location of ACME storage directory"`
	Email       string   `name:"email" description:"Email address to register with the ACME account"`
	Hosts       []string `name:"hosts" description:"Hosts to enable automatic certificates for"`
	DefaultHost string   `name:"default-host" description:"Default host to assume for clients without SNI"`
}

var (
	errMissingACMEDir      = errors.Define("missing_acme_dir", "missing ACME storage directory")
	errMissingACMEEndpoint = errors.Define("missing_acme_endpoint", "missing ACME endpoint")
)

// Initialize initializes the autocert manager for the ACME configuration.
// If it was already initialized, any changes after the previous initialization
// are ignored.
func (a *ACME) Initialize() (*autocert.Manager, error) {
	if a.manager != nil {
		return a.manager, nil
	}
	if a.Endpoint == "" {
		return nil, errMissingACMEEndpoint.New()
	}
	if a.Dir == "" {
		return nil, errMissingACMEDir.New()
	}
	a.manager = &autocert.Manager{
		Cache:      autocert.DirCache(a.Dir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(a.Hosts...),
		Client: &acme.Client{
			DirectoryURL: a.Endpoint,
		},
		Email: a.Email,
	}
	return a.manager, nil
}

// IsZero returns whether the ACME configuration is empty.
func (a ACME) IsZero() bool {
	return !a.Enable &&
		a.Endpoint == "" &&
		a.Dir == "" &&
		a.Email == "" &&
		len(a.Hosts) == 0
}

// KeyVault defines configuration for loading a certificate from the key vault.
type KeyVault struct {
	KeyVault interface {
		ExportCertificate(ctx context.Context, id string) (*tls.Certificate, error)
	} `name:"-"`

	ID string `name:"id" description:"ID of the certificate"`
}

// IsZero returns whether the TLS KeyVault is empty.
func (t KeyVault) IsZero() bool {
	return t.ID == ""
}

// Config represents TLS configuration.
type Config struct {
	Client     `name:",squash"`
	ServerAuth `name:",squash"`
}

// FileReader is the interface used to read TLS certificates and keys.
type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

// Client is client-side configuration for server TLS.
type Client struct {
	FileReader         FileReader `json:"-" yaml:"-" name:"-"`
	loadRootCA         sync.Once
	rootCABytes        []byte
	rootCABytesError   error
	RootCA             string `json:"root-ca" yaml:"root-ca" name:"root-ca" description:"Location of TLS root CA certificate (optional)"` //nolint:lll
	InsecureSkipVerify bool   `name:"insecure-skip-verify" description:"Skip verification of certificate chains (insecure)"`              //nolint:lll
}

// Equals checks if the other configuration is equivalent to this.
func (c Client) Equals(other Client) bool {
	return c.RootCA == other.RootCA &&
		c.InsecureSkipVerify == other.InsecureSkipVerify
}

// ApplyTo applies the client configuration options to the given TLS configuration.
// If tlsConfig is nil, this is a no-op.
func (c *Client) ApplyTo(tlsConfig *tls.Config) error {
	if tlsConfig == nil {
		return nil
	}
	c.loadRootCA.Do(func() {
		if c.RootCA != "" {
			readFile := os.ReadFile
			if c.FileReader != nil {
				readFile = c.FileReader.ReadFile
			}
			c.rootCABytes, c.rootCABytesError = readFile(c.RootCA)
		}
	})
	if c.rootCABytesError != nil {
		return c.rootCABytesError
	}

	if len(c.rootCABytes) > 0 {
		var err error
		if tlsConfig.RootCAs == nil {
			if tlsConfig.RootCAs, err = x509.SystemCertPool(); err != nil {
				tlsConfig.RootCAs = x509.NewCertPool()
			}
		}
		tlsConfig.RootCAs.AppendCertsFromPEM(c.rootCABytes)
	}
	tlsConfig.InsecureSkipVerify = c.InsecureSkipVerify
	return nil
}

func readCert(fileReader FileReader, certFile, keyFile string) (*tls.Certificate, error) {
	readFile := os.ReadFile
	if fileReader != nil {
		readFile = fileReader.ReadFile
	}
	certPEM, err := readFile(certFile)
	if err != nil {
		return nil, err
	}
	keyPEM, err := readFile(keyFile)
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// ServerAuth is configuration for TLS server authentication.
type ServerAuth struct {
	Source       string     `name:"source" description:"Source of the TLS certificate (file, acme, key-vault)"`
	FileReader   FileReader `json:"-" yaml:"-" name:"-"`
	Certificate  string     `json:"certificate" yaml:"certificate" name:"certificate" description:"Location of TLS certificate"` //nolint:lll
	Key          string     `json:"key" yaml:"key" name:"key" description:"Location of TLS private key"`
	ACME         ACME       `name:"acme"`
	KeyVault     KeyVault   `name:"key-vault"`
	CipherSuites []string   `name:"cipher-suites" description:"DEPRECATED: List of IANA names of TLS cipher suites to use"` //nolint:lll
}

var (
	errInvalidTLSConfigSource = errors.DefineFailedPrecondition(
		"tls_config_source_invalid", "invalid TLS configuration source `{source}`",
	)
	errEmptyTLSSource = errors.DefineFailedPrecondition(
		"tls_source_empty", "empty TLS source",
	)
	errTLSKeyVaultID = errors.DefineFailedPrecondition(
		"tls_key_vault_id", "invalid TLS key vault ID",
	)
	errInvalidCipherSuite = errors.DefineFailedPrecondition(
		"tls_cipher_suite_invalid", "invalid TLS cipher suite {cipher}",
	)
)

// GetCipherSuites returns a list of IDs of cipher suites in configuration.
// This list can be passed to tls.Config.
func (c *ServerAuth) GetCipherSuites() ([]uint16, error) {
	cs := make(map[string]uint16, len(tls.CipherSuites())+len(tls.InsecureCipherSuites()))
	for _, c := range tls.CipherSuites() {
		cs[c.Name] = c.ID
	}
	for _, c := range tls.InsecureCipherSuites() {
		cs[c.Name] = c.ID
	}
	cipherSuites := make([]uint16, 0, len(c.CipherSuites))
	for _, c := range c.CipherSuites {
		cipher, got := cs[c]
		if !got {
			return nil, errInvalidCipherSuite.WithAttributes("cipher", c)
		}
		cipherSuites = append(cipherSuites, cipher)
	}
	if len(cipherSuites) == 0 {
		return nil, nil
	}
	return cipherSuites, nil
}

// ApplyTo applies the TLS authentication configuration options to the given TLS configuration.
// If tlsConfig is nil, this is a no-op.
func (c *ServerAuth) ApplyTo(tlsConfig *tls.Config) error {
	if tlsConfig == nil {
		return nil
	}
	switch c.Source {
	case "":
		return errEmptyTLSSource.New()
	case "file":
		var atomicCert atomic.Value
		cert, err := readCert(c.FileReader, c.Certificate, c.Key)
		if err != nil {
			return err
		}
		atomicCert.Store(cert)
		// TODO: Reload certificates on signal.
		tlsConfig.GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert := atomicCert.Load().(*tls.Certificate)
			return cert, nil
		}
	case "acme":
		manager, err := c.ACME.Initialize()
		if err != nil {
			return err
		}
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, acme.ALPNProto)
		tlsConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			if hello.ServerName == "" {
				hello.ServerName = c.ACME.DefaultHost
			}
			return manager.GetCertificate(hello)
		}
	case "key-vault":
		if c.KeyVault.ID == "" {
			return errTLSKeyVaultID.New()
		}
		tlsConfig.GetCertificate = func(inf *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return c.KeyVault.KeyVault.ExportCertificate(inf.Context(), c.KeyVault.ID)
		}
	default:
		return errInvalidTLSConfigSource.WithAttributes("source", c.Source)
	}
	return nil
}

// ClientAuth is (client-side) configuration for TLS client authentication.
type ClientAuth struct {
	Source      string     `name:"source" description:"Source of the TLS certificate (file, key-vault)"`
	FileReader  FileReader `json:"-" yaml:"-" name:"-"`
	Certificate string     `json:"certificate" yaml:"certificate" name:"certificate" description:"Location of TLS certificate"` //nolint:lll
	Key         string     `json:"key" yaml:"key" name:"key" description:"Location of TLS private key"`
	KeyVault    KeyVault   `name:"key-vault"`
}

// ApplyTo applies the TLS authentication configuration options to the given TLS configuration.
// If tlsConfig is nil, this is a no-op.
func (c *ClientAuth) ApplyTo(tlsConfig *tls.Config) error {
	if tlsConfig == nil {
		return nil
	}
	switch c.Source {
	case "":
		return errEmptyTLSSource.New()
	case "file":
		var atomicCert atomic.Value
		cert, err := readCert(c.FileReader, c.Certificate, c.Key)
		if err != nil {
			return err
		}
		atomicCert.Store(cert)
		// TODO: Reload certificates on signal.
		tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert := atomicCert.Load().(*tls.Certificate)
			return cert, nil
		}
	case "key-vault":
		if c.KeyVault.ID == "" {
			return errTLSKeyVaultID.New()
		}
		tlsConfig.GetClientCertificate = func(r *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return c.KeyVault.KeyVault.ExportCertificate(r.Context(), c.KeyVault.ID)
		}
	default:
		return errInvalidTLSConfigSource.WithAttributes("source", c.Source)
	}
	return nil
}
