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

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	evtCreateEndDevice = events.Define(
		"ns.end_device.create", "create end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
	evtUpdateEndDevice = events.Define(
		"ns.end_device.update", "update end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
	evtDeleteEndDevice = events.Define(
		"ns.end_device.delete", "delete end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
	evtBatchDeleteEndDevices = events.Define(
		"ns.end_device.batch.delete", "batch delete end devices",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithDataType(&ttnpb.EndDeviceIdentifiersList{}),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
)

const maxRequiredDeviceReadRightCount = 3

func appendRequiredDeviceReadRights(rights []ttnpb.Right, gets ...string) []ttnpb.Right {
	if len(gets) == 0 {
		return rights
	}
	rights = append(rights,
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
	)
	if ttnpb.HasAnyField(gets,
		"pending_session.queued_application_downlinks",
		"queued_application_downlinks",
		"session.queued_application_downlinks",
	) {
		rights = append(rights, ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ)
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
		rights = append(rights, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS)
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

func unwrapSelectedSessionKeys(ctx context.Context, kv crypto.KeyService, dev *ttnpb.EndDevice, paths ...string) error {
	if dev.PendingSession != nil && ttnpb.HasAnyField(paths,
		"pending_session.keys.f_nwk_s_int_key.key",
		"pending_session.keys.nwk_s_enc_key.key",
		"pending_session.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.PendingSession.Keys, "pending_session.keys", paths...)
		if err != nil {
			return err
		}
		dev.PendingSession.Keys = sk
	}
	if dev.Session != nil && ttnpb.HasAnyField(paths,
		"session.keys.f_nwk_s_int_key.key",
		"session.keys.nwk_s_enc_key.key",
		"session.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.Session.Keys, "session.keys", paths...)
		if err != nil {
			return err
		}
		dev.Session.Keys = sk
	}
	if dev.PendingMacState.GetQueuedJoinAccept() != nil && ttnpb.HasAnyField(paths,
		"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key",
		"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key",
		"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key",
	) {
		sk, err := cryptoutil.UnwrapSelectedSessionKeys(ctx, kv, dev.PendingMacState.QueuedJoinAccept.Keys, "pending_mac_state.queued_join_accept.keys", paths...)
		if err != nil {
			return err
		}
		dev.PendingMacState.QueuedJoinAccept.Keys = sk
	}
	return nil
}

// Get implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.EndDeviceIds.ApplicationIds, appendRequiredDeviceReadRights(
		make([]ttnpb.Right, 0, maxRequiredDeviceReadRightCount),
		req.FieldMask.GetPaths()...,
	)...); err != nil {
		return nil, err
	}

	dev, ctx, err := ns.devices.GetByID(ctx, req.EndDeviceIds.ApplicationIds, req.EndDeviceIds.DeviceId, addDeviceGetPaths(req.FieldMask.GetPaths()...))
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to get device from registry")
		return nil, err
	}
	if err := unwrapSelectedSessionKeys(ctx, ns.KeyService(), dev, req.FieldMask.GetPaths()...); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to unwrap selected keys")
		return nil, err
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.GetPaths()...)
}

func newInvalidFieldValueError(field string) *errors.Error {
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

// hasAnyField caches the result of ttnpb.HasAnyField in the provided cache map
// in order to avoid redundant lookups.
//
// NOTE: If the search paths are not bottom level fields, hasAnyField may have unexpected
// results, as ttnpb.HasAnyField does not consider higher search paths as being part of
// the requested paths - i.e ttnpb.HasAnyField([]string{"a.b"}, "a") == false.
func hasAnyField(fs []string, cache map[string]bool, paths ...string) bool {
	for _, p := range paths {
		for i := len(p); i > 0; i = strings.LastIndex(p[:i], ".") {
			p := p[:i]
			v, ok := cache[p]
			if !ok {
				continue
			}
			if !v {
				continue
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

// WithFields calls f when path is available.
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

// WithFields calls f when all paths in paths are available.
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

// ValidateField ensures that isValid(dev), where dev is the device containing path evaluates to true.
func (st *setDeviceState) ValidateField(isValid func(*ttnpb.EndDevice) bool, path string) error {
	return st.WithField(func(dev *ttnpb.EndDevice) error {
		if !isValid(dev) {
			return newInvalidFieldValueError(path)
		}
		return nil
	}, path)
}

var errFieldNotZero = errors.DefineInvalidArgument("field_not_zero", "field `{name}` is not zero")

// ValidateFieldIsZero ensures that path is zero.
func (st *setDeviceState) ValidateFieldIsZero(path string) error {
	if st.HasSetField(path) {
		if !st.Device.FieldIsZero(path) {
			return newInvalidFieldValueError(path).WithCause(errFieldNotZero.WithAttributes("name", path))
		}
		return nil
	}
	v, ok := st.zeroPaths[path]
	if !ok {
		st.zeroPaths[path] = true
		st.AddGetFields(path)
		return nil
	}
	if !v {
		panic(fmt.Sprintf("path `%s` requested to be both zero and not zero", path))
	}
	return nil
}

var errFieldIsZero = errors.DefineInvalidArgument("field_is_zero", "field `{name}` is zero")

// ValidateFieldIsNotZero ensures that path is not zero.
func (st *setDeviceState) ValidateFieldIsNotZero(path string) error {
	if st.HasSetField(path) {
		if st.Device.FieldIsZero(path) {
			return newInvalidFieldValueError(path).WithCause(errFieldIsZero.WithAttributes("name", path))
		}
		return nil
	}
	v, ok := st.zeroPaths[path]
	if !ok {
		st.zeroPaths[path] = false
		st.AddGetFields(path)
		return nil
	}
	if v {
		panic(fmt.Sprintf("path `%s` requested to be both zero and not zero", path))
	}
	return nil
}

// ValidateFieldsAreZero ensures that each p in paths is zero.
func (st *setDeviceState) ValidateFieldsAreZero(paths ...string) error {
	for _, p := range paths {
		if err := st.ValidateFieldIsZero(p); err != nil {
			return err
		}
	}
	return nil
}

// ValidateFieldsAreNotZero ensures none of p in paths is zero.
func (st *setDeviceState) ValidateFieldsAreNotZero(paths ...string) error {
	for _, p := range paths {
		if err := st.ValidateFieldIsNotZero(p); err != nil {
			return err
		}
	}
	return nil
}

// ValidateFields calls isValid with a map path -> *ttnpb.EndDevice, where the value stored under the key
// is either a pointer to stored device or to device being set in request, depending on the request fieldmask.
// isValid is only executed once all fields are present. That means that if request sets all fields in paths
// isValid is executed immediately, otherwise it is called later (after device fetch) by SetFunc.
func (st *setDeviceState) ValidateFields(isValid func(map[string]*ttnpb.EndDevice) (bool, string), paths ...string) error {
	return st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
		ok, p := isValid(m)
		if !ok {
			return newInvalidFieldValueError(p)
		}
		return nil
	}, paths...)
}

// ValidateSetField validates the field iff path is being set in request.
func (st *setDeviceState) ValidateSetField(isValid func() bool, path string) error {
	if !st.HasSetField(path) {
		return nil
	}
	if !isValid() {
		return newInvalidFieldValueError(path)
	}
	return nil
}

// ValidateSetField is like ValidateSetField, but allows the validator callback to return an error
// and propagates it to the caller as the cause.
func (st *setDeviceState) ValidateSetFieldWithCause(isValid func() error, path string) error {
	if !st.HasSetField(path) {
		return nil
	}
	if err := isValid(); err != nil {
		return newInvalidFieldValueError(path).WithCause(err)
	}
	return nil
}

// ValidateSetFields validates the fields iff at least one of p in paths is being set in request.
func (st *setDeviceState) ValidateSetFields(isValid func(map[string]*ttnpb.EndDevice) (bool, string), paths ...string) error {
	if !st.HasSetField(paths...) {
		return nil
	}
	return st.ValidateFields(isValid, paths...)
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
		if ke := get(dev); !types.MustAES128Key(ke.GetKey()).OrZero().IsZero() {
			return false
		}
	}
	if dev, ok := m[path+".encrypted_key"]; ok {
		if ke := get(dev); len(ke.GetEncryptedKey()) != 0 {
			return false
		}
	}
	return true
}

func setKeyEqual(m map[string]*ttnpb.EndDevice, getA, getB func(*ttnpb.EndDevice) *ttnpb.KeyEnvelope, pathA, pathB string) bool {
	if a, b := getA(m[pathA+".key"]).GetKey(), getB(m[pathB+".key"]).GetKey(); a == nil && b != nil ||
		a != nil && b == nil ||
		a != nil && b != nil && !types.MustAES128Key(a).Equal(*types.MustAES128Key(b)) {
		return false
	}
	if a, b := getA(m[pathA+".encrypted_key"]).GetEncryptedKey(), getB(m[pathB+".encrypted_key"]).GetEncryptedKey(); !bytes.Equal(a, b) {
		return false
	}
	if a, b := getA(m[pathA+".kek_label"]).GetKekLabel(), getB(m[pathB+".kek_label"]).GetKekLabel(); a != b {
		return false
	}
	return true
}

// ifThenFuncFieldRight represents the RHS of a functional implication.
type ifThenFuncFieldRight struct {
	Func   func(m map[string]*ttnpb.EndDevice) (bool, string)
	Fields []string
}

var (
	ifZeroThenZeroFields = map[string][]string{
		"supports_join": {
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
			"pending_mac_state.current_parameters.relay.mode.served.backoff",
			"pending_mac_state.current_parameters.relay.mode.served.mode.always",
			"pending_mac_state.current_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
			"pending_mac_state.current_parameters.relay.mode.served.mode.end_device_controlled",
			"pending_mac_state.current_parameters.relay.mode.served.second_channel.ack_offset",
			"pending_mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
			"pending_mac_state.current_parameters.relay.mode.served.second_channel.frequency",
			"pending_mac_state.current_parameters.relay.mode.served.serving_device_id",
			"pending_mac_state.current_parameters.relay.mode.serving.cad_periodicity",
			"pending_mac_state.current_parameters.relay.mode.serving.default_channel_index",
			"pending_mac_state.current_parameters.relay.mode.serving.limits.join_requests.bucket_size",
			"pending_mac_state.current_parameters.relay.mode.serving.limits.join_requests.reload_rate",
			"pending_mac_state.current_parameters.relay.mode.serving.limits.notifications.bucket_size",
			"pending_mac_state.current_parameters.relay.mode.serving.limits.notifications.reload_rate",
			"pending_mac_state.current_parameters.relay.mode.serving.limits.overall.bucket_size",
			"pending_mac_state.current_parameters.relay.mode.serving.limits.overall.reload_rate",
			"pending_mac_state.current_parameters.relay.mode.serving.limits.reset_behavior",
			"pending_mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
			"pending_mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
			"pending_mac_state.current_parameters.relay.mode.serving.second_channel.ack_offset",
			"pending_mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
			"pending_mac_state.current_parameters.relay.mode.serving.second_channel.frequency",
			"pending_mac_state.current_parameters.relay.mode.serving.uplink_forwarding_rules",
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
			"pending_mac_state.desired_parameters.relay.mode.served.backoff",
			"pending_mac_state.desired_parameters.relay.mode.served.mode.always",
			"pending_mac_state.desired_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
			"pending_mac_state.desired_parameters.relay.mode.served.mode.end_device_controlled",
			"pending_mac_state.desired_parameters.relay.mode.served.second_channel.ack_offset",
			"pending_mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
			"pending_mac_state.desired_parameters.relay.mode.served.second_channel.frequency",
			"pending_mac_state.desired_parameters.relay.mode.served.serving_device_id",
			"pending_mac_state.desired_parameters.relay.mode.serving.cad_periodicity",
			"pending_mac_state.desired_parameters.relay.mode.serving.default_channel_index",
			"pending_mac_state.desired_parameters.relay.mode.serving.limits.join_requests.bucket_size",
			"pending_mac_state.desired_parameters.relay.mode.serving.limits.join_requests.reload_rate",
			"pending_mac_state.desired_parameters.relay.mode.serving.limits.notifications.bucket_size",
			"pending_mac_state.desired_parameters.relay.mode.serving.limits.notifications.reload_rate",
			"pending_mac_state.desired_parameters.relay.mode.serving.limits.overall.bucket_size",
			"pending_mac_state.desired_parameters.relay.mode.serving.limits.overall.reload_rate",
			"pending_mac_state.desired_parameters.relay.mode.serving.limits.reset_behavior",
			"pending_mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
			"pending_mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
			"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.ack_offset",
			"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
			"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.frequency",
			"pending_mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
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
			"pending_mac_state.recent_mac_command_identifiers",
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
		},
	}

	ifZeroThenNotZeroFields = map[string][]string{
		"supports_join": {
			"session.dev_addr",
			"session.keys.f_nwk_s_int_key.key",
			// NOTE: LoRaWAN-version specific fields are validated within Set directly.
		},
	}

	ifNotZeroThenZeroFields = map[string][]string{
		"multicast": {
			"mac_settings.desired_relay.mode.served.backoff",
			"mac_settings.desired_relay.mode.served.mode.always",
			"mac_settings.desired_relay.mode.served.mode.dynamic.smart_enable_level",
			"mac_settings.desired_relay.mode.served.mode.end_device_controlled",
			"mac_settings.desired_relay.mode.served.second_channel.ack_offset",
			"mac_settings.desired_relay.mode.served.second_channel.data_rate_index",
			"mac_settings.desired_relay.mode.served.second_channel.frequency",
			"mac_settings.desired_relay.mode.served.serving_device_id",
			"mac_settings.desired_relay.mode.serving.cad_periodicity",
			"mac_settings.desired_relay.mode.serving.default_channel_index",
			"mac_settings.desired_relay.mode.serving.limits.join_requests.bucket_size",
			"mac_settings.desired_relay.mode.serving.limits.join_requests.reload_rate",
			"mac_settings.desired_relay.mode.serving.limits.notifications.bucket_size",
			"mac_settings.desired_relay.mode.serving.limits.notifications.reload_rate",
			"mac_settings.desired_relay.mode.serving.limits.overall.bucket_size",
			"mac_settings.desired_relay.mode.serving.limits.overall.reload_rate",
			"mac_settings.desired_relay.mode.serving.limits.reset_behavior",
			"mac_settings.desired_relay.mode.serving.limits.uplink_messages.bucket_size",
			"mac_settings.desired_relay.mode.serving.limits.uplink_messages.reload_rate",
			"mac_settings.desired_relay.mode.serving.second_channel.ack_offset",
			"mac_settings.desired_relay.mode.serving.second_channel.data_rate_index",
			"mac_settings.desired_relay.mode.serving.second_channel.frequency",
			"mac_settings.desired_relay.mode.serving.uplink_forwarding_rules",
			"mac_settings.relay.mode.served.backoff",
			"mac_settings.relay.mode.served.mode.always",
			"mac_settings.relay.mode.served.mode.dynamic.smart_enable_level",
			"mac_settings.relay.mode.served.mode.end_device_controlled",
			"mac_settings.relay.mode.served.second_channel.ack_offset",
			"mac_settings.relay.mode.served.second_channel.data_rate_index",
			"mac_settings.relay.mode.served.second_channel.frequency",
			"mac_settings.relay.mode.served.serving_device_id",
			"mac_settings.relay.mode.serving.cad_periodicity",
			"mac_settings.relay.mode.serving.default_channel_index",
			"mac_settings.relay.mode.serving.limits.join_requests.bucket_size",
			"mac_settings.relay.mode.serving.limits.join_requests.reload_rate",
			"mac_settings.relay.mode.serving.limits.notifications.bucket_size",
			"mac_settings.relay.mode.serving.limits.notifications.reload_rate",
			"mac_settings.relay.mode.serving.limits.overall.bucket_size",
			"mac_settings.relay.mode.serving.limits.overall.reload_rate",
			"mac_settings.relay.mode.serving.limits.reset_behavior",
			"mac_settings.relay.mode.serving.limits.uplink_messages.bucket_size",
			"mac_settings.relay.mode.serving.limits.uplink_messages.reload_rate",
			"mac_settings.relay.mode.serving.second_channel.ack_offset",
			"mac_settings.relay.mode.serving.second_channel.data_rate_index",
			"mac_settings.relay.mode.serving.second_channel.frequency",
			"mac_settings.relay.mode.serving.uplink_forwarding_rules",
			"mac_settings.schedule_downlinks.value",
			"mac_state.current_parameters.relay.mode.served.backoff",
			"mac_state.current_parameters.relay.mode.served.mode.always",
			"mac_state.current_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
			"mac_state.current_parameters.relay.mode.served.mode.end_device_controlled",
			"mac_state.current_parameters.relay.mode.served.second_channel.ack_offset",
			"mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
			"mac_state.current_parameters.relay.mode.served.second_channel.frequency",
			"mac_state.current_parameters.relay.mode.served.serving_device_id",
			"mac_state.current_parameters.relay.mode.serving.cad_periodicity",
			"mac_state.current_parameters.relay.mode.serving.default_channel_index",
			"mac_state.current_parameters.relay.mode.serving.limits.join_requests.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.join_requests.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.limits.notifications.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.notifications.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.limits.overall.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.overall.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.limits.reset_behavior",
			"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.second_channel.ack_offset",
			"mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
			"mac_state.current_parameters.relay.mode.serving.second_channel.frequency",
			"mac_state.current_parameters.relay.mode.serving.uplink_forwarding_rules",
			"mac_state.desired_parameters.relay.mode.served.backoff",
			"mac_state.desired_parameters.relay.mode.served.mode.always",
			"mac_state.desired_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
			"mac_state.desired_parameters.relay.mode.served.mode.end_device_controlled",
			"mac_state.desired_parameters.relay.mode.served.second_channel.ack_offset",
			"mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
			"mac_state.desired_parameters.relay.mode.served.second_channel.frequency",
			"mac_state.desired_parameters.relay.mode.served.serving_device_id",
			"mac_state.desired_parameters.relay.mode.serving.cad_periodicity",
			"mac_state.desired_parameters.relay.mode.serving.default_channel_index",
			"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.limits.notifications.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.notifications.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.limits.overall.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.overall.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.limits.reset_behavior",
			"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.second_channel.ack_offset",
			"mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
			"mac_state.desired_parameters.relay.mode.serving.second_channel.frequency",
			"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
			"mac_state.last_adr_change_f_cnt_up",
			"mac_state.last_confirmed_downlink_at",
			"mac_state.last_dev_status_f_cnt_up",
			"mac_state.pending_application_downlink",
			"mac_state.pending_requests",
			"mac_state.queued_responses",
			"mac_state.recent_mac_command_identifiers",
			"mac_state.recent_uplinks",
			"mac_state.rejected_adr_data_rate_indexes",
			"mac_state.rejected_adr_tx_power_indexes",
			"mac_state.rejected_data_rate_ranges",
			"mac_state.rejected_frequencies",
			"mac_state.rx_windows_available",
			"session.last_conf_f_cnt_down",
			"session.last_f_cnt_up",
			"supports_join",
		},
	}

	ifNotZeroThenNotZeroFields = map[string][]string{
		"supports_join": {
			"ids.dev_eui",
			"ids.join_eui",
		},
	}

	ifZeroThenFuncFields = map[string][]ifThenFuncFieldRight{
		"supports_join": {
			{
				Func: func(m map[string]*ttnpb.EndDevice) (bool, string) {
					if dev, ok := m["ids.dev_eui"]; ok && !types.MustEUI64(dev.Ids.DevEui).OrZero().IsZero() {
						return true, ""
					}
					if m["lorawan_version"].GetLorawanVersion() == ttnpb.MACVersion_MAC_UNKNOWN {
						return false, "lorawan_version"
					}
					if macspec.RequireDevEUIForABP(m["lorawan_version"].LorawanVersion) && !m["multicast"].GetMulticast() {
						return false, "ids.dev_eui"
					}
					return true, ""
				},
				Fields: []string{
					"ids.dev_eui",
					"lorawan_version",
					"multicast",
				},
			},

			{
				Func: func(m map[string]*ttnpb.EndDevice) (bool, string) {
					if !m["supports_class_b"].GetSupportsClassB() ||
						m["mac_settings.ping_slot_periodicity.value"].GetMacSettings().GetPingSlotPeriodicity() != nil {
						return true, ""
					}
					return false, "mac_settings.ping_slot_periodicity.value"
				},
				Fields: []string{
					"mac_settings.ping_slot_periodicity.value",
					"supports_class_b",
				},
			},
		},
	}

	ifNotZeroThenFuncFields = map[string][]ifThenFuncFieldRight{
		"multicast": append(func() (rs []ifThenFuncFieldRight) {
			for s, eq := range map[string]func(*ttnpb.MACParameters, *ttnpb.MACParameters) bool{
				"adr_ack_delay_exponent.value": func(a, b *ttnpb.MACParameters) bool {
					return proto.Equal(a.AdrAckDelayExponent, b.AdrAckDelayExponent)
				},
				"adr_ack_limit_exponent.value": func(a, b *ttnpb.MACParameters) bool {
					return proto.Equal(a.AdrAckLimitExponent, b.AdrAckLimitExponent)
				},
				"adr_data_rate_index": func(a, b *ttnpb.MACParameters) bool {
					return a.AdrDataRateIndex == b.AdrDataRateIndex
				},
				"adr_nb_trans": func(a, b *ttnpb.MACParameters) bool {
					return a.AdrNbTrans == b.AdrNbTrans
				},
				"adr_tx_power_index": func(a, b *ttnpb.MACParameters) bool {
					return a.AdrTxPowerIndex == b.AdrTxPowerIndex
				},
				"beacon_frequency": func(a, b *ttnpb.MACParameters) bool {
					return a.BeaconFrequency == b.BeaconFrequency
				},
				"channels": func(a, b *ttnpb.MACParameters) bool {
					if len(a.Channels) != len(b.Channels) {
						return false
					}
					for i, ch := range a.Channels {
						if !proto.Equal(ch, b.Channels[i]) {
							return false
						}
					}
					return true
				},
				"downlink_dwell_time.value": func(a, b *ttnpb.MACParameters) bool {
					return proto.Equal(a.DownlinkDwellTime, b.DownlinkDwellTime)
				},
				"max_duty_cycle": func(a, b *ttnpb.MACParameters) bool {
					return a.MaxDutyCycle == b.MaxDutyCycle
				},
				"max_eirp": func(a, b *ttnpb.MACParameters) bool {
					return a.MaxEirp == b.MaxEirp
				},
				"ping_slot_data_rate_index_value.value": func(a, b *ttnpb.MACParameters) bool {
					return proto.Equal(a.PingSlotDataRateIndexValue, b.PingSlotDataRateIndexValue)
				},
				"ping_slot_frequency": func(a, b *ttnpb.MACParameters) bool {
					return a.PingSlotFrequency == b.PingSlotFrequency
				},
				"rejoin_count_periodicity": func(a, b *ttnpb.MACParameters) bool {
					return a.RejoinCountPeriodicity == b.RejoinCountPeriodicity
				},
				"rejoin_time_periodicity": func(a, b *ttnpb.MACParameters) bool {
					return a.RejoinTimePeriodicity == b.RejoinTimePeriodicity
				},
				"rx1_data_rate_offset": func(a, b *ttnpb.MACParameters) bool {
					return a.Rx1DataRateOffset == b.Rx1DataRateOffset
				},
				"rx1_delay": func(a, b *ttnpb.MACParameters) bool {
					return a.Rx1Delay == b.Rx1Delay
				},
				"rx2_data_rate_index": func(a, b *ttnpb.MACParameters) bool {
					return a.Rx2DataRateIndex == b.Rx2DataRateIndex
				},
				"rx2_frequency": func(a, b *ttnpb.MACParameters) bool {
					return a.Rx2Frequency == b.Rx2Frequency
				},
				"uplink_dwell_time.value": func(a, b *ttnpb.MACParameters) bool {
					return proto.Equal(a.UplinkDwellTime, b.UplinkDwellTime)
				},
			} {
				curPath := "mac_state.current_parameters." + s
				desPath := "mac_state.desired_parameters." + s
				eq := eq
				rs = append(rs, ifThenFuncFieldRight{
					Func: func(m map[string]*ttnpb.EndDevice) (bool, string) {
						curDev := m[curPath]
						desDev := m[desPath]
						if curDev == nil || desDev == nil {
							if curDev != desDev {
								return false, desPath
							}
							return true, ""
						}
						if !eq(curDev.MacState.CurrentParameters, desDev.MacState.DesiredParameters) {
							return false, desPath
						}
						return true, ""
					},
					Fields: []string{
						curPath,
						desPath,
					},
				})
			}
			return rs
		}(),

			ifThenFuncFieldRight{
				Func: func(m map[string]*ttnpb.EndDevice) (bool, string) {
					if !m["supports_class_b"].GetSupportsClassB() && !m["supports_class_c"].GetSupportsClassC() {
						return false, "supports_class_b"
					}
					return true, ""
				},
				Fields: []string{
					"supports_class_b",
					"supports_class_c",
				},
			},

			ifThenFuncFieldRight{
				Func: func(m map[string]*ttnpb.EndDevice) (bool, string) {
					if !m["supports_class_b"].GetSupportsClassB() ||
						m["mac_settings.ping_slot_periodicity.value"].GetMacSettings().GetPingSlotPeriodicity() != nil {
						return true, ""
					}
					return false, "mac_settings.ping_slot_periodicity.value"
				},
				Fields: []string{
					"mac_settings.ping_slot_periodicity.value",
					"supports_class_b",
				},
			},
		),
	}

	// downlinkInfluencingSetFields contains fields that can influence downlink scheduling, e.g. trigger one or make a scheduled slot obsolete.
	downlinkInfluencingSetFields = [...]string{
		"last_dev_status_received_at",
		"mac_settings.schedule_downlinks.value",
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
		"mac_state.current_parameters.relay.mode.served.backoff",
		"mac_state.current_parameters.relay.mode.served.mode.always",
		"mac_state.current_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
		"mac_state.current_parameters.relay.mode.served.mode.end_device_controlled",
		"mac_state.current_parameters.relay.mode.served.second_channel.ack_offset",
		"mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
		"mac_state.current_parameters.relay.mode.served.second_channel.frequency",
		"mac_state.current_parameters.relay.mode.serving.cad_periodicity",
		"mac_state.current_parameters.relay.mode.serving.default_channel_index",
		"mac_state.current_parameters.relay.mode.serving.limits.join_requests.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.join_requests.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.limits.notifications.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.notifications.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.limits.overall.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.overall.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.limits.reset_behavior",
		"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.second_channel.ack_offset",
		"mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
		"mac_state.current_parameters.relay.mode.serving.second_channel.frequency",
		"mac_state.current_parameters.relay.mode.serving.uplink_forwarding_rules",
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
		"mac_state.desired_parameters.relay.mode.served.backoff",
		"mac_state.desired_parameters.relay.mode.served.mode.always",
		"mac_state.desired_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
		"mac_state.desired_parameters.relay.mode.served.mode.end_device_controlled",
		"mac_state.desired_parameters.relay.mode.served.second_channel.ack_offset",
		"mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
		"mac_state.desired_parameters.relay.mode.served.second_channel.frequency",
		"mac_state.desired_parameters.relay.mode.serving.cad_periodicity",
		"mac_state.desired_parameters.relay.mode.serving.default_channel_index",
		"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.limits.notifications.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.notifications.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.limits.overall.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.overall.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.limits.reset_behavior",
		"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.ack_offset",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.frequency",
		"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
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
		"mac_state.recent_mac_command_identifiers",
		"mac_state.recent_uplinks",
		"mac_state.rejected_adr_data_rate_indexes",
		"mac_state.rejected_adr_tx_power_indexes",
		"mac_state.rejected_data_rate_ranges",
		"mac_state.rejected_frequencies",
		"mac_state.rx_windows_available",
	}

	legacyADRSettingsFields = []string{
		"mac_settings.adr_margin",
		"mac_settings.use_adr.value",
		"mac_settings.use_adr",
	}

	adrSettingsFields = []string{
		"mac_settings.adr.mode.disabled",
		"mac_settings.adr.mode.dynamic.channel_steering.mode.disabled",
		"mac_settings.adr.mode.dynamic.channel_steering.mode.lora_narrow",
		"mac_settings.adr.mode.dynamic.channel_steering.mode",
		"mac_settings.adr.mode.dynamic.channel_steering",
		"mac_settings.adr.mode.dynamic.margin",
		"mac_settings.adr.mode.dynamic.max_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.max_data_rate_index",
		"mac_settings.adr.mode.dynamic.max_nb_trans",
		"mac_settings.adr.mode.dynamic.max_tx_power_index",
		"mac_settings.adr.mode.dynamic.min_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.min_data_rate_index",
		"mac_settings.adr.mode.dynamic.min_nb_trans",
		"mac_settings.adr.mode.dynamic.min_tx_power_index",
		"mac_settings.adr.mode.dynamic",
		"mac_settings.adr.mode.static.data_rate_index",
		"mac_settings.adr.mode.static.nb_trans",
		"mac_settings.adr.mode.static.tx_power_index",
		"mac_settings.adr.mode.static",
		"mac_settings.adr.mode",
		"mac_settings.adr",
	}

	dynamicADRSettingsFields = []string{
		"mac_settings.adr.mode.dynamic.channel_steering.mode.disabled",
		"mac_settings.adr.mode.dynamic.channel_steering.mode.lora_narrow",
		"mac_settings.adr.mode.dynamic.channel_steering.mode",
		"mac_settings.adr.mode.dynamic.channel_steering",
		"mac_settings.adr.mode.dynamic.margin",
		"mac_settings.adr.mode.dynamic.max_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.max_nb_trans",
		"mac_settings.adr.mode.dynamic.max_tx_power_index",
		"mac_settings.adr.mode.dynamic.min_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.min_nb_trans",
		"mac_settings.adr.mode.dynamic.min_tx_power_index",
		"mac_settings.adr.mode.dynamic",
	}
)

// Set implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	st := newSetDeviceState(req.EndDevice, req.FieldMask.GetPaths()...)

	requiredRights := append(make([]ttnpb.Right, 0, 2),
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
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
		requiredRights = append(requiredRights, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS)
	}
	if err := rights.RequireApplication(ctx, st.Device.Ids.ApplicationIds, requiredRights...); err != nil {
		return nil, err
	}

	// Account for CLI not sending ids.* paths.
	st.AddSetFields(
		"ids.application_ids",
		"ids.device_id",
	)
	if st.Device.Ids.JoinEui != nil {
		st.AddSetFields(
			"ids.join_eui",
		)
	}
	if st.Device.Ids.DevEui != nil {
		st.AddSetFields(
			"ids.dev_eui",
		)
	}
	if st.Device.Ids.DevAddr != nil {
		st.AddSetFields(
			"ids.dev_addr",
		)
	}

	if err := st.ValidateSetField(
		func() bool { return st.Device.FrequencyPlanId != "" },
		"frequency_plan_id",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		st.Device.LorawanPhyVersion.Validate,
		"lorawan_phy_version",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		st.Device.LorawanVersion.Validate,
		"lorawan_version",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		func() error {
			if st.Device.MacState == nil {
				return nil
			}
			return st.Device.MacState.LorawanVersion.Validate()
		},
		"mac_state.lorawan_version",
	); err != nil {
		return nil, err
	}
	if err := st.ValidateSetFieldWithCause(
		func() error {
			if st.Device.PendingMacState == nil {
				return nil
			}
			return st.Device.PendingMacState.LorawanVersion.Validate()
		},
		"pending_mac_state.lorawan_version",
	); err != nil {
		return nil, err
	}

	// Ensure ids.dev_addr and session.dev_addr are consistent.
	if st.HasSetField("ids.dev_addr") {
		if err := st.ValidateField(func(dev *ttnpb.EndDevice) bool {
			if st.Device.Ids.DevAddr == nil {
				return dev.GetSession() == nil
			}
			return dev.GetSession() != nil && bytes.Equal(dev.Session.DevAddr, st.Device.Ids.DevAddr)
		}, "session.dev_addr"); err != nil {
			return nil, err
		}
	} else if st.HasSetField("session.dev_addr") {
		st.Device.Ids.DevAddr = nil
		if devAddr := types.MustDevAddr(st.Device.GetSession().GetDevAddr()); devAddr != nil {
			st.Device.Ids.DevAddr = devAddr.Bytes()
		}
		st.AddSetFields(
			"ids.dev_addr",
		)
	}

	// Ensure FieldIsZero(left) -> FieldIsZero(r), for each r in right.
	for left, right := range ifZeroThenZeroFields {
		if st.HasSetField(left) {
			if !st.Device.FieldIsZero(left) {
				continue
			}
			if err := st.ValidateFieldsAreZero(right...); err != nil {
				return nil, err
			}
		}
		for _, r := range right {
			if !st.HasSetField(r) || st.Device.FieldIsZero(r) {
				continue
			}
			if err := st.ValidateFieldIsNotZero(left); err != nil {
				return nil, err
			}
		}
	}

	// Ensure FieldIsZero(left) -> !FieldIsZero(r), for each r in right.
	for left, right := range ifZeroThenNotZeroFields {
		if st.HasSetField(left) {
			if !st.Device.FieldIsZero(left) {
				continue
			}
			if err := st.ValidateFieldsAreNotZero(right...); err != nil {
				return nil, err
			}
		}
		for _, r := range right {
			if !st.HasSetField(r) || !st.Device.FieldIsZero(r) {
				continue
			}
			if err := st.ValidateFieldIsNotZero(left); err != nil {
				return nil, err
			}
		}
	}

	// Ensure FieldIsZero(left) -> r.Func(map rr -> *ttnpb.EndDevice), for each rr in r.Fields for each r in rs.
	for left, rs := range ifZeroThenFuncFields {
		for _, r := range rs {
			if st.HasSetField(left) {
				if !st.Device.FieldIsZero(left) {
					continue
				}
				if err := st.ValidateFields(r.Func, r.Fields...); err != nil {
					return nil, err
				}
			}
			if !st.HasSetField(r.Fields...) {
				continue
			}

			left := left
			r := r
			if err := st.ValidateFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
				if !m[left].FieldIsZero(left) {
					return true, ""
				}
				return r.Func(m)
			}, append([]string{left}, r.Fields...)...); err != nil {
				return nil, err
			}
		}
	}

	// Ensure !FieldIsZero(left) -> FieldIsZero(r), for each r in right.
	for left, right := range ifNotZeroThenZeroFields {
		if st.HasSetField(left) {
			if st.Device.FieldIsZero(left) {
				continue
			}
			if err := st.ValidateFieldsAreZero(right...); err != nil {
				return nil, err
			}
		}
		for _, r := range right {
			if !st.HasSetField(r) || st.Device.FieldIsZero(r) {
				continue
			}
			if err := st.ValidateFieldIsZero(left); err != nil {
				return nil, err
			}
		}
	}

	// Ensure !FieldIsZero(left) -> !FieldIsZero(r), for each r in right.
	for left, right := range ifNotZeroThenNotZeroFields {
		if st.HasSetField(left) {
			if st.Device.FieldIsZero(left) {
				continue
			}
			if err := st.ValidateFieldsAreNotZero(right...); err != nil {
				return nil, err
			}
		}
		for _, r := range right {
			if !st.HasSetField(r) || !st.Device.FieldIsZero(r) {
				continue
			}
			if err := st.ValidateFieldIsZero(left); err != nil {
				return nil, err
			}
		}
	}

	// Ensure !FieldIsZero(left) -> r.Func(map rr -> *ttnpb.EndDevice), for each rr in r.Fields for each r in rs.
	for left, rs := range ifNotZeroThenFuncFields {
		for _, r := range rs {
			if st.HasSetField(left) {
				if st.Device.FieldIsZero(left) {
					continue
				}
				if err := st.ValidateFields(r.Func, r.Fields...); err != nil {
					return nil, err
				}
			}
			if !st.HasSetField(r.Fields...) {
				continue
			}

			left := left
			r := r
			if err := st.ValidateFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
				if m[left].FieldIsZero(left) {
					return true, ""
				}
				return r.Func(m)
			}, append([]string{left}, r.Fields...)...); err != nil {
				return nil, err
			}
		}
	}

	// Ensure parameters are consistent with band specifications.
	if st.HasSetField(
		"frequency_plan_id",
		"lorawan_phy_version",
		"mac_settings.adr.mode.disabled",
		"mac_settings.adr.mode.dynamic.channel_steering.mode.disabled",
		"mac_settings.adr.mode.dynamic.channel_steering.mode.lora_narrow",
		"mac_settings.adr.mode.dynamic.channel_steering.mode",
		"mac_settings.adr.mode.dynamic.channel_steering",
		"mac_settings.adr.mode.dynamic.margin",
		"mac_settings.adr.mode.dynamic.max_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.max_data_rate_index",
		"mac_settings.adr.mode.dynamic.max_nb_trans",
		"mac_settings.adr.mode.dynamic.max_tx_power_index",
		"mac_settings.adr.mode.dynamic.min_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.min_data_rate_index",
		"mac_settings.adr.mode.dynamic.min_nb_trans",
		"mac_settings.adr.mode.dynamic.min_tx_power_index",
		"mac_settings.adr.mode.dynamic",
		"mac_settings.adr.mode.static.data_rate_index",
		"mac_settings.adr.mode.static.nb_trans",
		"mac_settings.adr.mode.static.tx_power_index",
		"mac_settings.adr.mode.static",
		"mac_settings.adr.mode",
		"mac_settings.adr",
		"mac_settings.desired_ping_slot_data_rate_index.value",
		"mac_settings.desired_relay.mode.served.second_channel.data_rate_index",
		"mac_settings.desired_relay.mode.serving.default_channel_index",
		"mac_settings.desired_relay.mode.serving.second_channel.data_rate_index",
		"mac_settings.desired_rx2_data_rate_index.value",
		"mac_settings.downlink_dwell_time.value",
		"mac_settings.factory_preset_frequencies",
		"mac_settings.ping_slot_data_rate_index.value",
		"mac_settings.ping_slot_frequency.value",
		"mac_settings.relay.mode.served.second_channel.data_rate_index",
		"mac_settings.relay.mode.serving.default_channel_index",
		"mac_settings.relay.mode.serving.second_channel.data_rate_index",
		"mac_settings.rx2_data_rate_index.value",
		"mac_settings.uplink_dwell_time.value",
		"mac_settings.use_adr.value",
		"mac_state.current_parameters.adr_data_rate_index",
		"mac_state.current_parameters.adr_tx_power_index",
		"mac_state.current_parameters.channels",
		"mac_state.current_parameters.ping_slot_data_rate_index_value.value",
		"mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
		"mac_state.current_parameters.relay.mode.serving.default_channel_index",
		"mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
		"mac_state.current_parameters.rx2_data_rate_index",
		"mac_state.desired_parameters.adr_data_rate_index",
		"mac_state.desired_parameters.adr_tx_power_index",
		"mac_state.desired_parameters.channels",
		"mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
		"mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
		"mac_state.desired_parameters.relay.mode.serving.default_channel_index",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
		"mac_state.desired_parameters.rx2_data_rate_index",
		"pending_mac_state.current_parameters.adr_data_rate_index",
		"pending_mac_state.current_parameters.adr_tx_power_index",
		"pending_mac_state.current_parameters.channels",
		"pending_mac_state.current_parameters.ping_slot_data_rate_index_value.value",
		"pending_mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
		"pending_mac_state.current_parameters.relay.mode.serving.default_channel_index",
		"pending_mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
		"pending_mac_state.current_parameters.rx2_data_rate_index",
		"pending_mac_state.desired_parameters.adr_data_rate_index",
		"pending_mac_state.desired_parameters.adr_tx_power_index",
		"pending_mac_state.desired_parameters.channels",
		"pending_mac_state.desired_parameters.ping_slot_data_rate_index_value.value",
		"pending_mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
		"pending_mac_state.desired_parameters.relay.mode.serving.default_channel_index",
		"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
		"pending_mac_state.desired_parameters.rx2_data_rate_index",
		"supports_class_b",
	) {
		var deferredPHYValidations []func(*band.Band, *frequencyplans.FrequencyPlan) error
		withPHY := func(f func(*band.Band, *frequencyplans.FrequencyPlan) error) error {
			deferredPHYValidations = append(deferredPHYValidations, f)
			return nil
		}
		if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
			fps, err := ns.FrequencyPlansStore(ctx)
			if err != nil {
				return err
			}
			fp, phy, err := DeviceFrequencyPlanAndBand(&ttnpb.EndDevice{
				FrequencyPlanId:   m["frequency_plan_id"].GetFrequencyPlanId(),
				LorawanPhyVersion: m["lorawan_phy_version"].GetLorawanPhyVersion(),
			}, fps)
			if err != nil {
				return err
			}
			withPHY = func(f func(*band.Band, *frequencyplans.FrequencyPlan) error) error {
				return f(phy, fp)
			}
			for _, f := range deferredPHYValidations {
				if err := f(phy, fp); err != nil {
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

		hasPHYUpdate := st.HasSetField(
			"frequency_plan_id",
			"lorawan_phy_version",
		)
		hasSetField := func(field string) (fieldToRetrieve string, validate bool) {
			return field, st.HasSetField(field) || hasPHYUpdate
		}

		setFields := func(fields ...string) []string {
			setFields := make([]string, 0, len(fields))
			for _, field := range fields {
				if st.HasSetField(field) {
					setFields = append(setFields, field)
				}
			}
			return setFields
		}

		if st.HasSetField(
			"frequency_plan_id",
			"version_ids.band_id",
		) {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, fp *frequencyplans.FrequencyPlan) error {
					if devBandID := dev.GetVersionIds().GetBandId(); devBandID != "" && devBandID != fp.BandID {
						return newInvalidFieldValueError("version_ids.band_id").WithCause(
							errDeviceAndFrequencyPlanBandMismatch.WithAttributes(
								"dev_band_id", devBandID,
								"fp_band_id", fp.BandID,
							),
						)
					}
					return nil
				})
			}, "version_ids.band_id"); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.rx2_data_rate_index.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetRx2DataRateIndex() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacSettings.Rx2DataRateIndex.Value]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.desired_rx2_data_rate_index.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetDesiredRx2DataRateIndex() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacSettings.DesiredRx2DataRateIndex.Value]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.ping_slot_data_rate_index.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetPingSlotDataRateIndex() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacSettings.PingSlotDataRateIndex.Value]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.desired_ping_slot_data_rate_index.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetDesiredPingSlotDataRateIndex() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacSettings.DesiredPingSlotDataRateIndex.Value]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.desired_relay.mode.served.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetDesiredRelay().GetServed().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacSettings.DesiredRelay.GetServed().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.desired_relay.mode.serving.default_channel_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetDesiredRelay().GetServing() == nil {
						return nil
					}
					chIdx := dev.MacSettings.DesiredRelay.GetServing().DefaultChannelIndex
					if chIdx >= uint32(len(phy.Relay.WORChannels)) {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.desired_relay.mode.serving.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetDesiredRelay().GetServing().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacSettings.DesiredRelay.GetServing().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.adr.mode.dynamic.max_data_rate_index.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetAdr().GetDynamic().GetMaxDataRateIndex() == nil {
						return nil
					}
					drIdx := dev.MacSettings.Adr.GetDynamic().MaxDataRateIndex.Value
					_, ok := phy.DataRates[drIdx]
					if !ok || drIdx > phy.MaxADRDataRateIndex {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.adr.mode.dynamic.min_data_rate_index.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetAdr().GetDynamic().GetMinDataRateIndex() == nil {
						return nil
					}
					drIdx := dev.MacSettings.Adr.GetDynamic().MinDataRateIndex.Value
					_, ok := phy.DataRates[drIdx]
					if !ok || drIdx > phy.MaxADRDataRateIndex {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.adr.mode.dynamic.max_tx_power_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetAdr().GetDynamic().GetMaxTxPowerIndex() == nil {
						return nil
					}
					if dev.MacSettings.Adr.GetDynamic().MaxTxPowerIndex.Value > uint32(phy.MaxTxPowerIndex()) {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.adr.mode.dynamic.min_tx_power_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetAdr().GetDynamic().GetMinTxPowerIndex() == nil {
						return nil
					}
					if dev.MacSettings.Adr.GetDynamic().MinTxPowerIndex.Value > uint32(phy.MaxTxPowerIndex()) {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if setFields := setFields(dynamicADRSettingsFields...); hasPHYUpdate || len(setFields) > 0 {
			fields := setFields
			if hasPHYUpdate {
				fields = append(fields, "mac_settings.adr.mode")
			}
			if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if phy.SupportsDynamicADR {
						return nil
					}
					for _, field := range fields {
						if m[field].GetMacSettings().GetAdr().GetDynamic() != nil {
							return newInvalidFieldValueError(field)
						}
					}
					return nil
				})
			},
				fields...,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.adr.mode.static.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetAdr().GetStatic() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacSettings.Adr.GetStatic().DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.adr.mode.static.tx_power_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetAdr().GetStatic() == nil {
						return nil
					}
					if dev.MacSettings.Adr.GetStatic().TxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.uplink_dwell_time.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetUplinkDwellTime() == nil {
						return nil
					}
					if !phy.TxParamSetupReqSupport {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.downlink_dwell_time.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetDownlinkDwellTime() == nil {
						return nil
					}
					if !phy.TxParamSetupReqSupport {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.relay.mode.served.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetRelay().GetServed().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacSettings.Relay.GetServed().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.relay.mode.serving.default_channel_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetRelay().GetServing() == nil {
						return nil
					}
					chIdx := dev.MacSettings.Relay.GetServing().DefaultChannelIndex
					if chIdx >= uint32(len(phy.Relay.WORChannels)) {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_settings.relay.mode.serving.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacSettings().GetRelay().GetServing().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacSettings.Relay.GetServing().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_state.current_parameters.rx2_data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				if dev.GetMacState() == nil {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					_, ok := phy.DataRates[dev.MacState.CurrentParameters.Rx2DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_state.desired_parameters.rx2_data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				if dev.GetMacState() == nil {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					_, ok := phy.DataRates[dev.MacState.DesiredParameters.Rx2DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("pending_mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetPendingMacState().GetCurrentParameters().GetRelay().GetServed().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.PendingMacState.CurrentParameters.Relay.GetServed().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("pending_mac_state.current_parameters.relay.mode.serving.default_channel_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetPendingMacState().GetCurrentParameters().GetRelay().GetServing() == nil {
						return nil
					}
					chIdx := dev.PendingMacState.CurrentParameters.Relay.GetServing().DefaultChannelIndex
					if chIdx >= uint32(len(phy.Relay.WORChannels)) {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("pending_mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetPendingMacState().GetCurrentParameters().GetRelay().GetServing().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.PendingMacState.CurrentParameters.Relay.GetServing().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("pending_mac_state.current_parameters.rx2_data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				if dev.GetPendingMacState() == nil {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					_, ok := phy.DataRates[dev.PendingMacState.CurrentParameters.Rx2DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("pending_mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetPendingMacState().GetDesiredParameters().GetRelay().GetServed().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.PendingMacState.DesiredParameters.Relay.GetServed().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("pending_mac_state.desired_parameters.relay.mode.serving.default_channel_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetPendingMacState().GetDesiredParameters().GetRelay().GetServing() == nil {
						return nil
					}
					chIdx := dev.PendingMacState.DesiredParameters.Relay.GetServing().DefaultChannelIndex
					if chIdx >= uint32(len(phy.Relay.WORChannels)) {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("pending_mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetPendingMacState().GetDesiredParameters().GetRelay().GetServing().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.PendingMacState.DesiredParameters.Relay.GetServing().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("pending_mac_state.desired_parameters.rx2_data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				if dev.GetPendingMacState() == nil {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					_, ok := phy.DataRates[dev.PendingMacState.DesiredParameters.Rx2DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_state.current_parameters.ping_slot_data_rate_index_value.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				if dev.GetMacState() == nil || dev.MacState.CurrentParameters.PingSlotDataRateIndexValue == nil {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					_, ok := phy.DataRates[dev.MacState.CurrentParameters.PingSlotDataRateIndexValue.Value]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacState().GetCurrentParameters().GetRelay().GetServed().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacState.CurrentParameters.Relay.GetServed().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_state.current_parameters.relay.mode.serving.default_channel_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacState().GetCurrentParameters().GetRelay().GetServing() == nil {
						return nil
					}
					chIdx := dev.MacState.CurrentParameters.Relay.GetServing().DefaultChannelIndex
					if chIdx >= uint32(len(phy.Relay.WORChannels)) {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacState().GetCurrentParameters().GetRelay().GetServing().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacState.CurrentParameters.Relay.GetServing().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_state.desired_parameters.ping_slot_data_rate_index_value.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				if dev.GetMacState() == nil || dev.MacState.DesiredParameters.PingSlotDataRateIndexValue == nil {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					_, ok := phy.DataRates[dev.MacState.DesiredParameters.PingSlotDataRateIndexValue.Value]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacState().GetDesiredParameters().GetRelay().GetServed().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacState.DesiredParameters.Relay.GetServed().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_state.desired_parameters.relay.mode.serving.default_channel_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacState().GetDesiredParameters().GetRelay().GetServing() == nil {
						return nil
					}
					chIdx := dev.MacState.DesiredParameters.Relay.GetServing().DefaultChannelIndex
					if chIdx >= uint32(len(phy.Relay.WORChannels)) {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if dev.GetMacState().GetDesiredParameters().GetRelay().GetServing().GetSecondChannel() == nil {
						return nil
					}
					_, ok := phy.DataRates[dev.MacState.DesiredParameters.Relay.GetServing().SecondChannel.DataRateIndex]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}

		if field, validate := hasSetField("pending_mac_state.current_parameters.ping_slot_data_rate_index_value.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				if dev.GetPendingMacState() == nil || dev.PendingMacState.CurrentParameters.PingSlotDataRateIndexValue == nil {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					_, ok := phy.DataRates[dev.PendingMacState.CurrentParameters.PingSlotDataRateIndexValue.Value]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}
		if field, validate := hasSetField("pending_mac_state.desired_parameters.ping_slot_data_rate_index_value.value"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				if dev.GetPendingMacState() == nil || dev.PendingMacState.DesiredParameters.PingSlotDataRateIndexValue == nil {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					_, ok := phy.DataRates[dev.PendingMacState.DesiredParameters.PingSlotDataRateIndexValue.Value]
					if !ok {
						return newInvalidFieldValueError(field)
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}

		if field, validate := hasSetField("mac_settings.factory_preset_frequencies"); validate {
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				if dev.GetMacSettings() == nil || len(dev.MacSettings.FactoryPresetFrequencies) == 0 {
					return nil
				}
				return withPHY(func(phy *band.Band, fp *frequencyplans.FrequencyPlan) error {
					switch phy.CFListType {
					case ttnpb.CFListType_FREQUENCIES:
						// Factory preset frequencies in bands which provide frequencies as part of the CFList
						// are interpreted as being used both for uplinks and downlinks.
						for _, frequency := range dev.MacSettings.FactoryPresetFrequencies {
							_, inSubBand := fp.FindSubBand(frequency)
							for _, sb := range phy.SubBands {
								if sb.MinFrequency <= frequency && frequency <= sb.MaxFrequency {
									inSubBand = true
									break
								}
							}
							if !inSubBand {
								return newInvalidFieldValueError(field)
							}
						}
					case ttnpb.CFListType_CHANNEL_MASKS:
						// Factory preset frequencies in bands which provide channel masks as part of the CFList
						// are interpreted as enabling explicit uplink channels.
						uplinkChannels := make(map[uint64]struct{}, len(phy.UplinkChannels))
						for _, ch := range phy.UplinkChannels {
							uplinkChannels[ch.Frequency] = struct{}{}
						}
						for _, frequency := range dev.MacSettings.FactoryPresetFrequencies {
							if _, ok := uplinkChannels[frequency]; !ok {
								return newInvalidFieldValueError(field)
							}
						}
					default:
						panic("unreachable")
					}
					return nil
				})
			},
				field,
			); err != nil {
				return nil, err
			}
		}

		if hasPHYUpdate || st.HasSetField(
			"mac_settings.ping_slot_frequency.value",
			"supports_class_b",
		) {
			if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
				if !m["supports_class_b"].GetSupportsClassB() ||
					m["mac_settings.ping_slot_frequency.value"].GetMacSettings().GetPingSlotFrequency().GetValue() > 0 {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if len(phy.PingSlotFrequencies) == 0 {
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

		if hasPHYUpdate || st.HasSetField(
			"mac_settings.desired_ping_slot_frequency.value",
			"supports_class_b",
		) {
			if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
				if !m["supports_class_b"].GetSupportsClassB() ||
					m["mac_settings.desired_ping_slot_frequency.value"].GetMacSettings().GetDesiredPingSlotFrequency().GetValue() > 0 {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if len(phy.PingSlotFrequencies) == 0 {
						return newInvalidFieldValueError("mac_settings.desired_ping_slot_frequency.value")
					}
					return nil
				})
			},
				"mac_settings.desired_ping_slot_frequency.value",
				"supports_class_b",
			); err != nil {
				return nil, err
			}
		}

		if hasPHYUpdate || st.HasSetField(
			"mac_settings.beacon_frequency.value",
			"supports_class_b",
		) {
			if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
				if !m["supports_class_b"].GetSupportsClassB() ||
					m["mac_settings.beacon_frequency.value"].GetMacSettings().GetBeaconFrequency().GetValue() > 0 {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if len(phy.Beacon.Frequencies) == 0 {
						return newInvalidFieldValueError("mac_settings.beacon_frequency.value")
					}
					return nil
				})
			},
				"mac_settings.beacon_frequency.value",
				"supports_class_b",
			); err != nil {
				return nil, err
			}
		}

		if hasPHYUpdate || st.HasSetField(
			"mac_settings.desired_beacon_frequency.value",
			"supports_class_b",
		) {
			if err := st.WithFields(func(m map[string]*ttnpb.EndDevice) error {
				if !m["supports_class_b"].GetSupportsClassB() ||
					m["mac_settings.desired_beacon_frequency.value"].GetMacSettings().GetDesiredBeaconFrequency().GetValue() > 0 {
					return nil
				}
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
					if len(phy.Beacon.Frequencies) == 0 {
						return newInvalidFieldValueError("mac_settings.desired_beacon_frequency.value")
					}
					return nil
				})
			},
				"mac_settings.desired_beacon_frequency.value",
				"supports_class_b",
			); err != nil {
				return nil, err
			}
		}

		for p, isValid := range map[string]func(*ttnpb.EndDevice, *band.Band) bool{
			"mac_settings.use_adr.value": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return !dev.GetMacSettings().GetUseAdr().GetValue() || phy.SupportsDynamicADR
			},
			"mac_state.current_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetMacState().GetCurrentParameters().GetAdrDataRateIndex() <= phy.MaxADRDataRateIndex
			},
			"mac_state.current_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetMacState().GetCurrentParameters().GetAdrTxPowerIndex() <= uint32(phy.MaxTxPowerIndex())
			},
			"mac_state.current_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return len(dev.GetMacState().GetCurrentParameters().GetChannels()) <= int(phy.MaxUplinkChannels)
			},
			"mac_state.desired_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetMacState().GetDesiredParameters().GetAdrDataRateIndex() <= phy.MaxADRDataRateIndex
			},
			"mac_state.desired_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetMacState().GetDesiredParameters().GetAdrTxPowerIndex() <= uint32(phy.MaxTxPowerIndex())
			},
			"mac_state.desired_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return len(dev.GetMacState().GetDesiredParameters().GetChannels()) <= int(phy.MaxUplinkChannels)
			},
			"pending_mac_state.current_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetPendingMacState().GetCurrentParameters().GetAdrDataRateIndex() <= phy.MaxADRDataRateIndex
			},
			"pending_mac_state.current_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetPendingMacState().GetCurrentParameters().GetAdrTxPowerIndex() <= uint32(phy.MaxTxPowerIndex())
			},
			"pending_mac_state.current_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return len(dev.GetPendingMacState().GetCurrentParameters().GetChannels()) <= int(phy.MaxUplinkChannels)
			},
			"pending_mac_state.desired_parameters.adr_data_rate_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetPendingMacState().GetDesiredParameters().GetAdrDataRateIndex() <= phy.MaxADRDataRateIndex
			},
			"pending_mac_state.desired_parameters.adr_tx_power_index": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return dev.GetPendingMacState().GetDesiredParameters().GetAdrTxPowerIndex() <= uint32(phy.MaxTxPowerIndex())
			},
			"pending_mac_state.desired_parameters.channels": func(dev *ttnpb.EndDevice, phy *band.Band) bool {
				return len(dev.GetPendingMacState().GetDesiredParameters().GetChannels()) <= int(phy.MaxUplinkChannels)
			},
		} {
			if !hasPHYUpdate && !st.HasSetField(p) {
				continue
			}
			p, isValid := p, isValid
			if err := st.WithField(func(dev *ttnpb.EndDevice) error {
				return withPHY(func(phy *band.Band, _ *frequencyplans.FrequencyPlan) error {
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

	// Ensure ADR dynamic parameters are monotonic.
	// If one of the extrema is missing, the other extrema is considered to be valid.
	if err := st.ValidateSetFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		{
			min := m["mac_settings.adr.mode.dynamic.min_data_rate_index.value"].GetMacSettings().GetAdr().GetDynamic().GetMinDataRateIndex()
			max := m["mac_settings.adr.mode.dynamic.max_data_rate_index.value"].GetMacSettings().GetAdr().GetDynamic().GetMaxDataRateIndex()

			if min != nil && max != nil && max.Value < min.Value {
				return false, "mac_settings.adr.mode.dynamic.max_data_rate_index.value"
			}
		}
		{
			min := m["mac_settings.adr.mode.dynamic.min_tx_power_index"].GetMacSettings().GetAdr().GetDynamic().GetMinTxPowerIndex()
			max := m["mac_settings.adr.mode.dynamic.max_tx_power_index"].GetMacSettings().GetAdr().GetDynamic().GetMaxTxPowerIndex()

			if min != nil && max != nil && max.Value < min.Value {
				return false, "mac_settings.adr.mode.dynamic.max_tx_power_index"
			}
		}
		{
			min := m["mac_settings.adr.mode.dynamic.min_nb_trans"].GetMacSettings().GetAdr().GetDynamic().GetMinNbTrans()
			max := m["mac_settings.adr.mode.dynamic.max_nb_trans"].GetMacSettings().GetAdr().GetDynamic().GetMaxNbTrans()

			if min != nil && max != nil && max.Value < min.Value {
				return false, "mac_settings.adr.mode.dynamic.max_nb_trans"
			}
		}
		return true, ""
	},
		"mac_settings.adr.mode.dynamic.max_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.max_nb_trans",
		"mac_settings.adr.mode.dynamic.max_tx_power_index",
		"mac_settings.adr.mode.dynamic.min_data_rate_index.value",
		"mac_settings.adr.mode.dynamic.min_nb_trans",
		"mac_settings.adr.mode.dynamic.min_tx_power_index",
	); err != nil {
		return nil, err
	}

	var getTransforms []func(*ttnpb.EndDevice)
	if st.Device.Session != nil {
		for p, isZero := range map[string]func() bool{
			"session.dev_addr":                 types.MustDevAddr(st.Device.Session.DevAddr).OrZero().IsZero,
			"session.keys.f_nwk_s_int_key.key": st.Device.Session.Keys.GetFNwkSIntKey().IsZero,
			"session.keys.nwk_s_enc_key.key": func() bool {
				return st.Device.Session.Keys.GetNwkSEncKey() != nil && st.Device.Session.Keys.NwkSEncKey.IsZero()
			},
			"session.keys.s_nwk_s_int_key.key": func() bool {
				return st.Device.Session.Keys.GetSNwkSIntKey() != nil && st.Device.Session.Keys.SNwkSIntKey.IsZero()
			},
		} {
			p, isZero := p, isZero
			if err := st.ValidateSetField(func() bool { return !isZero() }, p); err != nil {
				return nil, err
			}
		}
		if st.HasSetField("session.keys.f_nwk_s_int_key.key") {
			k := st.Device.Session.Keys.FNwkSIntKey.Key
			fNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.Session.Keys.FNwkSIntKey = fNwkSIntKey
			st.AddSetFields(
				"session.keys.f_nwk_s_int_key.encrypted_key",
				"session.keys.f_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.Session.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if k := st.Device.Session.Keys.GetNwkSEncKey().GetKey(); k != nil && st.HasSetField("session.keys.nwk_s_enc_key.key") {
			nwkSEncKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.Session.Keys.NwkSEncKey = nwkSEncKey
			st.AddSetFields(
				"session.keys.nwk_s_enc_key.encrypted_key",
				"session.keys.nwk_s_enc_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.Session.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if k := st.Device.Session.Keys.GetSNwkSIntKey().GetKey(); k != nil && st.HasSetField("session.keys.s_nwk_s_int_key.key") {
			sNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.Session.Keys.SNwkSIntKey = sNwkSIntKey
			st.AddSetFields(
				"session.keys.s_nwk_s_int_key.encrypted_key",
				"session.keys.s_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.Session.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
	}
	if st.Device.PendingSession != nil {
		for p, isZero := range map[string]func() bool{
			"pending_session.dev_addr":                 types.MustDevAddr(st.Device.PendingSession.DevAddr).OrZero().IsZero,
			"pending_session.keys.f_nwk_s_int_key.key": st.Device.PendingSession.Keys.GetFNwkSIntKey().IsZero,
			"pending_session.keys.nwk_s_enc_key.key":   st.Device.PendingSession.Keys.GetNwkSEncKey().IsZero,
			"pending_session.keys.s_nwk_s_int_key.key": st.Device.PendingSession.Keys.GetSNwkSIntKey().IsZero,
			"pending_session.keys.session_key_id": func() bool {
				return len(st.Device.PendingSession.Keys.GetSessionKeyId()) == 0
			},
		} {
			p, isZero := p, isZero
			if err := st.ValidateSetField(func() bool { return !isZero() }, p); err != nil {
				return nil, err
			}
		}
		if st.HasSetField("pending_session.keys.f_nwk_s_int_key.key") {
			k := st.Device.PendingSession.Keys.FNwkSIntKey.Key
			fNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingSession.Keys.FNwkSIntKey = fNwkSIntKey
			st.AddSetFields(
				"pending_session.keys.f_nwk_s_int_key.encrypted_key",
				"pending_session.keys.f_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingSession.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_session.keys.nwk_s_enc_key.key") {
			k := st.Device.PendingSession.Keys.NwkSEncKey.Key
			nwkSEncKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingSession.Keys.NwkSEncKey = nwkSEncKey
			st.AddSetFields(
				"pending_session.keys.nwk_s_enc_key.encrypted_key",
				"pending_session.keys.nwk_s_enc_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingSession.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_session.keys.s_nwk_s_int_key.key") {
			k := st.Device.PendingSession.Keys.SNwkSIntKey.Key
			sNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingSession.Keys.SNwkSIntKey = sNwkSIntKey
			st.AddSetFields(
				"pending_session.keys.s_nwk_s_int_key.encrypted_key",
				"pending_session.keys.s_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingSession.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
	}
	if st.Device.PendingMacState.GetQueuedJoinAccept() != nil {
		for p, isZero := range map[string]func() bool{
			"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key": st.Device.PendingMacState.QueuedJoinAccept.Keys.GetFNwkSIntKey().IsZero,
			"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key":   st.Device.PendingMacState.QueuedJoinAccept.Keys.GetNwkSEncKey().IsZero,
			"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key": st.Device.PendingMacState.QueuedJoinAccept.Keys.GetSNwkSIntKey().IsZero,
			"pending_mac_state.queued_join_accept.keys.session_key_id":      func() bool { return len(st.Device.PendingMacState.QueuedJoinAccept.Keys.GetSessionKeyId()) == 0 },
			"pending_mac_state.queued_join_accept.payload":                  func() bool { return len(st.Device.PendingMacState.QueuedJoinAccept.Payload) == 0 },
			"pending_mac_state.queued_join_accept.dev_addr": types.MustDevAddr(
				st.Device.PendingMacState.QueuedJoinAccept.DevAddr,
			).OrZero().IsZero,
		} {
			p, isZero := p, isZero
			if err := st.ValidateSetField(func() bool { return !isZero() }, p); err != nil {
				return nil, err
			}
		}
		if st.HasSetField("pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key") {
			k := st.Device.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey.Key
			fNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey = fNwkSIntKey
			st.AddSetFields(
				"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key") {
			k := st.Device.PendingMacState.QueuedJoinAccept.Keys.NwkSEncKey.Key
			nwkSEncKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingMacState.QueuedJoinAccept.Keys.NwkSEncKey = nwkSEncKey
			st.AddSetFields(
				"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingMacState.QueuedJoinAccept.Keys.NwkSEncKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
		if st.HasSetField("pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key") {
			k := st.Device.PendingMacState.QueuedJoinAccept.Keys.SNwkSIntKey.Key
			sNwkSIntKey, err := cryptoutil.WrapAES128Key(ctx, types.MustAES128Key(k).OrZero(), ns.deviceKEKLabel, ns.KeyService())
			if err != nil {
				return nil, err
			}
			st.Device.PendingMacState.QueuedJoinAccept.Keys.SNwkSIntKey = sNwkSIntKey
			st.AddSetFields(
				"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.encrypted_key",
				"pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.kek_label",
			)
			getTransforms = append(getTransforms, func(dev *ttnpb.EndDevice) {
				dev.PendingMacState.QueuedJoinAccept.Keys.SNwkSIntKey = &ttnpb.KeyEnvelope{
					Key: k,
				}
			})
		}
	}

	var (
		// hasSession indicates whether the effective device model contains a non-zero session.
		hasSession bool

		// hasMACState indicates whether the effective device model contains a non-zero MAC state.
		hasMACState bool
	)
	if err := st.ValidateSetFields(func(m map[string]*ttnpb.EndDevice) (bool, string) {
		for k, v := range m {
			switch {
			case strings.HasPrefix(k, "mac_state."):
				if v.MacState != nil {
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
		switch {
		case !hasMACState && !hasSession && !isMulticast:
			return true, ""

		case !hasSession:
			return false, "session"

		case !hasMACState && st.HasSetField("mac_state"):
			return false, "mac_state"
		}

		var macVersion ttnpb.MACVersion
		if hasMACState {
			// NOTE: If not set, this will be derived from top-level device model.
			if isMulticast {
				if dev, ok := m["mac_state.device_class"]; ok && dev.MacState.GetDeviceClass() == ttnpb.Class_CLASS_A {
					return false, "mac_state.device_class"
				}
			}
			// NOTE: If not set, this will be derived from top-level device model.
			if dev, ok := m["mac_state.lorawan_version"]; ok && dev.MacState == nil {
				return false, "mac_state.lorawan_version"
			} else if !ok {
				macVersion = m["lorawan_version"].LorawanVersion
			} else {
				macVersion = dev.MacState.LorawanVersion
			}
		} else {
			macVersion = m["lorawan_version"].LorawanVersion
		}

		if dev, ok := m["session.dev_addr"]; !ok || dev.Session == nil {
			return false, "session.dev_addr"
		}

		getFNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetSession().GetKeys().GetFNwkSIntKey()
		}
		if setKeyIsZero(m, getFNwkSIntKey, "session.keys.f_nwk_s_int_key") {
			return false, "session.keys.f_nwk_s_int_key.key"
		}

		getNwkSEncKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetSession().GetKeys().GetNwkSEncKey()
		}
		getSNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
			return dev.GetSession().GetKeys().GetSNwkSIntKey()
		}
		isZero := struct {
			NwkSEncKey  bool
			SNwkSIntKey bool
		}{
			NwkSEncKey:  setKeyIsZero(m, getNwkSEncKey, "session.keys.nwk_s_enc_key"),
			SNwkSIntKey: setKeyIsZero(m, getSNwkSIntKey, "session.keys.s_nwk_s_int_key"),
		}
		if macspec.UseNwkKey(macVersion) {
			if isZero.NwkSEncKey {
				return false, "session.keys.nwk_s_enc_key.key"
			}
			if isZero.SNwkSIntKey {
				return false, "session.keys.s_nwk_s_int_key.key"
			}
		} else {
			if st.HasSetField("session.keys.nwk_s_enc_key.key") &&
				!setKeyEqual(m, getFNwkSIntKey, getNwkSEncKey, "session.keys.f_nwk_s_int_key", "session.keys.nwk_s_enc_key") {
				return false, "session.keys.nwk_s_enc_key.key"
			}
			if st.HasSetField("session.keys.s_nwk_s_int_key.key") &&
				!setKeyEqual(m, getFNwkSIntKey, getSNwkSIntKey, "session.keys.f_nwk_s_int_key", "session.keys.s_nwk_s_int_key") {
				return false, "session.keys.s_nwk_s_int_key.key"
			}
		}
		if m["supports_join"].GetSupportsJoin() {
			if dev, ok := m["session.keys.session_key_id"]; !ok || dev.Session == nil {
				return false, "session.keys.session_key_id"
			}
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
		"mac_state.current_parameters.relay.mode.served.backoff",
		"mac_state.current_parameters.relay.mode.served.mode.always",
		"mac_state.current_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
		"mac_state.current_parameters.relay.mode.served.mode.end_device_controlled",
		"mac_state.current_parameters.relay.mode.served.second_channel.ack_offset",
		"mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
		"mac_state.current_parameters.relay.mode.served.second_channel.frequency",
		"mac_state.current_parameters.relay.mode.served.serving_device_id",
		"mac_state.current_parameters.relay.mode.serving.cad_periodicity",
		"mac_state.current_parameters.relay.mode.serving.default_channel_index",
		"mac_state.current_parameters.relay.mode.serving.limits.join_requests.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.join_requests.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.limits.notifications.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.notifications.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.limits.overall.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.overall.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.limits.reset_behavior",
		"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
		"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
		"mac_state.current_parameters.relay.mode.serving.second_channel.ack_offset",
		"mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
		"mac_state.current_parameters.relay.mode.serving.second_channel.frequency",
		"mac_state.current_parameters.relay.mode.serving.uplink_forwarding_rules",
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
		"mac_state.desired_parameters.relay.mode.served.backoff",
		"mac_state.desired_parameters.relay.mode.served.mode.always",
		"mac_state.desired_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
		"mac_state.desired_parameters.relay.mode.served.mode.end_device_controlled",
		"mac_state.desired_parameters.relay.mode.served.second_channel.ack_offset",
		"mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
		"mac_state.desired_parameters.relay.mode.served.second_channel.frequency",
		"mac_state.desired_parameters.relay.mode.served.serving_device_id",
		"mac_state.desired_parameters.relay.mode.serving.cad_periodicity",
		"mac_state.desired_parameters.relay.mode.serving.default_channel_index",
		"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.limits.notifications.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.notifications.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.limits.overall.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.overall.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.limits.reset_behavior",
		"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
		"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.ack_offset",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
		"mac_state.desired_parameters.relay.mode.serving.second_channel.frequency",
		"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
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
		"mac_state.pending_relay_downlink.raw_payload",
		"mac_state.pending_relay_downlink",
		"mac_state.pending_requests",
		"mac_state.ping_slot_periodicity.value",
		"mac_state.queued_responses",
		"mac_state.recent_downlinks",
		"mac_state.recent_mac_command_identifiers",
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
		"supports_join",
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
				if v.PendingMacState != nil {
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
		switch {
		case !hasPendingMACState && !hasPendingSession:
			return true, ""
		case !hasPendingMACState:
			return false, "pending_mac_state"
		}
		for k, v := range m {
			if strings.HasPrefix(k, "pending_mac_state.queued_join_accept.") && v.PendingMacState.GetQueuedJoinAccept() != nil {
				hasQueuedJoinAccept = true
				break
			}
		}

		var macVersion ttnpb.MACVersion
		if dev, ok := m["pending_mac_state.lorawan_version"]; !ok || dev.PendingMacState == nil {
			return false, "pending_mac_state.lorawan_version"
		} else {
			macVersion = dev.PendingMacState.LorawanVersion
		}
		useNwkKey := macspec.UseNwkKey(macVersion)

		if hasPendingSession {
			// NOTE: PendingMACState may be set before PendingSession is set by downlink routine.
			if dev, ok := m["pending_session.dev_addr"]; !ok || dev.PendingSession == nil {
				return false, "pending_session.dev_addr"
			}

			getFNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				return dev.GetPendingSession().GetKeys().GetFNwkSIntKey()
			}
			if setKeyIsZero(m, getFNwkSIntKey, "pending_session.keys.f_nwk_s_int_key") {
				return false, "pending_session.keys.f_nwk_s_int_key.key"
			}
			getNwkSEncKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				return dev.GetPendingSession().GetKeys().GetNwkSEncKey()
			}
			if setKeyIsZero(m, getNwkSEncKey, "pending_session.keys.nwk_s_enc_key") {
				return false, "pending_session.keys.nwk_s_enc_key.key"
			}
			getSNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				return dev.GetPendingSession().GetKeys().GetSNwkSIntKey()
			}
			if setKeyIsZero(m, getSNwkSIntKey, "pending_session.keys.s_nwk_s_int_key") {
				return false, "pending_session.keys.s_nwk_s_int_key.key"
			}
			if !useNwkKey {
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
		} else if !hasQueuedJoinAccept {
			return false, "pending_mac_state.queued_join_accept"
		}

		if hasQueuedJoinAccept {
			getFNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				keys := dev.GetPendingMacState().GetQueuedJoinAccept().GetKeys()
				return keys.GetFNwkSIntKey()
			}
			if setKeyIsZero(m, getFNwkSIntKey, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key") {
				return false, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key"
			}
			getNwkSEncKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				keys := dev.GetPendingMacState().GetQueuedJoinAccept().GetKeys()
				return keys.GetNwkSEncKey()
			}
			if setKeyIsZero(m, getNwkSEncKey, "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key") {
				return false, "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key"
			}
			getSNwkSIntKey := func(dev *ttnpb.EndDevice) *ttnpb.KeyEnvelope {
				keys := dev.GetPendingMacState().GetQueuedJoinAccept().GetKeys()
				return keys.GetSNwkSIntKey()
			}
			if setKeyIsZero(m, getSNwkSIntKey, "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key") {
				return false, "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key"
			}

			if !useNwkKey {
				if !setKeyEqual(m, getFNwkSIntKey, getNwkSEncKey, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key", "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key") {
					return false, "pending_mac_state.queued_join_accept.keys.nwk_s_enc_key.key"
				}
				if !setKeyEqual(m, getFNwkSIntKey, getSNwkSIntKey, "pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key", "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key") {
					return false, "pending_mac_state.queued_join_accept.keys.s_nwk_s_int_key.key"
				}
			}

			if dev, ok := m["pending_mac_state.queued_join_accept.keys.session_key_id"]; !ok || dev.PendingMacState.GetQueuedJoinAccept() == nil {
				return false, "pending_mac_state.queued_join_accept.keys.session_key_id"
			}
			if dev, ok := m["pending_mac_state.queued_join_accept.payload"]; !ok || dev.PendingMacState.GetQueuedJoinAccept() == nil {
				return false, "pending_mac_state.queued_join_accept.payload"
			}
			if dev, ok := m["pending_mac_state.queued_join_accept.request.dev_addr"]; !ok || dev.PendingMacState.GetQueuedJoinAccept() == nil {
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
		"pending_mac_state.current_parameters.relay.mode.served.backoff",
		"pending_mac_state.current_parameters.relay.mode.served.mode.always",
		"pending_mac_state.current_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
		"pending_mac_state.current_parameters.relay.mode.served.mode.end_device_controlled",
		"pending_mac_state.current_parameters.relay.mode.served.second_channel.ack_offset",
		"pending_mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
		"pending_mac_state.current_parameters.relay.mode.served.second_channel.frequency",
		"pending_mac_state.current_parameters.relay.mode.served.serving_device_id",
		"pending_mac_state.current_parameters.relay.mode.serving.cad_periodicity",
		"pending_mac_state.current_parameters.relay.mode.serving.default_channel_index",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.join_requests.bucket_size",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.join_requests.reload_rate",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.notifications.bucket_size",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.notifications.reload_rate",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.overall.bucket_size",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.overall.reload_rate",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.reset_behavior",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
		"pending_mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
		"pending_mac_state.current_parameters.relay.mode.serving.second_channel.ack_offset",
		"pending_mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
		"pending_mac_state.current_parameters.relay.mode.serving.second_channel.frequency",
		"pending_mac_state.current_parameters.relay.mode.serving.uplink_forwarding_rules",
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
		"pending_mac_state.desired_parameters.relay.mode.served.backoff",
		"pending_mac_state.desired_parameters.relay.mode.served.mode.always",
		"pending_mac_state.desired_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
		"pending_mac_state.desired_parameters.relay.mode.served.mode.end_device_controlled",
		"pending_mac_state.desired_parameters.relay.mode.served.second_channel.ack_offset",
		"pending_mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
		"pending_mac_state.desired_parameters.relay.mode.served.second_channel.frequency",
		"pending_mac_state.desired_parameters.relay.mode.served.serving_device_id",
		"pending_mac_state.desired_parameters.relay.mode.serving.cad_periodicity",
		"pending_mac_state.desired_parameters.relay.mode.serving.default_channel_index",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.join_requests.bucket_size",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.join_requests.reload_rate",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.notifications.bucket_size",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.notifications.reload_rate",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.overall.bucket_size",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.overall.reload_rate",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.reset_behavior",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
		"pending_mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
		"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.ack_offset",
		"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
		"pending_mac_state.desired_parameters.relay.mode.serving.second_channel.frequency",
		"pending_mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
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
		"pending_mac_state.recent_mac_command_identifiers",
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

	needsDownlinkCheck := st.HasSetField(downlinkInfluencingSetFields[:]...)
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
			"mac_state.current_parameters.relay.mode.served.backoff",
			"mac_state.current_parameters.relay.mode.served.mode.always",
			"mac_state.current_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
			"mac_state.current_parameters.relay.mode.served.mode.end_device_controlled",
			"mac_state.current_parameters.relay.mode.served.second_channel.ack_offset",
			"mac_state.current_parameters.relay.mode.served.second_channel.data_rate_index",
			"mac_state.current_parameters.relay.mode.served.second_channel.frequency",
			"mac_state.current_parameters.relay.mode.served.serving_device_id",
			"mac_state.current_parameters.relay.mode.serving.cad_periodicity",
			"mac_state.current_parameters.relay.mode.serving.default_channel_index",
			"mac_state.current_parameters.relay.mode.serving.limits.join_requests.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.join_requests.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.limits.notifications.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.notifications.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.limits.overall.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.overall.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.limits.reset_behavior",
			"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
			"mac_state.current_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
			"mac_state.current_parameters.relay.mode.serving.second_channel.ack_offset",
			"mac_state.current_parameters.relay.mode.serving.second_channel.data_rate_index",
			"mac_state.current_parameters.relay.mode.serving.second_channel.frequency",
			"mac_state.current_parameters.relay.mode.serving.uplink_forwarding_rules",
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
			"mac_state.desired_parameters.relay.mode.served.backoff",
			"mac_state.desired_parameters.relay.mode.served.mode.always",
			"mac_state.desired_parameters.relay.mode.served.mode.dynamic.smart_enable_level",
			"mac_state.desired_parameters.relay.mode.served.mode.end_device_controlled",
			"mac_state.desired_parameters.relay.mode.served.second_channel.ack_offset",
			"mac_state.desired_parameters.relay.mode.served.second_channel.data_rate_index",
			"mac_state.desired_parameters.relay.mode.served.second_channel.frequency",
			"mac_state.desired_parameters.relay.mode.served.serving_device_id",
			"mac_state.desired_parameters.relay.mode.serving.cad_periodicity",
			"mac_state.desired_parameters.relay.mode.serving.default_channel_index",
			"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.join_requests.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.limits.notifications.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.notifications.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.limits.overall.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.overall.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.limits.reset_behavior",
			"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.bucket_size",
			"mac_state.desired_parameters.relay.mode.serving.limits.uplink_messages.reload_rate",
			"mac_state.desired_parameters.relay.mode.serving.second_channel.ack_offset",
			"mac_state.desired_parameters.relay.mode.serving.second_channel.data_rate_index",
			"mac_state.desired_parameters.relay.mode.serving.second_channel.frequency",
			"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
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
			"mac_state.recent_mac_command_identifiers",
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
	dev, ctx, err := ns.devices.SetByID(ctx, st.Device.Ids.ApplicationIds, st.Device.Ids.DeviceId, ttnpb.EndDeviceFieldPathsTopLevel, st.SetFunc(func(ctx context.Context, stored *ttnpb.EndDevice) error {
		if nonZeroFields := ttnpb.NonZeroFields(stored, st.GetFields()...); len(nonZeroFields) > 0 {
			newStored := &ttnpb.EndDevice{}
			if err := newStored.SetFields(stored, nonZeroFields...); err != nil {
				return err
			}
			stored = newStored
		}
		if hasSession {
			macVersion := stored.GetMacState().GetLorawanVersion()
			if stored.GetMacState() == nil && !st.HasSetField("mac_state") {
				fps, err := ns.FrequencyPlansStore(ctx)
				if err != nil {
					return err
				}
				macState, err := mac.NewState(st.Device, fps, ns.defaultMACSettings)
				if err != nil {
					return err
				}
				if macSets := ttnpb.FieldsWithoutPrefix("mac_state", st.SetFields()...); len(macSets) != 0 {
					if err := macState.SetFields(st.Device.MacState, macSets...); err != nil {
						return err
					}
				}
				st.Device.MacState = macState
				st.AddSetFields(
					"mac_state",
				)
				macVersion = macState.LorawanVersion
			} else if st.HasSetField("mac_state.lorawan_version") {
				macVersion = st.Device.MacState.LorawanVersion
			}

			if st.HasSetField("session.keys.f_nwk_s_int_key.key") && !macspec.UseNwkKey(macVersion) {
				st.Device.Session.Keys.NwkSEncKey = st.Device.Session.Keys.FNwkSIntKey
				st.Device.Session.Keys.SNwkSIntKey = st.Device.Session.Keys.FNwkSIntKey
				st.AddSetFields(
					"session.keys.nwk_s_enc_key.encrypted_key",
					"session.keys.nwk_s_enc_key.kek_label",
					"session.keys.nwk_s_enc_key.key",
					"session.keys.s_nwk_s_int_key.encrypted_key",
					"session.keys.s_nwk_s_int_key.kek_label",
					"session.keys.s_nwk_s_int_key.key",
				)
			}
			if st.HasSetField("session.started_at") && st.Device.GetSession().GetStartedAt() == nil ||
				st.HasSetField("session.session_key_id") && !bytes.Equal(st.Device.GetSession().GetKeys().GetSessionKeyId(), stored.GetSession().GetKeys().GetSessionKeyId()) ||
				stored.GetSession().GetStartedAt() == nil {
				st.Device.Session.StartedAt = timestamppb.New(time.Now()) // NOTE: This is not equivalent to timestamppb.Now().
				st.AddSetFields(
					"session.started_at",
				)
			}
		}
		if hasPendingSession {
			var macVersion ttnpb.MACVersion
			if st.HasSetField("pending_mac_state.lorawan_version") {
				macVersion = st.Device.GetPendingMacState().GetLorawanVersion()
			} else {
				macVersion = stored.GetPendingMacState().GetLorawanVersion()
			}

			useNwkKey := macspec.UseNwkKey(macVersion)
			if st.HasSetField("pending_session.keys.f_nwk_s_int_key.key") && !useNwkKey {
				st.Device.PendingSession.Keys.NwkSEncKey = st.Device.PendingSession.Keys.FNwkSIntKey
				st.Device.PendingSession.Keys.SNwkSIntKey = st.Device.PendingSession.Keys.FNwkSIntKey
				st.AddSetFields(
					"pending_session.keys.nwk_s_enc_key.encrypted_key",
					"pending_session.keys.nwk_s_enc_key.kek_label",
					"pending_session.keys.nwk_s_enc_key.key",
					"pending_session.keys.s_nwk_s_int_key.encrypted_key",
					"pending_session.keys.s_nwk_s_int_key.kek_label",
					"pending_session.keys.s_nwk_s_int_key.key",
				)
			}
			if st.HasSetField("pending_mac_state.queued_join_accept.keys.f_nwk_s_int_key.key") && hasQueuedJoinAccept && !useNwkKey {
				st.Device.PendingMacState.QueuedJoinAccept.Keys.NwkSEncKey = st.Device.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey
				st.Device.PendingMacState.QueuedJoinAccept.Keys.SNwkSIntKey = st.Device.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey
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

		if stored == nil {
			evt = evtCreateEndDevice.NewWithIdentifiersAndData(ctx, st.Device.Ids, nil)
			return nil
		}

		evt = evtUpdateEndDevice.NewWithIdentifiersAndData(ctx, st.Device.Ids, req.FieldMask.GetPaths())
		if st.HasSetField("multicast") && st.Device.Multicast != stored.Multicast {
			return newInvalidFieldValueError("multicast")
		}
		if st.HasSetField("supports_join") && st.Device.SupportsJoin != stored.SupportsJoin {
			return newInvalidFieldValueError("supports_join")
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
		return ttnpb.FilterGetEndDevice(dev, req.FieldMask.GetPaths()...)
	}

	if err := ns.updateDataDownlinkTask(ctx, dev, time.Time{}); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to update downlink task queue after device set")
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.GetPaths()...)
}

// ResetFactoryDefaults implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) ResetFactoryDefaults(ctx context.Context, req *ttnpb.ResetAndGetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.EndDeviceIds.ApplicationIds, appendRequiredDeviceReadRights(
		append(make([]ttnpb.Right, 0, 1+maxRequiredDeviceReadRightCount), ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE),
		req.FieldMask.GetPaths()...,
	)...); err != nil {
		return nil, err
	}

	dev, _, err := ns.devices.SetByID(ctx, req.EndDeviceIds.ApplicationIds, req.EndDeviceIds.DeviceId, addDeviceGetPaths(ttnpb.AddFields(append(req.FieldMask.GetPaths()[:0:0], req.FieldMask.GetPaths()...),
		"frequency_plan_id",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"multicast",
		"session.dev_addr",
		"session.keys",
		"session.queued_application_downlinks",
		"supports_class_b",
		"supports_class_c",
		"supports_join",
	)...), func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if stored == nil {
			return nil, nil, errDeviceNotFound.New()
		}

		stored.BatteryPercentage = nil
		stored.DownlinkMargin = 0
		stored.LastDevStatusReceivedAt = nil
		stored.MacState = nil
		stored.PendingMacState = nil
		stored.PendingSession = nil
		stored.PowerState = ttnpb.PowerState_POWER_UNKNOWN
		if stored.SupportsJoin {
			stored.Session = nil
		} else {
			if stored.Session == nil {
				return nil, nil, ErrCorruptedMACState.
					WithCause(ErrSession)
			}

			fps, err := ns.FrequencyPlansStore(ctx)
			if err != nil {
				return nil, nil, err
			}
			macState, err := mac.NewState(stored, fps, ns.defaultMACSettings)
			if err != nil {
				return nil, nil, err
			}
			stored.MacState = macState
			stored.Session = &ttnpb.Session{
				DevAddr:                    stored.Session.DevAddr,
				Keys:                       stored.Session.Keys,
				StartedAt:                  timestamppb.New(time.Now()), // NOTE: This is not equivalent to timestamppb.Now().
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
	if err := unwrapSelectedSessionKeys(ctx, ns.KeyService(), dev, req.FieldMask.GetPaths()...); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to unwrap selected keys")
		return nil, err
	}
	return ttnpb.FilterGetEndDevice(dev, req.FieldMask.GetPaths()...)
}

// Delete implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Delete(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*emptypb.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	var evt events.Event
	_, _, err := ns.devices.SetByID(ctx, req.ApplicationIds, req.DeviceId, nil, func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
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

type nsEndDeviceBatchRegistry struct {
	ttnpb.UnimplementedNsEndDeviceBatchRegistryServer

	NS *NetworkServer
}

// Delete implements ttipb.NsEndDeviceBatchRegistryServer.
func (srv *nsEndDeviceBatchRegistry) Delete(
	ctx context.Context,
	req *ttnpb.BatchDeleteEndDevicesRequest,
) (*emptypb.Empty, error) {
	// Check if the user has rights on the application.
	if err := rights.RequireApplication(
		ctx,
		req.ApplicationIds,
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	deleted, err := srv.NS.devices.BatchDelete(ctx, req.ApplicationIds, req.DeviceIds)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to delete device from registry")
		return nil, err
	}

	if len(deleted) != 0 {
		events.Publish(
			evtBatchDeleteEndDevices.NewWithIdentifiersAndData(
				ctx, req.ApplicationIds, &ttnpb.EndDeviceIdentifiersList{
					EndDeviceIds: deleted,
				},
			),
		)
	}

	return ttnpb.Empty, nil
}

func init() {
	// The legacy and modern ADR fields should be mutually exclusive.
	// As such, specifying one of the fields means that every other field of the opposite
	// type should be zero.
	for _, field := range adrSettingsFields {
		ifNotZeroThenZeroFields[field] = append(ifNotZeroThenZeroFields[field], legacyADRSettingsFields...)
	}
	for _, field := range legacyADRSettingsFields {
		ifNotZeroThenZeroFields[field] = append(ifNotZeroThenNotZeroFields[field], "mac_settings.adr")
	}
}
