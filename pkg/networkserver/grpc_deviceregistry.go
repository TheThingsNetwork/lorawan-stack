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

package networkserver

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var (
	evtCreateEndDevice = events.Define(
		"ns.end_device.create", "create end device",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateEndDevice = events.Define(
		"ns.end_device.update", "update end device",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_DEVICES_READ),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteEndDevice = events.Define(
		"ns.end_device.delete", "delete end device",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

const maxRequiredDeviceReadRightCount = 3

func appendRequiredDeviceReadRights(rights []ttnpb.Right, gets ...string) []ttnpb.Right {
	if len(gets) == 0 {
		return rights
	}
	rights = append(rights,
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
	if ttnpb.HasAnyField(gets,
		"pending_session.queued_application_downlinks",
		"queued_application_downlinks",
		"session.queued_application_downlinks",
	) {
		rights = append(rights, ttnpb.RIGHT_APPLICATION_LINK)
	}
	if ttnpb.HasAnyField(gets,
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
	) {
		rights = append(rights, ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS)
	}
	return rights
}

func addDeviceGetPaths(paths ...string) []string {
	gets := paths
	if ttnpb.HasAnyField(paths,
		"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
	) {
		if ttnpb.HasAnyField(paths,
			"pending_session.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_session.keys.f_nwk_s_int_key.encrypted_key",
				"pending_session.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"pending_session.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_session.keys.nwk_s_enc_key.encrypted_key",
				"pending_session.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"pending_session.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_session.keys.s_nwk_s_int_key.encrypted_key",
				"pending_session.keys.s_nwk_s_int_key.kek_label",
			)
		}

		if ttnpb.HasAnyField(paths,
			"session.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"session.keys.f_nwk_s_int_key.encrypted_key",
				"session.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"session.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"session.keys.nwk_s_enc_key.encrypted_key",
				"session.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"session.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"session.keys.s_nwk_s_int_key.encrypted_key",
				"session.keys.s_nwk_s_int_key.kek_label",
			)
		}

		if ttnpb.HasAnyField(paths,
			"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.kek_label",
			)
		}

		if ttnpb.HasAnyField(paths,
			"mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"mac_state.queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
				"mac_state.queued_join_accept.keys.f_nwk_s_int_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
				"mac_state.queued_join_accept.keys.nwk_s_enc_key.kek_label",
			)
		}
		if ttnpb.HasAnyField(paths,
			"mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		) {
			gets = ttnpb.AddFields(gets,
				"mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
				"mac_state.queued_join_accept.keys.s_nwk_s_int_key.kek_label",
			)
		}
	}
	return gets
}

func unwrapSelectedSessionKeys(ctx context.Context, kv crypto.KeyVault, dev *ttnpb.EndDevice, paths ...string) error {
	if dev.PendingSession != nil && ttnpb.HasAnyField(paths,
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.PendingSession.SessionKeys, "pending_session.keys", paths...)
		if err != nil {
			return err
		}
		dev.PendingSession.SessionKeys = sk
	}
	if dev.Session != nil && ttnpb.HasAnyField(paths,
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.Session.SessionKeys, "session.keys", paths...)
		if err != nil {
			return err
		}
		dev.Session.SessionKeys = sk
	}
	if dev.PendingMACState.GetQueuedJoinAccept() != nil && ttnpb.HasAnyField(paths,
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.PendingMACState.QueuedJoinAccept.Keys, "pending_mac_state.queued_join_accept.keys", paths...)
		if err != nil {
			return err
		}
		dev.PendingMACState.QueuedJoinAccept.Keys = sk
	}
	return nil
}

// Get implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, appendRequiredDeviceReadRights(
		make([]ttnpb.Right, 0, maxRequiredDeviceReadRightCount),
		req.FieldMask.Paths...,
	)...); err != nil {
		return nil, err
	}

	dev, ctx, err := ns.devices.GetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, addDeviceGetPaths(req.FieldMask.Paths...))
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to get device from registry")
		return nil, err
	}
	if err := unwrapSelectedSessionKeys(ctx, ns.KeyVault, dev, req.FieldMask.Paths...); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to unwrap selected keys")
		return nil, err
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
}

func newInvalidFieldValueError(field string) errors.Error {
	return errInvalidFieldValue.WithAttributes("field", field)
}

type setDeviceState struct {
	Device *ttnpb.EndDevice

	paths     []string
	extraSets []string
	extraGets []string

	pathsCache     map[string]bool
	extraSetsCache map[string]bool
	extraGetsCache map[string]bool

	zeroPaths map[string]bool
	onGet     []func(*ttnpb.EndDevice) error
}

func hasAnyField(fs []string, cache map[string]bool, paths ...string) bool {
outer:
	for _, p := range paths {
		i := len(p)
		for ; i > 0; i = strings.LastIndex(p[:i], ".") {
			p := p[:i]
			v, ok := cache[p]
			if !ok {
				continue
			}
			if !v {
				continue outer
			}
			return true
		}
		v := ttnpb.HasAnyField(fs, p)
		cache[p] = v
		if v {
			return v
		}
	}
	return false
}

func (st *setDeviceState) hasPathField(paths ...string) bool {
	return hasAnyField(st.paths, st.pathsCache, paths...)
}

func (st *setDeviceState) HasSetField(paths ...string) bool {
	return st.hasPathField(paths...) || hasAnyField(st.extraSets, st.extraSetsCache, paths...)
}

func (st *setDeviceState) HasGetField(paths ...string) bool {
	return st.hasPathField(paths...) || hasAnyField(st.extraGets, st.extraGetsCache, paths...)
}

func addFields(hasField func(...string) bool, selected []string, cache map[string]bool, paths ...string) []string {
	for _, p := range paths {
		if hasField(p) {
			continue
		}
		cache[p] = true
		selected = append(selected, p)
	}
	return selected
}

func (st *setDeviceState) AddSetFields(paths ...string) {
	st.extraSets = addFields(st.HasSetField, st.extraSets, st.extraSetsCache, paths...)
}

func (st *setDeviceState) AddGetFields(paths ...string) {
	st.extraGets = addFields(st.HasGetField, st.extraGets, st.extraGetsCache, paths...)
}

func (st *setDeviceState) SetFields() []string {
	return append(st.paths, st.extraSets...)
}

func (st *setDeviceState) GetFields() []string {
	return append(st.paths, st.extraGets...)
}

func (st *setDeviceState) RequireFields(paths ...string) error {
	if err := ttnpb.RequireFields(st.SetFields(), paths...); err != nil {
		return errInvalidFieldMask.WithCause(err)
	}
	return nil
}

func (st *setDeviceState) WithField(f func(*ttnpb.EndDevice) error, path string) error {
	if st.HasSetField(path) {
		return f(st.Device)
	}
	st.AddGetFields(path)
	st.onGet = append(st.onGet, func(stored *ttnpb.EndDevice) error {
		return f(stored)
	})
	return nil
}

func (st *setDeviceState) WithFields(f func(map[string]*ttnpb.EndDevice) error, paths ...string) error {
	storedPaths := make([]string, 0, len(paths))
	m := make(map[string]*ttnpb.EndDevice, len(paths))
	for _, p := range paths {
		if st.HasSetField(p) {
			m[p] = st.Device
		} else {
			storedPaths = append(storedPaths, p)
		}
	}
	if len(storedPaths) == 0 {
		return f(m)
	}
	st.AddGetFields(storedPaths...)
	st.onGet = append(st.onGet, func(stored *ttnpb.EndDevice) error {
		if stored == nil {
			return f(m)
		}
		for _, p := range storedPaths {
			m[p] = stored
		}
		return f(m)
	})
	return nil
}

func (st *setDeviceState) ValidateField(isValid func(*ttnpb.EndDevice) bool, path string) error {
	return st.WithField(func(dev *ttnpb.EndDevice) error {
		if !isValid(dev) {
			return newInvalidFieldValueError(path)
		}
		return nil
	}, path)
}

func (st *setDeviceState) ValidateFieldIsZero(path string) error {
	if st.HasSetField(path) {
		if !st.Device.FieldIsZero(path) {
			return newInvalidFieldValueError(path)
		}
		return nil
	}
	v, ok := st.zeroPaths[path]
	if !ok {
		st.zeroPaths[path] = true
		return nil
	}
	if !v {
		panic(fmt.Sprintf("path `%s` requested to be both zero and not zero", path))
	}
	return nil
}

func (st *setDeviceState) ValidateFieldIsNotZero(path string) error {
	if st.HasSetField(path) {
		if st.Device.FieldIsZero(path) {
			return newInvalidFieldValueError(path)
		}
		return nil
	}
	v, ok := st.zeroPaths[path]
	if !ok {
		st.zeroPaths[path] = false
		return nil
	}
	if v {
		panic(fmt.Sprintf("path `%s` requested to be both zero and not zero", path))
	}
	return nil
}

func (st *setDeviceState) ValidateFieldsAreZero(paths ...string) error {
	for _, p := range paths {
		if err := st.ValidateFieldIsZero(p); err != nil {
			return err
		}
	}
	return nil
}

func (st *setDeviceState) ValidateFieldsAreNotZero(paths ...string) error {
	for _, p := range paths {
		if err := st.ValidateFieldIsNotZero(p); err != nil {
			return err
		}
	}
	return nil
}

func (st *setDeviceState) ValidateFields(isValid func(map[string]*ttnpb.EndDevice) (bool, string), paths ...string) error {
	return st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
		ok, p := isValid(m)
		if !ok {
			return newInvalidFieldValueError(p)
		}
		return nil
	}, paths...)
}

func (st *setDeviceState) ValidateSetField(isValid func() bool, path string) error {
	if !st.HasSetField(path) {
		return nil
	}
	if !isValid() {
		return newInvalidFieldValueError(path)
	}
	return nil
}

func (st *setDeviceState) ValidateSetFieldWithCause(isValid func() error, path string) error {
	if !st.HasSetField(path) {
		return nil
	}
	if err := isValid(); err != nil {
		return newInvalidFieldValueError(path).WithCause(err)
	}
	return nil
}

func (st *setDeviceState) ValidateSetFields(isValid func(map[string]*ttnpb.EndDevice) (bool, string), paths ...string) error {
	if !st.HasSetField(paths...) {
		return nil
	}
	return st.ValidateFields(isValid, paths...)
}

// ValidateIfZeroThenZero ensures FieldIsZero(left) -> FieldIsZero(r), for each r in right.
func (st *setDeviceState) ValidateIfZeroThenZero(left string, right ...string) error {
	if st.HasSetField(left) {
		if !st.Device.FieldIsZero(left) {
			return nil
		}
		return st.ValidateFieldsAreZero(right...)
	}
	for _, r := range right {
		if !st.HasSetField(r) || st.Device.FieldIsZero(r) {
			continue
		}
		if err := st.ValidateFieldIsNotZero(left); err != nil {
			return err
		}
	}
	return nil
}

// ValidateIfZeroThenNotZero ensures FieldIsZero(left) -> !FieldIsZero(r), for each r in right.
func (st *setDeviceState) ValidateIfZeroThenNotZero(left string, right ...string) error {
	if st.HasSetField(left) {
		if !st.Device.FieldIsZero(left) {
			return nil
		}
		return st.ValidateFieldsAreNotZero(right...)
	}
	for _, r := range right {
		if !st.HasSetField(r) || !st.Device.FieldIsZero(r) {
			continue
		}
		if err := st.ValidateFieldIsNotZero(left); err != nil {
			return err
		}
	}
	return nil
}

// ValidateIfNotZeroThenZero ensures !FieldIsZero(left) -> FieldIsZero(r), for each r in right.
func (st *setDeviceState) ValidateIfNotZeroThenZero(left string, right ...string) error {
	if st.HasSetField(left) {
		if st.Device.FieldIsZero(left) {
			return nil
		}
		return st.ValidateFieldsAreZero(right...)
	}
	for _, r := range right {
		if !st.HasSetField(r) || st.Device.FieldIsZero(r) {
			continue
		}
		if err := st.ValidateFieldIsZero(left); err != nil {
			return err
		}
	}
	return nil
}

// ValidateIfNotZeroThenNotZero ensures !FieldIsZero(left) -> !FieldIsZero(r), for each r in right.
func (st *setDeviceState) ValidateIfNotZeroThenNotZero(left string, right ...string) error {
	if st.HasSetField(left) {
		if st.Device.FieldIsZero(left) {
			return nil
		}
		return st.ValidateFieldsAreNotZero(right...)
	}
	for _, r := range right {
		if !st.HasSetField(r) || !st.Device.FieldIsZero(r) {
			continue
		}
		if err := st.ValidateFieldIsZero(left); err != nil {
			return err
		}
	}
	return nil
}

// ValidateIfZeroThenFunc ensures FieldIsZero(left) -> f(map r -> *ttnpb.EndDevice), for each r in right.
func (st *setDeviceState) ValidateIfZeroThenFunc(isValid func(map[string]*ttnpb.EndDevice) (bool, string), left string, right ...string) error {
	if st.HasSetField(left) {
		if !st.Device.FieldIsZero(left) {
			return nil
		}
		return st.ValidateFields(isValid, right...)
	}
	if !st.HasSetField(right...) {
		return nil
	}
	return st.ValidateFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		if !m[left].FieldIsZero(left) {
			return true, ""
		}
		return isValid(m)
	}, append([]string{left}, right...)...)
}

// ValidateIfNotZeroThenFunc ensures !FieldIsZero(left) -> f(map r -> *ttnpb.EndDevice), for each r in right.
func (st *setDeviceState) ValidateIfNotZeroThenFunc(isValid func(map[string]*ttnpb.EndDevice) (bool, string), left string, right ...string) error {
	if st.HasSetField(left) {
		if st.Device.FieldIsZero(left) {
			return nil
		}
		return st.ValidateFields(isValid, right...)
	}
	if !st.HasSetField(right...) {
		return nil
	}
	return st.ValidateFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		if m[left].FieldIsZero(left) {
			return true, ""
		}
		return isValid(m)
	}, append([]string{left}, right...)...)
}

// SetFunc is the function meant to be passed to SetByID.
func (st *setDeviceState) SetFunc(f func(context.Context, *ttnpb.EndDevice) error) func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
	return func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		for p, shouldBeZero := range st.zeroPaths {
			if stored.FieldIsZero(p) != shouldBeZero {
				return nil, nil, newInvalidFieldValueError(p)
			}
		}
		for _, g := range st.onGet {
			if err := g(stored); err != nil {
				return nil, nil, err
			}
		}
		if err := f(ctx, stored); err != nil {
			return nil, nil, err
		}
		return st.Device, st.SetFields(), nil
	}
}

func newSetDeviceState(dev *ttnpb.EndDevice, paths ...string) *setDeviceState {
	return &setDeviceState{
		Device: dev,
		paths:  paths,

		pathsCache:     make(map[string]bool),
		extraSetsCache: make(map[string]bool),
		extraGetsCache: make(map[string]bool),

		zeroPaths: make(map[string]bool),
	}
}

func setKeyIsZero(m map[string]*ttnpb.EndDevice, get func(*ttnpb.EndDevice) *ttnpb.KeyEnvelope, path string) bool {
	if dev, ok := m[path+".key"]; ok {
		if ke := get(dev); !ke.Key.IsZero() {
			return false
		}
	}
	if dev, ok := m[path+".encrypted_key"]; ok {
		if ke := get(dev); len(ke.EncryptedKey) != 0 {
			return false
		}
	}
	return true
}

func setKeyEqual(m map[string]*ttnpb.EndDevice, getA, getB func(*ttnpb.EndDevice) *ttnpb.KeyEnvelope, pathA, pathB string) bool {
	if a, b := getA(m[pathA+".key"]).GetKey(), getB(m[pathB+".key"]).GetKey(); a == nil && b != nil || a != nil && b == nil || a != nil && b != nil && !a.Equal(*b) {
		return false
	}
	if a, b := getA(m[pathA+".encrypted_key"]).GetEncryptedKey(), getB(m[pathB+".encrypted_key"]).GetEncryptedKey(); !bytes.Equal(a, b) {
		return false
	}
	if a, b := getA(m[pathA+".kek_label"]).GetKEKLabel(), getB(m[pathB+".kek_label"]).GetKEKLabel(); a != b {
		return false
	}
	return true
}

// Set implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	st := newSetDeviceState(&req.EndDevice, req.FieldMask.Paths...)

	requiredRights := append(make([]ttnpb.Right, 0, 2),
		ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	)
	if st.HasSetField(
		"pending_mac_state.queued_join_accept.keys.app_s_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.app_s_key.kek_label",
		"pending_mac_state.queued_join_accept.keys.app_s_key.key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.session_key_id",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
		"pending_session.keys.session_key_id",
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
		"session.keys.session_key_id",
	) {
		requiredRights = append(requiredRights, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS)
	}
	if err := rights.RequireApplication(ctx, st.Device.ApplicationIdentifiers, requiredRights...); err != nil {
		return nil, err
	}

	// Account for CLI not sending ids.* paths.
	st.AddSetFields(
		"ids.application_ids",
		"ids.device_id",
	)
	if st.Device.JoinEUI != nil {
		st.AddSetFields(
			"ids.join_eui",
		)
	}
	if st.Device.DevEUI != nil {
		st.AddSetFields(
			"ids.dev_eui",
		)
	}
	if st.Device.DevAddr != nil {
		st.AddSetFields(
			"ids.dev_addr",
		)
	}

	if err := st.ValidateSetField(
		func() bool { return st.Device.FrequencyPlanID != "" },
		"frequency_plan_id",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		st.Device.LoRaWANPHYVersion.Validate,
		"lorawan_phy_version",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		st.Device.LoRaWANVersion.Validate,
		"lorawan_version",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		func() error {
			if st.Device.MACState == nil {
				return nil
			}
			return st.Device.MACState.LoRaWANVersion.Validate()
		},
		"mac_state.lorawan_version",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		func() error {
			if st.Device.PendingMACState == nil {
				return nil
			}
			return st.Device.PendingMACState.LoRaWANVersion.Validate()
		},
		"pending_mac_state.lorawan_version",
	); err != nil {
		return nil, err
	}

	switch {
	case st.HasSetField(
		"ids.dev_eui",
	):
		if st.Device.DevEUI != nil && !st.Device.DevEUI.IsZero() {
			break
		}
		if st.HasSetField(
			"supports_join",
		) {
			if st.Device.SupportsJoin {
				return nil, errNoDevEUI.New()
			} else {
				if err := st.ValidateField(func(dev *ttnpb.EndDevice) bool {
					return !dev.LoRaWANVersion.RequireDevEUIForABP()
				}, "lorawan_version"); err != nil {
					return nil, err
				}
			}
		} else {
			if err := st.ValidateFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
				if m["supports_join"].GetSupportsJoin() {
					return false, "supports_join"
				} else if m["lorawan_version"].LoRaWANVersion.RequireDevEUIForABP() {
					return false, "lorawan_version"
				}
				return true, ""
			},
				"lorawan_version",
				"supports_join",
			); err != nil {
				return nil, err
			}
		}
	case st.HasSetField(
		"lorawan_version",
		"supports_join",
	):
		if err := st.ValidateFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
			if dev, ok := m["ids.dev_eui"]; ok && dev.DevEUI != nil && !dev.DevEUI.IsZero() {
				return true, ""
			}
			if m["supports_join"].GetSupportsJoin() || m["lorawan_version"].LoRaWANVersion.RequireDevEUIForABP() {
				return false, "ids.dev_eui"
			}
			return true, ""
		},
			"ids.dev_eui",
			"lorawan_version",
			"supports_join",
		); err != nil {
			return nil, err
		}
	}

	if st.HasSetField("ids.dev_addr") {
		if err := st.ValidateField(func(dev *ttnpb.EndDevice) bool {
			if st.Device.DevAddr == nil {
				return dev.Session == nil
			}
			return dev.GetSession() != nil && dev.Session.DevAddr.Equal(*st.Device.DevAddr)
		}, "session.dev_addr"); err != nil {
			return nil, err
		}
	} else if st.HasSetField("session.dev_addr") {
		var devAddr *types.DevAddr
		if st.Device.Session != nil {
			devAddr = &st.Device.Session.DevAddr
		}
		st.Device.DevAddr = devAddr
		st.AddSetFields(
			"ids.dev_addr",
		)
	}

	if err := st.ValidateIfNotZeroThenNotZero("supports_join",
		"ids.join_eui",
	); err != nil {
		return nil, err
	}

	if err := st.ValidateIfZeroThenZero("supports_join",
		"pending_mac_state.current_parameters.adr_ack_delay_exponent.value",
		"pending_mac_state.current_parameters.adr_ack_limit_exponent.value",
		"pending_mac_state.current_parameters.adr_data_rate_index",
		"pending_mac_state.current_parameters.adr_nb_trans",
		"pending_mac_state.current_parameters.adr_tx_power_index",
		"pending_mac_state.current_parameters.beacon_frequency",
		"pending_mac_state.current_parameters.channels",
		"pending_mac_state.current_parameters.downlink_dwell_time.value",
		"pending_mac_state.current_parameters.max_duty_cycle",
		"pending_mac_state.current_parameters.max_eirp",
		"pending_mac_state.current_parameters.ping_slot_data_rate_index_value.value",
		"pending_mac_state.current_parameters.ping_slot_frequency",
		"pending_mac_state.current_parameters.rejoin_count_periodicity",
		"pending_mac_state.current_parameters.rejoin_time_periodicity",
		"pending_mac_state.current_parameters.rx1_data_rate_offset",
		"pending_mac_state.current_parameters.rx1_delay",
		"pending_mac_state.current_parameters.rx2_data_rate_index",
		"pending_mac_state.current_parameters.rx2_frequency",
		"pending_mac_state.current_parameters.uplink_dwell_time.value",
		"pending_mac_state.desired_parameters.adr_ack_delay_exponent.value",
		"pending_mac_state.desired_parameters.adr_ack_limit_exponent.value",
		"pending_mac_state.desired_parameters.adr_data_rate_index",
		"pending_mac_state.desired_parameters.adr_nb_trans",
		"pending_mac_state.desired_parameters.adr_tx_power_index",
		"pending_mac_state.desired_parameters.beacon_frequency",
		"pending_mac_state.desired_parameters.channels",
		"pending_mac_state.desired_parameters.downlink_dwell_time.value",
		"pending_mac_state.desired_parameters.max_duty_cycle",
		"pending_mac_state.desired_parameters.max_eirp",
		"pending_mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
		"pending_mac_state.desired_parameters.ping_slot_frequency",
		"pending_mac_state.desired_parameters.rejoin_count_periodicity",
		"pending_mac_state.desired_parameters.rejoin_time_periodicity",
		"pending_mac_state.desired_parameters.rx1_data_rate_offset",
		"pending_mac_state.desired_parameters.rx1_delay",
		"pending_mac_state.desired_parameters.rx2_data_rate_index",
		"pending_mac_state.desired_parameters.rx2_frequency",
		"pending_mac_state.desired_parameters.uplink_dwell_time.value",
		"pending_mac_state.device_class",
		"pending_mac_state.last_adr_change_f_cnt_up",
		"pending_mac_state.last_confirmed_downlink_at",
		"pending_mac_state.last_dev_status_f_cnt_up",
		"pending_mac_state.last_downlink_at",
		"pending_mac_state.last_network_initiated_downlink_at",
		"pending_mac_state.lorawan_version",
		"pending_mac_state.pending_join_request.cf_list.ch_masks",
		"pending_mac_state.pending_join_request.cf_list.freq",
		"pending_mac_state.pending_join_request.cf_list.type",
		"pending_mac_state.pending_join_request.downlink_settings.opt_neg",
		"pending_mac_state.pending_join_request.downlink_settings.rx1_dr_offset",
		"pending_mac_state.pending_join_request.downlink_settings.rx2_dr",
		"pending_mac_state.pending_join_request.rx_delay",
		"pending_mac_state.ping_slot_periodicity.value",
		"pending_mac_state.queued_join_accept.correlation_ids",
		"pending_mac_state.queued_join_accept.keys.app_s_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.app_s_key.kek_label",
		"pending_mac_state.queued_join_accept.keys.app_s_key.key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.session_key_id",
		"pending_mac_state.queued_join_accept.payload",
		"pending_mac_state.queued_join_accept.request.cf_list.ch_masks",
		"pending_mac_state.queued_join_accept.request.cf_list.freq",
		"pending_mac_state.queued_join_accept.request.cf_list.type",
		"pending_mac_state.queued_join_accept.request.dev_addr",
		"pending_mac_state.queued_join_accept.request.downlink_settings.opt_neg",
		"pending_mac_state.queued_join_accept.request.downlink_settings.rx1_dr_offset",
		"pending_mac_state.queued_join_accept.request.downlink_settings.rx2_dr",
		"pending_mac_state.queued_join_accept.request.net_id",
		"pending_mac_state.queued_join_accept.request.rx_delay",
		"pending_mac_state.recent_downlinks",
		"pending_mac_state.recent_uplinks",
		"pending_mac_state.rejected_adr_data_rate_indexes",
		"pending_mac_state.rejected_adr_tx_power_indexes",
		"pending_mac_state.rejected_data_rate_ranges",
		"pending_mac_state.rejected_frequencies",
		"pending_mac_state.rx_windows_available",
		"pending_session.dev_addr",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
		"pending_session.keys.session_key_id",
		"session.keys.session_key_id",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateIfZeroThenNotZero("supports_join",
		"session.dev_addr",
		"session.keys.f_nwk_s_int_key.key",
	); err != nil {
		return nil, err
	}

	if err := st.ValidateIfNotZeroThenZero("multicast",
		"mac_state.last_adr_change_f_cnt_up",
		"mac_state.last_confirmed_downlink_at",
		"mac_state.last_dev_status_f_cnt_up",
		"mac_state.pending_application_downlink",
		"mac_state.pending_requests",
		"mac_state.queued_responses",
		"mac_state.recent_uplinks",
		"mac_state.rejected_adr_data_rate_indexes",
		"mac_state.rejected_adr_tx_power_indexes",
		"mac_state.rejected_data_rate_ranges",
		"mac_state.rejected_frequencies",
		"mac_state.rx_windows_available",
		"session.last_conf_f_cnt_down",
		"session.last_f_cnt_up",
		"supports_join",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateIfNotZeroThenFunc(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		if !m["supports_class_b"].GetSupportsClassB() && !m["supports_class_c"].GetSupportsClassC() {
			return false, "supports_class_b"
		}
		return true, ""
	},
		"multicast",

		"supports_class_b",
		"supports_class_c",
	); err != nil {
		return nil, err
	}

	for s, eq := range map[string]func(ttnpb.MACParameters, ttnpb.MACParameters) bool{
		"adr_ack_delay_exponent.value": func(a, b ttnpb.MACParameters) bool {
			return a.ADRAckDelayExponent.Equal(b.ADRAckDelayExponent)
		},
		"adr_ack_limit_exponent.value": func(a, b ttnpb.MACParameters) bool {
			return a.ADRAckLimitExponent.Equal(b.ADRAckLimitExponent)
		},
		"adr_data_rate_index": func(a, b ttnpb.MACParameters) bool {
			return a.ADRDataRateIndex == b.ADRDataRateIndex
		},
		"adr_nb_trans": func(a, b ttnpb.MACParameters) bool {
			return a.ADRNbTrans == b.ADRNbTrans
		},
		"adr_tx_power_index": func(a, b ttnpb.MACParameters) bool {
			return a.ADRTxPowerIndex == b.ADRTxPowerIndex
		},
		"beacon_frequency": func(a, b ttnpb.MACParameters) bool {
			return a.BeaconFrequency == b.BeaconFrequency
		},
		"channels": func(a, b ttnpb.MACParameters) bool {
			if len(a.Channels) != len(b.Channels) {
				return false
			}
			for i, ch := range a.Channels {
				if !ch.Equal(b.Channels[i]) {
					return false
				}
			}
			return true
		},
		"downlink_dwell_time.value": func(a, b ttnpb.MACParameters) bool {
			return a.DownlinkDwellTime.Equal(b.DownlinkDwellTime)
		},
		"max_duty_cycle": func(a, b ttnpb.MACParameters) bool {
			return a.MaxDutyCycle == b.MaxDutyCycle
		},
		"max_eirp": func(a, b ttnpb.MACParameters) bool {
			return a.MaxEIRP == b.MaxEIRP
		},
		"ping_slot_data_rate_index_value.value": func(a, b ttnpb.MACParameters) bool {
			return a.PingSlotDataRateIndexValue.Equal(b.PingSlotDataRateIndexValue)
		},
		"ping_slot_frequency": func(a, b ttnpb.MACParameters) bool {
			return a.PingSlotFrequency == b.PingSlotFrequency
		},
		"rejoin_count_periodicity": func(a, b ttnpb.MACParameters) bool {
			return a.RejoinCountPeriodicity == b.RejoinCountPeriodicity
		},
		"rejoin_time_periodicity": func(a, b ttnpb.MACParameters) bool {
			return a.RejoinTimePeriodicity == b.RejoinTimePeriodicity
		},
		"rx1_data_rate_offset": func(a, b ttnpb.MACParameters) bool {
			return a.Rx1DataRateOffset == b.Rx1DataRateOffset
		},
		"rx1_delay": func(a, b ttnpb.MACParameters) bool {
			return a.Rx1Delay == b.Rx1Delay
		},
		"rx2_data_rate_index": func(a, b ttnpb.MACParameters) bool {
			return a.Rx2DataRateIndex == b.Rx2DataRateIndex
		},
		"rx2_frequency": func(a, b ttnpb.MACParameters) bool {
			return a.Rx2Frequency == b.Rx2Frequency
		},
		"uplink_dwell_time.value": func(a, b ttnpb.MACParameters) bool {
			return a.UplinkDwellTime.Equal(b.UplinkDwellTime)
		},
	} {
		curPath := "mac_state.current_parameters." + s
		desPath := "mac_state.desired_parameters." + s
		eq := eq
		if err := st.ValidateIfNotZeroThenFunc(func(m map[string]*ttnpb.EndDevice) (bool, string) {
			curDev := m[curPath]
			desDev := m[desPath]
			if curDev == nil || desDev == nil {
				if curDev != desDev {
					return false, desPath
				}
				return true, ""
			}
			if !eq(curDev.MACState.CurrentParameters, desDev.MACState.DesiredParameters) {
				return false, desPath
			}
			return true, ""
		},
			"multicast",

			curPath,
			desPath,
		); err != nil {
			return nil, err
		}
	}

	if err := st.ValidateIfNotZeroThenFunc(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		if !m["supports_class_b"].GetSupportsClassB() ||
			m["mac_settings.ping_slot_periodicity.value"].GetMACSettings().GetPingSlotPeriodicity() != nil {
			return true, ""
		}
		return false, "mac_settings.ping_slot_periodicity.value"
	},
		"multicast",

		"supports_class_b",
		"mac_settings.ping_slot_periodicity.value",
	); err != nil {
		return nil, err
	}

	if st.HasSetField(
		"frequency_plan_id",
		"lorawan_phy_version",
		"mac_settings.ping_slot_frequency.value",
		"mac_settings.use_adr.value",
		"mac_state.current_parameters.adr_data_rate_index",
		"mac_state.current_parameters.adr_tx_power_index",
		"mac_state.current_parameters.channels",
		"mac_state.desired_parameters.adr_data_rate_index",
		"mac_state.desired_parameters.adr_tx_power_index",
		"mac_state.desired_parameters.channels",
		"pending_mac_state.current_parameters.adr_data_rate_index",
		"pending_mac_state.current_parameters.adr_tx_power_index",
		"pending_mac_state.current_parameters.channels",
		"pending_mac_state.desired_parameters.adr_data_rate_index",
		"pending_mac_state.desired_parameters.adr_tx_power_index",
		"pending_mac_state.desired_parameters.channels",
		"supports_class_b",
	) {
		var deferredPHYValidations []func(*band.Band) error
		withPHY := func(f func(*band.Band) error) error {
			deferredPHYValidations = append(deferredPHYValidations, f)
			return nil
		}
		if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
			phy, err := DeviceBand(&ttnpb.EndDevice{
				FrequencyPlanID:   m["frequency_plan_id"].FrequencyPlanID,
				LoRaWANPHYVersion: m["lorawan_phy_version"].LoRaWANPHYVersion,
			}, ns.FrequencyPlans)
			if err != nil {
				return err
			}
			withPHY = func(f func(*band.Band) error) error {
				return f(phy)
			}
			for _, f := range deferredPHYValidations {
				if err := f(phy); err != nil {
					return err
				}
			}
			return nil
		},
			"frequency_plan_id",
			"lorawan_phy_version",
		); err != nil {
			return nil, err
		}

		if st.HasSetField(
			"frequency_plan_id",
			"lorawan_phy_version",
			"mac_settings.ping_slot_frequency.value",
			"supports_class_b",
		) {
			if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
				if !m["supports_class_b"].GetSupportsClassB() ||
					m["mac_settings.ping_slot_frequency.value"].GetMACSettings().GetPingSlotFrequency().GetValue() > 0 {
					return nil
				}
				return withPHY(func(phy *band.Band) error {
					if phy.PingSlotFrequency == nil {
						return newInvalidFieldValueError("mac_settings.ping_slot_frequency.value")
					}
					return nil
				})
			},
				"mac_settings.ping_slot_frequency.value",
				"supports_class_b",
			); err != nil {
				return nil, err
			}
		}

		for p, isValid := range map[string]func(*ttnpb.EndDevice, *band.Band) bool{
			"mac_settings.use_adr.value": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return !dev.GetMACSettings().GetUseADR().GetValue() || phy.EnableADR
			},
			"mac_state.current_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetMACState().GetCurrentParameters().ADRDataRateIndex <= phy.MaxADRDataRateIndex
			},
			"mac_state.current_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetMACState().GetCurrentParameters().ADRTxPowerIndex <= uint32(phy.MaxTxPowerIndex())
			},
			"mac_state.current_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return len(dev.GetMACState().GetCurrentParameters().Channels) <= int(phy.MaxUplinkChannels)
			},
			"mac_state.desired_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetMACState().GetDesiredParameters().ADRDataRateIndex <= phy.MaxADRDataRateIndex
			},
			"mac_state.desired_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetMACState().GetDesiredParameters().ADRTxPowerIndex <= uint32(phy.MaxTxPowerIndex())
			},
			"mac_state.desired_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return len(dev.GetMACState().GetDesiredParameters().Channels) <= int(phy.MaxUplinkChannels)
			},
			"pending_mac_state.current_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetPendingMACState().GetCurrentParameters().ADRDataRateIndex <= phy.MaxADRDataRateIndex
			},
			"pending_mac_state.current_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetPendingMACState().GetCurrentParameters().ADRTxPowerIndex <= uint32(phy.MaxTxPowerIndex())
			},
			"pending_mac_state.current_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return len(dev.GetPendingMACState().GetCurrentParameters().Channels) <= int(phy.MaxUplinkChannels)
			},
			"pending_mac_state.desired_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetPendingMACState().GetDesiredParameters().ADRDataRateIndex <= phy.MaxADRDataRateIndex
			},
			"pending_mac_state.desired_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetPendingMACState().GetDesiredParameters().ADRTxPowerIndex <= uint32(phy.MaxTxPowerIndex())
			},
			"pending_mac_state.desired_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return len(dev.GetPendingMACState().GetDesiredParameters().Channels) <= int(phy.MaxUplinkChannels)
			},
		} {
			if !st.HasSetField(
				p,
				"frequency_plan_id",
				"lorawan_phy_version",
			) {
				continue
			}
			p := p
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band) error {
					if !isValid(dev, phy) {
						return newInvalidFieldValueError(p)
					}
					return nil
				})
			}, p); err != nil {
				return nil, err
			}
		}
	}

	var getTransforms []func(*ttnpb.EndDevice)
	if st.Device.Session != nil {
		for p, isZero := range map[string]func() bool{
			"session.dev_addr":                 st.Device.Session.DevAddr.IsZero,
			"session.keys.f_nwk_s_int_key.key": st.Device.Session.FNwkSIntKey.IsZero,
			"session.keys.nwk_s_enc_key.key": func() bool {
				return st.Device.Session.NwkSEncKey != nil && st.Device.Session.NwkSEncKey.IsZero()
			},
			"session.keys.s_nwk_s_int_key.key": func() bool {
				return st.Device.Session.SNwkSIntKey != nil && st.Device.Session.SNwkSIntKey.IsZero()
			},
			"session.keys.session_key_id": func() bool {
				return len(st.Device.Session.SessionKeyID) == 0
			},
		} {
			if err := st.ValidateSetField(func() bool { return !isZero() }, p); err != nil {
				return nil, err
			}
		}
		if st.HasSetField("session.keys.f_nwk_s_int_key.key") {
			k := st.Device.Session.FNwkSIntKey.Key
			fNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, *k, ns.deviceKEKLabel, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			st.Device.Session.FNwkSIntKey = fNwkSIntKey
			st.AddSetFields(
				"session.keys.f_nwk_s_int_key.encrypted_key",
				"session.keys.f_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.Session.FNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if k := st.Device.Session.NwkSEncKey.GetKey(); k != nil && st.HasSetField("session.keys.nwk_s_enc_key.key") {
			nwkSEncKey, err := cryptoutil.WrapAES128Key(ctx, *k, ns.deviceKEKLabel, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			st.Device.Session.NwkSEncKey = nwkSEncKey
			st.AddSetFields(
				"session.keys.nwk_s_enc_key.encrypted_key",
				"session.keys.nwk_s_enc_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.Session.NwkSEncKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if k := st.Device.Session.SNwkSIntKey.GetKey(); k != nil && st.HasSetField("session.keys.s_nwk_s_int_key.key") {
			sNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, *k, ns.deviceKEKLabel, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			st.Device.Session.SNwkSIntKey = sNwkSIntKey
			st.AddSetFields(
				"session.keys.s_nwk_s_int_key.encrypted_key",
				"session.keys.s_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.Session.SNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
	}
	if st.Device.PendingSession != nil {
		for p, isZero := range map[string]func() bool{
			"pending_session.dev_addr":                 st.Device.PendingSession.DevAddr.IsZero,
			"pending_session.keys.f_nwk_s_int_key.key": st.Device.PendingSession.FNwkSIntKey.IsZero,
			"pending_session.keys.nwk_s_enc_key.key":   st.Device.PendingSession.NwkSEncKey.IsZero,
			"pending_session.keys.s_nwk_s_int_key.key": st.Device.PendingSession.SNwkSIntKey.IsZero,
			"pending_session.keys.session_key_id": func() bool {
				return len(st.Device.PendingSession.SessionKeyID) == 0
			},
		} {
			if err := st.ValidateSetField(func() bool { return !isZero() }, p); err != nil {
				return nil, err
			}
		}
		if st.HasSetField("pending_session.keys.f_nwk_s_int_key.key") {
			k := st.Device.PendingSession.FNwkSIntKey.Key
			fNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, *k, ns.deviceKEKLabel, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			st.Device.PendingSession.FNwkSIntKey = fNwkSIntKey
			st.AddSetFields(
				"pending_session.keys.f_nwk_s_int_key.encrypted_key",
				"pending_session.keys.f_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingSession.FNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_session.keys.nwk_s_enc_key.key") {
			k := st.Device.PendingSession.NwkSEncKey.Key
			nwkSEncKey, err := cryptoutil.WrapAES128Key(ctx, *k, ns.deviceKEKLabel, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			st.Device.PendingSession.NwkSEncKey = nwkSEncKey
			st.AddSetFields(
				"pending_session.keys.nwk_s_enc_key.encrypted_key",
				"pending_session.keys.nwk_s_enc_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingSession.NwkSEncKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_session.keys.s_nwk_s_int_key.key") {
			k := st.Device.PendingSession.SNwkSIntKey.Key
			sNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, *k, ns.deviceKEKLabel, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			st.Device.PendingSession.SNwkSIntKey = sNwkSIntKey
			st.AddSetFields(
				"pending_session.keys.s_nwk_s_int_key.encrypted_key",
				"pending_session.keys.s_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingSession.SNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
	}
	if st.Device.PendingMACState.GetQueuedJoinAccept() != nil {
		for p, isZero := range map[string]func() bool{
			"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key": st.Device.PendingMACState.QueuedJoinAccept.Keys.FNwkSIntKey.IsZero,
			"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key":   st.Device.PendingMACState.QueuedJoinAccept.Keys.NwkSEncKey.IsZero,
			"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key": st.Device.PendingMACState.QueuedJoinAccept.Keys.SNwkSIntKey.IsZero,
			"pending_mac_state.queued_join_accept.keys.session_key_id":      func() bool { return len(st.Device.PendingMACState.QueuedJoinAccept.Keys.SessionKeyID) == 0 },
			"pending_mac_state.queued_join_accept.payload":                  func() bool { return len(st.Device.PendingMACState.QueuedJoinAccept.Payload) == 0 },
			"pending_mac_state.queued_join_accept.dev_addr":                 st.Device.PendingMACState.QueuedJoinAccept.DevAddr.IsZero,
		} {
			if err := st.ValidateSetField(func() bool { return !isZero() }, p); err != nil {
				return nil, err
			}
		}
		if st.HasSetField("pending_mac_state_queued_join_accept.keys.f_nwk_s_int_key.key") {
			k := st.Device.PendingMACState.QueuedJoinAccept.Keys.FNwkSIntKey.Key
			fNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, *k, ns.deviceKEKLabel, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			st.Device.PendingMACState.QueuedJoinAccept.Keys.FNwkSIntKey = fNwkSIntKey
			st.AddSetFields(
				"pending_mac_state_queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
				"pending_mac_state_queued_join_accept.keys.f_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingMACState.QueuedJoinAccept.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_mac_state_queued_join_accept.keys.nwk_s_enc_key.key") {
			k := st.Device.PendingMACState.QueuedJoinAccept.Keys.NwkSEncKey.Key
			nwkSEncKey, err := cryptoutil.WrapAES128Key(ctx, *k, ns.deviceKEKLabel, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			st.Device.PendingMACState.QueuedJoinAccept.Keys.NwkSEncKey = nwkSEncKey
			st.AddSetFields(
				"pending_mac_state_queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
				"pending_mac_state_queued_join_accept.keys.nwk_s_enc_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingMACState.QueuedJoinAccept.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_mac_state_queued_join_accept.keys.s_nwk_s_int_key.key") {
			k := st.Device.PendingMACState.QueuedJoinAccept.Keys.SNwkSIntKey.Key
			sNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, *k, ns.deviceKEKLabel, ns.KeyVault)
			if err != nil {
				return nil, err
			}
			st.Device.PendingMACState.QueuedJoinAccept.Keys.SNwkSIntKey = sNwkSIntKey
			st.AddSetFields(
				"pending_mac_state_queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
				"pending_mac_state_queued_join_accept.keys.s_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingMACState.QueuedJoinAccept.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
	}

	// hasSession indicates whether the effective device model contains a non-zero session.
	var hasSession bool
	if err := st.ValidateSetFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		var hasMACState bool
		for k, v := range m {
			switch {
			case strings.HasPrefix(k, "mac_state."):
				if v.MACState != nil {
					hasMACState = true
				}
			case strings.HasPrefix(k, "session."):
				if v.Session != nil {
					hasSession = true
				}
			}
			if hasMACState && hasSession {
				break
			}
		}

		isMulticast := m["multicast"].GetMulticast()
		if !hasMACState && !hasSession && !isMulticast {
			return true, ""
		}

		var macVersion ttnpb.MACVersion
		if hasMACState {
			if isMulticast {
				if dev, ok := m["mac_state.device_class"]; !ok || dev.MACState == nil || dev.MACState.DeviceClass == ttnpb.CLASS_A {
					return false, "mac_state.device_class"
				}
			}
			if dev, ok := m["mac_state.lorawan_version"]; !ok || dev.MACState == nil {
				return false, "mac_state.lorawan_version"
			} else {
				macVersion = dev.MACState.LoRaWANVersion
			}
		} else {
			macVersion = m["lorawan_version"].LoRaWANVersion
		}

		if dev, ok := m["session.dev_addr"]; !ok || dev.Session == nil {
			return false, "session.dev_addr"
		}

		getFNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetSession().GetSessionKeys().GetFNwkSIntKey()
		}
		if setKeyIsZero(m, getFNwkSIntKey, "session.keys.f_nwk_s_int_key") {
			return false, "session.keys.f_nwk_s_int_key.key"
		}

		getNwkSEncKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetSession().GetSessionKeys().GetNwkSEncKey()
		}
		getSNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetSession().GetSessionKeys().GetSNwkSIntKey()
		}
		isZero := struct {
			NwkSEncKey  bool
			SNwkSIntKey bool
		}{
			NwkSEncKey:  setKeyIsZero(m, getNwkSEncKey, "session.keys.nwk_s_enc_key"),
			SNwkSIntKey: setKeyIsZero(m, getSNwkSIntKey, "session.keys.s_nwk_s_int_key"),
		}
		if macVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			if isZero.NwkSEncKey {
				return false, "session.keys.nwk_s_enc_key.key"
			}
			if isZero.SNwkSIntKey {
				return false, "session.keys.s_nwk_s_int_key.key"
			}
		} else {
			if !isZero.NwkSEncKey && !setKeyEqual(m, getFNwkSIntKey, getNwkSEncKey, "session.keys.f_nwk_s_int_key", "session.keys.nwk_s_enc_key") {
				return false, "session.keys.nwk_s_enc_key.key"
			}
			if !isZero.SNwkSIntKey && !setKeyEqual(m, getFNwkSIntKey, getSNwkSIntKey, "session.keys.f_nwk_s_int_key", "session.keys.s_nwk_s_int_key") {
				return false, "session.keys.s_nwk_s_int_key.key"
			}
		}
		if dev, ok := m["session.keys.session_key_id"]; !ok || dev.Session == nil {
			return false, "session.keys.session_key_id"
		}
		return true, ""
	},
		"lorawan_version",
		"mac_state.current_parameters.adr_ack_delay_exponent.value",
		"mac_state.current_parameters.adr_ack_limit_exponent.value",
		"mac_state.current_parameters.adr_data_rate_index",
		"mac_state.current_parameters.adr_nb_trans",
		"mac_state.current_parameters.adr_tx_power_index",
		"mac_state.current_parameters.beacon_frequency",
		"mac_state.current_parameters.channels",
		"mac_state.current_parameters.downlink_dwell_time.value",
		"mac_state.current_parameters.max_duty_cycle",
		"mac_state.current_parameters.max_eirp",
		"mac_state.current_parameters.ping_slot_data_rate_index_value.value",
		"mac_state.current_parameters.ping_slot_frequency",
		"mac_state.current_parameters.rejoin_count_periodicity",
		"mac_state.current_parameters.rejoin_time_periodicity",
		"mac_state.current_parameters.rx1_data_rate_offset",
		"mac_state.current_parameters.rx1_delay",
		"mac_state.current_parameters.rx2_data_rate_index",
		"mac_state.current_parameters.rx2_frequency",
		"mac_state.current_parameters.uplink_dwell_time.value",
		"mac_state.desired_parameters.adr_ack_delay_exponent.value",
		"mac_state.desired_parameters.adr_ack_limit_exponent.value",
		"mac_state.desired_parameters.adr_data_rate_index",
		"mac_state.desired_parameters.adr_nb_trans",
		"mac_state.desired_parameters.adr_tx_power_index",
		"mac_state.desired_parameters.beacon_frequency",
		"mac_state.desired_parameters.channels",
		"mac_state.desired_parameters.downlink_dwell_time.value",
		"mac_state.desired_parameters.max_duty_cycle",
		"mac_state.desired_parameters.max_eirp",
		"mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
		"mac_state.desired_parameters.ping_slot_frequency",
		"mac_state.desired_parameters.rejoin_count_periodicity",
		"mac_state.desired_parameters.rejoin_time_periodicity",
		"mac_state.desired_parameters.rx1_data_rate_offset",
		"mac_state.desired_parameters.rx1_delay",
		"mac_state.desired_parameters.rx2_data_rate_index",
		"mac_state.desired_parameters.rx2_frequency",
		"mac_state.desired_parameters.uplink_dwell_time.value",
		"mac_state.device_class",
		"mac_state.last_adr_change_f_cnt_up",
		"mac_state.last_confirmed_downlink_at",
		"mac_state.last_dev_status_f_cnt_up",
		"mac_state.last_downlink_at",
		"mac_state.last_network_initiated_downlink_at",
		"mac_state.lorawan_version",
		"mac_state.pending_application_downlink.class_b_c.absolute_time",
		"mac_state.pending_application_downlink.class_b_c.gateways",
		"mac_state.pending_application_downlink.confirmed",
		"mac_state.pending_application_downlink.correlation_ids",
		"mac_state.pending_application_downlink.f_cnt",
		"mac_state.pending_application_downlink.f_port",
		"mac_state.pending_application_downlink.frm_payload",
		"mac_state.pending_application_downlink.priority",
		"mac_state.pending_application_downlink.session_key_id",
		"mac_state.pending_requests",
		"mac_state.ping_slot_periodicity.value",
		"mac_state.queued_responses",
		"mac_state.recent_downlinks",
		"mac_state.recent_uplinks",
		"mac_state.rejected_adr_data_rate_indexes",
		"mac_state.rejected_adr_tx_power_indexes",
		"mac_state.rejected_data_rate_ranges",
		"mac_state.rejected_frequencies",
		"mac_state.rx_windows_available",
		"multicast",
		"session.dev_addr",
		"session.keys.f_nwk_s_int_key.encrypted_key",
		"session.keys.f_nwk_s_int_key.kek_label",
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.encrypted_key",
		"session.keys.nwk_s_enc_key.kek_label",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.encrypted_key",
		"session.keys.s_nwk_s_int_key.kek_label",
		"session.keys.s_nwk_s_int_key.key",
		"session.keys.session_key_id",
		"session.last_conf_f_cnt_down",
		"session.last_f_cnt_up",
		"session.last_n_f_cnt_down",
		"session.started_at",
	); err != nil {
		return nil, err
	}

	var (
		// hasPendingSession indicates whether the effective device model contains a non-zero pending session.
		hasPendingSession bool

		// hasQueuedJoinAccept indicates whether the effective device model contains a non-zero queued join-accept.
		hasQueuedJoinAccept bool
	)
	if err := st.ValidateSetFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		var hasPendingMACState bool
		for k, v := range m {
			switch {
			case strings.HasPrefix(k, "pending_mac_state."):
				if v.PendingMACState != nil {
					hasPendingMACState = true
				}
			case strings.HasPrefix(k, "pending_session."):
				if v.PendingSession != nil {
					hasPendingSession = true
				}
			}
			if hasPendingMACState && hasPendingSession {
				break
			}
		}
		if !hasPendingMACState && !hasPendingSession {
			return true, ""
		}

		var macVersion ttnpb.MACVersion
		if dev, ok := m["pending_mac_state.lorawan_version"]; !ok || dev.PendingMACState == nil {
			return false, "pending_mac_state.lorawan_version"
		} else {
			macVersion = dev.PendingMACState.LoRaWANVersion
		}
		if dev, ok := m["pending_session.dev_addr"]; !ok || dev.PendingSession == nil {
			return false, "pending_session.dev_addr"
		}

		getFNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetPendingSession().GetSessionKeys().GetFNwkSIntKey()
		}
		if setKeyIsZero(m, getFNwkSIntKey, "pending_session.keys.f_nwk_s_int_key") {
			return false, "pending_session.keys.f_nwk_s_int_key.key"
		}
		getNwkSEncKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetPendingSession().GetSessionKeys().GetNwkSEncKey()
		}
		if setKeyIsZero(m, getNwkSEncKey, "pending_session.keys.nwk_s_enc_key") {
			return false, "pending_session.keys.nwk_s_enc_key.key"
		}
		getSNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetPendingSession().GetSessionKeys().GetSNwkSIntKey()
		}
		if setKeyIsZero(m, getSNwkSIntKey, "pending_session.keys.s_nwk_s_int_key") {
			return false, "pending_session.keys.s_nwk_s_int_key.key"
		}

		supports1_1 := macVersion.Compare(ttnpb.MAC_V1_1) >= 0
		if !supports1_1 {
			if !setKeyEqual(m, getFNwkSIntKey, getNwkSEncKey, "pending_session.keys.f_nwk_s_int_key", "pending_session.keys.nwk_s_enc_key") {
				return false, "pending_session.keys.nwk_s_enc_key.key"
			}
			if !setKeyEqual(m, getFNwkSIntKey, getSNwkSIntKey, "pending_session.keys.f_nwk_s_int_key", "pending_session.keys.s_nwk_s_int_key") {
				return false, "pending_session.keys.s_nwk_s_int_key.key"
			}
		}
		if dev, ok := m["pending_session.keys.session_key_id"]; !ok || dev.PendingSession == nil {
			return false, "pending_session.keys.session_key_id"
		}

		for k, v := range m {
			if strings.HasPrefix(k, "pending_mac_state.queued_join_accept.") && v.PendingMACState.GetQueuedJoinAccept() != nil {
				hasQueuedJoinAccept = true
				break
			}
		}
		if hasQueuedJoinAccept {
			getFNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				keys := dev.GetPendingMACState().GetQueuedJoinAccept().GetKeys()
				return keys.GetFNwkSIntKey()
			}
			if setKeyIsZero(m, getFNwkSIntKey, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key") {
				return false, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key"
			}
			getNwkSEncKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				keys := dev.GetPendingMACState().GetQueuedJoinAccept().GetKeys()
				return keys.GetNwkSEncKey()
			}
			if setKeyIsZero(m, getNwkSEncKey, "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key") {
				return false, "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key"
			}
			getSNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				keys := dev.GetPendingMACState().GetQueuedJoinAccept().GetKeys()
				return keys.GetSNwkSIntKey()
			}
			if setKeyIsZero(m, getSNwkSIntKey, "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key") {
				return false, "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key"
			}

			if !supports1_1 {
				if !setKeyEqual(m, getFNwkSIntKey, getNwkSEncKey, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key", "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key") {
					return false, "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key"
				}
				if !setKeyEqual(m, getFNwkSIntKey, getSNwkSIntKey, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key", "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key") {
					return false, "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key"
				}
			}

			if dev, ok := m["pending_mac_state.queued_join_accept.keys.session_key_id"]; !ok || dev.PendingMACState.GetQueuedJoinAccept() == nil {
				return false, "pending_mac_state.queued_join_accept.keys.session_key_id"
			}
			if dev, ok := m["pending_mac_state.queued_join_accept.payload"]; !ok || dev.PendingMACState.GetQueuedJoinAccept() == nil {
				return false, "pending_mac_state.queued_join_accept.payload"
			}
			if dev, ok := m["pending_mac_state.queued_join_accept.request.dev_addr"]; !ok || dev.PendingMACState.GetQueuedJoinAccept() == nil {
				return false, "pending_mac_state.queued_join_accept.request.dev_addr"
			}
		}
		return true, ""
	},
		"pending_mac_state.current_parameters.adr_ack_delay_exponent.value",
		"pending_mac_state.current_parameters.adr_ack_limit_exponent.value",
		"pending_mac_state.current_parameters.adr_data_rate_index",
		"pending_mac_state.current_parameters.adr_nb_trans",
		"pending_mac_state.current_parameters.adr_tx_power_index",
		"pending_mac_state.current_parameters.beacon_frequency",
		"pending_mac_state.current_parameters.channels",
		"pending_mac_state.current_parameters.downlink_dwell_time.value",
		"pending_mac_state.current_parameters.max_duty_cycle",
		"pending_mac_state.current_parameters.max_eirp",
		"pending_mac_state.current_parameters.ping_slot_data_rate_index_value.value",
		"pending_mac_state.current_parameters.ping_slot_frequency",
		"pending_mac_state.current_parameters.rejoin_count_periodicity",
		"pending_mac_state.current_parameters.rejoin_time_periodicity",
		"pending_mac_state.current_parameters.rx1_data_rate_offset",
		"pending_mac_state.current_parameters.rx1_delay",
		"pending_mac_state.current_parameters.rx2_data_rate_index",
		"pending_mac_state.current_parameters.rx2_frequency",
		"pending_mac_state.current_parameters.uplink_dwell_time.value",
		"pending_mac_state.desired_parameters.adr_ack_delay_exponent.value",
		"pending_mac_state.desired_parameters.adr_ack_limit_exponent.value",
		"pending_mac_state.desired_parameters.adr_data_rate_index",
		"pending_mac_state.desired_parameters.adr_nb_trans",
		"pending_mac_state.desired_parameters.adr_tx_power_index",
		"pending_mac_state.desired_parameters.beacon_frequency",
		"pending_mac_state.desired_parameters.channels",
		"pending_mac_state.desired_parameters.downlink_dwell_time.value",
		"pending_mac_state.desired_parameters.max_duty_cycle",
		"pending_mac_state.desired_parameters.max_eirp",
		"pending_mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
		"pending_mac_state.desired_parameters.ping_slot_frequency",
		"pending_mac_state.desired_parameters.rejoin_count_periodicity",
		"pending_mac_state.desired_parameters.rejoin_time_periodicity",
		"pending_mac_state.desired_parameters.rx1_data_rate_offset",
		"pending_mac_state.desired_parameters.rx1_delay",
		"pending_mac_state.desired_parameters.rx2_data_rate_index",
		"pending_mac_state.desired_parameters.rx2_frequency",
		"pending_mac_state.desired_parameters.uplink_dwell_time.value",
		"pending_mac_state.device_class",
		"pending_mac_state.last_adr_change_f_cnt_up",
		"pending_mac_state.last_confirmed_downlink_at",
		"pending_mac_state.last_dev_status_f_cnt_up",
		"pending_mac_state.last_downlink_at",
		"pending_mac_state.last_network_initiated_downlink_at",
		"pending_mac_state.lorawan_version",
		"pending_mac_state.pending_join_request.cf_list.ch_masks",
		"pending_mac_state.pending_join_request.cf_list.freq",
		"pending_mac_state.pending_join_request.cf_list.type",
		"pending_mac_state.pending_join_request.downlink_settings.opt_neg",
		"pending_mac_state.pending_join_request.downlink_settings.rx1_dr_offset",
		"pending_mac_state.pending_join_request.downlink_settings.rx2_dr",
		"pending_mac_state.pending_join_request.rx_delay",
		"pending_mac_state.ping_slot_periodicity.value",
		"pending_mac_state.queued_join_accept.correlation_ids",
		"pending_mac_state.queued_join_accept.keys.app_s_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.app_s_key.kek_label",
		"pending_mac_state.queued_join_accept.keys.app_s_key.key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.session_key_id",
		"pending_mac_state.queued_join_accept.payload",
		"pending_mac_state.queued_join_accept.request.cf_list.ch_masks",
		"pending_mac_state.queued_join_accept.request.cf_list.freq",
		"pending_mac_state.queued_join_accept.request.cf_list.type",
		"pending_mac_state.queued_join_accept.request.dev_addr",
		"pending_mac_state.queued_join_accept.request.downlink_settings.opt_neg",
		"pending_mac_state.queued_join_accept.request.downlink_settings.rx1_dr_offset",
		"pending_mac_state.queued_join_accept.request.downlink_settings.rx2_dr",
		"pending_mac_state.queued_join_accept.request.net_id",
		"pending_mac_state.queued_join_accept.request.rx_delay",
		"pending_mac_state.queued_responses",
		"pending_mac_state.recent_downlinks",
		"pending_mac_state.recent_uplinks",
		"pending_mac_state.rejected_adr_data_rate_indexes",
		"pending_mac_state.rejected_adr_tx_power_indexes",
		"pending_mac_state.rejected_data_rate_ranges",
		"pending_mac_state.rejected_frequencies",
		"pending_mac_state.rx_windows_available",
		"pending_session.dev_addr",
		"pending_session.keys.f_nwk_s_int_key.encrypted_key",
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.encrypted_key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.encrypted_key",
		"pending_session.keys.s_nwk_s_int_key.key",
		"pending_session.keys.session_key_id",
	); err != nil {
		return nil, err
	}

	needsDownlinkCheck := st.HasSetField(
		"last_dev_status_received_at",
		"mac_state.current_parameters.adr_ack_delay_exponent.value",
		"mac_state.current_parameters.adr_ack_limit_exponent.value",
		"mac_state.current_parameters.adr_data_rate_index",
		"mac_state.current_parameters.adr_nb_trans",
		"mac_state.current_parameters.adr_tx_power_index",
		"mac_state.current_parameters.beacon_frequency",
		"mac_state.current_parameters.channels",
		"mac_state.current_parameters.downlink_dwell_time.value",
		"mac_state.current_parameters.max_duty_cycle",
		"mac_state.current_parameters.max_eirp",
		"mac_state.current_parameters.ping_slot_data_rate_index_value.value",
		"mac_state.current_parameters.ping_slot_frequency",
		"mac_state.current_parameters.rejoin_count_periodicity",
		"mac_state.current_parameters.rejoin_time_periodicity",
		"mac_state.current_parameters.rx1_data_rate_offset",
		"mac_state.current_parameters.rx1_delay",
		"mac_state.current_parameters.rx2_data_rate_index",
		"mac_state.current_parameters.rx2_frequency",
		"mac_state.current_parameters.uplink_dwell_time.value",
		"mac_state.desired_parameters.adr_ack_delay_exponent.value",
		"mac_state.desired_parameters.adr_ack_limit_exponent.value",
		"mac_state.desired_parameters.adr_data_rate_index",
		"mac_state.desired_parameters.adr_nb_trans",
		"mac_state.desired_parameters.adr_tx_power_index",
		"mac_state.desired_parameters.beacon_frequency",
		"mac_state.desired_parameters.channels",
		"mac_state.desired_parameters.downlink_dwell_time.value",
		"mac_state.desired_parameters.max_duty_cycle",
		"mac_state.desired_parameters.max_eirp",
		"mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
		"mac_state.desired_parameters.ping_slot_frequency",
		"mac_state.desired_parameters.rejoin_count_periodicity",
		"mac_state.desired_parameters.rejoin_time_periodicity",
		"mac_state.desired_parameters.rx1_data_rate_offset",
		"mac_state.desired_parameters.rx1_delay",
		"mac_state.desired_parameters.rx2_data_rate_index",
		"mac_state.desired_parameters.rx2_frequency",
		"mac_state.desired_parameters.uplink_dwell_time.value",
		"mac_state.device_class",
		"mac_state.last_confirmed_downlink_at",
		"mac_state.last_dev_status_f_cnt_up",
		"mac_state.last_downlink_at",
		"mac_state.last_network_initiated_downlink_at",
		"mac_state.lorawan_version",
		"mac_state.ping_slot_periodicity.value",
		"mac_state.queued_responses",
		"mac_state.recent_uplinks",
		"mac_state.rejected_adr_data_rate_indexes",
		"mac_state.rejected_adr_tx_power_indexes",
		"mac_state.rejected_data_rate_ranges",
		"mac_state.rejected_frequencies",
		"mac_state.rx_windows_available",
	)
	if needsDownlinkCheck {
		st.AddGetFields(
			"frequency_plan_id",
			"last_dev_status_received_at",
			"lorawan_phy_version",
			"mac_settings",
			"mac_state.current_parameters.adr_ack_delay_exponent.value",
			"mac_state.current_parameters.adr_ack_limit_exponent.value",
			"mac_state.current_parameters.adr_data_rate_index",
			"mac_state.current_parameters.adr_nb_trans",
			"mac_state.current_parameters.adr_tx_power_index",
			"mac_state.current_parameters.beacon_frequency",
			"mac_state.current_parameters.channels",
			"mac_state.current_parameters.downlink_dwell_time.value",
			"mac_state.current_parameters.max_duty_cycle",
			"mac_state.current_parameters.max_eirp",
			"mac_state.current_parameters.ping_slot_data_rate_index_value.value",
			"mac_state.current_parameters.ping_slot_frequency",
			"mac_state.current_parameters.rejoin_count_periodicity",
			"mac_state.current_parameters.rejoin_time_periodicity",
			"mac_state.current_parameters.rx1_data_rate_offset",
			"mac_state.current_parameters.rx1_delay",
			"mac_state.current_parameters.rx2_data_rate_index",
			"mac_state.current_parameters.rx2_frequency",
			"mac_state.current_parameters.uplink_dwell_time.value",
			"mac_state.desired_parameters.adr_ack_delay_exponent.value",
			"mac_state.desired_parameters.adr_ack_limit_exponent.value",
			"mac_state.desired_parameters.adr_data_rate_index",
			"mac_state.desired_parameters.adr_nb_trans",
			"mac_state.desired_parameters.adr_tx_power_index",
			"mac_state.desired_parameters.beacon_frequency",
			"mac_state.desired_parameters.channels",
			"mac_state.desired_parameters.downlink_dwell_time.value",
			"mac_state.desired_parameters.max_duty_cycle",
			"mac_state.desired_parameters.max_eirp",
			"mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
			"mac_state.desired_parameters.ping_slot_frequency",
			"mac_state.desired_parameters.rejoin_count_periodicity",
			"mac_state.desired_parameters.rejoin_time_periodicity",
			"mac_state.desired_parameters.rx1_data_rate_offset",
			"mac_state.desired_parameters.rx1_delay",
			"mac_state.desired_parameters.rx2_data_rate_index",
			"mac_state.desired_parameters.rx2_frequency",
			"mac_state.desired_parameters.uplink_dwell_time.value",
			"mac_state.device_class",
			"mac_state.last_confirmed_downlink_at",
			"mac_state.last_dev_status_f_cnt_up",
			"mac_state.last_downlink_at",
			"mac_state.last_network_initiated_downlink_at",
			"mac_state.lorawan_version",
			"mac_state.ping_slot_periodicity.value",
			"mac_state.queued_responses",
			"mac_state.recent_uplinks",
			"mac_state.rejected_adr_data_rate_indexes",
			"mac_state.rejected_adr_tx_power_indexes",
			"mac_state.rejected_data_rate_ranges",
			"mac_state.rejected_frequencies",
			"mac_state.rx_windows_available",
			"multicast",
			"session.dev_addr",
			"session.last_conf_f_cnt_down",
			"session.last_f_cnt_up",
			"session.last_n_f_cnt_down",
			"session.queued_application_downlinks",
			"supports_join",
		)
	}

	var evt events.Event
	dev, ctx, err := ns.devices.SetByID(ctx, st.Device.EndDeviceIdentifiers.ApplicationIdentifiers, st.Device.EndDeviceIdentifiers.DeviceID, st.GetFields(), st.SetFunc(func(ctx context.Context, stored *ttnpb.EndDevice) error {
		if stored != nil {
			evt = evtUpdateEndDevice.NewWithIdentifiersAndData(ctx, st.Device.EndDeviceIdentifiers, req.FieldMask.Paths)
			if st.HasSetField("multicast") && st.Device.Multicast != stored.Multicast {
				return errInvalidFieldValue.WithAttributes("field", "multicast")
			}
			if st.HasSetField("supports_join") && st.Device.SupportsJoin != stored.SupportsJoin {
				return errInvalidFieldValue.WithAttributes("field", "supports_join")
			}
			if hasSession {
				if st.HasSetField("session.keys.f_nwk_s_int_key.key") && st.Device.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
					st.Device.Session.NwkSEncKey = st.Device.Session.FNwkSIntKey
					st.Device.Session.SNwkSIntKey = st.Device.Session.FNwkSIntKey
					st.AddSetFields(
						"session.keys.nwk_s_enc_key.encrypted_key",
						"session.keys.nwk_s_enc_key.kek_label",
						"session.keys.nwk_s_enc_key.key",
						"session.keys.s_nwk_s_int_key.encrypted_key",
						"session.keys.s_nwk_s_int_key.kek_label",
						"session.keys.s_nwk_s_int_key.key",
					)
				}

				if st.HasSetField("session.started_at") && st.Device.GetSession().GetStartedAt().IsZero() ||
					st.HasSetField("session.session_key_id") && !bytes.Equal(st.Device.GetSession().GetSessionKeyID(), stored.GetSession().GetSessionKeyID()) ||
					stored.GetSession().GetStartedAt().IsZero() {
					st.Device.Session.StartedAt = time.Now().UTC()
					st.AddSetFields(
						"session.started_at",
					)
				}
			}
			return nil
		}

		evt = evtCreateEndDevice.NewWithIdentifiersAndData(ctx, st.Device.EndDeviceIdentifiers, nil)
		if err := st.RequireFields(
			"frequency_plan_id",
			"lorawan_phy_version",
			"lorawan_version",
		); err != nil {
			return err
		}

		if hasSession {
			if !st.HasSetField("mac_state") {
				macState, err := mac.NewState(st.Device, ns.FrequencyPlans, ns.defaultMACSettings)
				if err != nil {
					return err
				}
				if macSets := ttnpb.FieldsWithoutPrefix("mac_state", st.SetFields()...); len(macSets) != 0 {
					if err := macState.SetFields(st.Device.MACState, macSets...); err != nil {
						return err
					}
				}
				st.Device.MACState = macState
				st.AddSetFields(
					"mac_state",
				)
			}

			if st.Device.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
				st.Device.Session.NwkSEncKey = st.Device.Session.FNwkSIntKey
				st.Device.Session.SNwkSIntKey = st.Device.Session.FNwkSIntKey
				st.AddSetFields(
					"session.keys.nwk_s_enc_key.encrypted_key",
					"session.keys.nwk_s_enc_key.kek_label",
					"session.keys.nwk_s_enc_key.key",
					"session.keys.s_nwk_s_int_key.encrypted_key",
					"session.keys.s_nwk_s_int_key.kek_label",
					"session.keys.s_nwk_s_int_key.key",
				)
			}

			if !st.HasSetField("session.started_at") || st.Device.GetSession().GetStartedAt().IsZero() {
				st.Device.Session.StartedAt = time.Now().UTC()
				st.AddSetFields(
					"session.started_at",
				)
			}
		}
		if hasPendingSession {
			supports1_1 := st.Device.PendingMACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0
			if !supports1_1 {
				st.Device.PendingSession.NwkSEncKey = st.Device.PendingSession.FNwkSIntKey
				st.Device.PendingSession.SNwkSIntKey = st.Device.PendingSession.FNwkSIntKey
				st.AddSetFields(
					"pending_session.keys.nwk_s_enc_key.encrypted_key",
					"pending_session.keys.nwk_s_enc_key.kek_label",
					"pending_session.keys.nwk_s_enc_key.key",
					"pending_session.keys.s_nwk_s_int_key.encrypted_key",
					"pending_session.keys.s_nwk_s_int_key.kek_label",
					"pending_session.keys.s_nwk_s_int_key.key",
				)
			}
			if hasQueuedJoinAccept && !supports1_1 {
				st.Device.PendingMACState.QueuedJoinAccept.Keys.NwkSEncKey = st.Device.PendingMACState.QueuedJoinAccept.Keys.FNwkSIntKey
				st.Device.PendingMACState.QueuedJoinAccept.Keys.SNwkSIntKey = st.Device.PendingMACState.QueuedJoinAccept.Keys.FNwkSIntKey
				st.AddSetFields(
					"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
					"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.kek_label",
					"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
					"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
					"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.kek_label",
					"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
				)
			}
		}
		return nil
	}))
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to set device in registry")
		return nil, err
	}
	for _, f := range getTransforms {
		f(dev)
	}

	if evt != nil {
		events.Publish(evt)
	}

	if !needsDownlinkCheck {
		return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
	}

	if err := ns.updateDataDownlinkTask(ctx, dev, time.Time{}); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to update downlink task queue after device set")
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
}

// ResetFactoryDefaults implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) ResetFactoryDefaults(ctx context.Context, req *ttnpb.ResetAndGetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, appendRequiredDeviceReadRights(
		append(make([]ttnpb.Right, 0, 1+maxRequiredDeviceReadRightCount), ttnpb.RIGHT_APPLICATION_DEVICES_WRITE),
		req.FieldMask.Paths...,
	)...); err != nil {
		return nil, err
	}

	dev, _, err := ns.devices.SetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, addDeviceGetPaths(ttnpb.AddFields(append(req.FieldMask.Paths[:0:0], req.FieldMask.Paths...),
		"frequency_plan_id",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"session.dev_addr",
		"session.queued_application_downlinks",
		"session.keys",
		"supports_join",
	)...), func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if stored == nil {
			return nil, nil, errDeviceNotFound.New()
		}

		stored.BatteryPercentage = nil
		stored.DownlinkMargin = 0
		stored.LastDevStatusReceivedAt = nil
		stored.MACState = nil
		stored.PendingMACState = nil
		stored.PendingSession = nil
		stored.PowerState = ttnpb.PowerState_POWER_UNKNOWN
		if stored.SupportsJoin {
			stored.Session = nil
		} else {
			if stored.Session == nil {
				return nil, nil, errCorruptedMACState.New()
			}

			macState, err := mac.NewState(stored, ns.FrequencyPlans, ns.defaultMACSettings)
			if err != nil {
				return nil, nil, err
			}
			stored.MACState = macState
			stored.Session = &ttnpb.Session{
				DevAddr:                    stored.Session.DevAddr,
				SessionKeys:                stored.Session.SessionKeys,
				StartedAt:                  time.Now().UTC(),
				QueuedApplicationDownlinks: stored.Session.QueuedApplicationDownlinks,
			}
		}
		return stored, []string{
			"battery_percentage",
			"downlink_margin",
			"last_dev_status_received_at",
			"mac_state",
			"pending_mac_state",
			"pending_session",
			"session",
		}, nil
	})
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to reset device state in registry")
		return nil, err
	}
	if err := unwrapSelectedSessionKeys(ctx, ns.KeyVault, dev, req.FieldMask.Paths...); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to unwrap selected keys")
		return nil, err
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.Paths...)
}

// Delete implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Delete(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	var evt events.Event
	_, _, err := ns.devices.SetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, nil, func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev == nil {
			return nil, nil, errDeviceNotFound.New()
		}
		evt = evtDeleteEndDevice.NewWithIdentifiersAndData(ctx, req, nil)
		return nil, nil, nil
	})
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to delete device from registry")
		return nil, err
	}
	if evt != nil {
		events.Publish(evt)
	}
	return ttnpb.Empty, nil
}
