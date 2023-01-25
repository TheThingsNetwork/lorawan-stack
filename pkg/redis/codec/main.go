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

// Package codec provides a codec, which encodes and decodes protocol buffers
// stored in Redis to/from JSON.
package main

import (
	"flag"
	"io"
	"log"
	"os"

	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/proto"
)

var json = jsonpb.TTN()

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
}

func main() {
	typ := flag.String("type", "", "Proto type to be used")
	encode := flag.Bool("encode", false, "Whether encoding should be performed (default: false)")

	flag.Parse()

	var pb proto.Message
	switch *typ {
	case "":
		log.Print("Type cannot be empty")
		flag.Usage()
		os.Exit(1)

	case "ttnpb.EndDevice":
		pb = &ttnpb.EndDevice{}
	case "ttnpb.GatewayConnectionStats":
		pb = &ttnpb.GatewayConnectionStats{}
	case "ttnpb.ApplicationActivationSettings":
		pb = &ttnpb.ApplicationActivationSettings{}
	case "ttnpb.UplinkToken":
		pb = &ttnpb.UplinkToken{}
	default:
		log.Printf("Unknown type: `%s`", *typ)
		flag.Usage()
		os.Exit(1)
	}
	if *encode {
		if err := json.NewDecoder(os.Stdin).Decode(pb); err != nil {
			log.Fatalf("Failed to read proto as JSON from stdin: %v", err)
		}
		s, err := redis.MarshalProto(pb)
		if err != nil {
			log.Fatalf("Failed to marshal proto: %v", err)
		}
		if _, err := os.Stdout.Write([]byte(s)); err != nil {
			log.Fatalf("Failed to write to stdout: %v", err)
		}
		return
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read from stdin: %v", err)
	}
	if err := redis.UnmarshalProto(string(b), pb); err != nil {
		log.Fatalf("Failed to unmarshal proto: %v", err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(pb); err != nil {
		log.Fatalf("Failed to write proto as JSON to stdout: %v", err)
	}
}
