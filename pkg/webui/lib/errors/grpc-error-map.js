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

/* eslint-disable quote-props */

// Source: https://github.com/grpc/grpc/blob/master/doc/statuscodes.md

const errorMap = {
  '0': '200',
  '1': '499',
  '2': '500',
  '3': '400',
  '4': '504',
  '5': '404',
  '6': '409',
  '7': '403',
  '8': '429',
  '9': '400',
  '10': '409',
  '11': '400',
  '12': '501',
  '13': '500',
  '14': '503',
  '15': '500',
  '16': '401',
}

export default function getHttpErrorFromRpcError(rpcError) {
  if (typeof rpcError !== 'string' && typeof rpcError !== 'number') {
    return undefined
  }

  return errorMap[rpcError] || '520' // Fallback to 520 Unknown
}
