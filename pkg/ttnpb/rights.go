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

package ttnpb

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// MarshalText implements encoding.TextMarshaler interface.
func (v Right) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *Right) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := Right_value[s]; ok {
		*v = Right(i)
		return nil
	}
	if !strings.HasPrefix(s, "RIGHT_") {
		if i, ok := Right_value["RIGHT_"+s]; ok {
			*v = Right(i)
			return nil
		}
	}
	return errCouldNotParse("Right")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *Right) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("Right")(string(b)).WithCause(err)
	}
	*v = Right(i)
	return nil
}

var (
	AllUserRights         = &Rights{}
	AllApplicationRights  = &Rights{}
	AllClientRights       = &Rights{}
	AllEntityRights       = &Rights{}
	AllGatewayRights      = &Rights{}
	AllOrganizationRights = &Rights{}
	AllClusterRights      = &Rights{}
	AllAdminRights        = &Rights{}
	AllRights             = &Rights{}
)

func init() {
	for k, v := range Right_value {
		if v == 0 {
			continue
		}
		switch {
		case strings.HasPrefix(k, "RIGHT_USER_"):
			AllUserRights.Rights = append(AllUserRights.Rights, Right(v))
		case strings.HasPrefix(k, "RIGHT_APPLICATION_"):
			AllApplicationRights.Rights = append(AllApplicationRights.Rights, Right(v))
			AllEntityRights.Rights = append(AllEntityRights.Rights, Right(v))
		case strings.HasPrefix(k, "RIGHT_CLIENT_"):
			AllClientRights.Rights = append(AllClientRights.Rights, Right(v))
			AllEntityRights.Rights = append(AllEntityRights.Rights, Right(v))
		case strings.HasPrefix(k, "RIGHT_GATEWAY_"):
			AllGatewayRights.Rights = append(AllGatewayRights.Rights, Right(v))
			AllEntityRights.Rights = append(AllEntityRights.Rights, Right(v))
		case strings.HasPrefix(k, "RIGHT_ORGANIZATION_"):
			AllOrganizationRights.Rights = append(AllOrganizationRights.Rights, Right(v))
		}
		if strings.HasSuffix(k, "_READ") || strings.HasSuffix(k, "_INFO") {
			AllClusterRights.Rights = append(AllClusterRights.Rights, Right(v))
		}
		if !strings.HasSuffix(k, "_KEYS") && !strings.HasSuffix(k, "_ALL") {
			AllAdminRights.Rights = append(AllAdminRights.Rights, Right(v))
		}
		AllRights.Rights = append(AllRights.Rights, Right(v))
	}
	AllUserRights = AllUserRights.Sorted()
	AllApplicationRights = AllApplicationRights.Sorted()
	AllGatewayRights = AllGatewayRights.Sorted()
	AllOrganizationRights = AllOrganizationRights.Sorted()
	AllRights = AllRights.Sorted()
}

// Implied returns the Right together with its implied rights.
func (r Right) Implied() *Rights {
	// NOTE: Changes here require the documentation in rights.proto to be updated.
	switch r {
	case RIGHT_USER_ALL:
		return AllUserRights
	case RIGHT_APPLICATION_ALL:
		return AllApplicationRights
	case RIGHT_GATEWAY_ALL:
		return AllGatewayRights
	case RIGHT_ORGANIZATION_ALL:
		return AllOrganizationRights
	case RIGHT_ALL:
		return AllRights
	}
	return RightsFrom(r)
}

func makeRightsSet(rights ...*Rights) rightsSet {
	s := make(rightsSet)
	for _, r := range rights {
		if r == nil {
			continue
		}
		s.add(r.Rights...)
	}
	return s
}

type rightsSet map[Right]struct{}

func (s rightsSet) add(rights ...Right) {
	for _, right := range rights {
		s[right] = struct{}{}
	}
}

func (s rightsSet) rights() *Rights {
	res := make([]Right, 0, len(s))
	for right := range s {
		res = append(res, right)
	}
	return &Rights{Rights: res}
}

type rightsByString Rights

func (r rightsByString) Len() int           { return len(r.Rights) }
func (r rightsByString) Less(i, j int) bool { return r.Rights[i].String() < r.Rights[j].String() }
func (r rightsByString) Swap(i, j int)      { r.Rights[i], r.Rights[j] = r.Rights[j], r.Rights[i] }

// Sorted returns a sorted rights list by string value.
// The original rights list is not mutated.
func (r *Rights) Sorted() *Rights {
	if r == nil {
		return nil
	}
	res := Rights{Rights: make([]Right, len(r.Rights))}
	copy(res.Rights, r.Rights)
	sort.Sort(rightsByString(res))
	return &res
}

// Unique removes all duplicate rights from the rights list.
func (r *Rights) Unique() *Rights {
	return makeRightsSet(r).rights()
}

// Union returns the union of the rights lists.
func (r *Rights) Union(b ...*Rights) *Rights {
	return makeRightsSet(append(b, r)...).rights()
}

// Sub returns r without the rights in b.
func (r *Rights) Sub(b *Rights) *Rights {
	s := makeRightsSet(r)
	for _, right := range b.GetRights() {
		delete(s, right)
	}
	return s.rights()
}

// Intersect returns the rights that are contained in both r and b.
func (r *Rights) Intersect(b *Rights) *Rights {
	if r == nil {
		return nil
	}
	res := make([]Right, 0)
	rs, bs := makeRightsSet(r), makeRightsSet(b)
	for right := range rs {
		if _, ok := bs[right]; ok {
			res = append(res, right)
		}
	}
	return &Rights{Rights: res}
}

// Implied returns the rights together with their implied rights.
func (r *Rights) Implied() *Rights {
	s := makeRightsSet(r)
	for _, right := range r.GetRights() {
		s.add(right.Implied().GetRights()...)
	}
	return s.rights()
}

// IncludesAll returns true if r includes all given rights.
func (r *Rights) IncludesAll(search ...Right) bool {
	if r == nil {
		return len(search) == 0
	}
	return len(RightsFrom(search...).Sub(r).GetRights()) == 0
}

// RightsFrom returns a Rights message from a list of rights.
func RightsFrom(rights ...Right) *Rights { return &Rights{Rights: rights} }

// PrettyName returns the key ID (Name if present)
func (m *APIKey) PrettyName() string {
	identifier := m.GetID()
	if name := m.GetName(); name != "" {
		identifier = fmt.Sprintf("%v (%v)", identifier, name)
	}
	return identifier
}
