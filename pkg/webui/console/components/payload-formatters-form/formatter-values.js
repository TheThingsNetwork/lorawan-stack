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

const DEFAULT_UPLINK_JS_FORMATTER = `function Decoder(bytes, f_port) {
  return {
    raw: bytes,
    f_port: f_port
  };
}`
const DEFAULT_DOWNLINK_JS_FORMATTER = `function Encoder(payload, f_port) {
  return [];
}`

export const getDefaultJavascriptFormatter = uplink => uplink
  ? DEFAULT_UPLINK_JS_FORMATTER
  : DEFAULT_DOWNLINK_JS_FORMATTER

export const getDefaultGrpcServiceFormatter = () => undefined
