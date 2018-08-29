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

//+build ignore

package main

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/vanity"
	"github.com/gogo/protobuf/vanity/command"
)

func main() {
	req := command.Read()
	files := req.GetProtoFile()
	files = vanity.FilterFiles(files, vanity.NotGoogleProtobufDescriptorProto)

	for _, opt := range []func(*descriptor.FileDescriptorProto){
		vanity.TurnOffGoEnumStringerAll,
		vanity.TurnOffGoStringerAll,
		vanity.TurnOffGoUnrecognizedAll,
		vanity.TurnOnEnumStringerAll,
		vanity.TurnOnEqualAll,
		vanity.TurnOnGoRegistration,
		vanity.TurnOnMarshalerAll,
		vanity.TurnOnPopulateAll,
		vanity.TurnOnSizerAll,
		vanity.TurnOnStringerAll,
		// vanity.TurnOnTestGenAll,
		vanity.TurnOnUnmarshalerAll,
		vanity.TurnOnVerboseEqualAll,
	} {
		vanity.ForEachFile(files, opt)
	}
	command.Write(command.Generate(req))
}
