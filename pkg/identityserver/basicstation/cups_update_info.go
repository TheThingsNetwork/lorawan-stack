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

package basicstation

import (
	"crypto"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"hash/crc32"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

const (
	cupsAttribute               = "_cups"
	cupsURIAttribute            = "_cups_uri"
	cupsLastSeenAttribute       = "_cups_last_seen"
	cupsCredentialsIDAttribute  = "_cups_credentials_id"
	cupsCredentialsAttribute    = "_cups_credentials"
	cupsCredentialsCRCAttribute = "_cups_credentials_crc"
	cupsStationAttribute        = "_cups_station"
	cupsModelAttribute          = "_cups_model"
	cupsPackageAttribute        = "_cups_package"
	lnsCredentialsIDAttribute   = "_lns_credentials_id"
	lnsCredentialsAttribute     = "_lns_credentials"
	lnsCredentialsCRCAttribute  = "_lns_credentials_crc"
)

var (
	errUnauthenticated = errors.DefineUnauthenticated("unauthenticated", "call was not authenticated")
	errCUPSNotEnabled  = errors.DefinePermissionDenied("cups_not_enabled", "CUPS is not enabled for gateway `{gateway_uid}`")
	errInvalidToken    = errors.DefinePermissionDenied("invalid_token", "invalid provisioning token")
)

// UpdateInfo implements the CUPS update-info handler.
func (s *CUPSServer) UpdateInfo(c echo.Context) error {
	if c.Request().Header.Get(echo.HeaderContentType) == "" {
		c.Request().Header.Set(echo.HeaderContentType, "application/json")
	}

	var req UpdateInfoRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := getContext(c)

	md := rpcmetadata.FromIncomingContext(ctx)

	var (
		authorization   grpc.CallOption
		cupsCredentials = c.Request().Header.Get(echo.HeaderAuthorization)
	)

	switch strings.ToLower(md.AuthType) {
	case "":
		// TODO: Support TLS Client Auth (https://github.com/TheThingsNetwork/lorawan-stack/issues/137).
		return errUnauthenticated
	case "bearer":
		if _, _, _, err := auth.SplitToken(md.AuthValue); err == nil {
			authorization = grpc.PerRPCCredentials(md)
			cupsCredentials = ""
		}
	}

	if authorization == nil && s.fallbackAuth != nil {
		authorization = s.fallbackAuth(ctx, req.Router.EUI64, cupsCredentials)
	}
	if authorization == nil {
		return errUnauthenticated
	}

	var gtw *ttnpb.Gateway

	ids, err := s.gatewayRegistry.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
		EUI: req.Router.EUI64,
	}, authorization)
	if errors.IsNotFound(err) && s.registerUnknown {
		gtw, err = s.gatewayRegistry.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayID: fmt.Sprintf("eui-%s", req.Router.EUI64.String()),
					EUI:       &req.Router.EUI64,
				},
				Attributes: map[string]string{
					cupsAttribute:               "true",
					cupsURIAttribute:            req.CUPSURI,
					cupsCredentialsAttribute:    cupsCredentials,
					cupsCredentialsCRCAttribute: strconv.FormatUint(uint64(req.CUPSCredentialsCRC), 10),
					lnsCredentialsCRCAttribute:  strconv.FormatUint(uint64(req.LNSCredentialsCRC), 10),
				},
				GatewayServerAddress: req.LNSURI,
			},
			Collaborator: s.defaultOwner,
		}, authorization)
	} else if err != nil {
		return err
	} else {
		gtw, err = s.gatewayRegistry.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: *ids,
			FieldMask: types.FieldMask{Paths: []string{
				"attributes",
				"version_ids",
				"gateway_server_address",
				"auto_update",
				"update_channel",
				"frequency_plan_id",
			}},
		}, authorization)
	}
	if err != nil {
		return err
	}

	if gtw.Attributes == nil {
		gtw.Attributes = make(map[string]string)
	}

	if s.requireExplicitEnable {
		if cups, _ := strconv.ParseBool(gtw.Attributes[cupsAttribute]); !cups {
			return errCUPSNotEnabled.WithAttributes("gateway_uid", unique.ID(ctx, gtw.GatewayIdentifiers))
		}
	}

	if cupsCredentials != "" {
		registeredCredentials := gtw.Attributes[cupsCredentialsAttribute]
		if registeredCredentials != "" && registeredCredentials != cupsCredentials {
			return errInvalidToken
		}
	}

	res := UpdateInfoResponse{}

	if s.allowCUPSURIUpdate {
		if cupsURI := gtw.Attributes[cupsURIAttribute]; cupsURI != "" && cupsURI != req.CUPSURI {
			res.CUPSURI = cupsURI
		}
	}
	if gtw.GatewayServerAddress != "" && gtw.GatewayServerAddress != req.LNSURI {
		res.LNSURI = gtw.GatewayServerAddress // TODO: Clean / Format.
	}
	if gtw.Attributes[cupsCredentialsCRCAttribute] != strconv.FormatUint(uint64(req.CUPSCredentialsCRC), 10) {
		if gtw.Attributes[cupsCredentialsAttribute] == "" {
			if gtw.Attributes[cupsCredentialsIDAttribute] != "" {
				// TODO: Try deleting old CUPS credentials.
			}
			apiKey, err := s.gatewayAccess.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
				GatewayIdentifiers: gtw.GatewayIdentifiers,
				Name:               fmt.Sprintf("CUPS Key, generated %s", time.Now().UTC().Format(time.RFC3339)),
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_GATEWAY_INFO,              // We need to read private attributes.
					ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,    // We need to write attributes.
					ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS, // We need to create API keys.
					ttnpb.RIGHT_GATEWAY_LINK,              // We need to create the LNS API key with this right.
				},
			}, authorization)
			if err != nil {
				return err
			}
			gtw.Attributes[cupsCredentialsIDAttribute] = apiKey.ID
			gtw.Attributes[cupsCredentialsAttribute] = apiKey.Key
		}
		trust, err := s.getTrust(gtw.Attributes[cupsURIAttribute])
		if err != nil {
			return err
		}
		if trust != nil {
			creds, err := TokenCredentials(trust, gtw.Attributes[cupsCredentialsAttribute])
			if err != nil {
				return err
			}
			res.CUPSCredentials = creds
			gtw.Attributes[cupsCredentialsCRCAttribute] = strconv.FormatUint(uint64(crc32.ChecksumIEEE(res.CUPSCredentials)), 10)
		} else {
			delete(gtw.Attributes, cupsCredentialsCRCAttribute)
		}
	}
	if gtw.Attributes[lnsCredentialsCRCAttribute] != strconv.FormatUint(uint64(req.LNSCredentialsCRC), 10) {
		if gtw.Attributes[lnsCredentialsAttribute] == "" {
			if gtw.Attributes[lnsCredentialsIDAttribute] != "" {
				// TODO: Try deleting old LNS credentials.
			}
			apiKey, err := s.gatewayAccess.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
				GatewayIdentifiers: gtw.GatewayIdentifiers,
				Name:               fmt.Sprintf("LNS Key, generated %s", time.Now().UTC().Format(time.RFC3339)),
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_GATEWAY_INFO,
					ttnpb.RIGHT_GATEWAY_LINK,
				},
			}, authorization)
			if err != nil {
				return err
			}
			gtw.Attributes[lnsCredentialsIDAttribute] = apiKey.ID
			gtw.Attributes[lnsCredentialsAttribute] = apiKey.Key
		}
		trust, err := s.getTrust(gtw.GatewayServerAddress)
		if err != nil {
			return err
		}
		if trust != nil {
			creds, err := TokenCredentials(trust, gtw.Attributes[lnsCredentialsAttribute])
			if err != nil {
				return err
			}
			res.LNSCredentials = creds
			gtw.Attributes[lnsCredentialsCRCAttribute] = strconv.FormatUint(uint64(crc32.ChecksumIEEE(res.LNSCredentials)), 10)
		} else {
			delete(gtw.Attributes, lnsCredentialsCRCAttribute)
		}
	}
	if gtw.AutoUpdate {
		// TODO:
		// Check Station, Model, Package
		// Compare to version_ids, update_channel
		// Get update data
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

	gtw, err = s.gatewayRegistry.Update(ctx, &ttnpb.UpdateGatewayRequest{
		Gateway: *gtw,
		FieldMask: types.FieldMask{Paths: []string{
			"attributes",
		}},
	}, authorization)
	if err != nil {
		return err
	}

	b, err := res.MarshalBinary()
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, b)
}
