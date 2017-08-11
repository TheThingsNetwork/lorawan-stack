// Code copied from github.com/grpc-ecosystem/grpc-gateway/runtime

package jsonpb

import (
	"reflect"

	. "github.com/grpc-ecosystem/grpc-gateway/runtime"
)

var (
	convFromType = map[reflect.Kind]reflect.Value{
		reflect.String:  reflect.ValueOf(String),
		reflect.Bool:    reflect.ValueOf(Bool),
		reflect.Float64: reflect.ValueOf(Float64),
		reflect.Float32: reflect.ValueOf(Float32),
		reflect.Int64:   reflect.ValueOf(Int64),
		reflect.Int32:   reflect.ValueOf(Int32),
		reflect.Uint64:  reflect.ValueOf(Uint64),
		reflect.Uint32:  reflect.ValueOf(Uint32),
	}
)
