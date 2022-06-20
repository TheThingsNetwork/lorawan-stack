// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import PropTypes from '@ttn-lw/lib/prop-types'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import style from '../repository.styl'
import { useRepositoryContext } from '../context'
import { isOtherOption } from '../../utils'
import {
  selectBrand,
  selectModel,
  selectHwVersion,
  selectFwVersion,
  hasSelectedBrand,
  hasSelectedModel,
  hasSelectedHwVersion,
  hasSelectedFwVersion,
} from '../reducer'

import BrandSelect from './brand-select'
import ModelSelect from './model-select'
import HardwareVersionSelect from './hw-version-select'
import FirmwareVersionSelect from './fw-version-select'
import BandSelect from './band-select'

const Selection = props => {
  const { onBrandChange, onModelChange, onHwVersionChange, onFwVersionChange, onBandChange } = props
  const state = useRepositoryContext()

  const brand = selectBrand(state)
  const model = selectModel(state)
  const hardwareVersion = selectHwVersion(state)
  const firmwareVersion = selectFwVersion(state)

  const hasBrand = hasSelectedBrand(state) && !isOtherOption(brand)
  const hasModel = hasSelectedModel(state) && !isOtherOption(model)
  const hasHwVersion = hasSelectedHwVersion(state) && !isOtherOption(hardwareVersion)
  const hasFwVersion = hasSelectedFwVersion(state) && !isOtherOption(firmwareVersion)

  return (
    <div className={style.configurationSection}>
      <BrandSelect
        className={classnames(style.select, style.selectS)}
        name="version_ids.brand_id"
        required
        onChange={onBrandChange}
        tooltipId={tooltipIds.DEVICE_BRAND}
      />
      {hasBrand && (
        <ModelSelect
          className={classnames(style.select, style.selectS)}
          name="version_ids.model_id"
          required
          brandId={brand}
          onChange={onModelChange}
          tooltipId={tooltipIds.DEVICE_MODEL}
        />
      )}
      {hasModel && (
        <HardwareVersionSelect
          className={classnames(style.select, style.selectXs)}
          required
          brandId={brand}
          modelId={model}
          name="version_ids.hardware_version"
          onChange={onHwVersionChange}
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
          onChange={onFwVersionChange}
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
          onChange={onBandChange}
        />
      )}
    </div>
  )
}

Selection.propTypes = {
  onBandChange: PropTypes.func.isRequired,
  onBrandChange: PropTypes.func.isRequired,
  onFwVersionChange: PropTypes.func.isRequired,
  onHwVersionChange: PropTypes.func.isRequired,
  onModelChange: PropTypes.func.isRequired,
}

export default Selection
