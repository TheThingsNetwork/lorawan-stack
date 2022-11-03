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

package ttjsv1

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

type device struct {
	claimData
	claimedBy               string // AS-ID.
	locked                  bool
	claimAuthenticationCode string
}

type clientData struct {
	asID string
}

type mockTTJS struct {
	provisonedDevices map[types.EUI64]device
	joinEUIPrefixes   []types.EUI64Prefix
	lis               net.Listener
	clients           map[string]clientData // key is the auth token.
}

func (srv *mockTTJS) Start(ctx context.Context) error {
	r := mux.NewRouter()
	r.HandleFunc("/v1/claim/{devEUI}", srv.handleClaim)
	s := http.Server{
		Handler:           r,
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		s.Close()
	}()
	return s.Serve(srv.lis)
}

func writeResponse(w http.ResponseWriter, statusCode int, message string) {
	resp := errorResponse{
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func (srv *mockTTJS) handleClaim(w http.ResponseWriter, r *http.Request) { //nolint:gocyclo
	asID, password, ok := r.BasicAuth()
	if !ok {
		writeResponse(w, http.StatusUnauthorized, "API Key not found")
		return
	}
	var client *clientData
	for token, cl := range srv.clients {
		cl := cl
		if password == token && cl.asID == asID {
			client = &cl
			break
		}
	}
	if client == nil {
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
		if dev.claimedBy == "" {
			writeResponse(w, http.StatusNotFound, "Device not claimed")
			return
		}
		if dev.claimedBy != client.asID {
			writeResponse(w, http.StatusForbidden, "Client not allowed to unclaim")
			return
		}
		dev.claimedBy = ""
		dev.locked = false
		srv.provisonedDevices[reqDevEUI] = dev

	case http.MethodGet:
		// Check if the device exists and is claimed.
		dev, ok := srv.provisonedDevices[reqDevEUI]
		if !ok {
			writeResponse(w, http.StatusNotFound, "Device not provisoned")
			return
		}
		if dev.claimedBy == "" {
			writeResponse(w, http.StatusNotFound, "Device not claimed")
			return
		}
		if dev.claimedBy != client.asID {
			writeResponse(w, http.StatusForbidden, "Client not allowed to get status")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dev.claimData) //nolint:errcheck

	case http.MethodPost:
		var req claimRequest
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

		if dev.claimedBy != "" && dev.claimedBy != client.asID && dev.locked {
			writeResponse(w, http.StatusForbidden, "Client not allowed to claim")
			return
		}

		if dev.claimAuthenticationCode != req.OwnerToken {
			writeResponse(w, http.StatusUnauthorized, "Owner token mismatch")
			return
		}

		dev.claimedBy = client.asID
		dev.claimData = req.claimData
		dev.locked = req.Locked

		// Update
		srv.provisonedDevices[reqDevEUI] = dev
		w.WriteHeader(http.StatusCreated)
	}
}
