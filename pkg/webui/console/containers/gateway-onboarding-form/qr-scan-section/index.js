// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useState } from 'react'
import { useDispatch } from 'react-redux'
import { defineMessages } from 'react-intl'

import QRModalButton from '@ttn-lw/components/qr-modal-button'
import { useFormContext } from '@ttn-lw/components/form'
import Icon, { IconCheck, IconX } from '@ttn-lw/components/icon'
import ModalButton from '@ttn-lw/components/button/modal-button'
import ButtonGroup from '@ttn-lw/components/button/group'

import Message from '@ttn-lw/lib/components/message'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'

// TODO: Get the correct sdk (gateway qr code generator)
import { parseGatewayQRCode } from '@console/store/actions/qr-code-generator'

const m = defineMessages({
  hasGatewayQR:
    'Does your gateway have a LoRaWAN® Gateway Identification QR Code? Scan it to speed up onboarding.',
  gatewayGuide: 'Gateway registration help',
  invalidQRCode:
    'Invalid QR code data. Please note that only TTIGPRO1 Gateway Identification QR Code can be scanned. Some gateways have unrelated QR codes printed on them that cannot be used.',
})

const qrDataInitialState = {
  valid: false,
  approved: false,
  data: undefined,
  gateway: undefined,
}

const GatewayQRScanSection = () => {
  const dispatch = useDispatch()
  const { resetForm, setValues } = useFormContext()
  const [qrData, setQrData] = useState(qrDataInitialState)

  const handleReset = useCallback(() => {
    resetForm()
    setQrData(qrDataInitialState)
  }, [resetForm])

  const handleQRCodeApprove = useCallback(() => {
    const { gateway } = qrData

    setValues(values => ({
      ...values,
      _withQRdata: true,
      ids: {
        ...values.ids,
        eui: gateway.gateway_eui,
      },
      authenticated_identifiers: {
        gateway_eui: gateway.gateway_eui,
        authentication_code: gateway.owner_token ? btoa(gateway.owner_token) : '',
      },
    }))

    setQrData({ ...qrData, approved: true })
  }, [qrData, setValues])

  const handleQRCodeCancel = useCallback(() => {
    setQrData(qrDataInitialState)
  }, [])

  const handleQRCodeRead = useCallback(
    async qrCode => {
      try {
        // Get gateway from QR code
        const gateway = await dispatch(attachPromise(parseGatewayQRCode(qrCode)))

        const sheetData = [
          {
            header: sharedMessages.qrCodeData,
            items: [
              {
                key: sharedMessages.ownerToken,
                value: gateway.owner_token,
                type: 'code',
                sensitive: true,
              },
              {
                key: sharedMessages.gatewayEUI,
                value: gateway.gateway_eui,
                type: 'byte',
                sensitive: false,
              },
            ],
          },
        ]
        setQrData({
          ...qrData,
          valid: true,
          data: sheetData,
          gateway,
        })
      } catch (error) {
        setQrData({ ...qrData, data: [], valid: false })
      }
    },
    [dispatch, qrData],
  )

  return (
    <>
      {qrData.approved ? (
        <div className="mb-cs-xs d-flex al-center">
          <Icon icon={IconCheck} textPaddedRight className="c-text-success-normal" />
          <Message content={sharedMessages.scanSuccess} />
        </div>
      ) : (
        <div className="mb-cs-xs">
          <Message content={m.hasGatewayQR} />
        </div>
      )}
      <ButtonGroup>
        {qrData.approved ? (
          <ModalButton
            type="button"
            icon="close"
            onApprove={handleReset}
            message={sharedMessages.qrCodeDataReset}
            modalData={{
              title: sharedMessages.qrCodeDataReset,
              noTitleLine: true,
              buttonMessage: sharedMessages.qrCodeDataReset,
              children: <Message content={sharedMessages.resetConfirm} component="span" />,
              approveButtonProps: {
                icon: IconX,
              },
            }}
          />
        ) : (
          <QRModalButton
            message={sharedMessages.scanGatewayQR}
            invalidMessage={m.invalidQRCode}
            onApprove={handleQRCodeApprove}
            onCancel={handleQRCodeCancel}
            onRead={handleQRCodeRead}
            qrData={qrData}
          />
        )}
      </ButtonGroup>
      <hr className="mt-cs-l mb-cs-xl" />
    </>
  )
}

export default GatewayQRScanSection
