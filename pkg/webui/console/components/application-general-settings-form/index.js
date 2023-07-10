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

import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import Checkbox from '@ttn-lw/components/checkbox'
import Notification from '@ttn-lw/components/notification'

import CollaboratorSelect from '@ttn-lw/containers/collaborator-select'
import {
  decodeContact,
  encodeContact,
  organizationSchema,
  userSchema,
} from '@ttn-lw/containers/collaborator-select/util'

import Message from '@ttn-lw/lib/components/message'

import Require from '@console/lib/components/require'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  attributeValidCheck,
  attributeTooShortCheck,
  attributeKeyTooLongCheck,
  attributeValueTooLongCheck,
} from '@console/lib/attributes'

const m = defineMessages({
  basics: 'Basics',
  deleteApp: 'Delete application',
  useAlcsync: 'Use Application Layer Clock Synchronization',
  contactWarning:
    'Note that if no contact is provided, it will default to the first collaborator of the application.',
  adminContactDescription:
    'Administrative contact information for this application. Typically used to indicate who to contact with administrative questions about the application.',
  techContactDescription:
    'Technical contact information for this application. Typically used to indicate who to contact with technical/security questions about the application.',
})

const validationSchema = Yup.object().shape({
  name: Yup.string()
    .min(3, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  description: Yup.string().max(150, Yup.passValues(sharedMessages.validateTooLong)),
  attributes: Yup.object()
    .nullable()
    .test('has no null values', sharedMessages.attributesValidateRequired, attributeValidCheck)
    .test(
      'has key length longer than 2',
      sharedMessages.attributeKeyValidateTooShort,
      attributeTooShortCheck,
    )
    .test(
      'has key length less than 36',
      sharedMessages.attributeKeyValidateTooLong,
      attributeKeyTooLongCheck,
    )
    .test(
      'has value length less than 200',
      sharedMessages.attributeValueValidateTooLong,
      attributeValueTooLongCheck,
    ),
  skip_payload_crypto: Yup.boolean(),
  alcsync: Yup.boolean(),
  administrative_contact: Yup.object().when(['organization_ids'], {
    is: organizationIds => Boolean(organizationIds),
    then: schema => schema.concat(organizationSchema),
    otherwise: schema => schema.concat(userSchema),
  }),
  technical_contact: Yup.object().when(['organization_ids'], {
    is: organizationIds => Boolean(organizationIds),
    then: schema => schema.concat(organizationSchema),
    otherwise: schema => schema.concat(userSchema),
  }),
})

const encodeAttributes = formValue =>
  (Array.isArray(formValue) &&
    formValue.reduce(
      (result, { key, value }) => ({
        ...result,
        [key]: value,
      }),
      {},
    )) ||
  undefined

const decodeAttributes = attributesType =>
  (attributesType &&
    Object.keys(attributesType).reduce(
      (result, key) =>
        result.concat({
          key,
          value: attributesType[key],
        }),
      [],
    )) ||
  []

const ApplicationGeneralSettingsForm = ({
  error,
  handleSubmit,
  initialValues,
  mayViewApplicationLink,
  mayDeleteApplication,
  appId,
  applicationName,
  handleDelete,
  shouldConfirmDelete,
  mayPurge,
}) => (
  <Form
    error={error}
    onSubmit={handleSubmit}
    initialValues={initialValues}
    validationSchema={validationSchema}
    validateSync={false}
  >
    <Form.Field
      title={sharedMessages.appId}
      name="ids.application_id"
      required
      component={Input}
      disabled
    />
    <Form.Field title={sharedMessages.name} name="name" component={Input} />
    <Form.Field
      title={sharedMessages.description}
      type="textarea"
      name="description"
      component={Input}
    />
    {mayViewApplicationLink && (
      <Form.Field
        label={sharedMessages.skipCryptoTitle}
        name="skip_payload_crypto"
        component={Checkbox}
        tooltipId={tooltipIds.SKIP_PAYLOAD_CRYPTO_OVERRIDE}
      />
    )}
    <Form.Field
      label={m.useAlcsync}
      name="alcsync"
      component={Checkbox}
      tooltipId={tooltipIds.ALCSYNC}
    />
    <Form.Field
      name="attributes"
      title={sharedMessages.attributes}
      keyPlaceholder={sharedMessages.key}
      valuePlaceholder={sharedMessages.value}
      addMessage={sharedMessages.addAttributes}
      component={KeyValueMap}
      description={sharedMessages.attributeDescription}
      encode={encodeAttributes}
      decode={decodeAttributes}
    />
    <Form.SubTitle title={sharedMessages.contactInformation} className="mb-cs-s" />
    <Notification small warning content={m.contactWarning} />
    <CollaboratorSelect
      name="administrative_contact"
      title={sharedMessages.adminContact}
      placeholder={sharedMessages.contactFieldPlaceholder}
      entity="application"
      entityId={appId}
      encode={encodeContact}
      decode={decodeContact}
    />
    <Message
      content={m.adminContactDescription}
      component="p"
      className="mt-cs-xs tc-subtle-gray"
    />
    <CollaboratorSelect
      name="technical_contact"
      title={sharedMessages.technicalContact}
      placeholder={sharedMessages.contactFieldPlaceholder}
      entity="application"
      entityId={appId}
      encode={encodeContact}
      decode={decodeContact}
    />
    <Message content={m.techContactDescription} component="p" className="mt-cs-xs tc-subtle-gray" />
    <SubmitBar>
      <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      <Require featureCheck={mayDeleteApplication}>
        <DeleteModalButton
          message={m.deleteApp}
          entityId={appId}
          entityName={applicationName}
          onApprove={handleDelete}
          shouldConfirm={shouldConfirmDelete}
          mayPurge={mayPurge}
        />
      </Require>
    </SubmitBar>
  </Form>
)

ApplicationGeneralSettingsForm.propTypes = {
  appId: PropTypes.string.isRequired,
  applicationName: PropTypes.string,
  error: PropTypes.string,
  handleDelete: PropTypes.func.isRequired,
  handleSubmit: PropTypes.func.isRequired,
  initialValues: PropTypes.shape({
    name: PropTypes.string,
    description: PropTypes.string,
    attributes: PropTypes.shape({}),
    skip_payload_crypto: PropTypes.bool,
    alcsync: PropTypes.bool,
    _administrative_contact_id: PropTypes.string,
    _technical_contact_id: PropTypes.string,
  }).isRequired,
  mayDeleteApplication: PropTypes.shape({}).isRequired,
  mayPurge: PropTypes.bool.isRequired,
  mayViewApplicationLink: PropTypes.bool.isRequired,
  shouldConfirmDelete: PropTypes.bool.isRequired,
}

ApplicationGeneralSettingsForm.defaultProps = {
  applicationName: '',
  error: undefined,
}

export default ApplicationGeneralSettingsForm
