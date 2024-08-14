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

import React from 'react'
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Checkbox from '@ttn-lw/components/checkbox'

import CollaboratorSelect from '@ttn-lw/containers/collaborator-select'
import { decodeContact, encodeContact } from '@ttn-lw/containers/collaborator-select/util'

import Message from '@ttn-lw/lib/components/message'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import { id as organizationIdRegexp } from '@ttn-lw/lib/regexp'
import contactSchema from '@ttn-lw/lib/shared-schemas'

import { selectUserId } from '@console/store/selectors/user'
import { selectIsConfiguration } from '@console/store/selectors/identity-server'

const validationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    organization_id: Yup.string()
      .min(3, Yup.passValues(sharedMessages.validateTooShort))
      .max(36, Yup.passValues(sharedMessages.validateTooLong))
      .matches(organizationIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
      .required(sharedMessages.validateRequired),
  }),
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  description: Yup.string().max(2000, Yup.passValues(sharedMessages.validateTooLong)),
})

validationSchema.concat(contactSchema)

const m = defineMessages({
  orgDescPlaceholder: 'Description for my new organization',
  orgDescDescription:
    'Optional organization description; can also be used to save notes about the organization',
  orgIdPlaceholder: 'my-new-organization',
  orgNamePlaceholder: 'My new organization',
  adminContactDescription:
    'Administrative contact information for this organization. Typically used to indicate who to contact with administrative questions about the organization.',
  techContactDescription:
    'Technical contact information for this organization. Typically used to indicate who to contact with technical/security questions about the organization.',
})

const initialValues = {
  ids: {
    organization_id: '',
  },
  name: '',
  description: '',
}

const OrganizationForm = props => {
  const { onSubmit, error, submitBarItems, initialValues, submitMessage, update } = props
  const orgId = initialValues.ids?.organization_id
  const isUpdate = Boolean(initialValues.ids.organization_id)
  const userId = useSelector(selectUserId)
  const isConfig = useSelector(selectIsConfiguration)
  const isResctrictedUser =
    isConfig && isConfig.collaborator_rights?.set_others_as_contacts === false

  return (
    <Form
      error={error}
      onSubmit={onSubmit}
      initialValues={initialValues}
      validationSchema={validationSchema}
    >
      <Form.Field
        title={sharedMessages.organizationId}
        name="ids.organization_id"
        placeholder={m.orgIdPlaceholder}
        required
        component={Input}
        disabled={isUpdate}
        autoFocus={!isUpdate}
      />
      <Form.Field
        title={sharedMessages.name}
        name="name"
        placeholder={m.orgNamePlaceholder}
        component={Input}
      />
      <Form.Field
        title={sharedMessages.description}
        name="description"
        type="textarea"
        placeholder={m.orgDescPlaceholder}
        description={m.orgDescDescription}
        component={Input}
      />
      {update && (
        <>
          <Form.SubTitle title={sharedMessages.contactInformation} className="mb-cs-s" />
          <CollaboratorSelect
            name="administrative_contact"
            title={sharedMessages.adminContact}
            placeholder={sharedMessages.contactFieldPlaceholder}
            entity="organization"
            entityId={orgId}
            encode={encodeContact}
            decode={decodeContact}
            required
            isResctrictedUser={isResctrictedUser}
            userId={userId}
          />
          <Message
            content={m.adminContactDescription}
            component="p"
            className="mt-cs-xs c-text-neutral-light"
          />
          <CollaboratorSelect
            name="technical_contact"
            title={sharedMessages.technicalContact}
            placeholder={sharedMessages.contactFieldPlaceholder}
            entity="organization"
            entityId={orgId}
            encode={encodeContact}
            decode={decodeContact}
            required
            isResctrictedUser={isResctrictedUser}
            userId={userId}
          />
          <Message
            content={m.techContactDescription}
            component="p"
            className="mt-cs-xs c-text-neutral-light"
          />
          <Form.Field
            title={sharedMessages.gatewayFanoutNotificationsTitle}
            name="fanout_notifications"
            component={Checkbox}
            label={sharedMessages.gatewayFanoutNotificationsLabel}
            description={sharedMessages.gatewayFanoutNotificationsDescription}
            tooltipId={tooltipIds.GATEWAY_FANOUT_NOTIFICATIONS}
          />
          <Message
            content={m.gatewayFanoutNotificationsDescription}
            component="p"
            className="mt-cs-xs c-text-neutral-light"
          />
        </>
      )}
      <SubmitBar>
        <Form.Submit message={submitMessage} component={SubmitButton} />
        {submitBarItems}
      </SubmitBar>
    </Form>
  )
}
OrganizationForm.propTypes = {
  error: PropTypes.error,
  initialValues: PropTypes.shape({
    ids: PropTypes.shape({
      organization_id: PropTypes.string.isRequired,
    }),
    name: PropTypes.string,
    description: PropTypes.string,
  }),
  onSubmit: PropTypes.func.isRequired,
  submitBarItems: PropTypes.element,
  submitMessage: PropTypes.message,
  update: PropTypes.bool,
}

OrganizationForm.defaultProps = {
  initialValues,
  error: undefined,
  submitBarItems: null,
  submitMessage: sharedMessages.createOrganization,
  update: false,
}

export { OrganizationForm as default, initialValues }
