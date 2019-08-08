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

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"sync/atomic"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/events/fs"
	"go.thethings.network/lorawan-stack/pkg/log"
)

// ACME represents ACME configuration.
type ACME struct {
	Enable      bool     `name:"enable" description:"Enable automated certificate management (ACME)"`
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

// TLS represents TLS configuration.
type TLS struct {
	RootCA      string `name:"root-ca" description:"Location of TLS root CA certificate (optional)"`
	Certificate string `name:"certificate" description:"Location of TLS certificate"`
	Key         string `name:"key" description:"Location of TLS private key"`
	ACME        ACME   `name:"acme"`
}

// IsZero returns whether the TLS configuration is empty.
func (t TLS) IsZero() bool {
	return t.RootCA == "" &&
		t.Certificate == "" &&
		t.Key == "" &&
		t.ACME.IsZero()
}

var errNoKeyPair = errors.DefineFailedPrecondition("no_key_pair", "no TLS key pair")

// Config loads the key pair and returns the server TLS configuration.
// Config watches the certificate file and reloads the key pair on changes.
// NOTE: The configuration returned by Config cannot be used for client connections.
func (t TLS) Config(ctx context.Context) (*tls.Config, error) {
	logger := log.FromContext(ctx)
	if t.Certificate == "" || t.Key == "" {
		return nil, errNoKeyPair
	}
	var cv atomic.Value
	loadCertificate := func() error {
		cert, err := tls.LoadX509KeyPair(t.Certificate, t.Key)
		if err != nil {
			return err
		}
		cv.Store(&cert)
		logger.Debug("Loaded TLS certificate")
		return nil
	}
	if err := loadCertificate(); err != nil {
		return nil, err
	}
	var rootCAs *x509.CertPool
	if t.RootCA != "" {
		pem, err := ioutil.ReadFile(t.RootCA)
		if err != nil {
			return nil, err
		}
		rootCAs = x509.NewCertPool()
		rootCAs.AppendCertsFromPEM(pem)
	}

	debounce := make(chan struct{}, 1)
	fs.Watch(t.Certificate, events.HandlerFunc(func(evt events.Event) {
		if evt.Name() != "fs.write" {
			return
		}
		// We have to debounce this; OpenSSL typically causes a lot of write events.
		select {
		case debounce <- struct{}{}:
			time.AfterFunc(5*time.Second, func() {
				if err := loadCertificate(); err != nil {
					logger.WithError(err).Error("Could not reload TLS certificate")
					return
				}
				<-debounce
			})
		default:
		}
	}))

	return &tls.Config{
		RootCAs: rootCAs,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return cv.Load().(*tls.Certificate), nil
		},
	}, nil
}
