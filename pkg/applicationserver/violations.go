// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
	"google.golang.org/protobuf/types/known/structpb"
)

// ValueType is the type of value violation (NaN, Infinity, -Infinity).
type ValueType string

func (vt ValueType) String() string {
	return string(vt)
}

const (
	// ValueTypeNaN is a NaN violation.
	ValueTypeNaN ValueType = "NaN"

	// ValueTypePosInf is a positive Infinity violation.
	ValueTypePosInf ValueType = "Infinity"

	// ValueTypeNegInf is a negative Infinity violation.
	ValueTypeNegInf ValueType = "-Infinity"
)

// ValueContext is the context in which the value is located (struct, list).
type ValueContext string

func (vt ValueContext) String() string {
	return string(vt)
}

const (
	// ValueContextStruct indicates a violation in a struct.
	ValueContextStruct ValueContext = "struct"

	// ValueContextList indicates a violation in a list.
	ValueContextList ValueContext = "list"
)

// ValueViolation is a violation of a value.
type ValueViolation struct {
	Type    ValueType
	Context ValueContext
}

func findViolation(s string, valueContext ValueContext) *ValueViolation {
	violation := ValueType(s)
	switch violation {
	case ValueTypeNaN:
		return &ValueViolation{
			Type:    ValueTypeNaN,
			Context: valueContext,
		}
	case ValueTypePosInf:
		return &ValueViolation{
			Type:    ValueTypePosInf,
			Context: valueContext,
		}
	case ValueTypeNegInf:
		return &ValueViolation{
			Type:    ValueTypeNegInf,
			Context: valueContext,
		}
	}
	return nil
}

func findListViolations(l *structpb.ListValue) []ValueViolation {
	if l == nil {
		return nil
	}
	total := make([]ValueViolation, 0)
	for _, v := range l.Values {
		switch vv := v.GetKind().(type) {
		case *structpb.Value_StringValue:
			if vv == nil {
				break
			}
			violation := findViolation(vv.StringValue, ValueContextList)
			if violation != nil {
				total = append(total, *violation)
			}
		case *structpb.Value_StructValue:
			if vv == nil {
				break
			}
			total = append(total, findStructViolations(vv.StructValue)...)
		case *structpb.Value_ListValue:
			if vv == nil {
				break
			}
			total = append(total, findListViolations(vv.ListValue)...)
		}
	}
	return total
}

func findStructViolations(st *structpb.Struct) []ValueViolation {
	if st == nil {
		return nil
	}
	total := make([]ValueViolation, 0)
	for _, v := range st.Fields {
		switch vv := v.GetKind().(type) {
		case *structpb.Value_StringValue:
			if vv == nil {
				break
			}
			violation := findViolation(vv.StringValue, ValueContextStruct)
			if violation != nil {
				total = append(total, *violation)
			}
		case *structpb.Value_StructValue:
			if vv == nil {
				break
			}
			stViolations := findStructViolations(vv.StructValue)
			if len(stViolations) > 0 {
				total = append(total, stViolations...)
			}
		case *structpb.Value_ListValue:
			if vv == nil {
				break
			}
			listViolations := findListViolations(vv.ListValue)
			if len(listViolations) > 0 {
				total = append(total, listViolations...)
			}
		}
	}
	return total
}

// FindViolations recursively verifies if the struct contains any invalid values (NaN, -Infinity, Infinity)
// and creates ValueViolations for such fields.
func FindViolations(st *structpb.Struct) []ValueViolation {
	return findStructViolations(st)
}
