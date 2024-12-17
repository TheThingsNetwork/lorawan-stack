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

import React, { useCallback, useRef, useState } from 'react'
import { defineMessages, FormattedRelativeTime } from 'react-intl'
import { uniq } from 'lodash'

import tts from '@console/api/tts'

import { IconTrash, IconRefresh, IconPlayerPlay, IconPlayerPause } from '@ttn-lw/components/icon'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Button from '@ttn-lw/components/button'
import Notification from '@ttn-lw/components/notification'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import ModalButton from '@ttn-lw/components/button/modal-button'
import PortalledModal from '@ttn-lw/components/modal/portalled'
import Checkbox from '@ttn-lw/components/checkbox'
import Link from '@ttn-lw/components/link'
import Select from '@ttn-lw/components/select'

import Message from '@ttn-lw/lib/components/message'

import WebhookTemplateInfo from '@console/components/webhook-template-info'

import WebhookFormatSelector from '@console/containers/webhook-formats-select'

import Yup from '@ttn-lw/lib/yup'
import { url as urlRegexp, id as webhookIdRegexp } from '@ttn-lw/lib/regexp'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import { apiKey as webhookAPIKeyRegexp, duration as durationRegExp } from '@console/lib/regexp'

import {
  blankValues,
  encodeValues,
  decodeValues,
  decodeMessageType,
  encodeMessageType,
} from './mapping'

const units = {
  s: 'second',
  m: 'minute',
  h: 'hour',
}

const pathPlaceholder = '/path/to/webhook'

const m = defineMessages({
  idPlaceholder: 'my-new-webhook',
  messageInfo:
    'For each enabled event type an optional path can be defined which will be appended to the base URL',
  deleteWebhook: 'Delete Webhook',
  modalWarning:
    'Are you sure you want to delete webhook "{webhookId}"? Deleting a webhook cannot be undone.',
  additionalHeaders: 'Additional headers',
  downlinkAPIKey: 'Downlink API key',
  downlinkAPIKeyDesc:
    'The API key will be provided to the endpoint using the "X-Downlink-Apikey" header',
  templateInformation: 'Template information',
  updateErrorTitle: 'Could not update webhook',
  createErrorTitle: 'Could not create webhook',
  reactivateButtonMessage: 'Reactivate',
  suspendedWebhookMessage:
    'This webhook has been deactivated due to several unsuccessful forwarding attempts. It will be automatically reactivated after {webhookRetryInterval}. If you wish to reactivate right away, you can use the button below.',
  pendingInfo:
    'This webhook is currently pending until attempting its first regular request attempt. Note that webhooks can be restricted if they encounter too many request failures.',
  pausedInfo:
    'This webhook is currently paused and no messages are forwarded to the configured end point.',
  pauseWebhook: 'Pause webhook',
  pauseWebhookQuestion: 'Pause webhook?',
  pauseWebhookDescription:
    'When a webhook is paused, messages will not be forwarded to the configured end point and will be dropped.',
  messagePathValidateTooLong: 'Enabled message path must be at most 64 characters',
  basicAuthCheckbox: 'Use basic access authentication (basic auth)',
  requestBasicAuth: 'Request authentication',
  validateNoColon: 'Basic auth usernames may not contain colons',
  validateEmpty:
    'There must be no empty header entry. Please remove such entries before submitting.',
  validateNoDuplicate:
    'There must be no duplicate headers. Please remove or merge headers with the same key.',
  webhooksDescription:
    'The Webhooks feature allows The Things Stack to send application related messages to specific HTTP(S) endpoints. You can also use webhooks to schedule downlinks to an end device. Learn more in our <Link>Webhooks guide</Link>.',
  filterEventData: 'Filter event data',
  fieldMaskPlaceholder: 'Select a filter path',
  filtersAdd: 'Add filter path',
  pause: 'Pause',
  activate: 'Activate',
})

// We can use the allowed field masks of the `ApplicationUpStorage` API as
// options for the webhook field mask paths.
const filterOptions = uniq(
  tts.api.ApplicationUpStorage.GetStoredApplicationUpAllowedFieldMaskPaths,
).map(v => ({
  value: v,
  label: v,
}))

const isReadOnly = value => value.readOnly

const messageCheck = message => {
  if (message && 'path' in message) {
    if (message.path === undefined) {
      return true
    }
    return message.path.length <= 64
  }
  return true
}

const hasNoEmptyEntry = headers =>
  headers.findIndex(i => i.key === '' || i.key === undefined) === -1

const hasNoDuplicateEntry = headers => uniq(headers.map(i => i.key)).length === headers.length

const validationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    webhook_id: Yup.string()
      .min(3, Yup.passValues(sharedMessages.validateTooShort))
      .max(36, Yup.passValues(sharedMessages.validateTooLong))
      .matches(webhookIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
      .required(sharedMessages.validateRequired),
  }),
  format: Yup.string().required(sharedMessages.validateRequired),
  field_mask: Yup.object().shape({
    paths: Yup.array().of(Yup.string()).compact(),
  }),
  _headers: Yup.array()
    .of(
      Yup.object({
        key: Yup.string(),
        value: Yup.string(),
      }),
    )
    .test('has no duplicate entry', m.validateNoDuplicate, hasNoDuplicateEntry)
    .test('has no empty entry', m.validateEmpty, hasNoEmptyEntry)
    .default([]),
  _basic_auth_username: Yup.string()
    .test(
      'username does not have a colon',
      m.validateNoColon,
      val => val === undefined || !val.includes(':'),
    )
    .when('_basic_auth_enabled', {
      is: true,
      then: schema => schema.required(sharedMessages.validateRequired),
    }),
  _basic_auth_password: Yup.string().when('_basic_auth_enabled', {
    is: true,
    then: schema => schema.required(sharedMessages.validateRequired),
  }),
  base_url: Yup.string()
    .matches(urlRegexp, Yup.passValues(sharedMessages.validateUrl))
    .required(sharedMessages.validateRequired),
  downlink_api_key: Yup.string().matches(
    webhookAPIKeyRegexp,
    Yup.passValues(sharedMessages.validateApiKey),
  ),
  uplink_message: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
  uplink_normalized: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
  join_accept: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
  downlink_ack: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
  downlink_nack: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
  downlink_sent: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
  downlink_failed: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
  downlink_queued: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
  downlink_queue_invalidated: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
  location_solved: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
  service_data: Yup.object()
    .shape({
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck)
    .nullable(),
})

const WebhookForm = props => {
  const {
    update,
    initialWebhookValue,
    webhookTemplate,
    healthStatusEnabled,
    webhookRetryInterval,
    hasUnhealthyWebhookConfig,
    error: propsError,
    existCheck,
    onSubmit,
    onDelete,
    onDeleteSuccess,
    onDeleteFailure,
    onReactivate,
    onReactivateSuccess,
    onPause,
  } = props

  const form = useRef(null)
  const modalResolve = useRef(() => {})
  const modalReject = useRef(() => {})
  const [shouldShowCredentialsInput, setShouldShowCredentialsInput] = useState(
    Boolean(initialWebhookValue?.headers?.Authorization?.startsWith('Basic ')) &&
      Boolean(!decodeValues(initialWebhookValue)._headers.find(i => i.decodeError)?.decodeError),
  )
  const [showDecodeError, setShowDecodeError] = useState(
    Boolean(decodeValues(initialWebhookValue)._headers.find(i => i.decodeError)?.decodeError),
  )
  const [displayOverwriteModal, setDisplayOverwriteModal] = useState(false)
  const [existingId, setExistingId] = useState(undefined)
  const [error, setError] = useState(undefined)

  const retryIntervalValue = webhookRetryInterval?.match(durationRegExp)[0]
  const retryIntervalUnit = webhookRetryInterval?.match(durationRegExp)[1]
  const retryIntervalIntlUnit = units[retryIntervalUnit]

  let initialValues = blankValues
  if (update && initialWebhookValue) {
    initialValues = decodeValues({ ...blankValues, ...initialWebhookValue })
  }

  const hasTemplate = Boolean(webhookTemplate)

  const healthStatus = initialWebhookValue?.health_status
  const mayReactivate = update && hasUnhealthyWebhookConfig && healthStatus?.unhealthy
  const isPending = update && healthStatusEnabled && !healthStatus

  const isPaused = initialWebhookValue?.paused

  const handleReplaceModalDecision = useCallback(mayReplace => {
    if (mayReplace) {
      modalResolve.current()
    } else {
      modalReject.current()
    }
    setDisplayOverwriteModal(false)
  }, [])

  const handleSubmit = useCallback(
    async (values, { setSubmitting, resetForm }) => {
      const castedWebhookValues = validationSchema.cast(values)
      const encodedValues = encodeValues(castedWebhookValues)
      const webhookId = encodedValues.ids.webhook_id
      const exists = await existCheck(webhookId)
      setShowDecodeError(
        Boolean(decodeValues(encodedValues)._headers.find(i => i.decodeError)?.decodeError),
      )

      if (exists) {
        setDisplayOverwriteModal(true)
        setExistingId(webhookId)
        await new Promise((resolve, reject) => {
          modalResolve.current = resolve
          modalReject.current = reject
        })
      }
      await onSubmit(castedWebhookValues, encodedValues, { setSubmitting, resetForm })
    },
    [existCheck, onSubmit],
  )

  const handleDelete = useCallback(async () => {
    try {
      await onDelete()
      form.current.resetForm()
      onDeleteSuccess()
    } catch (error) {
      setError(error)
      onDeleteFailure()
    }
  }, [onDelete, onDeleteFailure, onDeleteSuccess])

  const handleReactivate = useCallback(async () => {
    const healthStatus = {
      health_status: null,
    }

    try {
      await onReactivate(healthStatus)
      onReactivateSuccess()
    } catch (error) {
      setError(error)
    }
  }, [onReactivate, onReactivateSuccess])

  const handleRequestAuthenticationChange = useCallback(event => {
    const currentHeaders = form.current.values._headers
    if (!event.target.checked) {
      form.current.setFieldValue(
        '_headers',
        currentHeaders.filter(i => !i.readOnly),
      )
    } else {
      form.current.setFieldValue('_headers', [
        { key: 'Authorization', value: 'Basic ...', readOnly: true },
      ])
    }
    setShouldShowCredentialsInput(event.target.checked)
  }, [])

  const handleHeadersChange = useCallback(() => {
    setShowDecodeError(!hasNoEmptyEntry)
  }, [])

  return (
    <>
      {!hasTemplate && (
        <>
          <Message
            content={m.webhooksDescription}
            values={{
              Link: val => (
                <Link.DocLink path="/integrations/webhooks" secondary>
                  {val}
                </Link.DocLink>
              ),
            }}
            component="p"
          />
          <hr className="mb-ls-m" />
        </>
      )}
      {isPaused ? (
        <Notification
          info
          content={m.pausedInfo}
          children={
            <Button
              icon={IconPlayerPlay}
              onClick={onPause}
              message={m.activate}
              className="mt-cs-m"
              secondary
            />
          }
          small
        />
      ) : (
        <>
          {mayReactivate && (
            <Notification
              warning
              content={m.suspendedWebhookMessage}
              messageValues={{
                webhookRetryInterval: (
                  <FormattedRelativeTime
                    style="long"
                    value={retryIntervalValue}
                    unit={retryIntervalIntlUnit}
                  />
                ),
              }}
              children={
                <Button
                  onClick={handleReactivate}
                  icon={IconRefresh}
                  message={m.reactivateButtonMessage}
                  className="mt-cs-m"
                  secondary
                />
              }
              small
            />
          )}
          {isPending && <Notification info content={m.pendingInfo} small />}
        </>
      )}

      <PortalledModal
        title={sharedMessages.idAlreadyExists}
        message={{
          ...sharedMessages.webhookAlreadyExistsModalMessage,
          values: { id: existingId },
        }}
        buttonMessage={sharedMessages.replaceWebhook}
        onComplete={handleReplaceModalDecision}
        approval
        visible={displayOverwriteModal}
      />
      {hasTemplate && <WebhookTemplateInfo webhookTemplate={webhookTemplate} update={update} />}
      <Form
        onSubmit={handleSubmit}
        validationSchema={validationSchema}
        initialValues={initialValues}
        error={error || propsError}
        errorTitle={update ? m.updateErrorTitle : m.createErrorTitle}
        formikRef={form}
      >
        <Form.SubTitle title={sharedMessages.generalSettings} />
        <Form.Field
          name="ids.webhook_id"
          title={sharedMessages.webhookId}
          placeholder={m.idPlaceholder}
          component={Input}
          required
          autoFocus
          disabled={update}
        />
        <WebhookFormatSelector name="format" required />
        <Form.Field
          name="base_url"
          title={sharedMessages.webhookBaseUrl}
          placeholder="https://example.com/webhooks"
          component={Input}
          required
        />
        <Form.Field
          name="downlink_api_key"
          title={m.downlinkAPIKey}
          component={Input}
          description={m.downlinkAPIKeyDesc}
          sensitive
          code
        />
        <Form.Field
          title={m.requestBasicAuth}
          name="_basic_auth_enabled"
          label={m.basicAuthCheckbox}
          onChange={handleRequestAuthenticationChange}
          component={Checkbox}
          tooltipId={tooltipIds.BASIC_AUTH}
        />
        {shouldShowCredentialsInput && (
          <Form.FieldContainer horizontal>
            <Form.Field
              data-test-id="basic-auth-username"
              required
              title={sharedMessages.username}
              name="_basic_auth_username"
              component={Input}
            />
            <Form.Field
              data-test-id="basic-auth-password"
              required
              title={sharedMessages.password}
              name="_basic_auth_password"
              component={Input}
              sensitive
            />
          </Form.FieldContainer>
        )}
        {showDecodeError && (
          <Notification
            warning
            content={
              'Something went wrong and the contents of the Authorization header could not be decoded.'
            }
            small
            className="mt-cs-xl"
          />
        )}
        <Form.Field
          name="_headers"
          title={m.additionalHeaders}
          keyPlaceholder={sharedMessages.authorization}
          valuePlaceholder={sharedMessages.bearerMyAuthToken}
          addMessage={sharedMessages.addHeaderEntry}
          component={KeyValueMap}
          isReadOnly={isReadOnly}
          onChange={handleHeadersChange}
        />
        <Form.Field
          name="field_mask.paths"
          title={m.filterEventData}
          valuePlaceholder={m.fieldMaskPlaceholder}
          component={KeyValueMap}
          tooltipId={tooltipIds.FILTER_EVENT_DATA}
          inputElement={Select}
          addMessage={m.filtersAdd}
          additionalInputProps={{ options: filterOptions }}
          indexAsKey
        />
        <Form.SubTitle title={sharedMessages.eventEnabledTypes} className="mb-0" />
        <Message component="p" content={m.messageInfo} className="mt-0 mb-ls-xxs" />
        <Form.Field
          name="uplink_message"
          type="toggled-input"
          enabledMessage={sharedMessages.uplinkMessage}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventUplinkMessageDesc}
        />
        <Form.Field
          name="uplink_normalized"
          type="toggled-input"
          enabledMessage={sharedMessages.uplinkNormalized}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventUplinkNormalizedDesc}
        />
        <Form.Field
          name="join_accept"
          type="toggled-input"
          enabledMessage={sharedMessages.joinAccept}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventJoinAcceptDesc}
        />
        <Form.Field
          name="downlink_ack"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkAck}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkAckDesc}
        />
        <Form.Field
          name="downlink_nack"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkNack}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkNackDesc}
        />
        <Form.Field
          name="downlink_sent"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkSent}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkSentDesc}
        />
        <Form.Field
          name="downlink_failed"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkFailed}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkFailedDesc}
        />
        <Form.Field
          name="downlink_queued"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkQueued}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkQueuedDesc}
        />
        <Form.Field
          name="downlink_queue_invalidated"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkQueueInvalidated}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkQueueInvalidatedDesc}
          tooltipId={tooltipIds.DOWNLINK_QUEUE_INVALIDATED}
        />
        <Form.Field
          name="location_solved"
          type="toggled-input"
          enabledMessage={sharedMessages.locationSolved}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventLocationSolvedDesc}
        />
        <Form.Field
          name="service_data"
          type="toggled-input"
          enabledMessage={sharedMessages.serviceData}
          placeholder={pathPlaceholder}
          decode={decodeMessageType}
          encode={encodeMessageType}
          component={Input.Toggled}
          description={sharedMessages.eventServiceDataDesc}
        />
        <SubmitBar>
          <Form.Submit
            component={SubmitButton}
            message={update ? sharedMessages.saveChanges : sharedMessages.addWebhook}
          />
          {update && (
            <div className="d-flex gap-cs-s">
              {isPaused ? (
                <Button secondary icon={IconPlayerPlay} onClick={onPause} message={m.activate} />
              ) : (
                <ModalButton
                  type="button"
                  icon={IconPlayerPause}
                  onApprove={onPause}
                  message={m.pause}
                  modalData={{
                    title: m.pauseWebhookQuestion,
                    noTitleLine: true,
                    buttonMessage: m.pauseWebhook,
                    children: <Message content={m.pauseWebhookDescription} component="span" />,
                    approveButtonProps: {
                      icon: 'close',
                      danger: false,
                    },
                  }}
                />
              )}
              <ModalButton
                type="button"
                icon={IconTrash}
                danger
                naked
                message={m.deleteWebhook}
                modalData={{
                  message: {
                    values: { webhookId: initialWebhookValue.ids.webhook_id },
                    ...m.modalWarning,
                  },
                }}
                onApprove={handleDelete}
              />
            </div>
          )}
        </SubmitBar>
      </Form>
    </>
  )
}

WebhookForm.propTypes = {
  error: PropTypes.error,
  existCheck: PropTypes.func,
  hasUnhealthyWebhookConfig: PropTypes.bool,
  healthStatusEnabled: PropTypes.bool,
  initialWebhookValue: PropTypes.shape({
    ids: PropTypes.shape({
      webhook_id: PropTypes.string,
    }),
    health_status: PropTypes.shape({
      healthy: PropTypes.shape({}),
      unhealthy: PropTypes.shape({}),
    }),
    headers: PropTypes.shape({
      Authorization: PropTypes.string,
    }),
    paused: PropTypes.bool,
  }),
  onDelete: PropTypes.func,
  onDeleteFailure: PropTypes.func,
  onDeleteSuccess: PropTypes.func,
  onPause: PropTypes.func.isRequired,
  onReactivate: PropTypes.func,
  onReactivateSuccess: PropTypes.func,
  onSubmit: PropTypes.func.isRequired,
  update: PropTypes.bool.isRequired,
  webhookRetryInterval: PropTypes.string,
  webhookTemplate: PropTypes.webhookTemplate,
}

WebhookForm.defaultProps = {
  initialWebhookValue: undefined,
  onReactivate: () => null,
  onReactivateSuccess: () => null,
  onDeleteFailure: () => null,
  onDeleteSuccess: () => null,
  onDelete: () => null,
  webhookTemplate: undefined,
  healthStatusEnabled: false,
  error: undefined,
  existCheck: () => null,
  webhookRetryInterval: null,
  hasUnhealthyWebhookConfig: false,
}
export default WebhookForm
