// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

const (
	ApplicationsList                 Scope = "applications:list"
	ApplicationsCreate               Scope = "applications:create"
	ApplicationInfo                  Scope = "application:info"
	ApplicationSettingsBasic         Scope = "application:settings:basic"
	ApplicationSettingsKeys          Scope = "application:settings:keys"
	ApplicationSettingsCollaborators Scope = "application:settings:collaborators"
	ApplicationDelete                Scope = "application:delete"
	ApplicationTrafficRead           Scope = "application:traffic:read"
	ApplicationTrafficWrite          Scope = "application:traffic:write"
	ApplicationDevicesRead           Scope = "application:devices:read"
	ApplicationDevicesWrite          Scope = "application:devices:write"
)
