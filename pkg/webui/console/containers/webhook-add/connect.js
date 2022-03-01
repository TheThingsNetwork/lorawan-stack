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
import { push } from 'connected-react-router'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import { createApplicationApiKeys } from '@console/store/actions/api-keys'
import { createWebhook } from '@console/store/actions/webhooks'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

const mapStateToProps = state => ({
  appId: selectSelectedApplicationId(state),
})

const mapDispatchToProps = dispatch => ({
  createApplicationApiKeys: (appId, key) =>
    dispatch(attachPromise(createApplicationApiKeys(appId, key))),
  navigateToList: appId => dispatch(push(`/applications/${appId}/integrations/webhooks`)),
  createWebhook: (appId, webhook) => dispatch(attachPromise(createWebhook(appId, webhook))),
})

export default WebhookTemplate => connect(mapStateToProps, mapDispatchToProps)(WebhookTemplate)
