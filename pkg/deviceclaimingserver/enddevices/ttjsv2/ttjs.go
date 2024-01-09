// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

// Package ttjsv2 provides the claiming client implementation for The Things Join Server 2.0 API.
package ttjsv2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	claimerrors "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/enddevices/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// ConfigFile defines the configuration file for The Things Join Server claiming client.
type ConfigFile struct {
	URL string    `yaml:"url"`
	TLS TLSConfig `yaml:"tls"`

	// BasicAuth is no longer used and is only kept for backwards compatibility.
	// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/6049)
	BasicAuth struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"basic-auth"`
}

// Config is the configuration for The Things Join Server claiming client.
type Config struct {
	NetID           types.NetID
	NSID            *types.EUI64
	ASID            string
	JoinEUIPrefixes []types.EUI64Prefix
	ConfigFile
}

// Component abstracts the component.
type Component interface {
	httpclient.Provider
	KeyService() crypto.KeyService
}

// TTJS is a client that claims end devices on a The Things Join Server.
type TTJS struct {
	Component

	fetcher fetch.Interface
	config  Config
}

// NewClient applies the config and returns a new TTJS client.
func NewClient(c Component, fetcher fetch.Interface, conf Config) *TTJS {
	return &TTJS{
		Component: c,
		fetcher:   fetcher,
		config:    conf,
	}
}

// SupportsJoinEUI implements EndDeviceClaimer.
func (c *TTJS) SupportsJoinEUI(eui types.EUI64) bool {
	for _, prefix := range c.config.JoinEUIPrefixes {
		if eui.HasPrefix(prefix) {
			return true
		}
	}
	return false
}

func (c *TTJS) httpClient(ctx context.Context) (*http.Client, error) {
	var opts []httpclient.Option
	if !c.config.TLS.IsZero() {
		tlsConf, err := c.config.TLS.TLSConfig(c.fetcher, c.KeyService())
		if err != nil {
			return nil, err
		}
		opts = append(opts, httpclient.WithTLSConfig(tlsConf))
	}
	// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/6049)
	if c.config.BasicAuth.Username != "" || c.config.BasicAuth.Password != "" {
		log.FromContext(ctx).Warn("Basic authentication with The Things Join Server is no longer supported and will be removed in a future version.") //nolint:lll
	}
	return c.HTTPClient(ctx, opts...)
}

var (
	errBadRequest           = errors.DefineInvalidArgument("bad_request", "bad request", "message")
	errDeviceNotProvisioned = errors.DefineNotFound("device_not_provisioned", "device with EUI `{dev_eui}` not provisioned") //nolint:lll
	errDeviceNotClaimed     = errors.DefineNotFound("device_not_claimed", "device with EUI `{dev_eui}` not claimed")
	errDeviceAccessDenied   = errors.DefinePermissionDenied("device_access_denied", "access to device with `{dev_eui}` denied: device is already claimed or the owner token is invalid") //nolint:lll
	errUnauthenticated      = errors.DefineUnauthenticated("unauthenticated", "unauthenticated")
	errUnclaimDevice        = errors.Define("unclaim_device", "unclaim device with EUI `{dev_eui}`", "message")
	errUnclaimDevices       = errors.Define("unclaim_devices", "unclaim devices")
)

// Claim implements EndDeviceClaimer.
func (c *TTJS) Claim(ctx context.Context, joinEUI, devEUI types.EUI64, claimAuthenticationCode string) error {
	reqURL := fmt.Sprintf("%s/api/v2/devices/%s/claim", c.config.URL, devEUI.String())
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_eui", devEUI,
		"join_eui", joinEUI,
		"url", reqURL,
	))

	claimReq := ClaimRequest{
		OwnerToken: claimAuthenticationCode,
		Lock:       boolValue(true),
		HomeNetID:  c.config.NetID.String(),
		ASID:       c.config.ASID,
	}
	if c.config.NSID != nil {
		claimReq.HomeNSID = stringValue(c.config.NSID.String())
	}
	claimReq, err := claimReq.Apply(ctx, c)
	if err != nil {
		return err
	}
	buf, err := json.Marshal(claimReq)
	if err != nil {
		return err
	}

	logger.Debug("Claim end device")
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client, err := c.httpClient(ctx)
	if err != nil {
		return err
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if isSuccess(resp.StatusCode) {
		return nil
	}

	var errResp ErrorResponse
	err = json.Unmarshal(respBody, &errResp)
	if err != nil {
		logger.WithError(err).Warn("Failed to decode error message")
	} else {
		logger.WithField("error", errResp.Message).Warn("Failed to claim end device")
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return errBadRequest.WithAttributes("message", errResp.Message)
	case http.StatusNotFound:
		return errDeviceNotProvisioned.WithAttributes("dev_eui", devEUI)
	case http.StatusForbidden:
		return errDeviceAccessDenied.WithAttributes("dev_eui", devEUI)
	case http.StatusUnauthorized:
		return errUnauthenticated.New()
	default:
		return errors.FromHTTPStatusCode(resp.StatusCode)
	}
}

// Unclaim implements EndDeviceClaimer.
func (c *TTJS) Unclaim(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) error {
	devEUI := types.MustEUI64(ids.DevEui)
	reqURL := fmt.Sprintf("%s/api/v2/devices/%s/claim", c.config.URL, devEUI.String())
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_eui", devEUI,
		"join_eui", ids.JoinEui,
		"url", reqURL,
	))

	logger.Debug("Unclaim end device")
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	client, err := c.httpClient(ctx)
	if err != nil {
		return err
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if isSuccess(resp.StatusCode) {
		return nil
	}

	var errResp ErrorResponse
	err = json.Unmarshal(respBody, &errResp)
	if err != nil {
		logger.WithError(err).Warn("Failed to decode error message")
	} else {
		logger.WithField("error", errResp.Message).Warn("Failed to unclaim end device")
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return errBadRequest.WithAttributes("message", errResp.Message)
	case http.StatusNotFound:
		return errDeviceNotClaimed.WithAttributes("dev_eui", devEUI)
	case http.StatusForbidden:
		return errDeviceAccessDenied.WithAttributes("dev_eui", devEUI)
	case http.StatusUnauthorized:
		return errUnauthenticated.New()
	default:
		return errors.FromHTTPStatusCode(resp.StatusCode)
	}
}

// GetClaimStatus implements EndDeviceClaimer.
func (c *TTJS) GetClaimStatus(
	ctx context.Context, ids *ttnpb.EndDeviceIdentifiers,
) (*ttnpb.GetClaimStatusResponse, error) {
	devEUI := types.MustEUI64(ids.DevEui)
	reqURL := fmt.Sprintf("%s/api/v2/devices/%s/claim", c.config.URL, devEUI.String())
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_eui", devEUI,
		"join_eui", ids.JoinEui,
		"url", reqURL,
	))

	logger.Debug("Get claim status for end device")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	client, err := c.httpClient(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if isSuccess(resp.StatusCode) {
		var (
			claimData ClaimData
			ret       = ttnpb.GetClaimStatusResponse{
				EndDeviceIds: ids,
			}
			homeNSID  types.EUI64
			homeNetID types.NetID
		)
		err = json.Unmarshal(respBody, &claimData)
		if err != nil {
			return nil, err
		}
		err = homeNetID.UnmarshalText([]byte(claimData.HomeNetID))
		if err != nil {
			return nil, err
		}
		ret.HomeNetId = homeNetID.Bytes()
		if claimData.HomeNSID != nil {
			err = homeNSID.UnmarshalText([]byte(*claimData.HomeNSID))
			if err != nil {
				return nil, err
			}
			ret.HomeNsId = homeNSID.Bytes()
		}
		return &ret, nil
	}

	var errResp ErrorResponse
	err = json.Unmarshal(respBody, &errResp)
	if err != nil {
		logger.WithError(err).Warn("Failed to decode error message")
	} else {
		logger.WithField("error", errResp.Message).Warn("Failed to get claim status")
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return nil, errBadRequest.WithAttributes("message", errResp.Message)
	case http.StatusNotFound:
		return nil, errDeviceNotClaimed.WithAttributes("dev_eui", devEUI)
	case http.StatusForbidden:
		return nil, errDeviceAccessDenied.WithAttributes("dev_eui", devEUI)
	case http.StatusUnauthorized:
		return nil, errUnauthenticated.New()
	default:
		return nil, errors.FromHTTPStatusCode(resp.StatusCode)
	}
}

// BatchUnclaim implements EndDeviceClaimer.
func (c *TTJS) BatchUnclaim(
	ctx context.Context,
	ids []*ttnpb.EndDeviceIdentifiers,
) error {
	if len(ids) == 0 {
		return errBadRequest.WithAttributes("message", "no devices in request")
	}
	var euis string
	joinEUI := types.MustEUI64(ids[0].JoinEui).OrZero()
	for _, ids := range ids {
		devEUI := types.MustEUI64(ids.DevEui).OrZero()
		if euis != "" {
			euis += ","
		}
		euis += devEUI.String()
	}

	reqURL := fmt.Sprintf("%s/api/v2/devices/%s/claims", c.config.URL, euis)
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"join_eui", joinEUI,
		"url", reqURL,
	))

	logger.Debug("Unclaim end devices")
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	client, err := c.httpClient(ctx)
	if err != nil {
		return err
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if isSuccess(resp.StatusCode) {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		var errMsg struct {
			Message string `json:"message"`
		}
		err = json.Unmarshal(respBody, &errMsg)
		if err != nil {
			return errUnclaimDevices.WithCause(err)
		}
		if errMsg.Message != "" {
			return errBadRequest.WithAttributes("message", errMsg.Message)
		}

		var resp EndDevicesErrors
		err = json.Unmarshal(respBody, &resp)
		if err != nil {
			return errUnclaimDevices.WithCause(err)
		}
		ret := claimerrors.DeviceErrors{
			Errors: make(map[types.EUI64]errors.ErrorDetails, len(resp)),
		}
		for euiText, r := range resp {
			var eui types.EUI64
			err := eui.UnmarshalText([]byte(euiText))
			if err != nil {
				logger.WithError(err).WithField("eui", euiText).Warn("Failed to decode EUI")
				continue
			}
			ret.Errors[eui] = errUnclaimDevice.WithAttributes("message", r.Message)
		}
		return ret
	case http.StatusUnauthorized:
		return errUnauthenticated.New()
	default:
		return errors.FromHTTPStatusCode(resp.StatusCode)
	}
}

// isSuccess returns true if the HTTP status code is 2xx.
func isSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}
