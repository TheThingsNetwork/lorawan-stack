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
import { defineMessages } from 'react-intl'
import { useNavigate, useParams } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'
import toast from '@ttn-lw/components/toast'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import PubsubForm from '@console/components/pubsub-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { deletePubsub, getPubsub, updatePubsub } from '@console/store/actions/pubsubs'

import {
  selectMqttProviderDisabled,
  selectNatsProviderDisabled,
} from '@console/store/selectors/application-server'
import { selectPubsubById, selectSelectedPubsub } from '@console/store/selectors/pubsubs'

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
  'uplink_normalized',
]

const m = defineMessages({
  editPubsub: 'Edit Pub/Sub',
  updateSuccess: 'Pub/Sub updated',
  deleteSuccess: 'Pub/Sub deleted',
})

const EditPubsubInner = () => {
  const { appId, pubsubId } = useParams()
  const pubsub = useSelector(selectSelectedPubsub)
  const mqttDisabled = useSelector(selectMqttProviderDisabled)
  const natsDisabled = useSelector(selectNatsProviderDisabled)
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const getPubsub = useCallback(() => {
    dispatch(getPubsub(appId, pubsubId, pubsubEntitySelector))
  }, [dispatch, appId, pubsubId])

  useBreadcrumbs(
    'apps.single.integrations.edit',
    <Breadcrumb
      path={`/applications/${appId}/integrations/${pubsubId}`}
      content={sharedMessages.edit}
    />,
  )

  const handleSubmit = useCallback(
    async patch => {
      await dispatch(attachPromise(updatePubsub(appId, pubsubId, patch)))
      toast({
        message: m.updateSuccess,
        type: toast.types.SUCCESS,
      })
    },
    [appId, dispatch, pubsubId],
  )

  const handleDelete = useCallback(async () => {
    await dispatch(attachPromise(deletePubsub(appId, pubsubId)))
    toast({
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })
    navigate(`/applications/${appId}/integrations/pubsubs`)
  }, [appId, dispatch, navigate, pubsubId])

  return (
    <div className="container container--xl grid">
      <PageTitle title={m.editPubsub} className="mb-0" />
      <div className="item-12">
        <PubsubForm
          update
          appId={appId}
          initialPubsubValue={pubsub}
          onSubmit={handleSubmit}
          onDelete={handleDelete}
          mqttDisabled={mqttDisabled}
          natsDisabled={natsDisabled}
        />
      </div>
    </div>
  )
}

const EditPubsub = () => {
  const { appId, pubsubId } = useParams()

  // Check if the pubsub exists after it was possibly deleted.
  const pubsub = useSelector(state => selectPubsubById(state, pubsubId))
  const hasPubsub = Boolean(pubsub)

  return (
    <RequireRequest requestAction={getPubsub(appId, pubsubId, pubsubEntitySelector)}>
      {hasPubsub && <EditPubsubInner />}
    </RequireRequest>
  )
}

export default EditPubsub
