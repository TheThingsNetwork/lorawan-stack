// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/errors/httperrors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/log/handler/memory"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/TheThingsNetwork/ttn/pkg/web"
	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	pemDir = filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "TheThingsNetwork", "ttn")

	certPem = filepath.Join(pemDir, "cert.pem")
	keyPem  = filepath.Join(pemDir, "key.pem")
)

func init() {
	for _, filepath := range []string{certPem, keyPem} {
		if _, err := os.Stat(filepath); err != nil {
			panic(fmt.Sprintf("Could not retrieve information about the %s file - if you haven't generated it, generate it with `make dev-cert`.", filepath))
		}
	}
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
	baseConfig := component.Config{
		ServiceBase: config.ServiceBase{HTTP: config.HTTP{PProf: true}},
	}

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
			a.So(httperrors.FromHTTP(resp), should.NotBeNil)

			// Registered path
			resp, err = http.Get(fmt.Sprintf("http://%s%s", httpAddress, workingRoutePath))
			a.So(err, should.BeNil)
			a.So(httperrors.FromHTTP(resp), should.BeNil)
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
			a.So(httperrors.FromHTTP(resp), should.NotBeNil)

			// Registered path
			resp, err = client.Get(fmt.Sprintf("https://localhost:9186%s", workingRoutePath))
			a.So(err, should.BeNil)
			a.So(httperrors.FromHTTP(resp), should.BeNil)
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
