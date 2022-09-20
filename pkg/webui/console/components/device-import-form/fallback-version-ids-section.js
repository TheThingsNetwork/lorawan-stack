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

import React from 'react'
import { Col, Row } from 'react-grid-system'
import classnames from 'classnames'

import { useFormContext } from '@ttn-lw/components/form'

import {
  hasSelectedDeviceRepositoryOther,
  isOtherOption,
} from '@console/containers/device-onboarding-form/utils'
import OtherHint from '@console/containers/device-onboarding-form/type-form-section/repository-form-section/hints/other-hint'
import BrandSelect from '@console/containers/device-onboarding-form/type-form-section/repository-form-section/device-selection/brand-select'
import ModelSelect from '@console/containers/device-onboarding-form/type-form-section/repository-form-section/device-selection/model-select'
import HardwareVersionSelect from '@console/containers/device-onboarding-form/type-form-section/repository-form-section/device-selection/hw-version-select'
import FirmwareVersionSelect from '@console/containers/device-onboarding-form/type-form-section/repository-form-section/device-selection/fw-version-select'
import BandSelect from '@console/containers/device-onboarding-form/type-form-section/repository-form-section/device-selection/band-select'
import style from '@console/containers/device-onboarding-form/type-form-section/repository-form-section/repository.styl'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

const brandValueSetter = ({ setValues }, { value }) =>
  setValues(values => ({
    ...values,
    version_ids: {
      ...initialValues.version_ids,
      brand_id: value,
    },
  }))
const modelValueSetter = ({ setValues }, { value }) =>
  setValues(values => ({
    ...values,
    version_ids: {
      ...initialValues.version_ids,
      brand_id: values.version_ids.brand_id,
      model_id: value,
    },
  }))

const initialValues = {
  version_ids: {
    brand_id: '',
    model_id: '',
    hardware_version: '',
    firmware_version: '',
    band_id: '',
  },
}

const FallbackVersionIdsSection = () => {
  const { values } = useFormContext()
  const { version_ids } = values

  const version = version_ids
  const brand = version_ids?.brand_id
  const model = version_ids?.model_id
  const hardwareVersion = version_ids?.hardware_version
  const firmwareVersion = version_ids?.firmware_version

  const hasBrand = Boolean(brand) && !isOtherOption(brand)
  const hasModel = Boolean(model) && !isOtherOption(model)
  const hasHwVersion = Boolean(hardwareVersion) && !isOtherOption(hardwareVersion)
  const hasFwVersion = Boolean(firmwareVersion) && !isOtherOption(firmwareVersion)

  const hasSelectedOther = hasSelectedDeviceRepositoryOther(version)
  const showOtherHint = hasSelectedOther

  return (
    <Row>
      <Col>
        <div className={style.configurationSection}>
          <BrandSelect
            className={classnames(style.select, style.selectS)}
            name="version_ids.brand_id"
            required
            tooltipId={tooltipIds.DEVICE_BRAND}
            valueSetter={brandValueSetter}
          />
          {hasBrand && (
            <ModelSelect
              className={classnames(style.select, style.selectS)}
              name="version_ids.model_id"
              required
              brandId={brand}
              tooltipId={tooltipIds.DEVICE_MODEL}
              valueSetter={modelValueSetter}
            />
          )}
          {hasModel && (
            <HardwareVersionSelect
              className={classnames(style.select, style.selectXs)}
              required
              brandId={brand}
              modelId={model}
              name="version_ids.hardware_version"
              tooltipId={tooltipIds.DEVICE_HARDWARE_VERSION}
            />
          )}
          {hasHwVersion && (
            <FirmwareVersionSelect
              className={classnames(style.select, style.selectXs)}
              required
              name="version_ids.firmware_version"
              brandId={brand}
              modelId={model}
              hwVersion={hardwareVersion}
              tooltipId={tooltipIds.DEVICE_FIRMWARE_VERSION}
            />
          )}
          {hasFwVersion && (
            <BandSelect
              className={classnames(style.select, style.selectS)}
              required
              name="version_ids.band_id"
              fwVersion={firmwareVersion}
              brandId={brand}
              modelId={model}
            />
          )}
        </div>
        {showOtherHint && <OtherHint manualGuideDocsPath="/devices/adding-devices/" />}
      </Col>
    </Row>
  )
}

export default FallbackVersionIdsSection
