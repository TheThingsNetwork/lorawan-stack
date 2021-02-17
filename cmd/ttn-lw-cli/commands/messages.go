// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

var (
	applicationPurgeWarning = []string{
		"This action will permanently delete the application and all related data (API keys, rights, attributes etc.)",
	}
	clientPurgeWarning = []string{
		"This action will permanently delete the client and all related data (rights, attributes etc.)",
	}
	gatewayPurgeWarning = []string{
		"This action will permanently delete the gateway and all related data (API keys, antennas, attributes etc.)",
	}
	organizationPurgeWarning = []string{
		"This action will permanently delete the organization and all related data (API keys, rights, attributes etc.)",
		"It might also cause entities to be orphaned if this organization is the only one that has full rights on the entity.",
	}
	userPurgeWarning = []string{
		"This action will permanently delete the user and all related data (API keys, entity rights, attributes etc.).",
		"It might also cause entities to be orphaned if this user is the only one that has full rights on the entity.",
	}
)
