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

import { useFormContext } from '@ttn-lw/components/form'

import OtherHint from '@console/containers/device-profile-section/hints/other-hint'
import VersionIdsSection from '@console/containers/device-profile-section'

import { hasSelectedDeviceRepositoryOther } from '@console/lib/device-utils'

const FallbackVersionIdsSection = () => {
  const { values } = useFormContext()
  const { version_ids } = values
  const version = version_ids
  const hasSelectedOther = hasSelectedDeviceRepositoryOther(version)
  const showOtherHint = hasSelectedOther

  return (
    <Row>
      <Col>
        <VersionIdsSection />
        {showOtherHint && <OtherHint manualGuideDocsPath="/devices/adding-devices/" />}
      </Col>
    </Row>
  )
}

export default FallbackVersionIdsSection
