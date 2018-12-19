// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package io

import "go.thethings.network/lorawan-stack/pkg/ttnpb"

// PathToUplinkToken returns an opaque uplink token from the given downlink path.
func PathToUplinkToken(path *ttnpb.DownlinkPath) ([]byte, error) {
	return path.Marshal()
}

// UplinkTokenToPath returns the downlink path from the giveno opaque uplink token.
func UplinkTokenToPath(buf []byte) (*ttnpb.DownlinkPath, error) {
	res := &ttnpb.DownlinkPath{}
	if err := res.Unmarshal(buf); err != nil {
		return nil, err
	}
	return res, nil
}
