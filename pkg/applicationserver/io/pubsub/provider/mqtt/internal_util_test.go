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

package mqtt

import (
	"context"
	"crypto/tls"

	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/server"
	"go.thethings.network/lorawan-stack/pkg/log"
)

func startMQTTServer(ctx context.Context, tlsConfig *tls.Config) (mqttnet.Listener, mqttnet.Listener, error) {
	logger := log.FromContext(ctx)
	s := server.New(ctx)

	lis, err := mqttnet.Listen("tcp", ":0")
	if err != nil {
		return nil, nil, err
	}
	logger.Infof("Listening on %v", lis.Addr())
	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				logger.WithError(err).Error("Could not accept connection")
				return
			}
			go s.Handle(conn)
		}
	}()

	if tlsConfig != nil {
		tlsTCPLis, err := tls.Listen("tcp", ":0", tlsConfig)
		if err != nil {
			lis.Close()
			return nil, nil, err
		}
		tlsLis := mqttnet.NewListener(tlsTCPLis, "tls")
		logger.Infof("Listening on TLS %v", tlsLis.Addr())
		go func() {
			for {
				conn, err := tlsLis.Accept()
				if err != nil {
					logger.WithError(err).Error("Could not accept connection")
					return
				}
				go s.Handle(conn)
			}
		}()
		return lis, tlsLis, nil
	}

	return lis, nil, nil
}
