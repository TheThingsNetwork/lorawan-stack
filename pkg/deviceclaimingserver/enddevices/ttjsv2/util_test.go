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

package ttjsv2_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/enddevices/ttjsv2"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

type device struct {
	homeNetID               types.NetID
	homeNSID                *types.EUI64
	asID                    string
	locked                  bool
	claimAuthenticationCode string
}

type clientData struct {
	cert *x509.Certificate
}

type mockTTJS struct {
	provisonedDevices map[types.EUI64]device
	joinEUIPrefixes   []types.EUI64Prefix
	lis               net.Listener
	clients           map[string]clientData // the key is the AS-ID
}

func (srv *mockTTJS) Start(ctx context.Context) error {
	r := mux.NewRouter()
	r.HandleFunc("/api/v2/devices/{devEUI}/claim", srv.handleClaim)
	r.HandleFunc("/api/v2/devices/{devEUIs}/claims", srv.handleBatchClaims)
	s := http.Server{
		Handler:           r,
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ClientAuth: tls.RequireAnyClientCert,
		},
	}
	go func() {
		<-ctx.Done()
		s.Close()
	}()
	return s.ServeTLS(srv.lis, "testdata/servercert.pem", "testdata/serverkey.pem")
}

func writeResponse(w http.ResponseWriter, statusCode int, message string) {
	resp := ttjsv2.ErrorResponse{
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func (srv *mockTTJS) handleClaim(w http.ResponseWriter, r *http.Request) { //nolint:gocyclo
	var (
		found bool
		asID  string
	)
	for registeredASID, data := range srv.clients {
		if len(r.TLS.PeerCertificates) > 0 && r.TLS.PeerCertificates[0].Equal(data.cert) {
			found = true
			asID = registeredASID
			break
		}
	}
	if !found {
		writeResponse(w, http.StatusUnauthorized, "Invalid API Key")
		return
	}

	devEUIVal := mux.Vars(r)["devEUI"]
	var reqDevEUI types.EUI64
	err := reqDevEUI.UnmarshalText([]byte(devEUIVal))
	if err != nil {
		writeResponse(w, http.StatusBadRequest, "DevEUI not found or is invalid")
		return
	}

	switch r.Method {
	case http.MethodDelete:
		// Check if the device exists and is claimed.
		dev, ok := srv.provisonedDevices[reqDevEUI]
		if !ok {
			writeResponse(w, http.StatusNotFound, "Device not provisoned")
			return
		}
		if dev.asID == "" {
			writeResponse(w, http.StatusNotFound, "Device not claimed")
			return
		}
		if dev.asID != asID {
			writeResponse(w, http.StatusForbidden, "Client not allowed to unclaim")
			return
		}
		dev.asID = ""
		dev.locked = false
		srv.provisonedDevices[reqDevEUI] = dev

	case http.MethodGet:
		// Check if the device exists and is claimed.
		dev, ok := srv.provisonedDevices[reqDevEUI]
		if !ok {
			writeResponse(w, http.StatusNotFound, "Device not provisoned")
			return
		}
		if dev.asID == "" {
			writeResponse(w, http.StatusNotFound, "Device not claimed")
			return
		}
		if dev.asID != asID {
			writeResponse(w, http.StatusForbidden, "Client not allowed to get status")
			return
		}
		res := ttjsv2.ClaimData{
			HomeNetID: dev.homeNetID.String(),
			Locked:    dev.locked,
		}
		if dev.homeNSID != nil {
			res.HomeNSID = new(string)
			*res.HomeNSID = dev.homeNSID.String()
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res) //nolint:errcheck

	case http.MethodPut:
		var req ttjsv2.ClaimRequest
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeResponse(w, http.StatusBadRequest, "Invalid request")
			return
		}
		// Check if the device is provisioned to be claimed.
		dev, ok := srv.provisonedDevices[reqDevEUI]
		if !ok {
			writeResponse(w, http.StatusNotFound, "Device not provisioned")
			return
		}

		if dev.asID != "" && dev.asID != asID && dev.locked {
			writeResponse(w, http.StatusForbidden, "Client not allowed to claim")
			return
		}

		if dev.claimAuthenticationCode != req.OwnerToken {
			writeResponse(w, http.StatusForbidden, "claim failed with given owner token")
			return
		}

		test.Must[any](nil, dev.homeNetID.UnmarshalText([]byte(req.HomeNetID)))
		if req.HomeNSID != nil {
			dev.homeNSID = new(types.EUI64)
			test.Must[any](nil, dev.homeNSID.UnmarshalText([]byte(*req.HomeNSID)))
		}
		dev.asID = asID
		dev.locked = req.Lock != nil && *req.Lock

		// Update
		srv.provisonedDevices[reqDevEUI] = dev
		w.WriteHeader(http.StatusCreated)
	}
}

func (srv *mockTTJS) handleBatchClaims(w http.ResponseWriter, r *http.Request) { // nolint:gocyclo
	var (
		found bool
		asID  string
	)
	for registeredASID, data := range srv.clients {
		if len(r.TLS.PeerCertificates) > 0 && r.TLS.PeerCertificates[0].Equal(data.cert) {
			found = true
			asID = registeredASID
			break
		}
	}
	if !found {
		writeResponse(w, http.StatusUnauthorized, "Invalid API Key")
		return
	}

	raw := mux.Vars(r)["devEUIs"]
	reqDevEUIs := strings.Split(raw, ",")
	if len(reqDevEUIs) == 0 {
		writeResponse(w, http.StatusNoContent, "No device EUIs in request")
		return
	}

	devEUIs := make([]types.EUI64, len(reqDevEUIs))
	for i, devEUIVal := range reqDevEUIs {
		err := devEUIs[i].UnmarshalText([]byte(devEUIVal))
		if err != nil {
			writeResponse(w, http.StatusBadRequest, "Invalid device EUI")
			return
		}
	}

	switch r.Method {
	case http.MethodDelete:
		retErrs := make(map[string]ttjsv2.ErrorResponse)
		for _, devEUI := range devEUIs {
			// Check if the device exists and is claimed.
			dev, ok := srv.provisonedDevices[devEUI]
			if !ok {
				retErrs[devEUI.String()] = ttjsv2.ErrorResponse{
					Message: "Device not provisioned",
				}
				continue
			}
			if dev.asID == "" {
				retErrs[devEUI.String()] = ttjsv2.ErrorResponse{
					Message: "Device not claimed",
				}
				continue
			}
			if dev.asID != asID {
				retErrs[devEUI.String()] = ttjsv2.ErrorResponse{
					Message: "Client not allowed to unclaim",
				}
				continue
			}
			dev.asID = ""
			dev.locked = false
			srv.provisonedDevices[devEUI] = dev
		}
		if len(retErrs) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(retErrs); err != nil {
				panic(err)
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
