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

package webui

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// GenerateNonce returns a nonce used for inline scripts.
func GenerateNonce() string {
	var b [20]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b[:])
}

// CleanCSP de-duplicates and removes empty entries from the CSP directive map.
func CleanCSP(csp map[string][]string) map[string][]string {
	for directive, entries := range csp {
		occurred := map[string]bool{}
		cleanedDirective := []string{}
		for i := range entries {
			if !occurred[entries[i]] && entries[i] != "" {
				occurred[entries[i]] = true
				cleanedDirective = append(cleanedDirective, entries[i])
			}
		}
		csp[directive] = cleanedDirective
	}
	return csp
}

// GenerateCSPNonce returns a final csp string from map of directives.
func GenerateCSPString(csp map[string][]string, nonce string) string {
	resultList := make([]string, 0)
	for key, value := range csp {
		if key == "default-src" {
			value = append(value, fmt.Sprintf("'nonce-%s'", nonce))
		}
		resultList = append(resultList, fmt.Sprintf("%s %s;", key, strings.Join(value, " ")))
	}
	return strings.Join(resultList, " ")
}
