// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	yaml "gopkg.in/yaml.v2"
)

const (
	// loRaAllianceDomain is the domain of LoRa Alliance.
	loRaAllianceDomain = "lora-alliance.org"

	// LoRaAllianceJoinEUIDomain is the LoRa Alliance domain used for JoinEUI resolution.
	LoRaAllianceJoinEUIDomain = "joineuis." + loRaAllianceDomain

	// LoRaAllianceNetIDDomain is the LoRa Alliance domain used for NetID resolution.
	LoRaAllianceNetIDDomain = "netids." + loRaAllianceDomain

	defaultHTTPSPort = 443
)

type jsRPCPaths struct {
	Join    string `yaml:"join"`
	Rejoin  string `yaml:"rejoin"`
	AppSKey string `yaml:"app-s-key"`
	HomeNS  string `yaml:"home-ns"`
}

func (p jsRPCPaths) join() string    { return p.Join }
func (p jsRPCPaths) appSKey() string { return p.AppSKey }

func serverURL(scheme, fqdn, path string, port uint32) string {
	if scheme == "" {
		scheme = "https"
	}
	if port == 0 {
		port = defaultHTTPSPort
	}
	if path != "" {
		path = fmt.Sprintf("/%s", path)
	}
	return fmt.Sprintf("%s://%s:%d%s", scheme, fqdn, port, path)
}

func newHTTPRequest(url string, pld interface{}, headers map[string]string) (*http.Request, error) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(pld); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

// JoinServerFQDN constructs Join Server FQDN using specified EUI under domain
// according to LoRaWAN Backend Interfaces specification.
// If domain is empty, LoRaAllianceJoinEUIDomain is used.
func JoinServerFQDN(eui types.EUI64, domain string) string {
	if domain == "" {
		domain = LoRaAllianceJoinEUIDomain
	}
	return fmt.Sprintf(
		"%01x.%01x.%01x.%01x.%01x.%01x.%01x.%01x.%01x.%01x.%01x.%01x.%01x.%01x.%01x.%01x.%s",
		eui[7]&0x0f, eui[7]>>4,
		eui[6]&0x0f, eui[6]>>4,
		eui[5]&0x0f, eui[5]>>4,
		eui[4]&0x0f, eui[4]>>4,
		eui[3]&0x0f, eui[3]>>4,
		eui[2]&0x0f, eui[2]>>4,
		eui[1]&0x0f, eui[1]>>4,
		eui[0]&0x0f, eui[0]>>4,
		domain,
	)
}

func httpExchange(ctx context.Context, httpReq *http.Request, res interface{}, do func(*http.Request) (*http.Response, error)) error {
	logger := log.FromContext(ctx).WithField("url", httpReq.URL)

	logger.Debug("Send interop HTTP request")
	httpRes, err := do(httpReq)
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()

	logger = logger.WithField("http_code", httpRes.StatusCode)
	logger.Debug("Receive interop HTTP response")

	b, err := io.ReadAll(httpRes.Body)
	if err != nil {
		if err == io.EOF && res == nil {
			return nil
		}
		logger.WithError(err).Warn("Failed to read HTTP response body")
		return errors.FromHTTPStatusCode(httpRes.StatusCode)
	}

	if err := json.Unmarshal(b, res); err != nil {
		logger.WithError(err).Warn("Failed to decode HTTP response body")
		return errors.FromHTTPStatusCode(httpRes.StatusCode)
	}
	return nil
}

type joinServerHTTPClient struct {
	Client         *http.Client
	NewRequestFunc func(types.EUI64, func(jsRPCPaths) string, interface{}) (*http.Request, error)
	Protocol       ProtocolVersion
}

func (cl joinServerHTTPClient) exchange(ctx context.Context, joinEUI types.EUI64, pathFunc func(jsRPCPaths) string, req, res interface{}) error {
	httpReq, err := cl.NewRequestFunc(joinEUI, pathFunc, req)
	if err != nil {
		return err
	}
	return httpExchange(ctx, httpReq.WithContext(ctx), res, cl.Client.Do)
}

func parseResult(r Result) error {
	if r.ResultCode == ResultSuccess {
		return nil
	}

	err, ok := resultErrors[r.ResultCode]
	if ok {
		return err
	}
	return errUnexpectedResult.WithAttributes("code", r.ResultCode)
}

// GetAppSKey performs AppSKey request according to LoRaWAN Backend Interfaces specification.
func (cl joinServerHTTPClient) GetAppSKey(ctx context.Context, asID string, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error) {
	interopAns := &AppSKeyAns{}
	if err := cl.exchange(ctx, req.JoinEui, jsRPCPaths.appSKey, &AppSKeyReq{
		AsJsMessageHeader: AsJsMessageHeader{
			MessageHeader: MessageHeader{
				ProtocolVersion: cl.Protocol,
				MessageType:     MessageTypeAppSKeyReq,
			},
			SenderID:   asID,
			ReceiverID: EUI64(req.JoinEui),
		},
		DevEUI:       EUI64(req.DevEui),
		SessionKeyID: Buffer(req.SessionKeyId),
	}, interopAns); err != nil {
		return nil, err
	}
	if err := parseResult(interopAns.Result); err != nil {
		return nil, err
	}

	return &ttnpb.AppSKeyResponse{
		AppSKey: (*ttnpb.KeyEnvelope)(&interopAns.AppSKey),
	}, nil
}

var (
	errNoJoinRequestPayload = errors.DefineInvalidArgument("no_join_request_payload", "no join-request payload")
	errGenerateSessionKeyID = errors.Define("generate_session_key_id", "failed to generate session key ID")

	generatedSessionKeyIDPrefix = []byte("ttn-lw-interop-generated:")
)

// HandleJoinRequest performs Join request according to LoRaWAN Backend Interfaces specification.
func (cl joinServerHTTPClient) HandleJoinRequest(ctx context.Context, netID types.NetID, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	pld := req.Payload.GetJoinRequestPayload()
	if pld == nil {
		return nil, errNoJoinRequestPayload.New()
	}

	dlSettings, err := lorawan.MarshalDLSettings(req.DownlinkSettings)
	if err != nil {
		return nil, err
	}

	var cfList []byte
	if req.CfList != nil {
		cfList, err = lorawan.MarshalCFList(*req.CfList)
		if err != nil {
			return nil, err
		}
	}

	interopAns := &JoinAns{}
	if err := cl.exchange(ctx, pld.JoinEui, jsRPCPaths.join, &JoinReq{
		NsJsMessageHeader: NsJsMessageHeader{
			MessageHeader: MessageHeader{
				ProtocolVersion: cl.Protocol,
				MessageType:     MessageTypeJoinReq,
			},
			SenderID:   NetID(netID),
			ReceiverID: EUI64(pld.JoinEui),
		},
		MACVersion: MACVersion(req.SelectedMacVersion),
		PHYPayload: Buffer(req.RawPayload),
		DevEUI:     EUI64(pld.DevEui),
		DevAddr:    DevAddr(req.DevAddr),
		DLSettings: Buffer(dlSettings),
		RxDelay:    req.RxDelay,
		CFList:     Buffer(cfList),
	}, interopAns); err != nil {
		return nil, err
	}
	if err := parseResult(interopAns.Result); err != nil {
		return nil, err
	}

	fNwkSIntKey := interopAns.FNwkSIntKey
	if req.SelectedMacVersion.Compare(ttnpb.MACVersion_MAC_V1_1) <= 0 {
		fNwkSIntKey = interopAns.NwkSKey
	}

	sessionKeyID := []byte(interopAns.SessionKeyID)
	if len(sessionKeyID) == 0 {
		log.FromContext(ctx).Debug("Interop join-accept does not contain session key ID, generate random ID")
		id, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
		if err != nil {
			return nil, errGenerateSessionKeyID.New()
		}
		sessionKeyID = make([]byte, 0, len(generatedSessionKeyIDPrefix)+len(id))
		sessionKeyID = append(sessionKeyID, generatedSessionKeyIDPrefix...)
		sessionKeyID = append(sessionKeyID, id[:]...)
	}
	return &ttnpb.JoinResponse{
		RawPayload: interopAns.PHYPayload,
		SessionKeys: &ttnpb.SessionKeys{
			SessionKeyId: sessionKeyID,
			FNwkSIntKey:  (*ttnpb.KeyEnvelope)(fNwkSIntKey),
			SNwkSIntKey:  (*ttnpb.KeyEnvelope)(interopAns.SNwkSIntKey),
			NwkSEncKey:   (*ttnpb.KeyEnvelope)(interopAns.NwkSEncKey),
			AppSKey:      (*ttnpb.KeyEnvelope)(interopAns.AppSKey),
		},
		Lifetime: ttnpb.ProtoDurationPtr(time.Duration(interopAns.Lifetime) * time.Second),
	}, nil
}

// GeneratedSessionKeyID returns whether the session key ID is generated locally and not by the Join Server.
func GeneratedSessionKeyID(id []byte) bool {
	return bytes.HasPrefix(id, generatedSessionKeyIDPrefix)
}

func makeJoinServerHTTPRequestFunc(scheme, dns, fqdn string, port uint32, rpcPaths jsRPCPaths, headers map[string]string) func(types.EUI64, func(jsRPCPaths) string, interface{}) (*http.Request, error) {
	if port == 0 {
		port = defaultHTTPSPort
	}
	return func(joinEUI types.EUI64, pathFunc func(jsRPCPaths) string, pld interface{}) (*http.Request, error) {
		fqdn := fqdn
		if fqdn == "" {
			fqdn = JoinServerFQDN(joinEUI, dns)
		}
		return newHTTPRequest(serverURL(scheme, fqdn, pathFunc(rpcPaths), port), pld, headers)
	}
}

type joinServerClient interface {
	HandleJoinRequest(ctx context.Context, netID types.NetID, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error)
	GetAppSKey(ctx context.Context, asID string, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error)
}

type prefixJoinServerClient struct {
	joinServerClient
	prefix types.EUI64Prefix
}

type Client struct {
	joinServers []prefixJoinServerClient // Sorted by JoinEUI prefix range length.
}

var errUnknownConfig = errors.DefineNotFound("unknown_config", "configuration is unknown")

// NewClient return new interop client.
func NewClient(ctx context.Context, conf config.InteropClient, httpClientProvider httpclient.Provider) (*Client, error) {
	fetcher, err := conf.Fetcher(ctx, httpClientProvider)
	if err != nil {
		return nil, err
	}
	if fetcher == nil {
		return nil, errUnknownConfig.New()
	}
	confFileBytes, err := fetcher.File(InteropClientConfigurationName)
	if err != nil {
		return nil, err
	}

	var yamlConf struct {
		JoinServers []struct {
			File     string              `yaml:"file"`
			JoinEUIs []types.EUI64Prefix `yaml:"join-euis"`
		} `yaml:"join-servers"`
	}
	if err := yaml.UnmarshalStrict(confFileBytes, &yamlConf); err != nil {
		return nil, err
	}

	type ComponentConfig struct {
		DNS     string            `yaml:"dns"`
		FQDN    string            `yaml:"fqdn"`
		Port    uint32            `yaml:"port"`
		Headers map[string]string `yaml:"headers"`
		TLS     tlsConfig         `yaml:"tls"`
	}

	jss := make([]prefixJoinServerClient, 0, len(yamlConf.JoinServers))
	for _, jsConf := range yamlConf.JoinServers {
		jsConfEls := strings.Split(filepath.ToSlash(jsConf.File), "/")

		fetcher := fetch.WithBasePath(fetcher, jsConfEls[:len(jsConfEls)-1]...)
		jsFileBytes, err := fetcher.File(jsConfEls[len(jsConfEls)-1])
		if err != nil {
			return nil, err
		}

		var yamlJSConf struct {
			ComponentConfig `yaml:",inline"`
			Paths           jsRPCPaths      `yaml:"paths"`
			Protocol        ProtocolVersion `yaml:"protocol"`
		}
		if err := yaml.UnmarshalStrict(jsFileBytes, &yamlJSConf); err != nil {
			return nil, err
		}

		var js joinServerClient
		switch yamlJSConf.Protocol {
		case ProtocolV1_0, ProtocolV1_1:
			var opts []httpclient.Option
			if !yamlJSConf.TLS.IsZero() {
				tlsConf, err := yamlJSConf.TLS.TLSConfig(fetcher)
				if err != nil {
					return nil, err
				}
				opts = append(opts, httpclient.WithTLSConfig(tlsConf))
			}

			httpClient, err := httpClientProvider.HTTPClient(ctx, opts...)
			if err != nil {
				return nil, err
			}

			js = &joinServerHTTPClient{
				Client:         httpClient,
				NewRequestFunc: makeJoinServerHTTPRequestFunc("https", yamlJSConf.DNS, yamlJSConf.FQDN, yamlJSConf.Port, yamlJSConf.Paths, yamlJSConf.Headers),
				Protocol:       yamlJSConf.Protocol,
			}
		default:
			return nil, errUnknownProtocol.New()
		}
		for _, pre := range jsConf.JoinEUIs {
			jss = append(jss, prefixJoinServerClient{
				joinServerClient: js,
				prefix:           pre,
			})
		}
	}
	sort.Slice(jss, func(i, j int) bool {
		pi, pj := jss[i].prefix, jss[j].prefix
		if pi.Length != pj.Length {
			return pi.Length > pj.Length
		}
		return pi.EUI64.MarshalNumber() > pj.EUI64.MarshalNumber()
	})
	return &Client{
		joinServers: jss,
	}, nil
}

func (cl Client) joinServer(joinEUI types.EUI64) (joinServerClient, bool) {
	// NOTE: joinServers slice is sorted by prefix length and the range start decreasing, hence the first match is the most specific one.
	for _, js := range cl.joinServers {
		if js.prefix.Matches(joinEUI) {
			return js.joinServerClient, true
		}
	}
	return nil, false
}

// GetAppSKey performs AppSKey request to Join Server associated with req.JoinEUI.
func (cl Client) GetAppSKey(ctx context.Context, asID string, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error) {
	js, ok := cl.joinServer(req.JoinEui)
	if !ok {
		return nil, errNotRegistered.New()
	}
	return js.GetAppSKey(ctx, asID, req)
}

// HandleJoinRequest performs Join request to Join Server associated with req.JoinEUI.
func (cl Client) HandleJoinRequest(ctx context.Context, netID types.NetID, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	pld := req.Payload.GetJoinRequestPayload()
	if pld == nil {
		return nil, errNoJoinRequestPayload.New()
	}
	js, ok := cl.joinServer(pld.JoinEui)
	if !ok {
		return nil, errNotRegistered.New()
	}
	return js.HandleJoinRequest(ctx, netID, req)
}
