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
import { useSelector } from 'react-redux'
import { Col, Row } from 'react-grid-system'
import classnames from 'classnames'

import { useFormContext } from '@ttn-lw/components/form'

import PropTypes from '@ttn-lw/lib/prop-types'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import { selectSupportLinkConfig } from '@ttn-lw/lib/selectors/env'

import { selectDeviceTemplate } from '@console/store/selectors/device-repository'

import { isOtherOption } from '../../utils'

import ProgressHint from './hints/progress-hint'
import OtherHint from './hints/other-hint'
import Card from './device-card'
import BrandSelect from './device-selection/brand-select'
import ModelSelect from './device-selection/model-select'
import HardwareVersionSelect from './device-selection/hw-version-select'
import FirmwareVersionSelect from './device-selection/fw-version-select'
import BandSelect from './device-selection/band-select'

import style from './repository.styl'

const initialValues = {
  version_ids: {
    brand_id: undefined,
    model_id: undefined,
    hardware_version: undefined,
    firmware_version: undefined,
    band_id: undefined,
  },
}

const DeviceTypeRepositoryFormSection = props => {
  const { appId, getRegistrationTemplate } = props

  const { values, setValues } = useFormContext()
  const { version_ids } = values

  const version = version_ids
  const brand = version_ids?.brand_id
  const model = version_ids?.model_id
  const hardwareVersion = version_ids?.hardware_version
  const firmwareVersion = version_ids?.firmware_version
  const template = useSelector(selectDeviceTemplate)
  const supportLink = useSelector(selectSupportLinkConfig)

  const hasBrand = Boolean(brand) && !isOtherOption(brand)
  const hasModel = Boolean(model) && !isOtherOption(model)
  const hasHwVersion = Boolean(hardwareVersion) && !isOtherOption(hardwareVersion)
  const hasFwVersion = Boolean(firmwareVersion) && !isOtherOption(firmwareVersion)

  const hasSelectedOther = version && Object.values(version).some(value => isOtherOption(value))
  const hasCompleted = version && Object.values(version).every(value => value)
  const showProgressHint = !hasSelectedOther && !hasCompleted
  const showDeviceCard = !hasSelectedOther && hasCompleted && Boolean(template)
  const showOtherHint = hasSelectedOther

  const handleBrandChange = React.useCallback(
    value => {
      if (Boolean(model) || Boolean(hardwareVersion) || Boolean(firmwareVersion)) {
        setValues({
          ...values,
          version_ids: {
            brand_id: value,
            model_id: undefined,
            hardware_version: undefined,
            firmware_version: undefined,
            band_id: undefined,
          },
        })
      }
    },
    [setValues, values, firmwareVersion, hardwareVersion, model],
  )

  const handleModelChange = React.useCallback(
    value => {
      if (Boolean(hardwareVersion) || Boolean(firmwareVersion)) {
        setValues({
          ...values,
          version_ids: {
            ...values.version_ids,
            model_id: value,
            hardware_version: undefined,
            firmware_version: undefined,
            band_id: undefined,
          },
        })
      }
    },
    [setValues, values, hardwareVersion, firmwareVersion],
  )

  React.useEffect(() => {
    // Fetch template after completing the selection step (select band, model, hw/fw versions and band).
    if (values && hasCompleted && !hasSelectedOther) {
      const {
        version_ids: { hardware_version, ...v },
      } = values

      getRegistrationTemplate(appId, v)
    }
  }, [appId, getRegistrationTemplate, hasCompleted, hasSelectedOther, values])

  return (
    <Row>
      <Col>
        <div className={style.configurationSection}>
          <BrandSelect
            className={classnames(style.select, style.selectS)}
            name="version_ids.brand_id"
            required
            onChange={handleBrandChange}
            tooltipId={tooltipIds.DEVICE_BRAND}
          />
          {hasBrand && (
            <ModelSelect
              className={classnames(style.select, style.selectS)}
              name="version_ids.model_id"
              required
              brandId={brand}
              onChange={handleModelChange}
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
        {showProgressHint && <ProgressHint supportLink={supportLink} />}
        {showOtherHint && <OtherHint manualGuideDocsPath="/devices/adding-devices/" />}
        {showDeviceCard && <Card brandId={brand} modelId={model} template={template} />}
      </Col>
    </Row>
  )
}

DeviceTypeRepositoryFormSection.propTypes = {
  appId: PropTypes.string.isRequired,
  getRegistrationTemplate: PropTypes.func.isRequired,
}

export { DeviceTypeRepositoryFormSection as default, initialValues }
