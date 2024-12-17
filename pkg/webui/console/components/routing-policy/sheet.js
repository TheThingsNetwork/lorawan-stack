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

import Icon, { IconCheck, IconX } from '@ttn-lw/components/icon'
import Tooltip from '@ttn-lw/components/tooltip'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import m from '@console/lib/packet-broker/messages'

import style from './routing-policy-sheet.styl'

const RoutingPolicy = ({ enabled, message, positiveMessage, negativeMessage }) => (
  <Tooltip content={<Message content={enabled ? positiveMessage : negativeMessage} />}>
    <span className={classnames(style.policy, 'd-flex al-center')} data-enabled={enabled}>
      <Icon
        icon={enabled ? IconCheck : IconX}
        className={classnames('mr-cs-xxs', {
          'c-text-success-normal': enabled,
          'c-text-error-normal': !enabled,
        })}
      />
      <Message content={message} />
    </span>
  </Tooltip>
)

RoutingPolicy.propTypes = {
  enabled: PropTypes.bool,
  message: PropTypes.message.isRequired,
  negativeMessage: PropTypes.message.isRequired,
  positiveMessage: PropTypes.message.isRequired,
}

RoutingPolicy.defaultProps = {
  enabled: false,
}

const RoutingPolicySheet = ({ policy }) => {
  const { uplink = {}, downlink = {} } = policy

  return (
    <div className="d-flex direction-row" data-test-id="routing-policy-sheet">
      <div className={classnames(style.uplink, 'mr-ls-m')}>
        <Message content={sharedMessages.uplink} component="h4" />
        <RoutingPolicy
          enabled={uplink.join_request}
          message={m.joinRequest}
          positiveMessage={m.forwardsJoinRequest}
          negativeMessage={m.doesNotForwardJoinRequest}
        />
        <RoutingPolicy
          enabled={uplink.mac_data}
          message={sharedMessages.macData}
          positiveMessage={m.forwardsMacData}
          negativeMessage={m.doesNotForwardMacData}
        />
        <RoutingPolicy
          enabled={uplink.application_data}
          message={sharedMessages.appData}
          positiveMessage={m.forwardsApplicationData}
          negativeMessage={m.doesNotForwardApplicationData}
        />
        <RoutingPolicy
          enabled={uplink.signal_quality}
          message={m.signalQualityInformation}
          positiveMessage={m.forwardsSignalQuality}
          negativeMessage={m.doesNotForwardSignalQuality}
        />
        <RoutingPolicy
          enabled={uplink.localization}
          message={m.localizationInformation}
          positiveMessage={m.forwardsLocalization}
          negativeMessage={m.doesNotForwardLocalization}
        />
      </div>
      <div className={style.downlink}>
        <Message content={sharedMessages.downlink} component="h4" />
        <RoutingPolicy
          enabled={downlink.join_accept}
          message={sharedMessages.joinAccept}
          positiveMessage={m.allowsJoinAccept}
          negativeMessage={m.doesNotAllowJoinAccept}
        />
        <RoutingPolicy
          enabled={downlink.mac_data}
          message={sharedMessages.macData}
          positiveMessage={m.allowsMacData}
          negativeMessage={m.doesNotAllowMacData}
        />
        <RoutingPolicy
          enabled={downlink.application_data}
          message={sharedMessages.appData}
          positiveMessage={m.allowsApplicationData}
          negativeMessage={m.doesNotAllowApplicationData}
        />
      </div>
    </div>
  )
}

RoutingPolicySheet.propTypes = {
  policy: PropTypes.routingPolicy.isRequired,
}

export default RoutingPolicySheet
