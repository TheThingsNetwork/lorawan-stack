// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import Form from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import policyMessages from '@console/lib/packet-broker/messages'

const DefaultRoutingPolicyForm = () => (
  <Row>
    <Col md={6} xs={12}>
      <Message content={sharedMessages.uplink} component="h4" className="mb-cs-xs" />
      <Form.Field
        name="policy.uplink.join_request"
        component={Checkbox}
        label={policyMessages.joinRequest}
        description={policyMessages.joinRequestDesc}
      />
      <Form.Field
        name="policy.uplink.mac_data"
        component={Checkbox}
        label={policyMessages.macData}
        description={policyMessages.macDataDesc}
      />
      <Form.Field
        name="policy.uplink.application_data"
        component={Checkbox}
        label={sharedMessages.appData}
        description={policyMessages.applicationDataDesc}
      />
      <Form.Field
        name="policy.uplink.signal_quality"
        component={Checkbox}
        label={policyMessages.signalQualityInformation}
        description={policyMessages.signalQualityInformationDesc}
      />
      <Form.Field
        name="policy.uplink.localization"
        component={Checkbox}
        label={policyMessages.localizationInformation}
        description={policyMessages.localizationInformationDesc}
      />
    </Col>
    <Col sm={6} xs={12}>
      <Message content={sharedMessages.downlink} component="h4" className="mb-cs-xs" />
      <Form.Field
        name="policy.downlink.join_accept"
        component={Checkbox}
        label={sharedMessages.joinAccept}
        description={policyMessages.joinAcceptDesc}
      />
      <Form.Field
        name="policy.downlink.mac_data"
        component={Checkbox}
        label={policyMessages.macData}
        description={policyMessages.macDataAllowDesc}
      />
      <Form.Field
        name="policy.downlink.application_data"
        component={Checkbox}
        label={sharedMessages.appData}
        description={policyMessages.applicationDataAllowDesc}
      />
    </Col>
  </Row>
)

export default DefaultRoutingPolicyForm
