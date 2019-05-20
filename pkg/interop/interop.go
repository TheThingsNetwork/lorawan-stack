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

package interop

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"

	echo "github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/web/middleware"
)

const (
	headerKey  = "header"
	messageKey = "message"
)

// Registerer allows components to register their interop services to the web server.
type Registerer interface {
	RegisterInterop(s *Server)
}

// JoinServer represents a Join Server.
type JoinServer interface {
	JoinRequest(req *JoinReq) (*JoinAns, error)
}

// HomeNetworkServer represents a Home Network Server.
type HomeNetworkServer interface {
}

// ServingNetworkServer represents a Serving Network Server.
type ServingNetworkServer interface {
}

// ForwardingNetworkServer represents a Forwarding Network Server.
type ForwardingNetworkServer interface {
}

// ApplicationServer represents an Application Server.
type ApplicationServer interface {
}

type noopServer struct{}

func (noopServer) JoinRequest(req *JoinReq) (*JoinAns, error) {
	return nil, errNotRegistered
}

// Server is the server.
type Server struct {
	SenderClientCAs map[string][]*x509.Certificate

	rootGroup *echo.Group
	server    *echo.Echo
	config    config.Interop

	js  JoinServer
	hNS HomeNetworkServer
	sNS ServingNetworkServer
	fNS ForwardingNetworkServer
	as  ApplicationServer
}

// New builds a new server.
func New(ctx context.Context, config config.Interop) (*Server, error) {
	logger := log.FromContext(ctx).WithField("namespace", "interop")

	server := echo.New()

	server.Logger = web.NewNoopLogger()
	server.HTTPErrorHandler = ErrorHandler

	server.Use(
		middleware.ID("interop"),
		echomiddleware.BodyLimit("16M"),
		middleware.Recover(),
	)

	senderClientCAs := make(map[string][]*x509.Certificate)
	for senderID, filename := range config.SenderClientCAs {
		pemCerts, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		for len(pemCerts) > 0 {
			var block *pem.Block
			block, pemCerts = pem.Decode(pemCerts)
			if block == nil {
				break
			}
			if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
				continue
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}
			senderClientCAs[senderID] = append(senderClientCAs[senderID], cert)
		}
	}

	noop := &noopServer{}
	s := &Server{
		SenderClientCAs: senderClientCAs,
		rootGroup: server.Group(
			"",
			middleware.Log(logger),
			middleware.Normalize(middleware.RedirectPermanent),
			parseMessage(),
			verifySenderID(senderClientCAs),
		),
		config: config,
		server: server,
		js:     noop,
		hNS:    noop,
		sNS:    noop,
		fNS:    noop,
		as:     noop,
	}

	// In 1.0, NS, JS and AS receive messages on the root path.
	// In 1.1, only JS and AS receive messages on the root path. Since NS can play various roles (hNS, sNS and fNS), their
	// group is created on registration of the handler.
	s.rootGroup.POST("/", s.handleRequest)

	return s, nil
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}

// RegisterJS registers the Join Server for AS-JS, hNS-JS and vNS-JS messages.
func (s *Server) RegisterJS(js JoinServer) {
	s.js = js
}

// RegisterHNS registers the Home Network Server for AS-hNS, JS-hNS and sNS-hNS messages.
func (s *Server) RegisterHNS(hNS HomeNetworkServer) {
	s.hNS = hNS
	s.rootGroup.POST("/hns", s.handleNsRequest)
}

// RegisterSNS registers the Serving Network Server for hNS-sNS, fNS-sNS and JS-vNS messages.
func (s *Server) RegisterSNS(sNS ServingNetworkServer) {
	s.sNS = sNS
	s.rootGroup.POST("/sns", s.handleNsRequest)
}

// RegisterFNS registers the Forwarding Network Server for sNS-fNS and JS-vNS messages.
func (s *Server) RegisterFNS(fNS ForwardingNetworkServer) {
	s.fNS = fNS
	s.rootGroup.POST("/fns", s.handleNsRequest)
}

// RegisterAS registers the Application Server for JS-AS messages.
func (s *Server) RegisterAS(as ApplicationServer) {
	s.as = as
}

func (s *Server) handleRequest(c echo.Context) error {
	var ans interface{}
	var err error
	switch req := c.Get(messageKey).(type) {
	case *JoinReq:
		ans, err = s.js.JoinRequest(req)
	default:
		panic(fmt.Sprintf("unexpected message type %T", c.Get(messageKey)))
	}
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, ans)
}

func (s *Server) handleNsRequest(c echo.Context) error {
	// TODO: Implement LoRaWAN roaming (https://github.com/TheThingsNetwork/lorawan-stack/issues/230)
	c.NoContent(http.StatusNotFound)
	return nil
}
