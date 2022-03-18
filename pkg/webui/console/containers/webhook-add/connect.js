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

import { connect } from 'react-redux'
import { push } from 'connected-react-router'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { getWebhook, createWebhook } from '@console/store/actions/webhooks'
import { createApplicationApiKey } from '@console/store/actions/api-keys'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

const mapStateToProps = state => ({
  appId: selectSelectedApplicationId(state),
})

const mapDispatchToProps = dispatch => ({
  createApplicationApiKey: (appId, key) =>
    dispatch(attachPromise(createApplicationApiKey(appId, key))),
  navigateToList: appId => dispatch(push(`/applications/${appId}/integrations/webhooks`)),
  createWebhook: (appId, webhook) => dispatch(attachPromise(createWebhook(appId, webhook))),
  getWebhook: (appId, webhookId, selector) =>
    dispatch(attachPromise(getWebhook(appId, webhookId, selector))),
})

export default WebhookTemplate => connect(mapStateToProps, mapDispatchToProps)(WebhookTemplate)
