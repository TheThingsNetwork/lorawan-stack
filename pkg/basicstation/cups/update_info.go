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
	"strings"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	cupsLastSeenAttribute = "cups-last-seen"
	cupsStationAttribute  = "cups-station"
	cupsModelAttribute    = "cups-model"
	cupsPackageAttribute  = "cups-package"
)

var (
	errUnauthenticated       = errors.DefineUnauthenticated("unauthenticated", "call was not authenticated")
	errTargetCUPSCredentials = errors.DefineNotFound("target_cups_credentials_not_found", "Target CUPS credentials not found for gateway `{gateway_uid}`")
	errLNSCredentials        = errors.DefineNotFound("lns_credentials_not_found", "LNS credentials not found for gateway `{gateway_uid}`")
	errServerTrust           = errors.Define("server_trust", "failed to fetch server trust for address `{address}`")
)

func getAuthHeader(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if authorization := md.Get("authorization"); len(authorization) > 0 {
			return authorization[len(authorization)-1]
		}
	}
	return ""
}

// registerGateway creates a new gateway for the default owner. It also creates the necessary CUPS and LNS credentials.
// `TargetCUPSURI` is set in order to make the gateway connect once again to this CUPS but using auth and then receive the LNS credentials.
func (s *Server) registerGateway(ctx context.Context, req UpdateInfoRequest) (*ttnpb.Gateway, error) {
	logger := log.FromContext(ctx)
	ids := ttnpb.GatewayIdentifiers{
		GatewayId: fmt.Sprintf("eui-%s", strings.ToLower(req.Router.EUI64.String())),
		Eui:       &req.Router.EUI64,
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
			ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
			ttnpb.RIGHT_GATEWAY_READ_SECRETS,
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
	_, err = registry.Update(ctx, &ttnpb.UpdateGatewayRequest{
		Gateway: ttnpb.Gateway{
			GatewayIdentifiers: ids,
			LBSLNSSecret: &ttnpb.Secret{
				Value: []byte(lnsKey.Key),
			},
			TargetCUPSURI: req.CUPSURI,
			TargetCUPSKey: &ttnpb.Secret{
				Value: []byte(cupsKey.Key),
			},
		},
		FieldMask: &pbtypes.FieldMask{
			Paths: []string{"lbs_lns_secret"},
		},
	}, auth)
	if err != nil {
		return nil, err
	}
	logger.WithField("api_key_id", lnsKey.ID).Info("Created gateway API key for LNS")
	return gtw, nil
}

var getGatewayMask = pbtypes.FieldMask{Paths: []string{
	"attributes",
	"version_ids",
	"gateway_server_address",
	"auto_update",
	"update_channel",
	"frequency_plan_id",
	"lbs_lns_secret",
	"target_cups_uri",
	"target_cups_key",
}}

// UpdateInfo implements the CUPS update-info handler.
func (s *Server) UpdateInfo(c echo.Context) error {
	// This is to account for older LBS gateways that don't set this header.
	c.Request().Header.Set(echo.HeaderContentType, "application/json")

	var req UpdateInfoRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := req.ValidateContext(c.Request().Context()); err != nil {
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

	var ids *ttnpb.GatewayIdentifiers
	ids, err = registry.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
		Eui: req.Router.EUI64,
	}, serverAuth)
	if err != nil {
		if errors.IsNotFound(err) && s.registerUnknown {
			gtw, err := s.registerGateway(ctx, req)
			if err != nil {
				return err
			}
			ids = &gtw.GatewayIdentifiers
			// Use the generated CUPS API Key for authenticating subsequent calls.
			md := metadata.New(map[string]string{
				"id":            ids.GatewayId,
				"authorization": fmt.Sprintf("Bearer %s", string(gtw.TargetCUPSKey.Value)),
			})
			if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
				md = metadata.Join(ctxMd, md)
			}
			ctx = metadata.NewIncomingContext(ctx, md)
			// This makes the server return the target CUPS URI and credentials to the gateway.
			req.CUPSURI = ""
		} else {
			return err
		}
	}

	uid := unique.ID(ctx, ids)
	logger.WithField("gateway_uid", uid).Debug("Found gateway for EUI")

	var md metadata.MD
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	if auth != "" {
		if !strings.HasPrefix(auth, "Bearer ") {
			auth = fmt.Sprintf("Bearer %s", auth)
		}
		md = metadata.New(map[string]string{
			"id":            ids.GatewayId,
			"authorization": auth,
		})
	}
	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	ctx = metadata.NewIncomingContext(ctx, md)

	var gatewayAuth grpc.CallOption
	if rights.RequireGateway(ctx, *ids,
		ttnpb.RIGHT_GATEWAY_INFO,
		ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
		ttnpb.RIGHT_GATEWAY_READ_SECRETS,
	) == nil {
		logger.Debug("Authorized with The Things Stack token")
	} else {
		return errUnauthenticated.New()
	}
	gatewayAuth, err = rpcmetadata.WithForwardedAuth(ctx, s.component.AllowInsecureForCredentials())
	if err != nil {
		return err
	}

	gtw, err := registry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIdentifiers: *ids,
		FieldMask:          &getGatewayMask,
	}, gatewayAuth)
	if err != nil {
		return err
	}

	if gtw.Attributes == nil {
		gtw.Attributes = make(map[string]string)
	}

	res := UpdateInfoResponse{}
	if s.allowCUPSURIUpdate && gtw.TargetCUPSURI != "" && gtw.TargetCUPSURI != req.CUPSURI {
		if gtw.TargetCUPSKey == nil || gtw.TargetCUPSKey.Value == nil {
			return errTargetCUPSCredentials.New()
		}
		logger := logger.WithField("cups_uri", gtw.TargetCUPSURI)
		logger.Debug("Configure CUPS")
		res.CUPSURI = gtw.TargetCUPSURI

		cupsTrust, err := s.getTrust(gtw.TargetCUPSURI)
		if err != nil {
			return errServerTrust.WithCause(err).WithAttributes("address", gtw.TargetCUPSURI)
		}
		cupsCredentials, err := TokenCredentials(cupsTrust, string(gtw.TargetCUPSKey.Value))
		if err != nil {
			return err
		}
		if crc32.ChecksumIEEE(cupsCredentials) != req.CUPSCredentialsCRC {
			res.CUPSCredentials = cupsCredentials
		}
	} else if gtw.TargetCUPSKey != nil && gtw.TargetCUPSKey.Value != nil {
		// Check if CUPS Key needs to be rotated.
		cupsTrust, err := s.getTrust(req.CUPSURI)
		if err != nil {
			return errServerTrust.WithCause(err).WithAttributes("address", req.CUPSURI)
		}
		cupsCredentials, err := TokenCredentials(cupsTrust, string(gtw.TargetCUPSKey.Value))
		if err != nil {
			return err
		}
		if crc32.ChecksumIEEE(cupsCredentials) != req.CUPSCredentialsCRC {
			logger.Debug("Update CUPS Credentials")
			res.CUPSCredentials = cupsCredentials
		}
	} else {
		logger := logger.WithField("lns_uri", gtw.GatewayServerAddress)
		logger.Debug("Configure LNS")
		if gtw.LBSLNSSecret == nil {
			return errLNSCredentials.WithAttributes("gateway_uid", gtw.GatewayId)
		}
		if gtw.GatewayServerAddress == "" {
			if req.LNSURI != "" {
				gtw.GatewayServerAddress = req.LNSURI
			} else {
				gtw.GatewayServerAddress = s.defaultLNSURI
			}
		}
		if gtw.GatewayServerAddress != req.LNSURI {
			scheme, host, port, err := parseAddress("wss", gtw.GatewayServerAddress)
			if err != nil {
				return err
			}
			address := host
			address = net.JoinHostPort(host, port)
			res.LNSURI = fmt.Sprintf("%s://%s", scheme, address)
		}

		lnsTrust, err := s.getTrust(gtw.GatewayServerAddress)
		if err != nil {
			return errServerTrust.WithCause(err).WithAttributes("address", gtw.GatewayServerAddress)
		}
		lnsCredentials, err := TokenCredentials(lnsTrust, string(gtw.LBSLNSSecret.Value))
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
	if req.Station != "" {
		gtw.Attributes[cupsStationAttribute] = req.Station
	}
	if req.Model != "" {
		gtw.Attributes[cupsModelAttribute] = req.Model
	}
	if req.Package != "" {
		gtw.Attributes[cupsPackageAttribute] = req.Package
	}

	registry, err = s.getRegistry(ctx, &gtw.GatewayIdentifiers)
	if err != nil {
		return err
	}
	gtw, err = registry.Update(ctx, &ttnpb.UpdateGatewayRequest{
		Gateway: *gtw,
		FieldMask: &pbtypes.FieldMask{Paths: []string{
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
