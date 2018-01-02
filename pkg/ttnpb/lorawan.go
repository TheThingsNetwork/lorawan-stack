// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

func (v MACVersion) Compare(other MACVersion) int {
	vStr := v.String()
	oStr := other.String()
	switch {
	case MACVersion_value[vStr] > MACVersion_value[oStr]:
		return 1
	case MACVersion_value[vStr] == MACVersion_value[oStr]:
		return 0
	default:
		return -1
	}
}
