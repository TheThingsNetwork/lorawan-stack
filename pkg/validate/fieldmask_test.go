// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package validate

import (
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestFieldMaskPaths(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name          string
		FieldMask     *types.FieldMask
		ExpectedError func(error) bool
	}{
		{
			Name: "ValidStrings",
			FieldMask: &types.FieldMask{
				Paths: []string{"uplink_messages", "gateway_status.time", "gateway_status.boot_time", "gateway_status.versions", "gateway_status.antenna_locations", "gateway_status.ip", "gateway_status.metrics", "gateway_status.advanced", "tx_acknowledgment.correlation_ids", "tx_acknowledgment.result"},
			},
			ExpectedError: nil,
		},
		{
			Name: "ValidMaxLength",
			FieldMask: &types.FieldMask{
				Paths: []string{"ozaj8qs0sait7oudxqbfyx6b14yuahcfrdlbdj0zxsjxfddmnvg.8gc5ppbvdy54bxh.angomzl2t9rd9lzq_s4yaa0se4sbkiqf44uede6l34cslwkb73mcci1dfo9cvrjo9m6trim7avvkmyxzhzs7b4k0bg_6ohp97.qlir.q3ib.0.3kv2sar4tple.z83.w8v5hb9sm3d.h8uy.4cd_j0qtk2oz8qp.9ocdffggyseq87jik3lovpu2emde"},
			},
			ExpectedError: nil,
		},
		{
			Name: "ValidNumerics",
			FieldMask: &types.FieldMask{
				Paths: []string{"1id1"},
			},
			ExpectedError: nil,
		},
		{
			Name: "ValidEmptyString",
			FieldMask: &types.FieldMask{
				Paths: []string{},
			},
			ExpectedError: nil,
		},
		{
			Name: "InvalidUnderscore",
			FieldMask: &types.FieldMask{
				Paths: []string{"_id"},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
		{
			Name: "InvalidTrailingUnderscore",
			FieldMask: &types.FieldMask{
				Paths: []string{"id_"},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
		{
			Name: "InvalidDot",
			FieldMask: &types.FieldMask{
				Paths: []string{".id"},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
		{
			Name: "InvalidTrailingDot",
			FieldMask: &types.FieldMask{
				Paths: []string{"id."},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
		{
			Name: "InvalidDashes",
			FieldMask: &types.FieldMask{
				Paths: []string{"-id-"},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
		{
			Name: "InvalidSpecialCharacter",
			FieldMask: &types.FieldMask{
				Paths: []string{"%id"},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
		{
			Name: "InvalidSpecialCharacters1",
			FieldMask: &types.FieldMask{
				Paths: []string{"-._"},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
		{
			Name: "InvalidDash",
			FieldMask: &types.FieldMask{
				Paths: []string{"gateway-id"},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
		{
			Name: "InvalidLength",
			FieldMask: &types.FieldMask{
				Paths: []string{"a"},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
		{
			Name: "InValidMaxLength",
			FieldMask: &types.FieldMask{
				Paths: []string{"ozaj8qs0sait7oudxqbfyx6b14yuahcfrdlbdj0zxsjxfddmnvg.8gc5ppbvdy54bxh.angomzl2t9rd9lzq_s4yaa0se4sbkiqf44uede6l34cslwkb73mcci1dfo9cvrjo9m6trim7avvkmyxzhzs7b4k0bg_6ohp97.qlir.q3ib.0.3kv2sar4tple.z83.w8v5hb9sm3d.h8uy.4cd_j0qtk2oz8qp.9ocdffggyseq87ik3lovpu2emde61"},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
		{
			Name: "InvalidSpecialCharacters",
			FieldMask: &types.FieldMask{
				Paths: []string{"a%!@#$%"},
			},
			ExpectedError: errors.IsInvalidArgument,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			err := FieldMaskPaths(tc.FieldMask)
			if tc.ExpectedError == nil {
				a.So(err, should.BeNil)
			} else {
				a.So(tc.ExpectedError(err), should.BeTrue)
			}
		})
	}
}
