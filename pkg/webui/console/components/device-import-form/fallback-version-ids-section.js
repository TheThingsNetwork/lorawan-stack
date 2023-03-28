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

import React, { useCallback, useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Col, Row } from 'react-grid-system'

import { useFormContext } from '@ttn-lw/components/form'

import OtherHint from '@console/containers/device-profile-section/hints/other-hint'
import VersionIdsSection from '@console/containers/device-profile-section'
import FreqPlansSelect from '@console/containers/device-freq-plans-select'
import {
  hasCompletedDeviceRepositorySelection,
  hasValidDeviceRepositoryType,
} from '@console/containers/device-onboarding-form/utils'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import { hasSelectedDeviceRepositoryOther } from '@console/lib/device-utils'

import { getTemplate } from '@console/store/actions/device-repository'

import { selectDeviceTemplate } from '@console/store/selectors/device-repository'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'

const FallbackVersionIdsSection = () => {
  const { values } = useFormContext()
  const { version_ids } = values
  const version = version_ids
  const template = useSelector(selectDeviceTemplate)
  const hasSelectedOther = hasSelectedDeviceRepositoryOther(version)
  const showOtherHint = hasSelectedOther
  const hasValidType = hasValidDeviceRepositoryType(version_ids, template)
  const hasCompleted = hasCompletedDeviceRepositorySelection(version_ids)
  const dispatch = useDispatch()
  const getRegistrationTemplate = useCallback(
    (appId, version) => dispatch(attachPromise(getTemplate(appId, version))),
    [dispatch],
  )
  const appId = useSelector(selectSelectedApplicationId)

  useEffect(() => {
    if (hasCompleted && !hasSelectedOther) {
      getRegistrationTemplate(appId, {
        brand_id: version_ids.brand_id,
        model_id: version_ids.model_id,
        firmware_version: version_ids.firmware_version,
        band_id: version_ids.band_id,
      })
    }
  }, [
    version_ids,
    appId,
    version_ids.band_id,
    version_ids.brand_id,
    version_ids.firmware_version,
    version_ids.model,
    getRegistrationTemplate,
    hasCompleted,
    hasSelectedOther,
  ])

  return (
    <Row>
      <Col>
        <VersionIdsSection />
        {showOtherHint && <OtherHint manualGuideDocsPath="/devices/adding-devices/" />}
        {hasValidType && (
          <FreqPlansSelect
            required
            className="mt-ls-xxs"
            tooltipId={tooltipIds.FREQUENCY_PLAN}
            name="frequency_plan_id"
            bandId={version_ids.band_id}
            autoFocus
          />
        )}
      </Col>
    </Row>
  )
}

export default FallbackVersionIdsSection
