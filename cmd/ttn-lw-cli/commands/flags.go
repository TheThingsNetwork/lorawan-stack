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

package commands

import (
	"strings"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func collaboratorFlags() *pflag.FlagSet {
	flagSet := new(pflag.FlagSet)
	flagSet.String("user-id", "", "")
	flagSet.String("organization-id", "", "")
	return flagSet
}

var errNoCollaborator = errors.DefineInvalidArgument("no_collaborator", "no collaborator set")

func getCollaborator(flagSet *pflag.FlagSet) *ttnpb.OrganizationOrUserIdentifiers {
	organizationID, _ := flagSet.GetString("organization-id")
	userID, _ := flagSet.GetString("user-id")
	if organizationID == "" && userID == "" {
		return nil
	}
	if organizationID != "" && userID != "" {
		logger.Warn("Don't set organization ID and user ID at the same time, assuming user ID")
	}
	if userID != "" {
		return ttnpb.UserIdentifiers{UserID: userID}.OrganizationOrUserIdentifiers()
	}
	return ttnpb.OrganizationIdentifiers{OrganizationID: organizationID}.OrganizationOrUserIdentifiers()
}

func attributesFlags() *pflag.FlagSet {
	flagSet := new(pflag.FlagSet)
	flagSet.StringSlice("attributes", nil, "key=value")
	return flagSet
}

func mergeAttributes(attributes map[string]string, flagSet *pflag.FlagSet) map[string]string {
	kv, _ := flagSet.GetStringSlice("attributes")
	out := make(map[string]string, len(attributes)+len(kv))
	for k, v := range attributes {
		out[k] = v
	}
	for _, kv := range kv {
		kv := strings.SplitN(kv, "=", 2)
		if kv[1] == "" {
			delete(out, kv[0])
		} else {
			out[kv[0]] = kv[1]
		}
	}
	return out
}
