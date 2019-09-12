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

package cups

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"hash/crc32"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

const (
	cupsAttribute              = "cups"
	cupsURIAttribute           = "cups-uri"
	cupsLastSeenAttribute      = "cups-last-seen"
	cupsCredentialsIDAttribute = "cups-credentials-id"
	cupsCredentialsAttribute   = "cups-credentials"
	cupsStationAttribute       = "cups-station"
	cupsModelAttribute         = "cups-model"
	cupsPackageAttribute       = "cups-package"
	lnsCredentialsIDAttribute  = "lns-credentials-id"
	lnsCredentialsAttribute    = "lns-credentials"
)

var (
	errUnauthenticated = errors.DefineUnauthenticated("unauthenticated", "call was not authenticated")
	errCUPSNotEnabled  = errors.DefinePermissionDenied("cups_not_enabled", "CUPS is not enabled for gateway `{gateway_uid}`")
	errInvalidToken    = errors.DefinePermissionDenied("invalid_token", "invalid provisioning token")
)

func (s *Server) registerGateway(ctx context.Context, req UpdateInfoRequest) (*ttnpb.Gateway, error) {
	logger := log.FromContext(ctx)
	ids := ttnpb.GatewayIdentifiers{
		GatewayID: fmt.Sprintf("eui-%s", strings.ToLower(req.Router.EUI64.String())),
		EUI:       &req.Router.EUI64,
	}
	logger = logger.WithField("gateway_uid", unique.ID(ctx, ids))
	registry, err := s.getRegistry(ctx, &ids)
	if err != nil {
		return nil, err
	}
	auth := s.defaultOwnerAuth(ctx)
	gtw, err := registry.Create(ctx, &ttnpb.CreateGatewayRequest{
		Gateway: ttnpb.Gateway{
			GatewayIdentifiers:   ids,
			GatewayServerAddress: s.defaultLNSURI,
		},
		Collaborator: s.defaultOwner,
	}, auth)
	if err != nil {
		return nil, err
	}
	logger.Info("Created new gateway")
	accessRegistry, err := s.getAccess(ctx, &gtw.GatewayIdentifiers)
	if err != nil {
		return nil, err
	}
	cupsKey, err := accessRegistry.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
		GatewayIdentifiers: gtw.GatewayIdentifiers,
		Name:               fmt.Sprintf("CUPS Key, generated %s", time.Now().UTC().Format(time.RFC3339)),
		Rights: []ttnpb.Right{
			ttnpb.RIGHT_GATEWAY_INFO,
			ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,    // We need to write attributes.
			ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS, // We need to create API keys.
			ttnpb.RIGHT_GATEWAY_LINK,              // We need to create the LNS API key with this right.
		},
	}, auth)
	if err != nil {
		return nil, err
	}
	logger.WithField("api_key_id", cupsKey.ID).Info("Created gateway API key for CUPS")
	lnsKey, err := accessRegistry.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
		GatewayIdentifiers: gtw.GatewayIdentifiers,
		Name:               fmt.Sprintf("LNS Key, generated %s", time.Now().UTC().Format(time.RFC3339)),
		Rights: []ttnpb.Right{
			ttnpb.RIGHT_GATEWAY_INFO,
			ttnpb.RIGHT_GATEWAY_LINK,
		},
	}, auth)
	if err != nil {
		return nil, err
	}
	logger.WithField("api_key_id", lnsKey.ID).Info("Created gateway API key for LNS")
	gtw.Attributes = map[string]string{
		cupsAttribute:              "true",
		cupsCredentialsIDAttribute: cupsKey.ID,
		cupsCredentialsAttribute:   fmt.Sprintf("Bearer %s", cupsKey.Key),
		lnsCredentialsIDAttribute:  lnsKey.ID,
		lnsCredentialsAttribute:    fmt.Sprintf("Bearer %s", lnsKey.Key),
	}
	return gtw, nil
}

var getGatewayMask = pbtypes.FieldMask{Paths: []string{
	"attributes",
	"version_ids",
	"gateway_server_address",
	"auto_update",
	"update_channel",
	"frequency_plan_id",
}}

// UpdateInfo implements the CUPS update-info handler.
func (s *Server) UpdateInfo(c echo.Context) error {
	if c.Request().Header.Get(echo.HeaderContentType) == "" {
		c.Request().Header.Set(echo.HeaderContentType, "application/json")
	}

	var req UpdateInfoRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := getContext(c)
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"gateway_eui", req.Router.EUI64.String(),
	))
	ctx = log.NewContext(ctx, logger)

	registry, err := s.getRegistry(ctx, nil)
	if err != nil {
		return err
	}
	serverAuth := s.getServerAuth(ctx)
	gatewayAuth, err := rpcmetadata.WithForwardedAuth(ctx, s.component.AllowInsecureForCredentials())
	if err != nil {
		return err
	}

	var gtw *ttnpb.Gateway
	ids, err := registry.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
		EUI: req.Router.EUI64,
	}, serverAuth)
	if err == nil {
		logger.WithField("gateway_uid", unique.ID(ctx, ids)).Debug("Found gateway for EUI")
		gtw, err = registry.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: *ids,
			FieldMask:          getGatewayMask,
		}, gatewayAuth)
	} else if errors.IsNotFound(err) && s.registerUnknown {
		gtw, err = s.registerGateway(ctx, req)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	logger = logger.WithField("gateway_uid", unique.ID(ctx, gtw.GatewayIdentifiers))

	if gtw.Attributes == nil {
		gtw.Attributes = make(map[string]string)
	}

	if s.requireExplicitEnable || gtw.Attributes[cupsAttribute] != "" {
		if cups, _ := strconv.ParseBool(gtw.Attributes[cupsAttribute]); !cups {
			return errCUPSNotEnabled.WithAttributes("gateway_uid", unique.ID(ctx, gtw.GatewayIdentifiers))
		}
	}

	res := UpdateInfoResponse{}

	if gtw.Attributes[cupsURIAttribute] == "" {
		gtw.Attributes[cupsURIAttribute] = req.CUPSURI
	}
	if s.allowCUPSURIUpdate && gtw.Attributes[cupsURIAttribute] != req.CUPSURI {
		res.CUPSURI = gtw.Attributes[cupsURIAttribute]
	}

	if credentials := gtw.Attributes[cupsCredentialsAttribute]; credentials != "" {
		cupsTrust, err := s.getTrust(gtw.Attributes[cupsURIAttribute])
		if err != nil {
			return err
		}
		cupsCredentials, err := TokenCredentials(cupsTrust, credentials)
		if err != nil {
			return err
		}
		if crc32.ChecksumIEEE(cupsCredentials) != req.CUPSCredentialsCRC {
			res.CUPSCredentials = cupsCredentials
		}
	}

	if gtw.GatewayServerAddress == "" {
		if req.LNSURI != "" {
			gtw.GatewayServerAddress = req.LNSURI
		} else {
			gtw.GatewayServerAddress = s.defaultLNSURI
		}
	}
	if gtw.GatewayServerAddress != req.LNSURI {
		scheme, host, port, err := parseAddress(gtw.GatewayServerAddress)
		if err != nil {
			return err
		}
		if scheme == "" {
			scheme = "wss"
		}
		address := host
		if port != "" {
			address = net.JoinHostPort(host, port)
		}
		res.LNSURI = fmt.Sprintf("%s://%s", scheme, address)
	}

	if credentials := gtw.Attributes[lnsCredentialsAttribute]; credentials != "" {
		lnsTrust, err := s.getTrust(gtw.GatewayServerAddress)
		if err != nil {
			return err
		}
		lnsCredentials, err := TokenCredentials(lnsTrust, credentials)
		if err != nil {
			return err
		}
		if crc32.ChecksumIEEE(lnsCredentials) != req.LNSCredentialsCRC {
			res.LNSCredentials = lnsCredentials
		}
	}

	if gtw.AutoUpdate {
		// TODO: Compare the Station, Model, Package, version_ids and update_channel in order to check if any updates are required
		// (https://github.com/TheThingsNetwork/lorawan-stack/issues/365)
		var updateData []byte
		if updateData != nil {
			var (
				keyCRC uint32
				signer crypto.Signer
			)
			for _, keyCRC = range req.KeyCRCs {
				if sig, ok := s.signers[keyCRC]; ok {
					signer = sig
					break
				}
			}
			if signer != nil {
				hash := sha512.Sum512(updateData)
				sig, err := signer.Sign(rand.Reader, hash[:], nil)
				if err != nil {
					return err
				}
				res.SignatureKeyCRC = keyCRC
				res.Signature = sig
				res.UpdateData = updateData
			}
		}
	}

	gtw.Attributes[cupsLastSeenAttribute] = time.Now().UTC().Format(time.RFC3339)
	gtw.Attributes[cupsStationAttribute] = req.Station
	gtw.Attributes[cupsModelAttribute] = req.Model
	gtw.Attributes[cupsPackageAttribute] = req.Package

	registry, err = s.getRegistry(ctx, &gtw.GatewayIdentifiers)
	if err != nil {
		return err
	}
	gtw, err = registry.Update(ctx, &ttnpb.UpdateGatewayRequest{
		Gateway: *gtw,
		FieldMask: pbtypes.FieldMask{Paths: []string{
			"attributes",
		}},
	}, gatewayAuth)
	if err != nil {
		return err
	}

	b, err := res.MarshalBinary()
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, b)
}
