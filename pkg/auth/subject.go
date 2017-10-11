// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import "strings"

const (
	UserPrefix        = "user"
	ApplicationPrefix = "application"
	GatewayPrefix     = "gateway"
	ClientPrefix      = "client"
	sep               = ":"
)

func splitprefix(prefix, sub string) string {
	p := prefix + sep
	if strings.HasPrefix(sub, p) {
		return strings.TrimPrefix(sub, p)
	}

	return ""
}

func UserSubject(username string) string {
	return UserPrefix + sep + username
}

func ApplicationSubject(appID string) string {
	return ApplicationPrefix + sep + appID
}

func GatewaySubject(gwID string) string {
	return ApplicationPrefix + sep + gwID
}

func ClientSubject(clientID string) string {
	return ApplicationPrefix + sep + clientID
}
