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

package joinserver

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type applicationActivationSettingsRegistryServer struct {
	JS       *JoinServer
	kekLabel string
}

var errApplicationActivationSettingsNotFound = errors.DefineNotFound("application_activation_settings_not_found", "application activation settings not found")

// Get implements ttnpb.ApplicationActivationSettingsRegistryServer.
func (srv applicationActivationSettingsRegistryServer) Get(ctx context.Context, req *ttnpb.GetApplicationActivationSettingsRequest) (*ttnpb.ApplicationActivationSettings, error) {
	if err := rights.RequireApplication(ctx, *req.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS); err != nil {
		return nil, err
	}
	sets, err := srv.JS.applicationActivationSettings.GetByID(ctx, req.ApplicationIds, req.FieldMask.GetPaths())
	if errors.IsNotFound(err) {
		return nil, errApplicationActivationSettingsNotFound.WithCause(err)
	}
	if err != nil {
		return nil, err
	}
	kek, err := cryptoutil.UnwrapKeyEnvelope(ctx, sets.Kek, srv.JS.KeyVault)
	if err != nil {
		return nil, errUnwrapKey.WithCause(err)
	}
	sets.Kek = kek
	return sets, nil
}

var (
	errNoPaths    = errors.DefineInvalidArgument("no_paths", "no paths specified")
	errNoKEKLabel = errors.DefineInvalidArgument("no_kek_label", "no KEK label specified")
)

// Set implements ttnpb.ApplicationActivationSettingsRegistryServer.
func (srv applicationActivationSettingsRegistryServer) Set(ctx context.Context, req *ttnpb.SetApplicationActivationSettingsRequest) (*ttnpb.ApplicationActivationSettings, error) {
	if len(req.FieldMask.GetPaths()) == 0 {
		return nil, errInvalidFieldMask.WithCause(errNoPaths)
	}

	reqKEK := req.Settings.Kek
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "kek.key") && reqKEK != nil {
		if reqKEK.Key.IsZero() {
			return nil, errInvalidFieldValue.WithAttributes("field", "kek.key")
		}
		if err := ttnpb.RequireFields(req.FieldMask.GetPaths(), "kek_label"); err != nil {
			return nil, errInvalidFieldMask.WithCause(err)
		}
		if req.Settings.KekLabel == "" {
			return nil, errNoKEKLabel.New()
		}
	}

	if err := rights.RequireApplication(ctx, *req.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS); err != nil {
		return nil, err
	}

	sets := req.FieldMask.GetPaths()
	if ttnpb.HasAnyField(sets, "kek.key") && reqKEK != nil {
		kek, err := cryptoutil.WrapAES128Key(ctx, *reqKEK.Key, srv.kekLabel, srv.JS.KeyVault)
		if err != nil {
			return nil, errWrapKey.WithCause(err)
		}
		req.Settings.Kek = kek
		sets = append(req.FieldMask.GetPaths()[:0:0], req.FieldMask.GetPaths()...)
		sets = ttnpb.AddFields(sets,
			"kek.encrypted_key",
			"kek.kek_label",
		)
	}
	v, err := srv.JS.applicationActivationSettings.SetByID(ctx, req.ApplicationIds, req.FieldMask.GetPaths(), func(stored *ttnpb.ApplicationActivationSettings) (*ttnpb.ApplicationActivationSettings, []string, error) {
		return req.Settings, sets, nil
	})
	if err != nil {
		return nil, err
	}
	v.Kek = reqKEK
	return v, nil
}

// Delete implements ttnpb.ApplicationActivationSettingsRegistryServer.
func (srv applicationActivationSettingsRegistryServer) Delete(ctx context.Context, req *ttnpb.DeleteApplicationActivationSettingsRequest) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, *req.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS); err != nil {
		return nil, err
	}
	_, err := srv.JS.applicationActivationSettings.SetByID(ctx, req.ApplicationIds, nil, func(stored *ttnpb.ApplicationActivationSettings) (*ttnpb.ApplicationActivationSettings, []string, error) {
		if stored == nil {
			return nil, nil, errApplicationActivationSettingsNotFound.New()
		}
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}
