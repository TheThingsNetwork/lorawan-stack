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
import { useDispatch, useSelector } from 'react-redux'

import Icon, { IconCheck, IconX } from '@ttn-lw/components/icon'
import QRModalButton from '@ttn-lw/components/qr-modal-button'
import { useFormContext } from '@ttn-lw/components/form'
import Link from '@ttn-lw/components/link'
import ModalButton from '@ttn-lw/components/button/modal-button'
import ButtonGroup from '@ttn-lw/components/button/group'

import Message from '@ttn-lw/lib/components/message'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { parseEndDeviceQRCode } from '@console/store/actions/qr-code-generator'
import { getTemplateByProfileIds } from '@console/store/actions/device-repository'

import { selectDeviceBrands } from '@console/store/selectors/device-repository'

import m from '../messages'

const qrDataInitialState = {
  valid: false,
  approved: false,
  data: undefined,
  device: undefined,
  versionIds: undefined,
}

const DeviceQRScanFormSection = () => {
  const dispatch = useDispatch()
  const { resetForm, setValues } = useFormContext()
  const brands = useSelector(selectDeviceBrands)
  const [qrData, setQrData] = useState(qrDataInitialState)

  const handleReset = useCallback(() => {
    resetForm()
    setQrData(qrDataInitialState)
  }, [resetForm])

  const getBrand = useCallback(
    (vendorId, brandId) =>
      brands.find(brand =>
        brandId ? brand?.brand_id === brandId : brand?.lora_alliance_vendor_id === vendorId,
      ),
    [brands],
  )

  // Resolve the full end device version identifiers (brand, model, hardware and firmware version
  // and band) from the LoRa Alliance vendor ID and vendor profile ID, if present in the QR code.
  // Not all vendors map vendor profile IDs to models, so failing to resolve is not an error.
  const getVersionIds = useCallback(
    async ({ vendor_id, vendor_profile_id } = {}) => {
      if (!vendor_id || !vendor_profile_id) {
        return undefined
      }
      try {
        const template = await dispatch(
          attachPromise(getTemplateByProfileIds(vendor_id, vendor_profile_id)),
        )
        return template?.end_device?.version_ids
      } catch {
        return undefined
      }
    },
    [dispatch],
  )

  const handleQRCodeApprove = useCallback(() => {
    const { device, versionIds } = qrData
    const { end_device } = device.end_device_template
    const { lora_alliance_profile_ids = {} } = end_device

    const brand = getBrand(lora_alliance_profile_ids.vendor_id, versionIds?.brand_id)

    setValues(values => ({
      ...values,
      _withQRdata: true,
      ids: {
        ...values.ids,
        join_eui: end_device.ids.join_eui,
        dev_eui: end_device.ids.dev_eui,
        device_id: `eui-${end_device.ids.dev_eui.toLowerCase()}`,
      },
      target_device_id: `eui-${end_device.ids.dev_eui.toLowerCase()}`,
      authenticated_identifiers: {
        dev_eui: end_device.ids.dev_eui,
        authentication_code: end_device.claim_authentication_code.value
          ? end_device.claim_authentication_code.value
          : '',
        join_eui: end_device.ids.join_eui,
      },
      version_ids:
        brand && versionIds?.model_id
          ? {
              brand_id: brand.brand_id,
              model_id: versionIds.model_id,
              hardware_version: versionIds.hardware_version || '',
              firmware_version: versionIds.firmware_version || '',
              band_id: versionIds.band_id || '',
            }
          : {
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
    async qrCode => {
      try {
        // Get end device template from QR code
        const device = await dispatch(attachPromise(parseEndDeviceQRCode(qrCode)))

        const { end_device } = device.end_device_template
        const { lora_alliance_profile_ids = {} } = end_device

        const versionIds = await getVersionIds(lora_alliance_profile_ids)
        const brand = getBrand(lora_alliance_profile_ids.vendor_id, versionIds?.brand_id)
        const sheetData = [
          {
            header: sharedMessages.qrCodeData,
            items: [
              {
                key: sharedMessages.claimAuthCode,
                value: end_device.claim_authentication_code.value,
                type: 'code',
                sensitive: true,
              },
              {
                key: sharedMessages.joinEUI,
                value: end_device.ids.join_eui,
                type: 'byte',
                sensitive: false,
              },
              {
                key: sharedMessages.devEUI,
                value: end_device.ids.dev_eui,
                type: 'byte',
                sensitive: false,
              },
              { key: sharedMessages.brand, value: brand?.name },
              ...(brand && versionIds?.model_id
                ? [
                    { key: sharedMessages.model, value: versionIds.model_id },
                    { key: sharedMessages.hardwareVersion, value: versionIds.hardware_version },
                    { key: sharedMessages.firmwareVersion, value: versionIds.firmware_version },
                  ].filter(({ value }) => Boolean(value))
                : []),
            ],
          },
        ]
        setQrData({
          ...qrData,
          valid: true,
          data: sheetData,
          device,
          versionIds,
        })
      } catch (error) {
        setQrData({ ...qrData, data: [], valid: false })
      }
    },
    [dispatch, getBrand, getVersionIds, qrData],
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
          <Message content={m.hasEndDeviceQR} />
        </div>
      )}
      <ButtonGroup>
        {qrData.approved ? (
          <ModalButton
            type="button"
            icon={IconX}
            onApprove={handleReset}
            message={sharedMessages.qrCodeDataReset}
            modalData={{
              title: sharedMessages.qrCodeDataReset,
              noTitleLine: true,
              buttonMessage: sharedMessages.qrCodeDataReset,
              children: <Message content={sharedMessages.resetConfirm} component="span" />,
              approveButtonProps: {
                primary: true,
                icon: IconX,
              },
            }}
          />
        ) : (
          <QRModalButton
            message={sharedMessages.scanEndDevice}
            invalidMessage={m.invalidData}
            onApprove={handleQRCodeApprove}
            onCancel={handleQRCodeCancel}
            onRead={handleQRCodeRead}
            qrData={qrData}
          />
        )}
        <Link.DocLink className="ml-cs-xs" path="/devices/adding-devices" secondary>
          <Message content={m.deviceGuide} />
        </Link.DocLink>
      </ButtonGroup>
      <hr className="mt-cs-m mb-0" />
    </>
  )
}

export default DeviceQRScanFormSection
