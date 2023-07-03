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
import { useSelector } from 'react-redux'
import { defineMessages, useIntl } from 'react-intl'
import classNames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectApplicationsTotalCount } from '@console/store/selectors/applications'
import { selectGatewaysTotalCount } from '@console/store/selectors/gateways'
import { selectOrganizationsTotalCount } from '@console/store/selectors/organizations'
import { selectUsersTotalCount } from '@console/store/selectors/users'

import styles from './network-information.styl'

const messages = defineMessages({
  applicationsUsed: 'Total applications',
  gatewaysUsed: 'Total gateways',
  registeredUsers: 'Registered users',
  endDevicesAdded: 'Total end devices',
})

const RegistryTotals = () => {
  const applicationTotal = useSelector(selectApplicationsTotalCount)
  const gatewayTotal = useSelector(selectGatewaysTotalCount)
  const userTotal = useSelector(selectUsersTotalCount)
  const organizationTotal = useSelector(selectOrganizationsTotalCount)

  const intl = useIntl()

  return (
    <>
      <div className={classNames('m-0', 'mt-ls-s', 'mb-cs-xl', styles.registryTotalsContainer)}>
        <div className={styles.registryTotalBox}>
          <Message content={messages.applicationsUsed} component="div" />
          <h2 className="mt-cs-xs mb-cs-xs">{intl.formatNumber(applicationTotal)}</h2>
        </div>
        <div className={styles.registryTotalBox}>
          <Message content={messages.gatewaysUsed} component="div" />
          <h2 className="mt-cs-xs mb-cs-xs">{intl.formatNumber(gatewayTotal)}</h2>
        </div>
        <div className={styles.registryTotalBox}>
          <Message content={messages.registeredUsers} component="div" />
          <h2 className="mt-cs-xs mb-cs-xs">{intl.formatNumber(userTotal)}</h2>
        </div>
        <div className={styles.registryTotalBox}>
          <Message content={sharedMessages.organizations} component="div" />
          <h2 className="mt-cs-xs mb-cs-xs">{intl.formatNumber(organizationTotal)}</h2>
        </div>
      </div>
    </>
  )
}

export default RegistryTotals
