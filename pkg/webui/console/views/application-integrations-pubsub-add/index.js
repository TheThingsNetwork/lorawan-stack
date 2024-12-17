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

import React, { useCallback } from 'react'
import { useSelector, useDispatch } from 'react-redux'
import { useNavigate, useParams } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import PubsubForm from '@console/components/pubsub-form'

import { isNotFoundError } from '@ttn-lw/lib/errors/utils'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { createPubsub, getPubsub } from '@console/store/actions/pubsubs'

import {
  selectMqttProviderDisabled,
  selectNatsProviderDisabled,
} from '@console/store/selectors/application-server'

// Inner Function
const ApplicationPubsubAdd = () => {
  const { appId } = useParams()
  const mqttDisabled = useSelector(selectMqttProviderDisabled)
  const natsDisabled = useSelector(selectNatsProviderDisabled)
  const dispatch = useDispatch()
  const navigate = useNavigate()

  useBreadcrumbs(
    'apps.single.integrations.add',
    <Breadcrumb path={`/applications/${appId}/integrations/add`} content={sharedMessages.add} />,
  )

  const existCheck = useCallback(
    async pubsubId => {
      try {
        await dispatch(attachPromise(getPubsub(appId, pubsubId, [])))
        return true
      } catch (error) {
        if (isNotFoundError(error)) {
          return false
        }
        throw error
      }
    },
    [appId, dispatch],
  )

  const handleSubmit = useCallback(
    async pubsub => {
      await dispatch(attachPromise(createPubsub(appId, pubsub)))
      navigate(`/applications/${appId}/integrations/pubsubs`)
    },
    [appId, dispatch, navigate],
  )

  return (
    <div className="container container--xl grid">
      <PageTitle title={sharedMessages.addPubsub} className="mb-0" />
      <div className="item-12">
        <PubsubForm
          appId={appId}
          update={false}
          onSubmit={handleSubmit}
          existCheck={existCheck}
          mqttDisabled={mqttDisabled}
          natsDisabled={natsDisabled}
        />
      </div>
    </div>
  )
}

export default ApplicationPubsubAdd
