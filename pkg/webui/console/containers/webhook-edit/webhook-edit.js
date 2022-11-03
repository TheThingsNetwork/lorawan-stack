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

import React, { useState } from 'react'
import { defineMessages } from 'react-intl'

import tts from '@console/api/tts'

import toast from '@ttn-lw/components/toast'

import WebhookForm from '@console/components/webhook-form'

import diff from '@ttn-lw/lib/diff'
import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  editWebhook: 'Edit webhook',
  updateSuccess: 'Webhook updated',
  deleteSuccess: 'Webhook deleted',
  reactivateSuccess: 'Webhook activated',
})

const WebhookEdit = props => {
  const {
    selectedWebhook,
    appId,
    webhookTemplate,
    updateWebhook,
    navigateToList,
    healthStatusEnabled,
    webhookId,
    updateHealthStatus,
    isUnhealthyWebhook,
    webhookRetryInterval,
  } = props

  const handleUpdateWebhook = React.useCallback(
    async updatedWebhook => {
      const patch = diff(selectedWebhook, updatedWebhook, ['ids'])

      // Ensure that the header prop is always patched fully, otherwise we loose
      // old header entries.
      if ('headers' in patch) {
        patch.headers = updatedWebhook.headers
      }

      if (Object.keys(patch).length === 0) {
        await updateWebhook(updatedWebhook)
      } else {
        await updateWebhook(patch)
      }
    },
    [updateWebhook, selectedWebhook],
  )
  const showSuccessToast = React.useCallback(() => {
    toast({
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }, [])

  const [error, setError] = useState()
  const handleWebhookSubmit = React.useCallback(
    async (values, newWebhook, { resetForm }) => {
      setError(undefined)
      try {
        const result = await handleUpdateWebhook(newWebhook)
        resetForm({ values })
        showSuccessToast(result)
      } catch (error) {
        resetForm({ values })
        setError(error)
      }
    },
    [handleUpdateWebhook, showSuccessToast],
  )

  const handleDelete = React.useCallback(async () => {
    await tts.Applications.Webhooks.deleteById(appId, webhookId)
  }, [appId, webhookId])
  const handleDeleteSuccess = React.useCallback(() => {
    toast({
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })

    navigateToList()
  }, [navigateToList])

  const handleReactivateSuccess = React.useCallback(() => {
    toast({
      message: m.reactivateSuccess,
      type: toast.types.SUCCESS,
    })
  }, [])
  const handleReactivate = React.useCallback(
    async updatedHealthStatus => {
      await updateHealthStatus(updatedHealthStatus)
    },
    [updateHealthStatus],
  )

  return (
    <WebhookForm
      update
      appId={appId}
      initialWebhookValue={selectedWebhook}
      webhookTemplate={webhookTemplate}
      onSubmit={handleWebhookSubmit}
      onDelete={handleDelete}
      onDeleteSuccess={handleDeleteSuccess}
      onReactivate={handleReactivate}
      onReactivateSuccess={handleReactivateSuccess}
      healthStatusEnabled={healthStatusEnabled}
      webhookRetryInterval={webhookRetryInterval}
      isUnhealthyWebhook={isUnhealthyWebhook}
      error={error}
    />
  )
}

WebhookEdit.propTypes = {
  appId: PropTypes.string.isRequired,
  healthStatusEnabled: PropTypes.bool.isRequired,
  isUnhealthyWebhook: PropTypes.bool.isRequired,
  navigateToList: PropTypes.func.isRequired,
  selectedWebhook: PropTypes.webhook.isRequired,
  updateHealthStatus: PropTypes.func.isRequired,
  updateWebhook: PropTypes.func.isRequired,
  webhookId: PropTypes.string.isRequired,
  webhookRetryInterval: PropTypes.string,
  webhookTemplate: PropTypes.webhookTemplate,
}

WebhookEdit.defaultProps = {
  webhookTemplate: undefined,
  webhookRetryInterval: null,
}

export default WebhookEdit
