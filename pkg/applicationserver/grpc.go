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

package applicationserver

import (
	"context"
	"fmt"

	pbtypes "github.com/gogo/protobuf/types"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func removeDeprecatedPaths(ctx context.Context, fieldMask *pbtypes.FieldMask) *pbtypes.FieldMask {
	validPaths := make([]string, 0, len(fieldMask.GetPaths()))
nextPath:
	for _, path := range fieldMask.GetPaths() {
		for _, deprecated := range []string{
			"api_key",
			"network_server_address",
			"tls",
		} {
			if path == deprecated {
				warning.Add(ctx, fmt.Sprintf("field %v is deprecated", deprecated))
				continue nextPath
			}
			validPaths = append(validPaths, path)
		}
	}
	return &pbtypes.FieldMask{
		Paths: validPaths,
	}
}

// getLink calls the underlying link registry in order to retrieve the link.
// If the link is not found, an empty link is returned instead.
func (as *ApplicationServer) getLink(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string) (*ttnpb.ApplicationLink, error) {
	link, err := as.linkRegistry.Get(ctx, ids, paths)
	if err != nil && errors.IsNotFound(err) {
		return &ttnpb.ApplicationLink{}, nil
	} else if err != nil {
		return nil, err
	}
	return link, nil
}

// GetLink implements ttnpb.AsServer.
func (as *ApplicationServer) GetLink(ctx context.Context, req *ttnpb.GetApplicationLinkRequest) (*ttnpb.ApplicationLink, error) {
	if err := rights.RequireApplication(ctx, *req.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	req.FieldMask = removeDeprecatedPaths(ctx, req.FieldMask)
	return as.linkRegistry.Get(ctx, req.ApplicationIds, req.FieldMask.GetPaths())
}

// SetLink implements ttnpb.AsServer.
func (as *ApplicationServer) SetLink(ctx context.Context, req *ttnpb.SetApplicationLinkRequest) (*ttnpb.ApplicationLink, error) {
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "default_formatters.up_formatter_parameter") {
		if size := len(req.Link.GetDefaultFormatters().GetUpFormatterParameter()); size > as.config.Formatters.MaxParameterLength {
			return nil, errInvalidFieldValue.WithAttributes("field", "default_formatters.up_formatter_parameter").WithCause(
				errFormatterScriptTooLarge.WithAttributes("size", size, "max_size", as.config.Formatters.MaxParameterLength),
			)
		}
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "default_formatters.down_formatter_parameter") {
		if size := len(req.Link.GetDefaultFormatters().GetDownFormatterParameter()); size > as.config.Formatters.MaxParameterLength {
			return nil, errInvalidFieldValue.WithAttributes("field", "default_formatters.down_formatter_parameter").WithCause(
				errFormatterScriptTooLarge.WithAttributes("size", size, "max_size", as.config.Formatters.MaxParameterLength),
			)
		}
	}
	if err := rights.RequireApplication(ctx, *req.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	req.FieldMask = removeDeprecatedPaths(ctx, req.FieldMask)
	return as.linkRegistry.Set(ctx, req.ApplicationIds, ttnpb.ApplicationLinkFieldPathsTopLevel,
		func(*ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
			return req.Link, req.FieldMask.GetPaths(), nil
		},
	)
}

// DeleteLink implements ttnpb.AsServer.
func (as *ApplicationServer) DeleteLink(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, *ids, ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	_, err := as.linkRegistry.Set(ctx, ids, nil, func(link *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) { return nil, nil, nil })
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

var errLinkingNotImplemented = errors.DefineUnimplemented("linking_not_implemented", "linking is not implemented")

// GetLinkStats implements ttnpb.AsServer.
func (as *ApplicationServer) GetLinkStats(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.ApplicationLinkStats, error) {
	return nil, errLinkingNotImplemented.New()
}

// GetConfiguration implements ttnpb.AsServer.
func (as *ApplicationServer) GetConfiguration(ctx context.Context, _ *ttnpb.GetAsConfigurationRequest) (*ttnpb.GetAsConfigurationResponse, error) {
	config, err := as.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &ttnpb.GetAsConfigurationResponse{
		Configuration: config.toProto(),
	}, nil
}

// HandleUplink implements ttnpb.NsAsServer.
func (as *ApplicationServer) HandleUplink(ctx context.Context, req *ttnpb.NsAsHandleUplinkRequest) (*pbtypes.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}
	link, err := as.getLink(ctx, req.ApplicationUps[0].EndDeviceIds.ApplicationIds, []string{
		"default_formatters",
		"skip_payload_crypto",
	})
	if err != nil {
		return nil, err
	}
	for _, up := range req.ApplicationUps {
		if err := as.processUp(ctx, up, link); err != nil {
			return nil, err
		}
	}
	return ttnpb.Empty, nil
}
