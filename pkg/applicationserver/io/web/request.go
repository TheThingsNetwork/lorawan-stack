// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package web

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/jtacoma/uritemplates"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

const (
	downlinkKeyHeader     = "X-Downlink-Apikey"
	downlinkPushHeader    = "X-Downlink-Push"
	downlinkReplaceHeader = "X-Downlink-Replace"

	domainHeader = "X-Tts-Domain"
)

func webhookMessage(
	msg *ttnpb.ApplicationUp, hook *ttnpb.ApplicationWebhook,
) *ttnpb.ApplicationWebhook_Message {
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_UplinkMessage:
		return hook.UplinkMessage
	case *ttnpb.ApplicationUp_UplinkNormalized:
		return hook.UplinkNormalized
	case *ttnpb.ApplicationUp_JoinAccept:
		return hook.JoinAccept
	case *ttnpb.ApplicationUp_DownlinkAck:
		return hook.DownlinkAck
	case *ttnpb.ApplicationUp_DownlinkNack:
		return hook.DownlinkNack
	case *ttnpb.ApplicationUp_DownlinkSent:
		return hook.DownlinkSent
	case *ttnpb.ApplicationUp_DownlinkFailed:
		return hook.DownlinkFailed
	case *ttnpb.ApplicationUp_DownlinkQueued:
		return hook.DownlinkQueued
	case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
		return hook.DownlinkQueueInvalidated
	case *ttnpb.ApplicationUp_LocationSolved:
		return hook.LocationSolved
	case *ttnpb.ApplicationUp_ServiceData:
		return hook.ServiceData
	}
	return nil
}

func webhookUplinkMessageMask(msg *ttnpb.ApplicationUp) string {
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_UplinkMessage:
		return "up.uplink_message"
	case *ttnpb.ApplicationUp_UplinkNormalized:
		return "up.uplink_normalized"
	case *ttnpb.ApplicationUp_JoinAccept:
		return "up.join_accept"
	case *ttnpb.ApplicationUp_DownlinkAck:
		return "up.downlink_ack"
	case *ttnpb.ApplicationUp_DownlinkNack:
		return "up.downlink_nack"
	case *ttnpb.ApplicationUp_DownlinkSent:
		return "up.downlink_sent"
	case *ttnpb.ApplicationUp_DownlinkFailed:
		return "up.downlink_failed"
	case *ttnpb.ApplicationUp_DownlinkQueued:
		return "up.downlink_queued"
	case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
		return "up.downlink_queue_invalidated"
	case *ttnpb.ApplicationUp_LocationSolved:
		return "up.location_solved"
	case *ttnpb.ApplicationUp_ServiceData:
		return "up.service_data"
	}
	return ""
}

func expandVariables(u string, up *ttnpb.ApplicationUp) (*url.URL, error) {
	var joinEUI, devEUI, devAddr string
	if up.EndDeviceIds.JoinEui != nil {
		joinEUI = types.MustEUI64(up.EndDeviceIds.JoinEui).String()
	}
	if up.EndDeviceIds.DevEui != nil {
		devEUI = types.MustEUI64(up.EndDeviceIds.DevEui).String()
	}
	if up.EndDeviceIds.DevAddr != nil {
		devAddr = types.MustDevAddr(up.EndDeviceIds.DevAddr).String()
	}
	tmpl, err := uritemplates.Parse(u)
	if err != nil {
		return nil, err
	}
	expanded, err := tmpl.Expand(map[string]any{
		"appID":         up.EndDeviceIds.ApplicationIds.ApplicationId,
		"applicationID": up.EndDeviceIds.ApplicationIds.ApplicationId,
		"appEUI":        joinEUI,
		"joinEUI":       joinEUI,
		"devID":         up.EndDeviceIds.DeviceId,
		"deviceID":      up.EndDeviceIds.DeviceId,
		"devEUI":        devEUI,
		"devAddr":       devAddr,
	})
	if err != nil {
		return nil, err
	}
	return url.Parse(expanded)
}

// NewRequest returns an HTTP request.
// This method returns nil, nil if the hook is not configured for the message.
func NewRequest(
	ctx context.Context, downlinks DownlinksConfig, msg *ttnpb.ApplicationUp, hook *ttnpb.ApplicationWebhook,
) (*http.Request, error) {
	cfg := webhookMessage(msg, hook)
	if cfg == nil {
		return nil, nil //nolint:nilnil
	}
	baseURL, err := expandVariables(hook.BaseUrl, msg)
	if err != nil {
		return nil, err
	}
	pathURL, err := expandVariables(cfg.Path, msg)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(pathURL.Path, "/") {
		// Trim the leading slash, in order to ensure that the path is not
		// interpreted as relative to the root of the URL.
		pathURL.Path = strings.TrimLeft(pathURL.Path, "/")
		// Add the "/" suffix here instead of the condition below in order
		// to treat the case in which the pathURL.Path is "/".
		if !strings.HasSuffix(baseURL.Path, "/") {
			baseURL.Path += "/"
		}
	}
	// If the path URL contains an actual path (i.e. is not only a query)
	// ensure that it does not override the top level path.
	if pathURL.Path != "" && !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	finalURL := baseURL.ResolveReference(pathURL)
	format, ok := formats[hook.Format]
	if !ok {
		return nil, errFormatNotFound.WithAttributes("format", hook.Format)
	}
	deviceIDs := msg.EndDeviceIds
	if paths := hook.FieldMask.GetPaths(); len(paths) > 0 {
		mask := webhookUplinkMessageMask(msg)
		included := ttnpb.IncludeFields(paths, mask)
		// Filter active oneof field paths by removing all `up` fields
		// and appending paths related to the active oneof `up` field
		paths = append(ttnpb.ExcludeSubFields(paths, "up"), included...)
		up := &ttnpb.ApplicationUp{}
		if err := up.SetFields(msg, paths...); err != nil {
			return nil, err
		}
		msg = up
	}
	buf, err := format.FromUp(msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, finalURL.String(), bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	for key, value := range hook.Headers {
		req.Header.Set(key, value)
	}
	if hook.DownlinkApiKey != "" {
		req.Header.Set(downlinkKeyHeader, hook.DownlinkApiKey)
		req.Header.Set(downlinkPushHeader, downlinks.URL(ctx, hook.Ids, deviceIDs, "push"))
		req.Header.Set(downlinkReplaceHeader, downlinks.URL(ctx, hook.Ids, deviceIDs, "replace"))
	}
	if domain := downlinks.Domain(ctx); domain != "" {
		req.Header.Set(domainHeader, domain)
	}
	req.Header.Set("Content-Type", format.ContentType)
	return req, nil
}
