// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { replace } from 'connected-react-router'

import withRequest from '@ttn-lw/lib/components/with-request'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { deletePubsub, getPubsub, updatePubsub } from '@console/store/actions/pubsubs'

import {
  selectSelectedPubsub,
  selectPubsubFetching,
  selectPubsubError,
} from '@console/store/selectors/pubsubs'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import {
  selectMqttProviderDisabled,
  selectNatsProviderDisabled,
} from '@console/store/selectors/application-server'

const pubsubEntitySelector = [
  'base_topic',
  'format',
  'provider.nats',
  'provider.mqtt',
  'downlink_ack',
  'downlink_failed',
  'downlink_nack',
  'downlink_push',
  'downlink_queued',
  'downlink_queue_invalidated',
  'downlink_replace',
  'downlink_sent',
  'join_accept',
  'location_solved',
  'service_data',
  'uplink_message',
]

const mapStateToProps = state => ({
  appId: selectSelectedApplicationId(state),
  pubsub: selectSelectedPubsub(state),
  fetching: selectPubsubFetching(state),
  error: selectPubsubError(state),
  mqttDisabled: selectMqttProviderDisabled(state),
  natsDisabled: selectNatsProviderDisabled(state),
})

const promisifiedUpdatePubsub = attachPromise(updatePubsub)
const mapDispatchToProps = (dispatch, { match }) => {
  const { appId, pubsubId } = match.params

  return {
    getPubsub: () => dispatch(getPubsub(appId, pubsubId, pubsubEntitySelector)),
    navigateToList: () => dispatch(replace(`/applications/${appId}/integrations/pubsubs`)),
    updatePubsub: patch => dispatch(promisifiedUpdatePubsub(appId, pubsubId, patch)),
    deletePubsub: (appId, pubsubId) => dispatch(attachPromise(deletePubsub(appId, pubsubId))),
  }
}

export default PubsubEdit =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
  )(withRequest(({ getPubsub }) => getPubsub())(PubsubEdit))
