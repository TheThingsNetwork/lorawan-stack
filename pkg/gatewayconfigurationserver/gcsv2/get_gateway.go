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

package gcsv2

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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

var errUnauthenticated = errors.DefineUnauthenticated("unauthenticated", "call was not authenticated")

func (s *Server) handleGetGateway(c echo.Context) error {
	ctx := c.Request().Context()

	gatewayIDs := ttnpb.GatewayIdentifiers{
		GatewayID: c.Param("gateway_id"),
	}
	if err := gatewayIDs.ValidateContext(ctx); err != nil {
		return err
	}

	registry, err := s.getRegistry(ctx, &gatewayIDs)
	if err != nil {
		return err
	}
	gateway, err := registry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIdentifiers: gatewayIDs,
		FieldMask: types.FieldMask{Paths: []string{
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
		return err
	}

	md := rpcmetadata.FromIncomingContext(ctx)
	switch md.AuthType {
	case "bearer":
		if err := rights.RequireGateway(ctx, gatewayIDs, ttnpb.RIGHT_GATEWAY_INFO); err != nil {
			return err
		}
	case "key":
		keyAttr, ok := gateway.Attributes["key"]
		if !ok || md.AuthValue != keyAttr {
			return errUnauthenticated
		}
	default:
		gateway = gateway.PublicSafe()
	}

	freqPlanURL := s.inferServerAddress(c) + compatAPIPrefix + "/frequency-plans/" + gateway.FrequencyPlanID

	var rtr *router
	if gateway.GatewayServerAddress != "" {
		rtr = &router{
			ID: gateway.GatewayServerAddress,
		}
		rtr.MQTTAddress, err = s.inferMQTTAddress(gateway.GatewayServerAddress)
		if err != nil {
			return err
		}
	}

	response := &gatewayInfoResponse{
		ID:               gateway.GatewayID,
		FrequencyPlan:    gateway.FrequencyPlanID,
		FrequencyPlanURL: freqPlanURL,
		AutoUpdate:       gateway.AutoUpdate,
	}

	if rtr != nil {
		response.Router = rtr
		response.FallbackRouters = []*router{rtr}
	}

	if gateway.Description != "" {
		response.Attributes = &attributes{
			Description: gateway.Description,
		}
	}

	if len(gateway.Antennas) > 0 {
		response.AntennaLocation = &antennaLocation{
			Latitude:  gateway.Antennas[0].Location.Latitude,
			Longitude: gateway.Antennas[0].Location.Longitude,
			Altitude:  gateway.Antennas[0].Location.Altitude,
		}
	}

	if c.Request().Header.Get("User-Agent") == "TTNGateway" {
		s.setTTKGFirmwareURL(response, gateway)
	}

	if token, ok := gateway.Attributes["token"]; ok {
		response.Token = &oauth2Token{
			AccessToken: token,
			ExpiresIn:   uint32((24 * time.Hour).Seconds()),
		}
		if tokenExpires, ok := gateway.Attributes["token_expires"]; ok {
			if t, err := time.Parse(time.RFC3339, tokenExpires); err == nil {
				response.Token.ExpiresIn = uint32(time.Until(t).Seconds())
			}
		}
	}

	if c.Request().Header.Get("User-Agent") == "TTNGateway" {
		// Filter out fields to reduce response size.
		response.ID = ""
		response.Attributes = nil
		response.AntennaLocation = nil
		response.Token = nil
		response.FallbackRouters = nil
		response.Router.ID = ""
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) setTTKGFirmwareURL(res *gatewayInfoResponse, gtw *ttnpb.Gateway) {
	updateChannel := gtw.UpdateChannel
	if updateChannel == "" {
		updateChannel = s.ttgConfig.Default.UpdateChannel
	}
	if updateChannel == "" {
		updateChannel = "stable"
	}
	if strings.HasPrefix(updateChannel, "http") {
		res.FirmwareURL = updateChannel
		return
	}
	firmwareBaseURL := s.ttgConfig.Default.FirmwareURL
	if firmwareBaseURL == "" {
		firmwareBaseURL = "https://thethingsproducts.blob.core.windows.net/the-things-gateway/v1"
	}
	res.FirmwareURL = fmt.Sprintf("%s/%s", firmwareBaseURL, updateChannel)
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

func (s *Server) inferServerAddress(c echo.Context) string {
	scheme := c.Scheme()
	if forwardedProto := c.Request().Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	}
	hostport := c.Request().Host
	if forwardedHost := c.Request().Header.Get("X-Forwarded-Host"); forwardedHost != "" {
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
