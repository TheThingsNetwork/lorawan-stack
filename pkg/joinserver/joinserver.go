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
	"encoding/binary"
	"io"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	ulid "github.com/oklog/ulid/v2"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
)

// JoinServer implements the Join Server component.
//
// The Join Server exposes the NsJs, AsJs, AppJs and DeviceRegistry services.
type JoinServer struct {
	*component.Component
	ctx context.Context

	devices                       DeviceRegistry
	keys                          KeyRegistry
	applicationActivationSettings ApplicationActivationSettingRegistry

	euiPrefixes []types.EUI64Prefix

	entropyMu *sync.Mutex
	entropy   io.Reader

	grpc struct {
		nsJs                          nsJsServer
		asJs                          asJsServer
		appJs                         appJsServer
		jsDevices                     jsEndDeviceRegistryServer
		js                            jsServer
		applicationActivationSettings applicationActivationSettingsRegistryServer
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

		devices:                       conf.Devices,
		keys:                          conf.Keys,
		applicationActivationSettings: conf.ApplicationActivationSettings,

		euiPrefixes: conf.JoinEUIPrefixes,

		entropyMu: &sync.Mutex{},
		entropy:   ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0),
	}

	js.grpc.applicationActivationSettings = applicationActivationSettingsRegistryServer{
		JS:       js,
		kekLabel: conf.DeviceKEKLabel,
	}
	js.grpc.jsDevices = jsEndDeviceRegistryServer{
		JS:       js,
		kekLabel: conf.DeviceKEKLabel,
	}
	js.grpc.nsJs = nsJsServer{JS: js}
	js.grpc.asJs = asJsServer{JS: js}
	js.grpc.appJs = appJsServer{JS: js}
	js.grpc.js = jsServer{JS: js}
	js.interop = interopServer{JS: js}

	// TODO: Support authentication from non-cluster-local NS and AS (https://github.com/TheThingsNetwork/lorawan-stack/issues/4).
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsJs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("joinserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsJs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("joinserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AppJs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("joinserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.Js", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("joinserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.ApplicationActivationSettingsRegistry", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("joinserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsJs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsJs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.Js", cluster.HookName, c.ClusterAuthUnaryHook())

	c.RegisterGRPC(js)
	c.RegisterInterop(js)
	return js, nil
}

// Roles of the gRPC service.
func (js *JoinServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_JOIN_SERVER}
}

// RegisterServices registers services provided by js at s.
func (js *JoinServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterAsJsServer(s, js.grpc.asJs)
	ttnpb.RegisterAppJsServer(s, js.grpc.appJs)
	ttnpb.RegisterNsJsServer(s, js.grpc.nsJs)
	ttnpb.RegisterJsEndDeviceRegistryServer(s, js.grpc.jsDevices)
	ttnpb.RegisterJsServer(s, js.grpc.js)
	ttnpb.RegisterApplicationActivationSettingRegistryServer(s, js.grpc.applicationActivationSettings)
}

// RegisterHandlers registers gRPC handlers.
func (js *JoinServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterJsHandler(js.Context(), s, conn)
	ttnpb.RegisterJsEndDeviceRegistryHandler(js.Context(), s, conn)
	ttnpb.RegisterApplicationActivationSettingRegistryHandler(js.Context(), s, conn)
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
	ttnpb.MAC_V1_0_4,
	ttnpb.MAC_V1_1,
}

// wrapKeyWithVault wraps the given key with the configured KEK label.
// If KEK label is empty or wrapping fails with err, for which plaintextCond(err) is true, the key is returned in the clear.
func wrapKeyWithVault(ctx context.Context, key types.AES128Key, kekLabel string, kv crypto.KeyVault, plaintextCond func(error) bool) (*ttnpb.KeyEnvelope, error) {
	if kekLabel == "" {
		return &ttnpb.KeyEnvelope{
			Key: &key,
		}, nil
	}
	ke, err := cryptoutil.WrapAES128Key(ctx, key, kekLabel, kv)
	if err != nil {
		if plaintextCond != nil && plaintextCond(err) {
			return &ttnpb.KeyEnvelope{
				Key: &key,
			}, nil
		}
		return nil, errWrapKey.WithAttributes("label", kekLabel).WithCause(err)
	}
	return ke, nil
}

// wrapKeyWithKEK wraps the given key with the configured KEK label.
// If KEK label is empty, the key is returned in the clear.
func wrapKeyWithKEK(ctx context.Context, key types.AES128Key, kekLabel string, kek types.AES128Key) (*ttnpb.KeyEnvelope, error) {
	if kekLabel == "" {
		return &ttnpb.KeyEnvelope{
			Key: &key,
		}, nil
	}
	ke, err := cryptoutil.WrapAES128KeyWithKEK(ctx, key, kekLabel, kek)
	if err != nil {
		return nil, errWrapKey.WithAttributes("label", kekLabel).WithCause(err)
	}
	return ke, nil
}

var (
	errGetApplicationActivationSettings = errors.Define("application_activation_settings", "failed to get application activation settings")
	errNoKEK                            = errors.DefineNotFound("kek", "KEK not found")
)

// HandleJoin handles the given join-request.
func (js *JoinServer) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest, authorizer Authorizer) (res *ttnpb.JoinResponse, err error) {
	if err := authorizer.Authorized(ctx); err != nil {
		return nil, err
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
		return nil, errNoJoinRequest.New()
	}
	if pld.DevEui.IsZero() {
		return nil, errNoDevEUI.New()
	}
	logger = logger.WithFields(log.Fields(
		"join_eui", pld.JoinEui,
		"dev_eui", pld.DevEui,
	))

	var match bool
	for _, p := range js.euiPrefixes {
		if p.Matches(pld.JoinEui) {
			match = true
			break
		}
	}
	if !match {
		return nil, errUnknownJoinEUI.New()
	}

	var handled bool
	dev, err := js.devices.SetByEUI(ctx, pld.JoinEui, pld.DevEui,
		[]string{
			"application_server_address",
			"application_server_id",
			"application_server_kek_label",
			"last_dev_nonce",
			"last_join_nonce",
			"net_id",
			"network_server_address",
			"network_server_kek_label",
			"provisioner_id",
			"provisioning_data",
			"resets_join_nonces",
			"root_keys",
			"used_dev_nonces",
		},
		func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			getAppSettings := func(ids ttnpb.ApplicationIdentifiers) func() (*ttnpb.ApplicationActivationSettings, error) {
				var (
					res *ttnpb.ApplicationActivationSettings
					err error
				)
				return func() (*ttnpb.ApplicationActivationSettings, error) {
					if res == nil && err == nil {
						res, err = js.applicationActivationSettings.GetByID(ctx, ids, []string{
							"home_net_id",
							"kek_label",
							"kek",
						})
					}
					return res, err
				}
			}(dev.ApplicationIdentifiers)

			if trustedOriginAuth, ok := authorizer.(TrustedOriginAuthorizer); ok {
				netID := dev.NetId
				if netID == nil {
					appSettings, err := getAppSettings()
					if err == nil {
						netID = appSettings.HomeNetId
					} else if !errors.IsNotFound(err) {
						return nil, nil, errLookupNetID.WithCause(err)
					}
				}
				if netID == nil {
					return nil, nil, errNoNetID.New()
				}
				if !req.NetId.Equal(*netID) {
					return nil, nil, errNetIDMismatch.WithAttributes("net_id", req.NetId)
				}
				if dev.NetworkServerAddress != "" {
					if err := trustedOriginAuth.RequireAddress(ctx, dev.NetworkServerAddress); err != nil {
						return nil, nil, err
					}
				}
			}

			paths := make([]string, 0, 3)

			dn := uint32(binary.BigEndian.Uint16(pld.DevNonce[:]))
			if req.SelectedMACVersion.IncrementDevNonce() {
				if (dn != 0 || dev.LastDevNonce != 0 || dev.LastJoinNonce != 0) && !dev.ResetsJoinNonces {
					if dn <= dev.LastDevNonce {
						return nil, nil, errDevNonceTooSmall.New()
					}
				}
				dev.LastDevNonce = dn
				paths = append(paths, "last_dev_nonce")
			} else {
				i := sort.Search(len(dev.UsedDevNonces), func(i int) bool { return dev.UsedDevNonces[i] >= dn })
				if i >= len(dev.UsedDevNonces) || dev.UsedDevNonces[i] != dn {
					dev.UsedDevNonces = append(dev.UsedDevNonces, 0)
					copy(dev.UsedDevNonces[i+1:], dev.UsedDevNonces[i:])
					dev.UsedDevNonces[i] = dn
					paths = append(paths, "used_dev_nonces")
				} else if !dev.ResetsJoinNonces {
					return nil, nil, errReuseDevNonce.New()
				}
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
				return nil, nil, errJoinNonceTooHigh.New()
			}
			dev.LastJoinNonce++
			paths = append(paths, "last_join_nonce")

			var jn types.JoinNonce
			nb := make([]byte, 4)
			binary.BigEndian.PutUint32(nb, dev.LastJoinNonce)
			copy(jn[:], nb[1:])

			b, err = lorawan.AppendJoinAcceptPayload(b, ttnpb.JoinAcceptPayload{
				NetId:      req.NetId,
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
				return nil, nil, errGenerateSessionKeyID.New()
			}

			cc, err := js.GetPeerConn(ctx, ttnpb.ClusterRole_CRYPTO_SERVER, &dev.EndDeviceIdentifiers)
			if err != nil {
				if !errors.IsNotFound(err) {
					logger.WithError(err).Debug("Crypto Server connection is not available")
				}
				cc = nil
			}

			var networkCryptoService cryptoservices.Network
			if req.SelectedMACVersion.UseNwkKey() && dev.RootKeys != nil && dev.RootKeys.NwkKey != nil {
				// LoRaWAN 1.1 and higher use a NwkKey.
				nwkKey, err := cryptoutil.UnwrapAES128Key(ctx, dev.RootKeys.NwkKey, js.KeyVault)
				if err != nil {
					return nil, nil, err
				}
				networkCryptoService = cryptoservices.NewMemory(&nwkKey, nil)
			} else if cc != nil && dev.ProvisionerID != "" {
				networkCryptoService = cryptoservices.NewNetworkRPCClient(cc, js.KeyVault, js.WithClusterAuth())
			}

			var applicationCryptoService cryptoservices.Application
			if dev.RootKeys != nil && dev.RootKeys.AppKey != nil {
				appKey, err := cryptoutil.UnwrapAES128Key(ctx, dev.RootKeys.AppKey, js.KeyVault)
				if err != nil {
					return nil, nil, err
				}
				applicationCryptoService = cryptoservices.NewMemory(nil, &appKey)
				if !req.SelectedMACVersion.UseNwkKey() {
					// LoRaWAN 1.0.x use the AppKey for network security operations.
					networkCryptoService = cryptoservices.NewMemory(nil, &appKey)
				}
			} else if cc != nil && dev.ProvisionerID != "" {
				applicationCryptoService = cryptoservices.NewApplicationRPCClient(cc, js.KeyVault, js.WithClusterAuth())
			}
			if networkCryptoService == nil {
				return nil, nil, errNoNwkKey.New()
			}
			if applicationCryptoService == nil {
				return nil, nil, errNoAppKey.New()
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
				return nil, nil, errMICMismatch.New()
			}
			resMIC, err := networkCryptoService.JoinAcceptMIC(ctx, cryptoDev, req.SelectedMACVersion, 0xff, pld.DevNonce, b)
			if err != nil {
				return nil, nil, errComputeMIC.WithCause(err)
			}
			enc, err := networkCryptoService.EncryptJoinAccept(ctx, cryptoDev, req.SelectedMACVersion, append(b[1:], resMIC[:]...))
			if err != nil {
				return nil, nil, errEncryptPayload.WithCause(err)
			}
			nwkSKeys, err := networkCryptoService.DeriveNwkSKeys(ctx, cryptoDev, req.SelectedMACVersion, jn, pld.DevNonce, req.NetId)
			if err != nil {
				return nil, nil, errDeriveNwkSKeys.WithCause(err)
			}
			appSKey, err := applicationCryptoService.DeriveAppSKey(ctx, cryptoDev, req.SelectedMACVersion, jn, pld.DevNonce, req.NetId)
			if err != nil {
				return nil, nil, errDeriveAppSKey.WithCause(err)
			}

			var (
				fNwkSIntKeyEnvelope *ttnpb.KeyEnvelope
				sNwkSIntKeyEnvelope *ttnpb.KeyEnvelope
				nwkSEncKeyEnvelope  *ttnpb.KeyEnvelope
				appSKeyEnvelope     *ttnpb.KeyEnvelope

				nsPlaintextCond func(error) bool
				asPlaintextCond func(error) bool
			)
			nsKEKLabel, asKEKLabel := dev.NetworkServerKEKLabel, dev.ApplicationServerKEKLabel
			if nsKEKLabel == "" {
				nsKEKLabel = js.KeyVault.NsKEKLabel(ctx, dev.NetId, dev.NetworkServerAddress)
				nsPlaintextCond = errors.IsNotFound
			}
			fNwkSIntKeyEnvelope, err = wrapKeyWithVault(ctx, nwkSKeys.FNwkSIntKey, nsKEKLabel, js.KeyVault, nsPlaintextCond)
			if err != nil {
				return nil, nil, err
			}
			if req.SelectedMACVersion.UseNwkKey() {
				sNwkSIntKeyEnvelope, err = wrapKeyWithVault(ctx, nwkSKeys.SNwkSIntKey, nsKEKLabel, js.KeyVault, nsPlaintextCond)
				if err != nil {
					return nil, nil, err
				}
				nwkSEncKeyEnvelope, err = wrapKeyWithVault(ctx, nwkSKeys.NwkSEncKey, nsKEKLabel, js.KeyVault, nsPlaintextCond)
				if err != nil {
					return nil, nil, err
				}
			}
			if asKEKLabel == "" {
				appSettings, err := getAppSettings()
				if err != nil {
					if !errors.IsNotFound(err) {
						return nil, nil, errGetApplicationActivationSettings.WithCause(err)
					}
					asKEKLabel = js.KeyVault.AsKEKLabel(ctx, dev.ApplicationServerAddress)
					asPlaintextCond = errors.IsNotFound
				} else {
					var kek types.AES128Key
					if appSettings.KEKLabel != "" {
						if appSettings.KEK == nil {
							return nil, nil, errNoKEK.New()
						}
						kek, err = cryptoutil.UnwrapAES128Key(ctx, appSettings.KEK, js.KeyVault)
						if err != nil {
							return nil, nil, errUnwrapKey.WithCause(err)
						}
					}
					appSKeyEnvelope, err = wrapKeyWithKEK(ctx, appSKey, appSettings.KEKLabel, kek)
					if err != nil {
						return nil, nil, err
					}
				}
			}
			if asKEKLabel != "" {
				appSKeyEnvelope, err = wrapKeyWithVault(ctx, appSKey, asKEKLabel, js.KeyVault, asPlaintextCond)
				if err != nil {
					return nil, nil, err
				}
			}

			sk := ttnpb.SessionKeys{
				SessionKeyID: skID[:],
				FNwkSIntKey:  fNwkSIntKeyEnvelope,
				NwkSEncKey:   nwkSEncKeyEnvelope,
				SNwkSIntKey:  sNwkSIntKeyEnvelope,
				AppSKey:      appSKeyEnvelope,
			}
			_, err = js.keys.SetByID(ctx, *dev.JoinEui, *dev.DevEui, sk.SessionKeyID,
				[]string{
					"session_key_id",
					"f_nwk_s_int_key",
					"s_nwk_s_int_key",
					"nwk_s_enc_key",
					"app_s_key",
				},
				func(stored *ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error) {
					if stored != nil {
						return nil, nil, errDuplicateIdentifiers.New()
					}
					return &sk, []string{
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
				SessionKeys: sk,
			}
			dev.EndDeviceIdentifiers.DevAddr = &req.DevAddr
			paths = append(paths, "session", "ids.dev_addr")

			handled = true
			res = &ttnpb.JoinResponse{
				RawPayload:  append(b[:1], enc...),
				SessionKeys: sk,
			}
			return dev, paths, nil
		},
	)
	if err != nil {
		logger := logger.WithError(err)
		if !handled {
			logger.Info("Join not accepted")
			return nil, err
		}
		logger.Error("Failed to update device")
		return nil, errRegistryOperation.WithCause(err)
	}

	registerAcceptJoin(dev.Context, dev.EndDevice, req)
	return res, nil
}

// GetNwkSKeys returns the requested network session keys.
func (js *JoinServer) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest, authorizer Authorizer) (*ttnpb.NwkSKeysResponse, error) {
	if err := authorizer.Authorized(ctx); err != nil {
		return nil, err
	}

	if trustedOriginAuth, ok := authorizer.(TrustedOriginAuthorizer); ok {
		dev, err := js.devices.GetByEUI(ctx, req.JoinEui, req.DevEui,
			[]string{
				"network_server_address",
			},
		)
		if err != nil {
			return nil, errRegistryOperation.WithCause(err)
		}
		if dev.NetworkServerAddress != "" {
			if err := trustedOriginAuth.RequireAddress(ctx, dev.NetworkServerAddress); err != nil {
				return nil, err
			}
		}
	}

	ks, err := js.keys.GetByID(ctx, req.JoinEui, req.DevEui, req.SessionKeyID,
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
		return nil, errNoNwkSEncKey.New()
	}
	if ks.FNwkSIntKey == nil {
		return nil, errNoFNwkSIntKey.New()
	}
	if ks.SNwkSIntKey == nil {
		return nil, errNoSNwkSIntKey.New()
	}

	return &ttnpb.NwkSKeysResponse{
		NwkSEncKey:  *ks.NwkSEncKey,
		FNwkSIntKey: *ks.FNwkSIntKey,
		SNwkSIntKey: *ks.SNwkSIntKey,
	}, nil
}

// GetAppSKey returns the requested application session key.
func (js *JoinServer) GetAppSKey(ctx context.Context, req *ttnpb.SessionKeyRequest, authorizer Authorizer) (*ttnpb.AppSKeyResponse, error) {
	if err := authorizer.Authorized(ctx); err != nil {
		return nil, err
	}

	if trustedOriginAuth, ok := authorizer.(TrustedOriginAuthorizer); ok {
		dev, err := js.devices.GetByEUI(ctx, req.JoinEui, req.DevEui,
			[]string{
				"application_server_address",
				"application_server_id",
			},
		)
		if err != nil {
			return nil, errRegistryOperation.WithCause(err)
		}
		if dev.ApplicationServerID != "" {
			if err := trustedOriginAuth.RequireID(ctx, dev.ApplicationServerID); err != nil {
				return nil, err
			}
		} else if dev.ApplicationServerAddress != "" {
			if err := trustedOriginAuth.RequireAddress(ctx, dev.ApplicationServerAddress); err != nil {
				return nil, err
			}
		} else {
			sets, err := js.applicationActivationSettings.GetByID(ctx, dev.ApplicationIdentifiers, []string{
				"application_server_id",
			})
			if err != nil {
				if !errors.IsNotFound(err) {
					return nil, errGetApplicationActivationSettings.WithCause(err)
				}
				return nil, errNoApplicationServerID.New()
			}
			if sets.ApplicationServerID == "" {
				return nil, errNoApplicationServerID.New()
			}
			if err := trustedOriginAuth.RequireID(ctx, sets.ApplicationServerID); err != nil {
				return nil, err
			}
		}
	}
	if appAuth, ok := authorizer.(ApplicationAccessAuthorizer); ok {
		dev, err := js.devices.GetByEUI(ctx, req.JoinEui, req.DevEui, nil)
		if err != nil {
			return nil, errRegistryOperation.WithCause(err)
		}
		if err := appAuth.RequireApplication(ctx, dev.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS); err != nil {
			return nil, err
		}
	}

	ks, err := js.keys.GetByID(ctx, req.JoinEui, req.DevEui, req.SessionKeyID,
		[]string{
			"app_s_key",
		},
	)
	if err != nil {
		return nil, errRegistryOperation.WithCause(err)
	}
	if ks.AppSKey == nil {
		return nil, errNoAppSKey.New()
	}
	return &ttnpb.AppSKeyResponse{
		AppSKey: *ks.AppSKey,
	}, nil
}

// GetHomeNetID returns the requested NetID.
func (js *JoinServer) GetHomeNetID(ctx context.Context, joinEUI, devEUI types.EUI64, authorizer Authorizer) (*types.NetID, error) {
	if err := authorizer.Authorized(ctx); err != nil {
		return nil, err
	}

	dev, err := js.devices.GetByEUI(ctx, joinEUI, devEUI,
		[]string{
			"net_id",
		},
	)
	if err != nil {
		return nil, errRegistryOperation.WithCause(err)
	}
	if dev.NetId != nil {
		return dev.NetId, nil
	}
	sets, err := js.applicationActivationSettings.GetByID(ctx, dev.ApplicationIdentifiers, []string{
		"home_net_id",
	})
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, errGetApplicationActivationSettings.WithCause(err)
		}
		return nil, nil
	}
	return sets.HomeNetId, nil
}
