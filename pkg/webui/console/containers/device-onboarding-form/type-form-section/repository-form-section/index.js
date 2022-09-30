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

import React, { useCallback } from 'react'
import { useSelector, useDispatch } from 'react-redux'
import { Col, Row } from 'react-grid-system'
import { useFormikContext } from 'formik'
import { get, set } from 'lodash'

import { useFormContext } from '@ttn-lw/components/form'

import FreqPlansSelect from '@console/containers/device-freq-plans-select'
import VersionIdsSection, { initialValues } from '@console/containers/device-profile-section'
import ProgressHint from '@console/containers/device-profile-section/hints/progress-hint'
import OtherHint from '@console/containers/device-profile-section/hints/other-hint'
import Card from '@console/containers/device-profile-section/device-card'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import { selectSupportLinkConfig } from '@ttn-lw/lib/selectors/env'

import { hasSelectedDeviceRepositoryOther } from '@console/lib/device-utils'

import { getTemplate } from '@console/store/actions/device-repository'

import { selectDeviceTemplate } from '@console/store/selectors/device-repository'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import { hasCompletedDeviceRepositorySelection, hasValidDeviceRepositoryType } from '../../utils'

const DeviceTypeRepositoryFormSection = () => {
  const appId = useSelector(selectSelectedApplicationId)
  const dispatch = useDispatch()
  const getRegistrationTemplate = useCallback(
    (appId, version) => dispatch(attachPromise(getTemplate(appId, version))),
    [dispatch],
  )

  const { addToFieldRegistry, removeFromFieldRegistry } = useFormikContext()
  const { values, setValues } = useFormContext()
  const { version_ids } = values

  const version = version_ids
  const brand = version_ids?.brand_id
  const model = version_ids?.model_id
  const firmwareVersion = version_ids?.firmware_version
  const band = version_ids?.band_id
  const template = useSelector(selectDeviceTemplate)
  const supportLink = useSelector(selectSupportLinkConfig)

  const hasSelectedOther = hasSelectedDeviceRepositoryOther(version)
  const hasCompleted = hasCompletedDeviceRepositorySelection(version)
  const hasValidType = hasValidDeviceRepositoryType(version, template)
  const showProgressHint = !hasSelectedOther && !hasCompleted
  const showDeviceCard = hasValidType
  const showFrequencyPlanSelector = hasValidType
  const showOtherHint = hasSelectedOther

  // Apply template once it is fetched and register the template fields so they don't get cleaned.
  React.useEffect(() => {
    if (template && hasCompleted) {
      // Since the template response will strip zero values, we cannot simply spread the result
      // over the existing form values. Instead we need to make all zero values explicit
      // by assigning them as `undefined`, using the provided field mask.
      const templateFields = template.field_mask.split(',')
      const endDeviceFill = templateFields.reduce(
        (device, path) => set(device, path, get(template.end_device, path)),
        {},
      )

      setValues(values => ({
        ...values,
        ...endDeviceFill,
        version_ids: values.version_ids,
      }))

      const hiddenFields = [
        ...templateFields,
        'network_server_address',
        'application_server_address',
        'join_server_address',
      ]

      addToFieldRegistry(...hiddenFields)
      return () => removeFromFieldRegistry(...hiddenFields)
    }
  }, [hasCompleted, setValues, template, addToFieldRegistry, removeFromFieldRegistry])

  // Fetch template after completing the selection step (select band, model, hw/fw versions and band).
  React.useEffect(() => {
    if (hasCompleted && !hasSelectedOther && values._isClaiming === undefined) {
      getRegistrationTemplate(appId, {
        brand_id: brand,
        model_id: model,
        firmware_version: firmwareVersion,
        band_id: band,
      })
    }
  }, [
    appId,
    band,
    brand,
    firmwareVersion,
    getRegistrationTemplate,
    hasCompleted,
    hasSelectedOther,
    model,
    values._isClaiming,
  ])

  return (
    <Row>
      <Col>
        <VersionIdsSection />
        {showProgressHint && <ProgressHint supportLink={supportLink} />}
        {showOtherHint && <OtherHint manualGuideDocsPath="/devices/adding-devices/" />}
        {showDeviceCard && <Card brandId={brand} modelId={model} template={template} />}
        {showFrequencyPlanSelector && (
          <FreqPlansSelect
            required
            className="mt-ls-xxs"
            tooltipId={tooltipIds.FREQUENCY_PLAN}
            name="frequency_plan_id"
            bandId={band}
            autoFocus
          />
        )}
        {hasCompleted && <hr />}
      </Col>
    </Row>
  )
}

export { DeviceTypeRepositoryFormSection as default, initialValues }
