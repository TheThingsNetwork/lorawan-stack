// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

// Package ttjs provides the claiming client implementation for The Things Join Server API.
package ttjs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
)

// BasicAuth contains HTTP basic auth settings.
type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Config is the configuration to communicate with The Things Join Server End Device Claming API.
type Config struct {
	NetID           types.NetID         `yaml:"-"`
	JoinEUIPrefixes []types.EUI64Prefix `yaml:"-"`

	BasicAuth          `yaml:"basic-auth"`
	ClaimingAPIVersion string            `yaml:"claiming-api-version"`
	URL                string            `yaml:"url"`
	TenantID           string            `yaml:"tenant-id"`
	HomeNSIDs          map[string]string `yaml:"home-ns-ids"`
}

// Component abstracts the underlying *component.Component.
type Component interface {
	httpclient.Provider
	GetBaseConfig(ctx context.Context) config.ServiceBase
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
}

// TTJS is a client that claims end devices on a The Things Join Server.
type TTJS struct {
	Component
	hsNSIDs     map[string]types.EUI64
	httpClient  *http.Client
	baseURL     *url.URL
	config      Config
	ttiVendorID OUI
}

// NewClient applies the config and returns a new TTJS client.
func (config Config) NewClient(ctx context.Context, c Component) (*TTJS, error) {
	httpClient, err := c.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	baseURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}
	// Parse the HomeNSIDs map.
	res := make(map[string]types.EUI64, len(config.HomeNSIDs))
	for nsAddress, nsIDString := range config.HomeNSIDs {
		var nsID types.EUI64
		err := nsID.UnmarshalText([]byte(nsIDString))
		if err != nil {
			return nil, err
		}
		res[nsAddress] = nsID
	}
	return &TTJS{
		config:      config,
		httpClient:  httpClient,
		baseURL:     baseURL,
		Component:   c,
		hsNSIDs:     res,
		ttiVendorID: OUI(interop.TTIVendorID.MarshalNumber()),
	}, nil
}

// SupportsJoinEUI implements EndDeviceClaimer.
func (client *TTJS) SupportsJoinEUI(eui types.EUI64) bool {
	for _, prefix := range client.config.JoinEUIPrefixes {
		if eui.HasPrefix(prefix) {
			return true
		}
	}
	return false
}

var (
	errDeviceNotProvisioned = errors.DefineNotFound("device_not_provisioned", "device with EUI `{dev_eui}` not provisioned")
	errDeviceNotClaimed     = errors.DefineNotFound("device_not_claimed", "device with EUI `{dev_eui}` not claimed")
	errDeviceAccessDenied   = errors.DefinePermissionDenied("device_access_denied", "access to device with `{dev_eui}` denied. Either device is already claimed or owner token is invalid")
	errUnauthorized         = errors.DefineUnauthenticated("unauthorized", "client API Key missing or invalid")
	errNoHomeNSID           = errors.DefineInvalidArgument("no_home_ns_id", "no HomeNSID configured for network server address `{address}`")
	errParseQRCode          = errors.Define("parse_qr_code", "parse QR code failed")
	errQRCodeData           = errors.DefineInvalidArgument("qr_code_data", "invalid QR code data")
	errNoJoinEUI            = errors.DefineInvalidArgument("no_join_eui", "failed to extract JoinEUI from request")
)

// Claim implements EndDeviceClaimer.
func (client *TTJS) Claim(ctx context.Context, req *ttnpb.ClaimEndDeviceRequest) (*ttnpb.EndDeviceIdentifiers, error) {
	htenantID := client.config.TenantID

	var (
		joinEUI, devEUI        types.EUI64
		hNSAddress, ownerToken string
	)
	if authenticatedIDs := req.GetAuthenticatedIdentifiers(); authenticatedIDs != nil {
		joinEUI = authenticatedIDs.JoinEui
		devEUI = authenticatedIDs.DevEui
		ownerToken = authenticatedIDs.AuthenticationCode
	} else if qrCode := req.GetQrCode(); qrCode != nil {
		conn, err := client.Component.GetPeerConn(ctx, ttnpb.ClusterRole_QR_CODE_GENERATOR, nil)
		if err != nil {
			return nil, err
		}
		qrg := ttnpb.NewEndDeviceQRCodeGeneratorClient(conn)
		data, err := qrg.Parse(ctx, &ttnpb.ParseEndDeviceQRCodeRequest{
			QrCode: qrCode,
		})
		if err != nil {
			return nil, errParseQRCode.WithCause(err)
		}

		dev := data.GetEndDeviceTempate().GetEndDevice()
		if dev == nil {
			return nil, errParseQRCode.New()
		}
		joinEUI = *dev.GetIds().JoinEui
		devEUI = *dev.GetIds().DevEui
		ownerToken = dev.ClaimAuthenticationCode.Value
	} else {
		return nil, errNoJoinEUI.New()
	}

	hNSAddress, _, err := net.SplitHostPort(req.TargetNetworkServerAddress)
	if err != nil {
		// TargetNetworkServerAddress is already validated by the API.
		// An error here means that it does not contain a port, so we use it directly.
		hNSAddress = req.TargetNetworkServerAddress
	}

	hNSID, ok := client.hsNSIDs[hNSAddress]
	if !ok {
		return nil, errNoHomeNSID.WithAttributes("address", hNSAddress)
	}

	claimReq := &claimRequest{
		OwnerToken: ownerToken,
		claimData: claimData{
			HomeNetID: client.config.NetID.String(),
			HomeNSID:  hNSID.String(),
			VendorSpecific: VendorSpecific{
				OUI: client.ttiVendorID,
				Data: struct {
					HTenantID  string
					HNSAddress string
				}{
					HTenantID:  htenantID,
					HNSAddress: hNSAddress,
				},
			},
		},
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(claimReq)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("%s/%s/claim/%s", client.baseURL.String(), client.config.ClaimingAPIVersion, devEUI.String())

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_eui", devEUI,
		"path", path,
	))

	logger.Debug("Claim end device")
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, path, &buf)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(client.config.BasicAuth.Username, client.config.BasicAuth.Password)
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if isSuccess(resp.StatusCode) {
		// Echo values from the Request
		return &ttnpb.EndDeviceIdentifiers{
			DeviceId:       req.TargetDeviceId,
			ApplicationIds: req.TargetApplicationIds,
			DevEui:         &devEUI,
			JoinEui:        &joinEUI,
		}, nil
	}

	// Unmarshal and log the error body.
	var errResp errorResponse
	err = json.Unmarshal(respBody, &errResp)
	if err != nil {
		logger.WithError(err).Warn("Failed to decode error message")
	} else {
		logger.WithField("error", errResp.Message).Warn("Failed to unclaim end device")
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, errDeviceNotProvisioned.WithAttributes("dev_eui", devEUI)
	case http.StatusForbidden:
		return nil, errDeviceAccessDenied.WithAttributes("dev_eui", devEUI)
	case http.StatusUnauthorized:
		return nil, errUnauthorized.New()
	default:
		return nil, errors.FromHTTPStatusCode(resp.StatusCode)
	}
}

// Unclaim implements EndDeviceClaimer.
func (client *TTJS) Unclaim(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) error {
	devEUI := *ids.DevEui
	path := fmt.Sprintf("%s/%s/claim/%s", client.baseURL.String(), client.config.ClaimingAPIVersion, devEUI.String())
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_eui", devEUI,
		"join_server_address", client.baseURL.Host,
		"path", path,
	))

	logger.Debug("Unclaim end device")
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	request.SetBasicAuth(client.config.BasicAuth.Username, client.config.BasicAuth.Password)
	resp, err := client.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if isSuccess(resp.StatusCode) {
		return nil
	}

	// Unmarshal and log the error body.
	var errResp errorResponse
	err = json.Unmarshal(respBody, &errResp)
	if err != nil {
		logger.WithError(err).Warn("Failed to decode error message")
	} else {
		logger.WithField("error", errResp.Message).Warn("Failed to unclaim end device")
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return errDeviceNotClaimed.WithAttributes("dev_eui", devEUI)
	case http.StatusForbidden:
		return errDeviceAccessDenied.WithAttributes("dev_eui", devEUI)
	case http.StatusUnauthorized:
		return errUnauthorized.New()
	default:
		return errors.FromHTTPStatusCode(resp.StatusCode)
	}
}

// GetClaimStatus implements EndDeviceClaimer.
func (client *TTJS) GetClaimStatus(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.GetClaimStatusResponse, error) {
	devEUI := ids.DevEui
	path := fmt.Sprintf("%s/%s/claim/%s", client.baseURL.String(), client.config.ClaimingAPIVersion, devEUI.String())
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_eui", devEUI,
		"join_eui", ids.JoinEui,
		"join_server_address", client.baseURL.Host,
		"path", path,
	))

	logger.Debug("Get claim status for end device")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	request.SetBasicAuth(client.config.BasicAuth.Username, client.config.BasicAuth.Password)
	resp, err := client.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if isSuccess(resp.StatusCode) {
		var (
			claimData claimData
			ret       = ttnpb.GetClaimStatusResponse{
				EndDeviceIds: ids,
				HomeNetId:    &types.NetID{},
				HomeNsId:     &types.EUI64{},
			}
		)
		err = json.Unmarshal(respBody, &claimData)
		if err != nil {
			return nil, err
		}
		err = ret.HomeNetId.UnmarshalText([]byte(claimData.HomeNetID))
		if err != nil {
			return nil, err
		}
		err = ret.HomeNsId.UnmarshalText([]byte(claimData.HomeNSID))
		if err != nil {
			return nil, err
		}
		ret.VendorSpecific = &ttnpb.GetClaimStatusResponse_VendorSpecific{
			OrganizationUniqueIdentifier: uint32(claimData.VendorSpecific.OUI),
		}
		return &ret, nil
	}

	// Unmarshal and log the error body.
	var errResp errorResponse
	err = json.Unmarshal(respBody, &errResp)
	if err != nil {
		logger.WithError(err).Warn("Failed to decode error message")
	} else {
		logger.WithField("error", errResp.Message).Warn("Failed to unclaim end device")
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, errDeviceNotClaimed.WithAttributes("dev_eui", devEUI)
	case http.StatusForbidden:
		return nil, errDeviceAccessDenied.WithAttributes("dev_eui", devEUI)
	case http.StatusUnauthorized:
		return nil, errUnauthorized.New()
	default:
		return nil, errors.FromHTTPStatusCode(resp.StatusCode)
	}
}

// isSuccess returns true if the HTTP status code is 2xx.
func isSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}
