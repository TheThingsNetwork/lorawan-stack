// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

func (ids *ApplicationIdentifiers) ExtractRequestFields(m map[string]interface{}) {
	if ids == nil {
		return
	}
	m["application_id"] = ids.ApplicationID
}

func (ids *ClientIdentifiers) ExtractRequestFields(m map[string]interface{}) {
	if ids == nil {
		return
	}
	m["client_id"] = ids.ClientID
}

func (ids *EndDeviceIdentifiers) ExtractRequestFields(m map[string]interface{}) {
	if ids == nil {
		return
	}
	m["application_id"] = ids.ApplicationID
	m["device_id"] = ids.DeviceID
}

func (ids *GatewayIdentifiers) ExtractRequestFields(m map[string]interface{}) {
	if ids == nil {
		return
	}
	m["gateway_id"] = ids.GatewayID
}

func (ids *OrganizationIdentifiers) ExtractRequestFields(m map[string]interface{}) {
	if ids == nil {
		return
	}
	m["organization_id"] = ids.OrganizationID
}

func (ids *UserIdentifiers) ExtractRequestFields(m map[string]interface{}) {
	if ids == nil {
		return
	}
	m["user_id"] = ids.UserID
}

func extractCollaboratorFields(m map[string]interface{}, ids *OrganizationOrUserIdentifiers) {
	if ids == nil {
		return
	}
	switch oneof := ids.Ids.(type) {
	case nil:
	case *OrganizationOrUserIdentifiers_OrganizationIDs:
		m["collaborator_organization_id"] = oneof.OrganizationIDs.OrganizationID
	case *OrganizationOrUserIdentifiers_UserIDs:
		m["collaborator_user_id"] = oneof.UserIDs.UserID
	default:
		panic("missed oneof type in extractCollaboratorFields()")
	}
}

func (req *CreateApplicationRequest) ExtractRequestFields(m map[string]interface{}) {
	if req == nil {
		return
	}
	req.Application.ExtractRequestFields(m)
	extractCollaboratorFields(m, &req.Collaborator)
}

func (req *CreateClientRequest) ExtractRequestFields(m map[string]interface{}) {
	if req == nil {
		return
	}
	req.Client.ExtractRequestFields(m)
	extractCollaboratorFields(m, &req.Collaborator)
}

func (req *CreateGatewayRequest) ExtractRequestFields(m map[string]interface{}) {
	if req == nil {
		return
	}
	req.Gateway.ExtractRequestFields(m)
	extractCollaboratorFields(m, &req.Collaborator)
}

func (req *CreateOrganizationRequest) ExtractRequestFields(m map[string]interface{}) {
	if req == nil {
		return
	}
	req.Organization.ExtractRequestFields(m)
	extractCollaboratorFields(m, &req.Collaborator)
}

func (req *SetApplicationCollaboratorRequest) ExtractRequestFields(m map[string]interface{}) {
	if req == nil {
		return
	}
	req.ApplicationIdentifiers.ExtractRequestFields(m)
	extractCollaboratorFields(m, &req.Collaborator.OrganizationOrUserIdentifiers)
}

func (req *SetClientCollaboratorRequest) ExtractRequestFields(m map[string]interface{}) {
	if req == nil {
		return
	}
	req.ClientIdentifiers.ExtractRequestFields(m)
	extractCollaboratorFields(m, &req.Collaborator.OrganizationOrUserIdentifiers)
}

func (req *SetGatewayCollaboratorRequest) ExtractRequestFields(m map[string]interface{}) {
	if req == nil {
		return
	}
	req.GatewayIdentifiers.ExtractRequestFields(m)
	extractCollaboratorFields(m, &req.Collaborator.OrganizationOrUserIdentifiers)
}

func (req *SetOrganizationCollaboratorRequest) ExtractRequestFields(m map[string]interface{}) {
	if req == nil {
		return
	}
	req.OrganizationIdentifiers.ExtractRequestFields(m)
	extractCollaboratorFields(m, &req.Collaborator.OrganizationOrUserIdentifiers)
}
