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
import { Container, Col, Row } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import DeviceOnboardingForm from '@console/containers/device-onboarding-form'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { listBrands } from '@console/store/actions/device-repository'

const DeviceAdd = props => {
  const { appId } = props
  return (
    <RequireRequest requestAction={listBrands(appId, {}, ['name', 'lora_alliance_vendor_id'])}>
      <Container>
        <Row>
          <Col>
            <PageTitle tall title={sharedMessages.registerEndDevice} className="mb-cs-m" />
            <DeviceOnboardingForm />
          </Col>
        </Row>
      </Container>
    </RequireRequest>
  )
}

DeviceAdd.propTypes = {
  appId: PropTypes.string.isRequired,
}

export default DeviceAdd
