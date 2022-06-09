// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
import { connect } from 'react-redux'
import { defineMessages } from 'react-intl'

import Status from '@ttn-lw/components/status'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getWebhooksList } from '@console/store/actions/webhooks'

import { selectWebhooksHealthStatusEnabled } from '@console/store/selectors/application-server'
import {
  selectWebhooks,
  selectWebhooksTotalCount,
  selectWebhooksFetching,
} from '@console/store/selectors/webhooks'

import style from './webhooks-table.styl'

const m = defineMessages({
  templateId: 'Template ID',
  format: 'Format',
  baseUrl: 'Base URL',
  healthy: 'Healthy',
  pending: 'Pending',
  requestsFailing: 'Requests failing',
})

@connect(state => ({
  healthStatusEnabled: selectWebhooksHealthStatusEnabled(state),
}))
export default class WebhooksTable extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    healthStatusEnabled: PropTypes.bool.isRequired,
  }

  constructor(props) {
    super(props)

    const { appId } = props
    this.getWebhooksList = () => getWebhooksList(appId, ['template_ids', 'health_status'])
  }

  baseDataSelector(state) {
    return {
      webhooks: selectWebhooks(state),
      totalCount: selectWebhooksTotalCount(state),
      fetching: selectWebhooksFetching(state),
    }
  }

  render() {
    const { healthStatusEnabled } = this.props

    const headers = [
      {
        name: 'ids.webhook_id',
        displayName: sharedMessages.id,
        width: 30,
        sortable: true,
      },
      {
        name: 'base_url',
        displayName: m.baseUrl,
        width: 40,
        sortable: true,
      },
      {
        name: 'template_ids.template_id',
        displayName: m.templateId,
        width: 12,
        render: value => value || <Message className={style.none} content={sharedMessages.none} />,
        sortable: true,
      },
    ]

    if (healthStatusEnabled) {
      headers.push({
        name: 'health_status',
        displayName: sharedMessages.status,
        width: 8,
        render: value => {
          let indicator = 'unknown'
          let label = sharedMessages.unknown

          if (value && value.healthy) {
            indicator = 'good'
            label = m.healthy
          } else if (value && value.unhealthy) {
            indicator = 'bad'
            label = m.requestsFailing
          } else {
            indicator = 'mediocre'
            label = m.pending
          }

          return <Status status={indicator} label={label} pulse={false} />
        },
      })
    }

    headers.push({
      name: 'created_at',
      displayName: sharedMessages.createdAt,
      width: 10,
      sortable: true,
      render: date => <DateTime.Relative value={date} />,
    })

    return (
      <FetchTable
        entity="webhooks"
        defaultOrder="-created_at"
        addMessage={sharedMessages.addWebhook}
        headers={headers}
        getItemsAction={this.getWebhooksList}
        baseDataSelector={this.baseDataSelector}
        tableTitle={<Message content={sharedMessages.webhooks} />}
        paginated={false}
        handlesSorting
        {...this.props}
      />
    )
  }
}
