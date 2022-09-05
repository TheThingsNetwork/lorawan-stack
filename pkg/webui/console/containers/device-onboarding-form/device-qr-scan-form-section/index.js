// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import QRModalButton from '@ttn-lw/components/qr-modal-button'
import { useFormContext } from '@ttn-lw/components/form'
import Link from '@ttn-lw/components/link'
import Icon from '@ttn-lw/components/icon'
import ModalButton from '@ttn-lw/components/button/modal-button'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { readDeviceQr } from '@console/lib/device-claiming-parse-qr'

import { selectDeviceBrands } from '@console/store/selectors/device-repository'

const hexToDecimal = hex => parseInt(hex, 16)

const m = defineMessages({
  hasEndDeviceQR: 'Does your end device have a QR code? Scan it to speed up onboarding.',
  learnMore: 'Learn more',
  scanEndDevice: 'Scan end device code',
  deviceInfo: 'Found QR code data',
  resetQRCodeData: 'Reset QR code data',
  resetConfirm:
    'Are you sure you want to discard QR code data? The scanned device will not be registered and the form will be reset.',
  scanSuccess: 'QR code scanned successfully',
})

const qrDataInitialState = {
  valid: false,
  approved: false,
  data: undefined,
  device: undefined,
}

const DeviceQRScanFormSection = () => {
  const { resetForm, setValues } = useFormContext()
  const brands = useSelector(selectDeviceBrands)
  const [qrData, setQrData] = useState(qrDataInitialState)

  const handleReset = useCallback(() => {
    resetForm()
    setQrData(qrDataInitialState)
  }, [resetForm])

  const getBrand = useCallback(
    vendorId => {
      const brand = brands.find(brand => brand?.lora_alliance_vendor_id === hexToDecimal(vendorId))

      return brand
    },
    [brands],
  )

  const handleQRCodeApprove = useCallback(() => {
    const { device } = qrData
    const brand = getBrand(device.profileID.vendorID)

    setValues(values => ({
      ...values,
      _withQRdata: true,
      ids: {
        ...values.ids,
        join_eui: device.joinEUI,
        dev_eui: device.devEUI,
      },
      authenticated_identifiers: {
        dev_eui: device.devEUI,
        authentication_code: device.ownerToken ? device.ownerToken : '',
      },
      version_ids: {
        ...values.version_ids,
        brand_id: brand ? brand.brand_id : values.version_ids.brand_id,
      },
    }))

    setQrData({ ...qrData, approved: true })
  }, [getBrand, qrData, setValues])

  const handleQRCodeCancel = useCallback(() => {
    setQrData(qrDataInitialState)
  }, [])

  const handleQRCodeRead = useCallback(
    qrCode => {
      const device = readDeviceQr(qrCode)
      if (device && device.devEUI) {
        const brand = getBrand(device.profileID.vendorID)
        const sheetData = [
          {
            header: m.deviceInfo,
            items: [
              {
                key: sharedMessages.claimAuthCode,
                value: device.ownerToken,
                type: 'code',
                sensitive: true,
              },
              {
                key: sharedMessages.joinEUI,
                value: device.joinEUI,
                type: 'byte',
                sensitive: false,
              },
              { key: sharedMessages.devEUI, value: device.devEUI, type: 'byte', sensitive: false },
              { key: sharedMessages.brand, value: brand?.name },
            ],
          },
        ]
        setQrData({
          ...qrData,
          valid: true,
          data: sheetData,
          device,
        })
      } else {
        setQrData({ ...qrData, data: [], valid: false })
      }
    },
    [getBrand, qrData],
  )

  return (
    <div className="mb-cs-l">
      {qrData.approved ? (
        <>
          <div className="mb-cs-xs">
            <Icon icon="check" textPaddedRight className="c-success" />
            <Message content={m.scanSuccess} />
          </div>
          <ModalButton
            type="button"
            icon="close"
            onApprove={handleReset}
            message={m.resetQRCodeData}
            modalData={{
              title: m.resetQRCodeData,
              buttonMessage: m.resetQRCodeData,
              children: <Message content={m.resetConfirm} component="span" />,
              approveButtonProps: {
                icon: 'close',
              },
            }}
          />
        </>
      ) : (
        <>
          <div className="mb-cs-xs">
            <Message content={m.hasEndDeviceQR} />
          </div>
          <QRModalButton
            message={m.scanEndDevice}
            onApprove={handleQRCodeApprove}
            onCancel={handleQRCodeCancel}
            onRead={handleQRCodeRead}
            qrData={qrData}
          />
        </>
      )}
      <Link.Anchor
        className="ml-cs-xs"
        href="https://www.thethingsindustries.com/docs"
        external
        secondary
      >
        <Message content={m.learnMore} />
      </Link.Anchor>
    </div>
  )
}

export default DeviceQRScanFormSection
