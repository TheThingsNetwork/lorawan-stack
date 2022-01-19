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

package gatewayconfigurationserver

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

type antennaLocation struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Altitude  int32   `json:"altitude,omitempty"`
}

type oauth2Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   uint32 `json:"expires_in"`
}

type attributes struct {
	Description string `json:"description"`
}

type router struct {
	ID          string `json:"id,omitempty"`
	MQTTAddress string `json:"mqtt_address"`
}

type gatewayInfoResponse struct {
	ID               string           `json:"id,omitempty"`
	Attributes       *attributes      `json:"attributes,omitempty"`
	FrequencyPlan    string           `json:"frequency_plan"`
	FrequencyPlanURL string           `json:"frequency_plan_url"`
	AutoUpdate       bool             `json:"auto_update"`
	FirmwareURL      string           `json:"firmware_url,omitempty"`
	AntennaLocation  *antennaLocation `json:"antenna_location,omitempty"`
	Token            *oauth2Token     `json:"token,omitempty"`
	Router           *router          `json:"router,omitempty"`
	FallbackRouters  []*router        `json:"fallback_routers,omitempty"`
}

var errUnauthenticated = errors.DefineUnauthenticated("unauthenticated", "unauthenticated")

func (s *Server) handleGetGateway(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := gatewayIDFromContext(ctx)

	registry, err := s.getRegistry(ctx, &id)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	gateway, err := registry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIds: &id,
		FieldMask: &types.FieldMask{Paths: []string{
			"description",
			"attributes",
			"frequency_plan_id",
			"auto_update",
			"update_channel",
			"antennas",
			"gateway_server_address",
		}},
	}, s.getAuth(ctx))
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	md := rpcmetadata.FromIncomingContext(ctx)
	switch md.AuthType {
	case "bearer":
		if err := rights.RequireGateway(ctx, id, ttnpb.Right_RIGHT_GATEWAY_INFO); err != nil {
			webhandlers.Error(w, r, err)
			return
		}
	case "key":
		keyAttr, ok := gateway.Attributes["key"]
		if !ok || md.AuthValue != keyAttr {
			webhandlers.Error(w, r, errUnauthenticated.New())
			return
		}
	default:
		gateway = gateway.PublicSafe()
	}

	freqPlanURL := s.inferServerAddress(r) + "/api/v2/frequency-plans/" + gateway.FrequencyPlanId

	var rtr *router
	if gateway.GatewayServerAddress != "" {
		rtr = &router{
			ID: gateway.GatewayServerAddress,
		}
		rtr.MQTTAddress, err = s.inferMQTTAddress(gateway.GatewayServerAddress)
		if err != nil {
			webhandlers.Error(w, r, err)
			return
		}
	}

	res := &gatewayInfoResponse{
		FrequencyPlan:    gateway.FrequencyPlanId,
		FrequencyPlanURL: freqPlanURL,
		AutoUpdate:       gateway.AutoUpdate,
	}

	switch r.Header.Get("User-Agent") {
	case "TTNGateway": // The Things Kickstarter Gateway
		res.FirmwareURL = s.ttkgFirmwareURL(gateway.UpdateChannel)
		if rtr != nil {
			res.Router = &router{
				MQTTAddress: rtr.MQTTAddress,
			}
		}
	default:
		res.ID = gateway.GetIds().GetGatewayId()
		if rtr != nil {
			res.Router = rtr
			res.FallbackRouters = []*router{rtr}
		}
		if gateway.Description != "" {
			res.Attributes = &attributes{
				Description: gateway.Description,
			}
		}
		if len(gateway.Antennas) > 0 {
			loc := gateway.Antennas[0].Location
			if loc != nil {
				res.AntennaLocation = &antennaLocation{
					Latitude:  loc.Latitude,
					Longitude: loc.Longitude,
					Altitude:  loc.Altitude,
				}
			}
		}
		if token, ok := gateway.Attributes["token"]; ok {
			res.Token = &oauth2Token{
				AccessToken: token,
				ExpiresIn:   uint32((24 * time.Hour).Seconds()),
			}
			if tokenExpires, ok := gateway.Attributes["token_expires"]; ok {
				if t, err := time.Parse(time.RFC3339, tokenExpires); err == nil {
					res.Token.ExpiresIn = uint32(time.Until(t).Seconds())
				}
			}
		}
	}

	webhandlers.JSON(w, r, res)
}

func (s *Server) ttkgFirmwareURL(updateChannel string) string {
	if updateChannel == "" {
		updateChannel = s.ttgConfig.Default.UpdateChannel
	}
	if updateChannel == "" {
		updateChannel = "stable"
	}
	if strings.HasPrefix(updateChannel, "http") {
		return updateChannel
	}
	firmwareBaseURL := s.ttgConfig.Default.FirmwareURL
	if firmwareBaseURL == "" {
		firmwareBaseURL = "https://thethingsproducts.blob.core.windows.net/the-things-gateway/v1"
	}
	return fmt.Sprintf("%s/%s", firmwareBaseURL, updateChannel)
}

func (s *Server) inferMQTTAddress(address string) (result string, err error) {
	if address == "" {
		return s.ttgConfig.Default.MQTTServer, nil
	}
	if strings.Contains(address, "://") {
		return address, nil
	}
	var host, port string
	if strings.Contains(address, ":") {
		host, port, err = net.SplitHostPort(address)
		if err != nil {
			return "", err
		}
	} else {
		host = address
	}
	if port == "" {
		port = "8881"
	}
	switch port {
	case "1881", "1882", "1883":
		return fmt.Sprintf("mqtt://%s:%s", host, port), nil
	case "8881", "8882", "8883":
		return fmt.Sprintf("mqtts://%s:%s", host, port), nil
	}
	return fmt.Sprintf("%s:%s", host, port), nil
}

func (s *Server) inferServerAddress(r *http.Request) string {
	scheme := r.URL.Scheme
	if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	}
	hostport := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		hostport = forwardedHost
	}
	if host, port, err := net.SplitHostPort(hostport); err == nil {
		switch port {
		case "80":
			scheme = "http"
			hostport = host
		case "1885":
			scheme = "http"
		case "443":
			scheme = "https"
			hostport = host
		case "8885":
			scheme = "https"
		default:
			scheme = "https"
		}
	}
	return fmt.Sprintf("%s://%s", scheme, hostport)
}
