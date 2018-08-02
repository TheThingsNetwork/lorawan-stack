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

package networkserver

import "go.thethings.network/lorawan-stack/pkg/ttnpb"

func handleMACResponse(cid ttnpb.MACCommandIdentifier, f func(*ttnpb.MACCommand) error, cmds ...*ttnpb.MACCommand) ([]*ttnpb.MACCommand, error) {
	for i, cmd := range cmds {
		if cmd.CID != cid {
			continue
		}
		if err := f(cmd); err != nil {
			return cmds, err
		}
		return append(cmds[:i], cmds[i+1:]...), nil
	}
	return cmds, errMACRequestNotFound
}

func handleMACResponseBlock(cid ttnpb.MACCommandIdentifier, f func(*ttnpb.MACCommand) error, cmds ...*ttnpb.MACCommand) ([]*ttnpb.MACCommand, error) {
	first := -1
	last := -1

outer:
	for i, cmd := range cmds {
		last = i

		switch {
		case first >= 0 && cmd.CID != cid:
			break outer
		case first < 0 && cmd.CID != cid:
			continue
		case first < 0:
			first = i
		}
		if err := f(cmd); err != nil {
			return cmds, err
		}
	}

	if first < 0 {
		return cmds, errMACRequestNotFound
	}
	return append(cmds[:first], cmds[last+1:]...), nil
}
