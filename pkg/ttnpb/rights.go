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
	"strings"
)

var (
	AllUserRights         = &Rights{}
	AllApplicationRights  = &Rights{}
	AllClientRights       = &Rights{}
	AllEntityRights       = &Rights{}
	AllGatewayRights      = &Rights{}
	AllOrganizationRights = &Rights{}
	AllClusterRights      = &Rights{}
	AllReadAdminRights    = &Rights{}
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
		if strings.HasSuffix(k, "_READ") ||
			strings.HasSuffix(k, "_INFO") ||
			strings.HasSuffix(k, "_LIST") ||
			strings.HasSuffix(k, "_GET") {
			AllReadAdminRights.Rights = append(AllReadAdminRights.Rights, Right(v))
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

// Implied returns the Right's implied rights.
func (r Right) Implied() *Rights {
	// NOTE: Changes here require the documentation in rights.proto to be updated.
	switch r {
	case Right_RIGHT_USER_ALL:
		return AllUserRights
	case Right_RIGHT_APPLICATION_ALL:
		return AllApplicationRights
	case Right_RIGHT_APPLICATION_LINK:
		return RightsFrom(
			Right_RIGHT_APPLICATION_INFO,
			Right_RIGHT_APPLICATION_TRAFFIC_READ,
			Right_RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
		)
	case Right_RIGHT_GATEWAY_ALL:
		return AllGatewayRights
	case Right_RIGHT_GATEWAY_LINK:
		return RightsFrom(
			Right_RIGHT_GATEWAY_INFO,
		)
	case Right_RIGHT_ORGANIZATION_ALL:
		return AllOrganizationRights
	case Right_RIGHT_ALL:
		return AllRights
	case Right_RIGHT_CLIENT_ALL:
		return RightsFrom(
			Right_RIGHT_CLIENT_INFO,
			Right_RIGHT_CLIENT_SETTINGS_BASIC,
			Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS,
			Right_RIGHT_CLIENT_DELETE,
		)
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

type rightsByString []Right

func (r rightsByString) Len() int           { return len(r) }
func (r rightsByString) Less(i, j int) bool { return r[i].String() < r[j].String() }
func (r rightsByString) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

// Sorted returns a sorted rights list by string value.
// The original rights list is not mutated.
func (r *Rights) Sorted() *Rights {
	if r == nil {
		return &Rights{}
	}
	res := &Rights{Rights: make([]Right, len(r.Rights))}
	copy(res.Rights, r.Rights)
	sort.Sort(rightsByString(res.Rights))
	return res
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
		return &Rights{}
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
	identifier := m.GetId()
	if name := m.GetName(); name != "" {
		identifier = fmt.Sprintf("%v (%v)", identifier, name)
	}
	return identifier
}
