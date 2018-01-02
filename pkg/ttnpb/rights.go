// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"strconv"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/gogo/protobuf/jsonpb"
)

// AllUserRights is the set thart contains all the rights that are to users.
var AllUserRights = []Right{
	RIGHT_USER_PROFILE_READ,
	RIGHT_USER_PROFILE_WRITE,
	RIGHT_USER_DELETE,
	RIGHT_USER_AUTHORIZEDCLIENTS,
	RIGHT_USER_APPLICATIONS_LIST,
	RIGHT_USER_APPLICATIONS_CREATE,
	RIGHT_USER_GATEWAYS_LIST,
	RIGHT_USER_GATEWAYS_CREATE,
	RIGHT_USER_CLIENTS_LIST,
	RIGHT_USER_CLIENTS_CREATE,
	RIGHT_USER_CLIENTS_MANAGE,
}

// AllApplicationRights is the set that contains all the rights that are to applications.
var AllApplicationRights = []Right{
	RIGHT_APPLICATION_INFO,
	RIGHT_APPLICATION_SETTINGS_BASIC,
	RIGHT_APPLICATION_SETTINGS_KEYS,
	RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
	RIGHT_APPLICATION_DELETE,
	RIGHT_APPLICATION_DEVICES_READ,
	RIGHT_APPLICATION_DEVICES_WRITE,
	RIGHT_APPLICATION_TRAFFIC_READ,
	RIGHT_APPLICATION_TRAFFIC_WRITE,
}

// AllGatewayRights is the set that contains all the rights that are to gateways.
var AllGatewayRights = []Right{
	RIGHT_GATEWAY_INFO,
	RIGHT_GATEWAY_SETTINGS_BASIC,
	RIGHT_GATEWAY_SETTINGS_KEYS,
	RIGHT_GATEWAY_SETTINGS_COLLABORATORS,
	RIGHT_GATEWAY_DELETE,
	RIGHT_GATEWAY_TRAFFIC,
	RIGHT_GATEWAY_STATUS,
	RIGHT_GATEWAY_LOCATION,
}

// ParseRight parses the string specified into a Right.
func ParseRight(str string) (Right, error) {
	val, ok := Right_value["RIGHT_"+strings.ToUpper(strings.Replace(str, ":", "_", -1))]
	if !ok {
		val, ok = Right_value[str]
		if !ok {
			return -1, errors.Errorf("Could not parse right `%s`", str)
		}
	}
	return Right(val), nil
}

// TextString returns a textual string representation of the right.
func (r Right) TextString() string {
	str, ok := Right_name[int32(r)]
	if ok {
		return strings.ToLower(strings.Replace(strings.TrimPrefix(str, "RIGHT_"), "_", ":", -1))
	}
	return strconv.Itoa(int(r))
}

// MarshalText implements encoding.TextMarshaler interface.
func (r Right) MarshalText() ([]byte, error) {
	return []byte(r.TextString()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (r Right) MarshalJSON() ([]byte, error) {
	txt, err := r.MarshalText()
	if err != nil {
		return nil, err
	}
	return []byte("\"" + string(txt) + "\""), nil
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (r Right) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if m.EnumsAsInts {
		return []byte("\"" + strconv.Itoa(int(r)) + "\""), nil
	}
	return r.MarshalJSON()
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (r *Right) UnmarshalText(b []byte) (err error) {
	*r, err = ParseRight(string(b))
	return
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (r *Right) UnmarshalJSON(b []byte) error {
	return r.UnmarshalText(b[1 : len(b)-1])
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (r *Right) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return r.UnmarshalJSON(b)
}
