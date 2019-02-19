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

var allowedFieldMaskPathsForRPC = map[string][]string{
	// Applications:
	"/ttn.lorawan.v3.ApplicationRegistry/Get":                 ApplicationFieldPathsNested,
	"/ttn.lorawan.v3.ApplicationRegistry/List":                ApplicationFieldPathsNested,
	"/ttn.lorawan.v3.ApplicationRegistry/Update":              ApplicationFieldPathsNested,
	"/ttn.lorawan.v3.EntityRegistrySearch/SearchApplications": ApplicationFieldPathsNested,

	// Application Webhooks:
	"/ttn.lorawan.v3.ApplicationWebhookRegistry/Get":  ApplicationWebhookFieldPathsNested,
	"/ttn.lorawan.v3.ApplicationWebhookRegistry/List": ApplicationWebhookFieldPathsNested,
	"/ttn.lorawan.v3.ApplicationWebhookRegistry/Set":  ApplicationWebhookFieldPathsNested,

	// Application Links:
	"/ttn.lorawan.v3.As/GetLink": ApplicationLinkFieldPathsNested,
	"/ttn.lorawan.v3.As/SetLink": ApplicationLinkFieldPathsNested,

	// Clients:
	"/ttn.lorawan.v3.ClientRegistry/Get":                 ClientFieldPathsNested,
	"/ttn.lorawan.v3.ClientRegistry/List":                ClientFieldPathsNested,
	"/ttn.lorawan.v3.ClientRegistry/Update":              ClientFieldPathsNested,
	"/ttn.lorawan.v3.EntityRegistrySearch/SearchClients": ClientFieldPathsNested,

	// End Devices:
	// TODO: Restrict field paths for IS/NS/AS/JS.
	"/ttn.lorawan.v3.AsEndDeviceRegistry/Get":                  EndDeviceFieldPathsNested,
	"/ttn.lorawan.v3.AsEndDeviceRegistry/Set":                  EndDeviceFieldPathsNested,
	"/ttn.lorawan.v3.EndDeviceRegistry/Get":                    EndDeviceFieldPathsNested,
	"/ttn.lorawan.v3.EndDeviceRegistry/List":                   EndDeviceFieldPathsNested,
	"/ttn.lorawan.v3.EndDeviceRegistry/Update":                 EndDeviceFieldPathsNested,
	"/ttn.lorawan.v3.EndDeviceRegistrySearch/SearchEndDevices": EndDeviceFieldPathsNested,
	"/ttn.lorawan.v3.JsEndDeviceRegistry/Get":                  EndDeviceFieldPathsNested,
	"/ttn.lorawan.v3.JsEndDeviceRegistry/Set":                  EndDeviceFieldPathsNested,
	"/ttn.lorawan.v3.NsEndDeviceRegistry/Get":                  EndDeviceFieldPathsNested,
	"/ttn.lorawan.v3.NsEndDeviceRegistry/Set":                  EndDeviceFieldPathsNested,

	// Gateways:
	"/ttn.lorawan.v3.EntityRegistrySearch/SearchGateways": GatewayFieldPathsNested,
	"/ttn.lorawan.v3.GatewayRegistry/Get":                 GatewayFieldPathsNested,
	"/ttn.lorawan.v3.GatewayRegistry/List":                GatewayFieldPathsNested,
	"/ttn.lorawan.v3.GatewayRegistry/Update":              GatewayFieldPathsNested,

	// Organizations:
	"/ttn.lorawan.v3.OrganizationRegistry/Get":                 OrganizationFieldPathsNested,
	"/ttn.lorawan.v3.OrganizationRegistry/List":                OrganizationFieldPathsNested,
	"/ttn.lorawan.v3.OrganizationRegistry/Update":              OrganizationFieldPathsNested,
	"/ttn.lorawan.v3.EntityRegistrySearch/SearchOrganizations": OrganizationFieldPathsNested,

	// Users:
	"/ttn.lorawan.v3.UserRegistry/Get":                 UserFieldPathsNested,
	"/ttn.lorawan.v3.UserRegistry/Update":              UserFieldPathsNested,
	"/ttn.lorawan.v3.EntityRegistrySearch/SearchUsers": UserFieldPathsNested,
}

// AllowedFieldMaskPathsForRPC returns the list of allowed field mask paths for
// the given RPC method.
func AllowedFieldMaskPathsForRPC(rpcFullMethod string) []string {
	return allowedFieldMaskPathsForRPC[rpcFullMethod]
}
