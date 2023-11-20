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
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
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

type nonceKeyType struct{}

var nonceKey nonceKeyType

// WithNonce constructs a *http.Request which has a nonce attached.
func WithNonce(r *http.Request) (*http.Request, string) {
	nonce := GenerateNonce()
	return r.WithContext(context.WithValue(r.Context(), nonceKey, nonce)), nonce
}

// ContentSecurityPolicy contains the Content Security Policy.
type ContentSecurityPolicy struct {
	ConnectionSource []string
	StyleSource      []string
	ScriptSource     []string
	BaseURI          []string
	FrameAncestors   []string
}

// Clean de-duplicates and removes empty entries from the policy.
func (csp ContentSecurityPolicy) Clean() ContentSecurityPolicy {
	cleanDirective := func(contents []string) []string {
		added := map[string]struct{}{}
		cleanContents := []string{}
		for _, entry := range contents {
			if entry == "" || strings.HasPrefix(entry, "/") {
				continue // Skip empty and relative locations.
			}
			if strings.HasPrefix(entry, "http://") || strings.HasPrefix(entry, "https://") {
				if parsed, err := url.Parse(entry); err == nil {
					entry = parsed.Host
				}
			}
			if strings.HasPrefix(entry, "ws://") || strings.HasPrefix(entry, "wss://") {
				if parsed, err := url.Parse(entry); err == nil {
					parsed.Path, parsed.RawPath = "", ""
					entry = parsed.String()
				}
			}
			if _, ok := added[entry]; ok {
				continue // Skip already added locations.
			}
			added[entry] = struct{}{}
			cleanContents = append(cleanContents, entry)
		}
		return cleanContents
	}
	derived := csp
	derived.ConnectionSource = cleanDirective(csp.ConnectionSource)
	derived.StyleSource = cleanDirective(csp.StyleSource)
	derived.ScriptSource = cleanDirective(csp.ScriptSource)
	derived.BaseURI = cleanDirective(csp.BaseURI)
	derived.FrameAncestors = cleanDirective(csp.FrameAncestors)
	return derived
}

// Merge merges the provided policies into the existing one.
func (csp ContentSecurityPolicy) Merge(others ...ContentSecurityPolicy) ContentSecurityPolicy {
	derived := csp
	for _, other := range others {
		derived.ConnectionSource = append(derived.ConnectionSource, other.ConnectionSource...)
		derived.StyleSource = append(derived.StyleSource, other.StyleSource...)
		derived.ScriptSource = append(derived.ScriptSource, other.ScriptSource...)
		derived.BaseURI = append(derived.BaseURI, other.BaseURI...)
		derived.FrameAncestors = append(derived.FrameAncestors, other.FrameAncestors...)
	}
	return derived
}

// String returns the policy in string form.
func (csp ContentSecurityPolicy) String() string {
	appendPolicy := func(all []string, name string, contents []string) []string {
		if len(contents) == 0 {
			return all
		}
		return append(all, fmt.Sprintf("%s %s;", name, strings.Join(contents, " ")))
	}
	result := make([]string, 0, 5)
	result = appendPolicy(result, "connect-src", csp.ConnectionSource)
	result = appendPolicy(result, "style-src", csp.StyleSource)
	result = appendPolicy(result, "script-src", csp.ScriptSource)
	result = appendPolicy(result, "base-uri", csp.BaseURI)
	result = appendPolicy(result, "frame-ancestors", csp.FrameAncestors)
	return strings.Join(result, " ")
}

// RewriteScheme rewrites the scheme of the provided URL if it matches a rewrite rule.
func RewriteScheme(rewrites map[string]string, baseURL string) []string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return []string{baseURL}
	}
	rewrite, ok := rewrites[u.Scheme]
	if !ok {
		return []string{baseURL}
	}
	u.Scheme = rewrite
	return append(make([]string, 0, 2), baseURL, u.String())
}

// RewriteSchemes rewrites the scheme of the provided URLs if they match a rewrite rule.
func RewriteSchemes(rewrites map[string]string, baseURLs ...string) []string {
	urls := make([]string, 0, 2*len(baseURLs))
	for _, baseURL := range baseURLs {
		urls = append(urls, RewriteScheme(rewrites, baseURL)...)
	}
	return urls
}

// WebsocketSchemeRewrites contains the rewrite rules for websocket schemes.
var WebsocketSchemeRewrites = map[string]string{
	"http":  "ws",
	"https": "wss",
}
