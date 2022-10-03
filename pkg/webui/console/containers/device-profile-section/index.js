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
import classnames from 'classnames'

import { useFormContext } from '@ttn-lw/components/form'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import PropTypes from '@ttn-lw/lib/prop-types'

import { isOtherOption } from '@console/lib/device-utils'

import BrandSelect from './device-selection/brand-select'
import ModelSelect from './device-selection/model-select'
import HardwareVersionSelect from './device-selection/hw-version-select'
import FirmwareVersionSelect from './device-selection/fw-version-select'
import BandSelect from './device-selection/band-select'

import style from './repository.styl'

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

const VersionIdsSection = ({ isImport }) => {
  const { values } = useFormContext()
  const { version_ids } = values
  const brand = version_ids?.brand_id
  const model = version_ids?.model_id
  const hardwareVersion = version_ids?.hardware_version
  const firmwareVersion = version_ids?.firmware_version

  const hasBrand = Boolean(brand) && !isOtherOption(brand)
  const hasModel = Boolean(model) && !isOtherOption(model)
  const hasHwVersion = Boolean(hardwareVersion) && !isOtherOption(hardwareVersion)
  const hasFwVersion = Boolean(firmwareVersion) && !isOtherOption(firmwareVersion)

  return (
    <>
      <div className={style.configurationSection}>
        <BrandSelect
          className={classnames(style.select, style.selectS)}
          name="version_ids.brand_id"
          required={!isImport}
          tooltipId={tooltipIds.DEVICE_BRAND}
          valueSetter={brandValueSetter}
        />
        {hasBrand && (
          <ModelSelect
            className={classnames(style.select, style.selectS)}
            name="version_ids.model_id"
            required={!isImport}
            brandId={brand}
            tooltipId={tooltipIds.DEVICE_MODEL}
            valueSetter={modelValueSetter}
          />
        )}
        {hasModel && (
          <HardwareVersionSelect
            className={classnames(style.select, style.selectXs)}
            required={!isImport}
            brandId={brand}
            modelId={model}
            name="version_ids.hardware_version"
            tooltipId={tooltipIds.DEVICE_HARDWARE_VERSION}
          />
        )}
        {hasHwVersion && (
          <FirmwareVersionSelect
            className={classnames(style.select, style.selectXs)}
            required={!isImport}
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
            required={!isImport}
            name="version_ids.band_id"
            fwVersion={firmwareVersion}
            brandId={brand}
            modelId={model}
          />
        )}
      </div>
    </>
  )
}

VersionIdsSection.propTypes = {
  isImport: PropTypes.bool,
}

VersionIdsSection.defaultProps = {
  isImport: false,
}

export { VersionIdsSection as default, initialValues }
