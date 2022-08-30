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

import { connect } from 'react-redux'

import withRequest from '@ttn-lw/lib/components/with-request'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import { mayViewApplicationEvents } from '@console/lib/feature-checks'

import { listWebhookTemplates } from '@console/store/actions/webhook-templates'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import {
  selectWebhookTemplates,
  selectWebhookTemplatesFetching,
  selectWebhookTemplatesError,
} from '@console/store/selectors/webhook-templates'

const selector = [
  'base_url',
  'create_downlink_api_key',
  'description',
  'documentation_url',
  'downlink_ack',
  'downlink_failed',
  'downlink_nack',
  'downlink_queue_invalidated',
  'downlink_queued',
  'downlink_sent',
  'fields',
  'format',
  'headers',
  'ids',
  'info_url',
  'join_accept',
  'location_solved',
  'logo_url',
  'name',
  'service_data',
  'uplink_message',
  'uplink_normalized',
]

const mapStateToProps = state => ({
  appId: selectSelectedApplicationId(state),
  webhookTemplates: selectWebhookTemplates(state),
  fetching: selectWebhookTemplatesFetching(state),
  error: selectWebhookTemplatesError(state),
})

const mapDispatchToProps = {
  listWebhookTemplates,
}

export default ApplicationWebhooks =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
  )(
    withFeatureRequirement(mayViewApplicationEvents, {
      redirect: ({ appId }) => `/applications/${appId}`,
    })(
      withRequest(({ listWebhookTemplates }) => listWebhookTemplates(selector))(
        ApplicationWebhooks,
      ),
    ),
  )
