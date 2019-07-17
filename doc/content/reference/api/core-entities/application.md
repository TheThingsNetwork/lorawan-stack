---
title: "Application"
description: "API core type Application reference."
weight: 2
category: [reference]
tags: [application, api]
---

* Services
 * [ApplicationAccess](#ApplicationAccess)
 * [ApplicationRegistry](#ApplicationRegistry)
* Messages
 * [Application](#ttn.lorawan.v3.Application)
 * [AttributesEntry](#ttn.lorawan.v3.Application.AttributesEntry)
 * [Applications](#ttn.lorawan.v3.Applications)
 * [CreateApplicationAPIKeyRequest](#ttn.lorawan.v3.CreateApplicationAPIKeyRequest)
 * [CreateApplicationRequest](#ttn.lorawan.v3.CreateApplicationRequest)
 * [GetApplicationAPIKeyRequest](#ttn.lorawan.v3.GetApplicationAPIKeyRequest)
 * [GetApplicationCollaboratorRequest](#ttn.lorawan.v3.GetApplicationCollaboratorRequest)
 * [GetApplicationRequest](#ttn.lorawan.v3.GetApplicationRequest)
 * [ListApplicationAPIKeysRequest](#ttn.lorawan.v3.ListApplicationAPIKeysRequest)
 * [ListApplicationCollaboratorsRequest](#ttn.lorawan.v3.ListApplicationCollaboratorsRequest)
 * [ListApplicationsRequest](#ttn.lorawan.v3.ListApplicationsRequest)
 * [SetApplicationCollaboratorRequest](#ttn.lorawan.v3.SetApplicationCollaboratorRequest)
 * [UpdateApplicationAPIKeyRequest](#ttn.lorawan.v3.UpdateApplicationAPIKeyRequest)
 * [UpdateApplicationRequest](#ttn.lorawan.v3.UpdateApplicationRequest)

## Services


    

### <a name="ApplicationAccess">ApplicationAccess</a>

    

    
#### <a name="ListRights">ListRights</a>



{{% reftab ListRights gRPCListRights HTTPListRights %}}

**Request**: [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers)

**Response**: [Rights](#ttn.lorawan.v3.Rights)

    $$$$$$

Method | Pattern | Body
------|-------|----
`GET` | `/api/v3/applications/{application_id}/rights` |  |

{{% /reftab %}}

    
#### <a name="CreateAPIKey">CreateAPIKey</a>



{{% reftab CreateAPIKey gRPCCreateAPIKey HTTPCreateAPIKey %}}

**Request**: [CreateApplicationAPIKeyRequest](#ttn.lorawan.v3.CreateApplicationAPIKeyRequest)

**Response**: [APIKey](#ttn.lorawan.v3.APIKey)

    $$$$$$

Method | Pattern | Body
------|-------|----
`POST` | `/api/v3/applications/{application_ids.application_id}/api-keys` | * |

{{% /reftab %}}

    
#### <a name="ListAPIKeys">ListAPIKeys</a>



{{% reftab ListAPIKeys gRPCListAPIKeys HTTPListAPIKeys %}}

**Request**: [ListApplicationAPIKeysRequest](#ttn.lorawan.v3.ListApplicationAPIKeysRequest)

**Response**: [APIKeys](#ttn.lorawan.v3.APIKeys)

    $$$$$$

Method | Pattern | Body
------|-------|----
`GET` | `/api/v3/applications/{application_ids.application_id}/api-keys` |  |

{{% /reftab %}}

    
#### <a name="GetAPIKey">GetAPIKey</a>



{{% reftab GetAPIKey gRPCGetAPIKey HTTPGetAPIKey %}}

**Request**: [GetApplicationAPIKeyRequest](#ttn.lorawan.v3.GetApplicationAPIKeyRequest)

**Response**: [APIKey](#ttn.lorawan.v3.APIKey)

    $$$$$$

Method | Pattern | Body
------|-------|----
`GET` | `/api/v3/applications/{application_ids.application_id}/api-keys/{key_id}` |  |

{{% /reftab %}}

    
#### <a name="UpdateAPIKey">UpdateAPIKey</a>

Update the rights of an existing application API key. To generate an API key,
the CreateAPIKey should be used. To delete an API key, update it
with zero rights.

{{% reftab UpdateAPIKey gRPCUpdateAPIKey HTTPUpdateAPIKey %}}

**Request**: [UpdateApplicationAPIKeyRequest](#ttn.lorawan.v3.UpdateApplicationAPIKeyRequest)

**Response**: [APIKey](#ttn.lorawan.v3.APIKey)

    $$$$$$

Method | Pattern | Body
------|-------|----
`PUT` | `/api/v3/applications/{application_ids.application_id}/api-keys/{api_key.id}` | * |

{{% /reftab %}}

    
#### <a name="GetCollaborator">GetCollaborator</a>

Get the rights of a collaborator (member) of the application.
Pseudo-rights in the response (such as the "_ALL" right) are not expanded.

{{% reftab GetCollaborator gRPCGetCollaborator HTTPGetCollaborator %}}

**Request**: [GetApplicationCollaboratorRequest](#ttn.lorawan.v3.GetApplicationCollaboratorRequest)

**Response**: [GetCollaboratorResponse](#ttn.lorawan.v3.GetCollaboratorResponse)

    $$$$$$

Method | Pattern | Body
------|-------|----
`GET` | `/api/v3/applications/{application_ids.application_id}/collaborator` |  |
`GET` | `/api/v3/applications/{application_ids.application_id}/collaborator/user/{collaborator.user_ids.user_id}` |  |
`GET` | `/api/v3/applications/{application_ids.application_id}/collaborator/organization/{collaborator.organization_ids.organization_id}` |  |

{{% /reftab %}}

    
#### <a name="SetCollaborator">SetCollaborator</a>

Set the rights of a collaborator (member) on the application.
Setting a collaborator without rights, removes them.

{{% reftab SetCollaborator gRPCSetCollaborator HTTPSetCollaborator %}}

**Request**: [SetApplicationCollaboratorRequest](#ttn.lorawan.v3.SetApplicationCollaboratorRequest)

**Response**: [.google.protobuf.Empty](#google.protobuf.Empty)

    $$$$$$

Method | Pattern | Body
------|-------|----
`PUT` | `/api/v3/applications/{application_ids.application_id}/collaborators` | * |

{{% /reftab %}}

    
#### <a name="ListCollaborators">ListCollaborators</a>



{{% reftab ListCollaborators gRPCListCollaborators HTTPListCollaborators %}}

**Request**: [ListApplicationCollaboratorsRequest](#ttn.lorawan.v3.ListApplicationCollaboratorsRequest)

**Response**: [Collaborators](#ttn.lorawan.v3.Collaborators)

    $$$$$$

Method | Pattern | Body
------|-------|----
`GET` | `/api/v3/applications/{application_ids.application_id}/collaborators` |  |

{{% /reftab %}}

    
  

### <a name="ApplicationRegistry">ApplicationRegistry</a>

    ApplicationRegistry is used to managed application

    
#### <a name="Create">Create</a>

Create a new application. This also sets the given organization or user as
first collaborator with all possible rights.

{{% reftab Create gRPCCreate HTTPCreate %}}

**Request**: [CreateApplicationRequest](#ttn.lorawan.v3.CreateApplicationRequest)

**Response**: [Application](#ttn.lorawan.v3.Application)

    $$$$$$

Method | Pattern | Body
------|-------|----
`POST` | `/api/v3/users/{collaborator.user_ids.user_id}/applications` | * |
`POST` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/applications` | * |

{{% /reftab %}}

    
#### <a name="Get">Get</a>

Get the application with the given identifiers, selecting the fields given
by the field mask. The method may return more or less fields, depending on
the rights of the caller.

{{% reftab Get gRPCGet HTTPGet %}}

**Request**: [GetApplicationRequest](#ttn.lorawan.v3.GetApplicationRequest)

**Response**: [Application](#ttn.lorawan.v3.Application)

    $$$$$$

Method | Pattern | Body
------|-------|----
`GET` | `/api/v3/applications/{application_ids.application_id}` |  |

{{% /reftab %}}

    
#### <a name="List">List</a>

List applications. See request message for details.

{{% reftab List gRPCList HTTPList %}}

**Request**: [ListApplicationsRequest](#ttn.lorawan.v3.ListApplicationsRequest)

**Response**: [Applications](#ttn.lorawan.v3.Applications)

    $$$$$$

Method | Pattern | Body
------|-------|----
`GET` | `/api/v3/applications` |  |
`GET` | `/api/v3/users/{collaborator.user_ids.user_id}/applications` |  |
`GET` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/applications` |  |

{{% /reftab %}}

    
#### <a name="Update">Update</a>



{{% reftab Update gRPCUpdate HTTPUpdate %}}

**Request**: [UpdateApplicationRequest](#ttn.lorawan.v3.UpdateApplicationRequest)

**Response**: [Application](#ttn.lorawan.v3.Application)

    $$$$$$

Method | Pattern | Body
------|-------|----
`PUT` | `/api/v3/applications/{application.ids.application_id}` | * |

{{% /reftab %}}

    
#### <a name="Delete">Delete</a>



{{% reftab Delete gRPCDelete HTTPDelete %}}

**Request**: [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers)

**Response**: [.google.protobuf.Empty](#google.protobuf.Empty)

    $$$$$$

Method | Pattern | Body
------|-------|----
`DELETE` | `/api/v3/applications/{application_id}` |  |

{{% /reftab %}}

    
  






































{{% refswitcher %}}

## Messages
 
### <a name="ttn.lorawan.v3.Application">Application</a>

  Application is the message that defines an Application in the network.

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
description | [string](#string) |  |  | <p>`string.max_len`: `2000`</p>
attributes | [Application.AttributesEntry](#ttn.lorawan.v3.Application.AttributesEntry) | repeated |  | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  | 
### <a name="ttn.lorawan.v3.Application.AttributesEntry">AttributesEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 
### <a name="ttn.lorawan.v3.Applications">Applications</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
applications | [Application](#ttn.lorawan.v3.Application) | repeated |  | 
### <a name="ttn.lorawan.v3.CreateApplicationAPIKeyRequest">CreateApplicationAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | <p>`repeated.items.enum.defined_only`: `true`</p>
### <a name="ttn.lorawan.v3.CreateApplicationRequest">CreateApplicationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application | [Application](#ttn.lorawan.v3.Application) |  |  | <p>`message.required`: `true`</p>
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created application. | <p>`message.required`: `true`</p>
### <a name="ttn.lorawan.v3.GetApplicationAPIKeyRequest">GetApplicationAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
key_id | [string](#string) |  | Unique public identifier for the API key. | 
### <a name="ttn.lorawan.v3.GetApplicationCollaboratorRequest">GetApplicationCollaboratorRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  | <p>`message.required`: `true`</p>
### <a name="ttn.lorawan.v3.GetApplicationRequest">GetApplicationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
### <a name="ttn.lorawan.v3.ListApplicationAPIKeysRequest">ListApplicationAPIKeysRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 
### <a name="ttn.lorawan.v3.ListApplicationCollaboratorsRequest">ListApplicationCollaboratorsRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 
### <a name="ttn.lorawan.v3.ListApplicationsRequest">ListApplicationsRequest</a>

  By default we list all applications the caller has rights on.
Set the user or the organization (not both) to instead list the applications
where the user or organization is collaborator on.

Field | Type | Label | Description | Validation
---|---|---|---|---
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  | 
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. | 
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 
### <a name="ttn.lorawan.v3.SetApplicationCollaboratorRequest">SetApplicationCollaboratorRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
collaborator | [Collaborator](#ttn.lorawan.v3.Collaborator) |  |  | <p>`message.required`: `true`</p>
### <a name="ttn.lorawan.v3.UpdateApplicationAPIKeyRequest">UpdateApplicationAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  | <p>`message.required`: `true`</p>
### <a name="ttn.lorawan.v3.UpdateApplicationRequest">UpdateApplicationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application | [Application](#ttn.lorawan.v3.Application) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

