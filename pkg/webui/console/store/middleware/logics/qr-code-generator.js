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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as QRCodeGenerator from '@console/store/actions/qr-code-generator'

const parseEndDeviceQRCodeLogic = createRequestLogic({
  type: QRCodeGenerator.PARSE_END_DEVICE_QR_CODE,
  process: async ({ action }) => {
    const { qrCode } = action.payload
    return await tts.QRCodeGenerator.parseEndDeviceQrCode(qrCode)
  },
})

const parseGatewayQRCodeLogic = createRequestLogic({
  type: QRCodeGenerator.PARSE_GATEWAY_QR_CODE,
  process: async ({ action }) => {
    const { qrCode } = action.payload
    return await tts.QRCodeGenerator.parseGatewayQrCode(qrCode)
  },
})

export default [parseEndDeviceQRCodeLogic, parseGatewayQRCodeLogic]
