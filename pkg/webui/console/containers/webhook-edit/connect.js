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
import { replace } from 'connected-react-router'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { updateWebhook } from '@console/store/actions/webhooks'

const mapDispatchToProps = (dispatch, props) => {
  const { webhookId, appId } = props
  return {
    navigateToList: () => dispatch(replace(`/applications/${appId}/integrations/webhooks`)),
    updateWebhook: patch => dispatch(attachPromise(updateWebhook(appId, webhookId, patch))),
    updateHealthStatus: updatedHealthStatus =>
      dispatch(
        attachPromise(updateWebhook(appId, webhookId, updatedHealthStatus, ['health_status'])),
      ),
  }
}

export default WebhookEdit => connect(null, mapDispatchToProps)(WebhookEdit)
