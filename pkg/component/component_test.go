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

package component_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/log/handler/memory"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	certPem string
	keyPem  string
)

func init() {
	certPem = "cert.pem"
	if _, err := os.Stat(certPem); err != nil {
		certPem = "../../cert.pem"
	}
	keyPem = "key.pem"
	if _, err := os.Stat(keyPem); err != nil {
		keyPem = "../../key.pem"
	}
	for _, filepath := range []string{certPem, keyPem} {
		if _, err := os.Stat(filepath); err != nil {
			panic(fmt.Sprintf("could not retrieve information about the %s file - if you haven't generated it, generate it with `./mage dev:certificates`.", filepath))
		}
	}
}

func TestNew(t *testing.T) {
	a := assertions.New(t)

	logger := test.GetLogger(t)

	_, err := component.New(logger, &component.Config{})
	a.So(err, should.BeNil)

	mustNew := func() { component.MustNew(logger, &component.Config{}) }
	a.So(mustNew, should.NotPanic)
}

func TestLogger(t *testing.T) {
	a := assertions.New(t)

	mem := memory.New()

	logger, err := log.NewLogger(log.WithHandler(mem))
	a.So(err, should.BeNil)

	// Component logger
	{
		c, err := component.New(logger, &component.Config{})
		a.So(err, should.BeNil)

		nbEntries := len(mem.Entries)
		c.Logger().Info("Hello world")
		a.So(mem.Entries, should.HaveLength, nbEntries+1)
	}
}

type registererFunc func(s *web.Server)

func (r registererFunc) RegisterRoutes(s *web.Server) {
	r(s)
}

func TestHTTP(t *testing.T) {
	a := assertions.New(t)

	httpAddress, httpsAddress := "0.0.0.0:9185", "0.0.0.0:9186"
	baseConfig := component.Config{}

	workingRoutePath := "/ok"
	workingRoute := registererFunc(func(s *web.Server) {
		s.GET(workingRoutePath, func(c echo.Context) error {
			c.JSON(http.StatusOK, "OK")
			return nil
		})
	})

	// HTTP
	{
		config := baseConfig
		config.HTTP.Listen = httpAddress
		config.HTTP.ListenTLS = ""

		c, err := component.New(test.GetLogger(t), &config)
		a.So(err, should.BeNil)
		c.RegisterWeb(workingRoute)

		err = c.Start()
		a.So(err, should.BeNil)

		{
			// Non-registered path
			resp, err := http.Get(fmt.Sprintf("http://%s/not found", httpAddress))
			a.So(err, should.BeNil)
			a.So(resp.StatusCode, should.Equal, http.StatusNotFound)

			// Registered path
			resp, err = http.Get(fmt.Sprintf("http://%s%s", httpAddress, workingRoutePath))
			a.So(err, should.BeNil)
			a.So(resp.StatusCode, should.Equal, http.StatusOK)
		}

		c.Close()
	}

	// Invalid HTTP port
	{
		config := baseConfig
		config.HTTP.Listen = "0.0.0.0:12391483"

		c, err := component.New(test.GetLogger(t), &config)
		a.So(err, should.BeNil)

		err = c.Start()
		a.So(err, should.NotBeNil)
	}

	// HTTPS
	{
		config := baseConfig

		config.HTTP.Listen = ""
		config.HTTP.ListenTLS = httpsAddress
		config.TLS.Certificate = certPem
		config.TLS.Key = keyPem

		c, err := component.New(test.GetLogger(t), &config)
		a.So(err, should.BeNil)
		c.RegisterWeb(workingRoute)

		err = c.Start()
		a.So(err, should.BeNil)

		certPool := x509.NewCertPool()
		certContent, err := ioutil.ReadFile(config.TLS.Certificate)
		a.So(err, should.BeNil)
		certPool.AppendCertsFromPEM(certContent)
		client := http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: certPool}},
		}

		{
			// Non-registered path
			resp, err := client.Get("https://localhost:9186/not found")
			a.So(err, should.BeNil)
			a.So(resp.StatusCode, should.Equal, http.StatusNotFound)

			// Registered path
			resp, err = client.Get(fmt.Sprintf("https://localhost:9186%s", workingRoutePath))
			a.So(err, should.BeNil)
			a.So(resp.StatusCode, should.Equal, http.StatusOK)
		}

		c.Close()
	}

	// Invalid HTTPS port
	{
		config := baseConfig
		config.HTTP.ListenTLS = "0.0.0.0:394823525"

		c, err := component.New(test.GetLogger(t), &config)
		a.So(err, should.BeNil)

		err = c.Start()
		a.So(err, should.NotBeNil)
	}
}

func TestGRPC(t *testing.T) {
	a := assertions.New(t)

	baseConfig := component.Config{
		ServiceBase: config.ServiceBase{GRPC: config.GRPC{}},
	}

	// gRPC without TLS
	{
		grpcPort := 9199
		config := baseConfig
		config.ServiceBase.GRPC.Listen = fmt.Sprintf("0.0.0.0:%d", grpcPort)

		c, err := component.New(test.GetLogger(t), &config)
		a.So(err, should.BeNil)

		err = c.Start()
		a.So(err, should.BeNil)

		client, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcPort),
			grpc.WithInsecure(),
			grpc.WithTimeout(time.Second*3),
			grpc.WithBlock())
		a.So(err, should.BeNil)
		client.Close()

		c.Close()
	}

	// gRPC with TLS
	{
		grpcPort := 9197

		config := baseConfig
		config.ServiceBase.GRPC.ListenTLS = fmt.Sprintf("0.0.0.0:%d", grpcPort)
		config.TLS.Certificate = certPem
		config.TLS.Key = keyPem

		c, err := component.New(test.GetLogger(t), &config)
		a.So(err, should.BeNil)

		err = c.Start()
		a.So(err, should.BeNil)

		tlsCredentials, err := credentials.NewClientTLSFromFile(config.TLS.Certificate, "")
		a.So(err, should.BeNil)

		client, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcPort),
			grpc.WithTimeout(time.Second*3),
			grpc.WithTransportCredentials(tlsCredentials))
		a.So(err, should.BeNil)
		client.Close()

		c.Close()
	}
}

func TestContext(t *testing.T) {
	a := assertions.New(t)

	config := component.Config{}

	c, err := component.New(test.GetLogger(t), &config)
	a.So(err, should.BeNil)

	baseCtx := context.Background()

	// Filler
	{
		c.AddContextFiller(func(ctx context.Context) context.Context {
			ctx = context.WithValue(ctx, "k1", "v1")
			ctx = context.WithValue(ctx, "k2", "v2")
			return ctx
		})
		ctx := c.FillContext(baseCtx)
		a.So(ctx, should.HaveParentContext, baseCtx)
		v1 := ctx.Value("k1")
		a.So(v1, should.Equal, "v1")
		v2 := ctx.Value("k2")
		a.So(v2, should.Equal, "v2")
	}
}
