// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpcgateway

import (
	"github.com/TheThingsNetwork/ttn/pkg/rpcgateway/internal/jsonpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// New returns a new gRPC HTTP Gateway
func New() *runtime.ServeMux {
	mux := runtime.NewServeMux(runtime.WithMarshalerOption("*", &jsonpb.GoGoJSONPb{
		OrigName: true,
	}))
	return mux
}
