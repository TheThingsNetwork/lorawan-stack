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

// Package joinserver provides a LoRaWAN-compliant Join Server implementation.
package joinserver

import (
	"bytes"
	"context"
	"crypto/x509/pkix"
	"encoding/binary"
	"io"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/oklog/ulid"
	"go.thethings.network/lorawan-stack/pkg/auth"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/interop"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
)

// Config represents the JoinServer configuration.
type Config struct {
	Devices         DeviceRegistry       `name:"-"`
	Keys            KeyRegistry          `name:"-"`
	JoinEUIPrefixes []*types.EUI64Prefix `name:"join-eui-prefix" description:"JoinEUI prefixes handled by this JS"`
}

// JoinServer implements the Join Server component.
//
// The Join Server exposes the NsJs and DeviceRegistry services.
type JoinServer struct {
	*component.Component
	ctx context.Context

	devices DeviceRegistry
	keys    KeyRegistry

	euiPrefixes []*types.EUI64Prefix

	entropyMu *sync.Mutex
	entropy   io.Reader

	grpc struct {
		nsJs      nsJsServer
		asJs      asJsServer
		jsDevices jsEndDeviceRegistryServer
		js        jsServer
	}
	interop interopServer
}

// Context returns the context of the Join Server.
func (js *JoinServer) Context() context.Context {
	return js.ctx
}

// New returns new *JoinServer.
func New(c *component.Component, conf *Config) (*JoinServer, error) {
	js := &JoinServer{
		Component: c,
		ctx:       log.NewContextWithField(c.Context(), "namespace", "joinserver"),

		devices: conf.Devices,
		keys:    conf.Keys,

		euiPrefixes: conf.JoinEUIPrefixes,

		entropyMu: &sync.Mutex{},
		entropy:   ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0),
	}

	js.grpc.jsDevices = jsEndDeviceRegistryServer{JS: js}
	js.grpc.asJs = asJsServer{JS: js}
	js.grpc.nsJs = nsJsServer{JS: js}
	js.grpc.js = jsServer{JS: js}
	js.interop = interopServer{JS: js}

	// TODO: Support authentication from non-cluster-local NS and AS (https://github.com/TheThingsNetwork/lorawan-stack/issues/4).
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsJs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("joinserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsJs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("joinserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.Js", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("joinserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsJs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsJs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.Js", cluster.HookName, c.ClusterAuthUnaryHook())

	c.RegisterGRPC(js)
	c.RegisterInterop(js)
	return js, nil
}

// Roles of the gRPC service.
func (js *JoinServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_JOIN_SERVER}
}

// RegisterServices registers services provided by js at s.
func (js *JoinServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterAsJsServer(s, js.grpc.asJs)
	ttnpb.RegisterNsJsServer(s, js.grpc.nsJs)
	ttnpb.RegisterJsEndDeviceRegistryServer(s, js.grpc.jsDevices)
	ttnpb.RegisterJsServer(s, js.grpc.js)
}

// RegisterHandlers registers gRPC handlers.
func (js *JoinServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterJsHandler(js.Context(), s, conn)
	ttnpb.RegisterJsEndDeviceRegistryHandler(js.Context(), s, conn)
}

// RegisterInterop registers the NS-JS and AS-JS interop services.
func (js *JoinServer) RegisterInterop(srv *interop.Server) {
	srv.RegisterJS(js.interop)
}

var supportedMACVersions = [...]ttnpb.MACVersion{
	ttnpb.MAC_V1_0,
	ttnpb.MAC_V1_0_1,
	ttnpb.MAC_V1_0_2,
	ttnpb.MAC_V1_0_3,
	ttnpb.MAC_V1_1,
}

func validateCaller(dn pkix.Name, addr string) error {
	if addr := strings.ToLower(addr); addr != "" && addr != dn.CommonName {
		return errAddressNotAuthorized.WithAttributes("address", dn.CommonName)
	}
	return nil
}

// wrapKeyIfKEKExists wraps the given key with the KEK label.
// If the configured key vault cannot find the KEK, the key is returned in the clear.
func (js *JoinServer) wrapKeyIfKEKExists(key types.AES128Key, kekLabel string) (*ttnpb.KeyEnvelope, error) {
	env, err := cryptoutil.WrapAES128Key(key, kekLabel, js.KeyVault)
	if err != nil {
		if errors.IsNotFound(err) {
			return &ttnpb.KeyEnvelope{
				Key: &key,
			}, nil
		}
		return nil, err
	}
	return &env, nil
}

// HandleJoin handles the given join-request.
func (js *JoinServer) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest) (res *ttnpb.JoinResponse, err error) {
	if _, ok := auth.X509DNFromContext(ctx); !ok {
		if err := clusterauth.Authorized(ctx); err != nil {
			return nil, err
		}
	}

	logger := log.FromContext(ctx)
	defer func() {
		if err != nil {
			registerRejectJoin(ctx, req, err)
		}
	}()

	supported := false
	for _, ver := range supportedMACVersions {
		if req.SelectedMACVersion == ver {
			supported = true
			break
		}
	}
	if !supported {
		return nil, errUnsupportedMACVersion.WithAttributes("version", req.SelectedMACVersion)
	}

	req.Payload = &ttnpb.Message{}
	if err = lorawan.UnmarshalMessage(req.RawPayload, req.Payload); err != nil {
		return nil, errDecodePayload.WithCause(err)
	}

	if req.Payload.Major != ttnpb.Major_LORAWAN_R1 {
		return nil, errUnsupportedLoRaWANMajorVersion.WithAttributes("major", req.Payload.Major)
	}
	if req.Payload.MType != ttnpb.MType_JOIN_REQUEST {
		return nil, errWrongPayloadType.WithAttributes("type", req.Payload.MType)
	}

	pld := req.Payload.GetJoinRequestPayload()
	if pld == nil {
		return nil, errNoJoinRequest
	}
	if pld.DevEUI.IsZero() {
		return nil, errNoDevEUI
	}
	if pld.JoinEUI.IsZero() {
		return nil, errNoJoinEUI
	}

	match := false
	for _, p := range js.euiPrefixes {
		if p.Matches(pld.JoinEUI) {
			match = true
			break
		}
	}
	switch {
	case !match && req.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) < 0:
		return nil, errUnknownAppEUI
	case !match:
		// TODO: Determine the cluster containing the device.
		// https://github.com/TheThingsNetwork/lorawan-stack/issues/4
		return nil, errForwardJoinRequest
	}

	dev, err := js.devices.SetByEUI(ctx, pld.JoinEUI, pld.DevEUI,
		[]string{
			"application_server_address",
			"last_dev_nonce",
			"last_join_nonce",
			"net_id",
			"network_server_address",
			"provisioner_id",
			"provisioning_data",
			"resets_join_nonces",
			"root_keys",
			"used_dev_nonces",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dn, ok := auth.X509DNFromContext(ctx); ok {
				if dev.NetID == nil {
					return nil, nil, errNoNetID
				}
				if !req.NetID.Equal(*dev.NetID) {
					return nil, nil, errNetIDMismatch.WithAttributes("net_id", req.NetID)
				}
				if err := validateCaller(dn, dev.NetworkServerAddress); err != nil {
					return nil, nil, err
				}
			}

			paths := make([]string, 0, 3)

			dn := uint32(binary.BigEndian.Uint16(pld.DevNonce[:]))
			if req.SelectedMACVersion.Compare(ttnpb.MAC_V1_0_3) >= 0 {
				if (dn != 0 || dev.LastDevNonce != 0 || dev.LastJoinNonce != 0) && !dev.ResetsJoinNonces {
					if dn <= dev.LastDevNonce {
						return nil, nil, errDevNonceTooSmall
					}
					if dn == math.MaxUint32 {
						return nil, nil, errDevNonceTooHigh
					}
				}
				dev.LastDevNonce = dn
				paths = append(paths, "last_dev_nonce")
			} else {
				i := sort.Search(len(dev.UsedDevNonces), func(i int) bool { return dev.UsedDevNonces[i] >= dn })
				if i < len(dev.UsedDevNonces) && dev.UsedDevNonces[i] == dn {
					return nil, nil, errReuseDevNonce
				}
				dev.UsedDevNonces = append(dev.UsedDevNonces, 0)
				copy(dev.UsedDevNonces[i+1:], dev.UsedDevNonces[i:])
				dev.UsedDevNonces[i] = dn
				paths = append(paths, "used_dev_nonces")
			}

			var b []byte
			if req.CFList == nil {
				b = make([]byte, 0, 17)
			} else {
				b = make([]byte, 0, 33)
			}
			b, err = lorawan.AppendMHDR(b, ttnpb.MHDR{
				MType: ttnpb.MType_JOIN_ACCEPT,
				Major: req.Payload.Major,
			})
			if err != nil {
				return nil, nil, errEncodePayload.WithCause(err)
			}

			if dev.LastJoinNonce >= 1<<24-1 {
				return nil, nil, errJoinNonceTooHigh
			}
			dev.LastJoinNonce++
			paths = append(paths, "last_join_nonce")

			var jn types.JoinNonce
			nb := make([]byte, 4)
			binary.BigEndian.PutUint32(nb, dev.LastJoinNonce)
			copy(jn[:], nb[1:])

			b, err = lorawan.AppendJoinAcceptPayload(b, ttnpb.JoinAcceptPayload{
				NetID:      req.NetID,
				JoinNonce:  jn,
				CFList:     req.CFList,
				DevAddr:    req.DevAddr,
				DLSettings: req.DownlinkSettings,
				RxDelay:    req.RxDelay,
			})
			if err != nil {
				return nil, nil, errEncodePayload.WithCause(err)
			}

			js.entropyMu.Lock()
			skID, err := ulid.New(ulid.Timestamp(time.Now()), js.entropy)
			js.entropyMu.Unlock()
			if err != nil {
				return nil, nil, errGenerateSessionKeyID
			}

			cs := js.GetPeer(ctx, ttnpb.PeerInfo_CRYPTO_SERVER, dev.EndDeviceIdentifiers)

			var networkCryptoService cryptoservices.Network
			if req.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) >= 0 && dev.RootKeys != nil && dev.RootKeys.NwkKey != nil {
				// LoRaWAN 1.1 and higher use a NwkKey.
				nwkKey, err := cryptoutil.UnwrapAES128Key(*dev.RootKeys.NwkKey, js.KeyVault)
				if err != nil {
					return nil, nil, err
				}
				networkCryptoService = cryptoservices.NewMemory(&nwkKey, nil)
			} else if cs != nil {
				networkCryptoService = cryptoservices.NewNetworkRPCClient(cs.Conn(), js.KeyVault, js.WithClusterAuth())
			}

			var applicationCryptoService cryptoservices.Application
			if dev.RootKeys != nil && dev.RootKeys.AppKey != nil {
				appKey, err := cryptoutil.UnwrapAES128Key(*dev.RootKeys.AppKey, js.KeyVault)
				if err != nil {
					return nil, nil, err
				}
				applicationCryptoService = cryptoservices.NewMemory(nil, &appKey)
				if req.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) < 0 {
					// LoRaWAN 1.0.x use the AppKey for network security operations.
					networkCryptoService = cryptoservices.NewMemory(nil, &appKey)
				}
			} else if cs != nil {
				applicationCryptoService = cryptoservices.NewApplicationRPCClient(cs.Conn(), js.KeyVault, js.WithClusterAuth())
			}
			if networkCryptoService == nil {
				return nil, nil, errNoNwkKey
			}
			if applicationCryptoService == nil {
				return nil, nil, errNoAppKey
			}

			cryptoDev := &ttnpb.EndDevice{}
			if err := cryptoDev.SetFields(dev, "ids", "provisioner_id", "provisioning_data"); err != nil {
				return nil, nil, err
			}
			reqMIC, err := networkCryptoService.JoinRequestMIC(ctx, cryptoDev, req.SelectedMACVersion, req.RawPayload[:19])
			if err != nil {
				return nil, nil, errComputeMIC.WithCause(err)
			}
			if !bytes.Equal(reqMIC[:], req.RawPayload[19:]) {
				return nil, nil, errMICMismatch
			}
			resMIC, err := networkCryptoService.JoinAcceptMIC(ctx, cryptoDev, req.SelectedMACVersion, 0xff, pld.DevNonce, b)
			if err != nil {
				return nil, nil, errComputeMIC.WithCause(err)
			}
			enc, err := networkCryptoService.EncryptJoinAccept(ctx, cryptoDev, req.SelectedMACVersion, append(b[1:], resMIC[:]...))
			if err != nil {
				return nil, nil, errEncryptPayload.WithCause(err)
			}
			nwkSKeys, err := networkCryptoService.DeriveNwkSKeys(ctx, cryptoDev, req.SelectedMACVersion, jn, pld.DevNonce, req.NetID)
			if err != nil {
				return nil, nil, errDeriveNwkSKeys.WithCause(err)
			}
			appSKey, err := applicationCryptoService.DeriveAppSKey(ctx, cryptoDev, req.SelectedMACVersion, jn, pld.DevNonce, req.NetID)
			if err != nil {
				return nil, nil, errDeriveAppSKey.WithCause(err)
			}

			sessionKeys := ttnpb.SessionKeys{
				SessionKeyID: skID[:],
			}
			sessionKeys.FNwkSIntKey, err = js.wrapKeyIfKEKExists(nwkSKeys.FNwkSIntKey, js.KeyVault.NsKEKLabel(ctx, dev.NetID, dev.NetworkServerAddress))
			if err != nil {
				return nil, nil, errWrapKey.WithCause(err)
			}
			sessionKeys.AppSKey, err = js.wrapKeyIfKEKExists(appSKey, js.KeyVault.AsKEKLabel(ctx, dev.ApplicationServerAddress))
			if err != nil {
				return nil, nil, errWrapKey.WithCause(err)
			}
			if req.SelectedMACVersion >= ttnpb.MAC_V1_1 {
				sessionKeys.SNwkSIntKey, err = js.wrapKeyIfKEKExists(nwkSKeys.SNwkSIntKey, js.KeyVault.NsKEKLabel(ctx, dev.NetID, dev.NetworkServerAddress))
				if err != nil {
					return nil, nil, errWrapKey.WithCause(err)
				}
				sessionKeys.NwkSEncKey, err = js.wrapKeyIfKEKExists(nwkSKeys.NwkSEncKey, js.KeyVault.NsKEKLabel(ctx, dev.NetID, dev.NetworkServerAddress))
				if err != nil {
					return nil, nil, errWrapKey.WithCause(err)
				}
			}

			res = &ttnpb.JoinResponse{
				RawPayload:  append(b[:1], enc...),
				SessionKeys: sessionKeys,
			}
			_, err = js.keys.SetByID(ctx, *dev.DevEUI, res.SessionKeys.SessionKeyID,
				[]string{
					"session_key_id",
					"f_nwk_s_int_key",
					"s_nwk_s_int_key",
					"nwk_s_enc_key",
					"app_s_key",
				},
				func(stored *ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error) {
					if stored != nil {
						return nil, nil, errDuplicateIdentifiers
					}
					return &res.SessionKeys, []string{
						"session_key_id",
						"f_nwk_s_int_key",
						"s_nwk_s_int_key",
						"nwk_s_enc_key",
						"app_s_key",
					}, nil
				},
			)
			if err != nil {
				return nil, nil, err
			}

			dev.Session = &ttnpb.Session{
				StartedAt:   time.Now().UTC(),
				DevAddr:     req.DevAddr,
				SessionKeys: res.SessionKeys,
			}
			dev.EndDeviceIdentifiers.DevAddr = &req.DevAddr
			paths = append(paths, "session", "ids.dev_addr")

			return dev, paths, nil
		})
	if err != nil {
		logger.WithFields(log.Fields(
			"join_eui", pld.JoinEUI,
			"dev_eui", pld.DevEUI,
		)).WithError(err).Error("Failed to update device")
		return nil, err
	}

	registerAcceptJoin(ctx, dev, req)
	return res, nil
}

// GetNwkSKeys returns the requested network session keys.
func (js *JoinServer) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error) {
	if dn, ok := auth.X509DNFromContext(ctx); ok {
		dev, err := js.devices.GetByEUI(ctx, req.JoinEUI, req.DevEUI,
			[]string{
				"network_server_address",
			},
		)
		if err != nil {
			return nil, errRegistryOperation.WithCause(err)
		}
		if err := validateCaller(dn, dev.NetworkServerAddress); err != nil {
			return nil, err
		}
	} else if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ks, err := js.keys.GetByID(ctx, req.DevEUI, req.SessionKeyID,
		[]string{
			"f_nwk_s_int_key",
			"nwk_s_enc_key",
			"s_nwk_s_int_key",
		},
	)
	if err != nil {
		return nil, errRegistryOperation.WithCause(err)
	}

	if ks.NwkSEncKey == nil {
		return nil, errNoNwkSEncKey
	}
	if ks.FNwkSIntKey == nil {
		return nil, errNoFNwkSIntKey
	}
	if ks.SNwkSIntKey == nil {
		return nil, errNoSNwkSIntKey
	}

	return &ttnpb.NwkSKeysResponse{
		NwkSEncKey:  *ks.NwkSEncKey,
		FNwkSIntKey: *ks.FNwkSIntKey,
		SNwkSIntKey: *ks.SNwkSIntKey,
	}, nil
}

// GetAppSKey returns the requested application session key.
func (js *JoinServer) GetAppSKey(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error) {
	if dn, ok := auth.X509DNFromContext(ctx); ok {
		dev, err := js.devices.GetByEUI(ctx, req.JoinEUI, req.DevEUI,
			[]string{
				"application_server_address",
			},
		)
		if err != nil {
			return nil, errRegistryOperation.WithCause(err)
		}
		if err := validateCaller(dn, dev.ApplicationServerAddress); err != nil {
			return nil, err
		}
	} else if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ks, err := js.keys.GetByID(ctx, req.DevEUI, req.SessionKeyID,
		[]string{
			"app_s_key",
		},
	)
	if err != nil {
		return nil, errRegistryOperation.WithCause(err)
	}
	if ks.AppSKey == nil {
		return nil, errNoAppSKey
	}
	return &ttnpb.AppSKeyResponse{
		AppSKey: *ks.AppSKey,
	}, nil
}

// GetHomeNetID returns the requested NetID.
func (js *JoinServer) GetHomeNetID(ctx context.Context, joinEUI, devEUI types.EUI64) (*types.NetID, error) {
	if _, ok := auth.X509DNFromContext(ctx); !ok {
		if err := clusterauth.Authorized(ctx); err != nil {
			return nil, err
		}
	}

	dev, err := js.devices.GetByEUI(ctx, joinEUI, devEUI,
		[]string{
			"net_id",
			"network_server_address",
		},
	)
	if err != nil {
		return nil, errRegistryOperation.WithCause(err)
	}

	return dev.NetID, nil
}
