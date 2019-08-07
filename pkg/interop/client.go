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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/afero"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
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

// JoinServerProtocol represents the protocol used for connection to Join Server by interop client.
type JoinServerProtocol uint8

const (
	LoRaWANJoinServerProtocol1_0 JoinServerProtocol = iota
	LoRaWANJoinServerProtocol1_1
)

// BackendInterfacesVersion returns the version of LoRaWAN Backend Interfaces specification version the protocol p is compliant with.
// BackendInterfacesVersion panics if p is not compliant with LoRaWAN Backend Interfaces specification.
func (p JoinServerProtocol) BackendInterfacesVersion() string {
	switch p {
	case LoRaWANJoinServerProtocol1_0:
		return "1.0"
	case LoRaWANJoinServerProtocol1_1:
		return "1.1"
	default:
		panic(fmt.Sprintf("Join Server protocol	`%v` is not compliant with Backend Interfaces specification", p))
	}
}

func (p *JoinServerProtocol) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	switch s {
	case "BI1.1":
		*p = LoRaWANJoinServerProtocol1_1
		return nil
	case "BI1.0":
		*p = LoRaWANJoinServerProtocol1_0
		return nil
	default:
		return errUnknownProtocol
	}
}

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

type joinServerHTTPClient struct {
	Client         http.Client
	NewRequestFunc func(joinEUI types.EUI64, pld interface{}) (*http.Request, error)
	Protocol       JoinServerProtocol
}

func (cl joinServerHTTPClient) exchange(joinEUI types.EUI64, req, res interface{}) error {
	httpReq, err := cl.NewRequestFunc(joinEUI, req)
	if err != nil {
		return err
	}

	httpRes, err := cl.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()
	return json.NewDecoder(httpRes.Body).Decode(res)
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

func (cl joinServerHTTPClient) GetAppSKey(ctx context.Context, asID string, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error) {
	interopAns := &AppSKeyAns{}
	if err := cl.exchange(req.JoinEUI, &AppSKeyReq{
		AsJsMessageHeader: AsJsMessageHeader{
			MessageHeader: MessageHeader{
				ProtocolVersion: cl.Protocol.BackendInterfacesVersion(),
				MessageType:     MessageTypeAppSKeyReq,
			},
			SenderID:   asID,
			ReceiverID: EUI64(req.JoinEUI),
		},
		DevEUI:       EUI64(req.DevEUI),
		SessionKeyID: Buffer(req.SessionKeyID),
	}, interopAns); err != nil {
		return nil, err
	}
	if err := parseResult(interopAns.Result); err != nil {
		return nil, err
	}

	return &ttnpb.AppSKeyResponse{
		AppSKey: ttnpb.KeyEnvelope(interopAns.AppSKey),
	}, nil
}

func (cl joinServerHTTPClient) HandleJoinRequest(ctx context.Context, netID types.NetID, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	pld := req.Payload.GetJoinRequestPayload()
	if pld == nil {
		return nil, ErrMalformedMessage
	}

	dlSettings, err := lorawan.MarshalDLSettings(req.DownlinkSettings)
	if err != nil {
		return nil, err
	}

	var cfList []byte
	if req.CFList != nil {
		cfList, err = lorawan.MarshalCFList(*req.CFList)
		if err != nil {
			return nil, err
		}
	}

	interopAns := &JoinAns{}
	if err := cl.exchange(pld.JoinEUI, &JoinReq{
		NsJsMessageHeader: NsJsMessageHeader{
			MessageHeader: MessageHeader{
				ProtocolVersion: cl.Protocol.BackendInterfacesVersion(),
				MessageType:     MessageTypeJoinReq,
			},
			SenderID:   NetID(netID),
			ReceiverID: EUI64(pld.JoinEUI),
			SenderNSID: NetID(netID),
		},
		MACVersion: MACVersion(req.SelectedMACVersion),
		PHYPayload: Buffer(req.RawPayload),
		DevEUI:     EUI64(pld.DevEUI),
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
	if req.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) <= 0 {
		fNwkSIntKey = interopAns.NwkSKey
	}
	return &ttnpb.JoinResponse{
		RawPayload: interopAns.PHYPayload,
		SessionKeys: ttnpb.SessionKeys{
			SessionKeyID: []byte(interopAns.SessionKeyID),
			FNwkSIntKey:  (*ttnpb.KeyEnvelope)(fNwkSIntKey),
			SNwkSIntKey:  (*ttnpb.KeyEnvelope)(interopAns.SNwkSIntKey),
			NwkSEncKey:   (*ttnpb.KeyEnvelope)(interopAns.NwkSEncKey),
			AppSKey:      (*ttnpb.KeyEnvelope)(interopAns.AppSKey),
		},
		Lifetime: time.Duration(interopAns.Lifetime) * time.Second,
	}, nil
}

func makeJoinServerHTTPRequestFunc(scheme string, dns, fqdn, path string, port uint32, headers map[string]string) func(types.EUI64, interface{}) (*http.Request, error) {
	if port == 0 {
		port = defaultHTTPSPort
	}
	if path != "" {
		path = fmt.Sprintf("/%s", path)
	}
	return func(joinEUI types.EUI64, pld interface{}) (*http.Request, error) {
		if fqdn == "" {
			fqdn = JoinServerFQDN(joinEUI, dns)
		}
		return newHTTPRequest(serverURL(scheme, fqdn, path, port), pld, headers)
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

var errUnknownProtocol = errors.DefineInvalidArgument("unknown_protocol", "unknown protocol")

// ClientConfig represents the client-side interoperability through LoRaWAN Backend Interfaces configuration.
type ClientConfig struct {
	ConfigURI string        `name:"config-uri" description:"URI of the configuration file"`
	CacheTime time.Duration `name:"cache-time"`
}

var errUnsupportedURIScheme = errors.DefineUnimplemented("unsupported_uri_scheme", "URI scheme `{scheme}` is not supported", "scheme")

type tlsConfig struct {
	RootCA      string `yaml:"root-ca"`
	Certificate string `yaml:"certificate"`
	Key         string `yaml:"key"`
}

func (conf tlsConfig) IsZero() bool {
	return conf == (tlsConfig{})
}

func openFile(fs afero.Fs, basePath, path string) (afero.File, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(basePath, path)
	}
	return fs.Open(path)
}

func readFile(fs afero.Fs, basePath, path string) ([]byte, error) {
	f, err := openFile(fs, basePath, path)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

func (conf tlsConfig) TLSConfig(fs afero.Fs, confDir string) (*tls.Config, error) {
	var rootCAs *x509.CertPool
	if conf.RootCA != "" {
		caFile, err := openFile(fs, confDir, conf.RootCA)
		if err != nil {
			return nil, err
		}

		caPEM, err := ioutil.ReadAll(caFile)
		if err != nil {
			return nil, err
		}
		rootCAs = x509.NewCertPool()
		rootCAs.AppendCertsFromPEM(caPEM)
	}

	var getCert func(*tls.CertificateRequestInfo) (*tls.Certificate, error)
	if conf.Certificate != "" || conf.Key != "" {
		getCert = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			certFile, err := openFile(fs, confDir, conf.Certificate)
			if err != nil {
				return nil, err
			}
			certPEM, err := ioutil.ReadAll(certFile)
			if err != nil {
				return nil, err
			}

			keyFile, err := openFile(fs, confDir, conf.Key)
			if err != nil {
				return nil, err
			}
			keyPEM, err := ioutil.ReadAll(keyFile)
			if err != nil {
				return nil, err
			}

			cert, err := tls.X509KeyPair(certPEM, keyPEM)
			if err != nil {
				return nil, err
			}
			return &cert, nil
		}
	}
	return &tls.Config{
		RootCAs:              rootCAs,
		GetClientCertificate: getCert,
	}, nil
}

func decodeYAML(r io.Reader, v interface{}) error {
	dec := yaml.NewDecoder(r)
	dec.SetStrict(true)
	return dec.Decode(v)
}

func NewClient(ctx context.Context, conf ClientConfig, fallbackTLS *tls.Config) (*Client, error) {
	u, err := url.Parse(conf.ConfigURI)
	if err != nil {
		return nil, err
	}

	var fs afero.Fs
	switch s := u.Scheme; s {
	case "file":
		fs = afero.NewOsFs()
	default:
		return nil, errUnsupportedURIScheme.WithAttributes("scheme", s)
	}
	confFile, err := fs.Open(u.Path)
	if err != nil {
		return nil, err
	}
	confDir := filepath.Dir(u.Path)
	fs = afero.NewCacheOnReadFs(fs, afero.NewMemMapFs(), conf.CacheTime)

	var yamlConf struct {
		JoinServers []struct {
			File    string              `yaml:"file"`
			JoinEUI []types.EUI64Prefix `yaml:"join-eui"`
		} `yaml:"join-servers"`
	}
	if err := decodeYAML(confFile, &yamlConf); err != nil {
		return nil, err
	}

	jss := make([]prefixJoinServerClient, 0, len(yamlConf.JoinServers))
	for _, jsConf := range yamlConf.JoinServers {
		jsFile, err := openFile(fs, confDir, jsConf.File)
		if err != nil {
			return nil, err
		}

		var yamlJSConf struct {
			DNS      string             `yaml:"dns"`
			FQDN     string             `yaml:"fqdn"`
			Path     string             `yaml:"path"`
			Port     uint32             `yaml:"port"`
			Protocol JoinServerProtocol `yaml:"protocol"`
			Headers  map[string]string  `yaml:"headers"`
			TLS      tlsConfig          `yaml:"tls"`
		}
		if err := decodeYAML(jsFile, &yamlJSConf); err != nil {
			return nil, err
		}

		var js joinServerClient
		switch yamlJSConf.Protocol {
		case LoRaWANJoinServerProtocol1_0, LoRaWANJoinServerProtocol1_1:
			tlsConf := fallbackTLS
			if !yamlJSConf.TLS.IsZero() {
				tlsConf, err = yamlJSConf.TLS.TLSConfig(fs, confDir)
				if err != nil {
					return nil, err
				}
			}

			var tr *http.Transport
			if tlsConf != nil {
				tr = &http.Transport{
					TLSClientConfig: tlsConf,
				}
			}
			js = &joinServerHTTPClient{
				Client: http.Client{
					Transport: tr,
				},
				NewRequestFunc: makeJoinServerHTTPRequestFunc("https", yamlJSConf.DNS, yamlJSConf.FQDN, yamlJSConf.Path, yamlJSConf.Port, yamlJSConf.Headers),
				Protocol:       yamlJSConf.Protocol,
			}
		default:
			return nil, errUnknownProtocol
		}
		for _, pre := range jsConf.JoinEUI {
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

func (cl Client) GetAppSKey(ctx context.Context, asID string, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error) {
	js, ok := cl.joinServer(req.JoinEUI)
	if !ok {
		return nil, errNotRegistered
	}
	return js.GetAppSKey(ctx, asID, req)
}

func (cl Client) HandleJoinRequest(ctx context.Context, netID types.NetID, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	pld := req.Payload.GetJoinRequestPayload()
	if pld == nil {
		return nil, ErrMalformedMessage
	}
	js, ok := cl.joinServer(pld.JoinEUI)
	if !ok {
		return nil, errNotRegistered
	}
	return js.HandleJoinRequest(ctx, netID, req)
}
