// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import { useIntl } from 'react-intl'
import { useSelector } from 'react-redux'
import { isEmpty } from 'lodash'

import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import Checkbox from '@ttn-lw/components/checkbox'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import Notification from '@ttn-lw/components/notification'
import Select from '@ttn-lw/components/select'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import RightsGroup from '@ttn-lw/components/rights-group'

import CollaboratorSelect from '@ttn-lw/containers/collaborator-select'
import { decodeContact, encodeContact } from '@ttn-lw/containers/collaborator-select/util'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import capitalizeMessage from '@ttn-lw/lib/capitalize-message'

import { checkFromState, mayEditBasicClientInformation } from '@console/lib/feature-checks'

import { selectSelectedClientId } from '@console/store/selectors/clients'

import { approvalStates, encodeGrants, decodeGrants } from './utils'
import validationSchema from './validation-schema'
import m from './messages'

const OAuthClientForm = props => {
  const {
    isAdmin,
    userId,
    rights,
    pseudoRights,
    error,
    update,
    initialValues: values,
    onSubmit,
    onDelete,
    isResctrictedUser,
    readOnly,
    mayDelete,
  } = props
  const { formatMessage } = useIntl()

  const clientId = useSelector(selectSelectedClientId)
  const mayEditBasicInformation = useSelector(state =>
    checkFromState(mayEditBasicClientInformation, state),
  )

  const approvalStateOptions = approvalStates.map(state => ({
    value: state,
    label: capitalizeMessage(formatMessage({ id: `enum:${state}` })),
  }))

  const handleSubmit = useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values)

      // If there are no contacts do not include them in the casted value.
      if (
        'administrative_contact' in castedValues &&
        isEmpty(castedValues.administrative_contact)
      ) {
        delete castedValues.administrative_contact
      }
      if ('technical_contact' in castedValues && isEmpty(castedValues.technical_contact)) {
        delete castedValues.technical_contact
      }

      await onSubmit(castedValues, resetForm, setSubmitting)
    },
    [onSubmit],
  )

  const handleDelete = useCallback(
    async shouldPurge => {
      await onDelete(shouldPurge, values.ids.client_id)
    },
    [onDelete, values.ids.client_id],
  )

  const initialValues = {
    owner_id: userId,
    rights: [...pseudoRights],
    ...values,
  }

  return (
    <Form
      error={error}
      onSubmit={handleSubmit}
      initialValues={initialValues}
      validationSchema={validationSchema}
      disabled={readOnly}
    >
      <Form.SubTitle title={sharedMessages.generalSettings} />
      <Form.Field
        title={sharedMessages.oauthClientId}
        name="ids.client_id"
        placeholder={m.clientIdPlaceholder}
        component={Input}
        disabled={update}
        autoFocus={!update}
        required
      />
      <Form.Field
        title={sharedMessages.name}
        name="name"
        placeholder={m.clientNamePlaceholder}
        component={Input}
      />
      <Form.Field
        title={sharedMessages.description}
        type="textarea"
        name="description"
        placeholder={m.clientDescPlaceholder}
        description={m.clientDescDescription}
        component={Input}
      />
      <Form.Field
        name="redirect_uris"
        title={m.redirectUrls}
        valuePlaceholder={m.urlsPlaceholder}
        addMessage={m.addRedirectUri}
        component={KeyValueMap}
        indexAsKey
        description={m.redirectUrlDescription}
      />
      <Form.Field
        name="logout_redirect_uris"
        title={m.logoutRedirectUrls}
        valuePlaceholder={m.urlsPlaceholder}
        addMessage={m.addLogoutRedirectUri}
        component={KeyValueMap}
        description={m.logoutRedirectUrlsDescription}
        indexAsKey
      />
      {isAdmin && (
        <>
          <Form.SubTitle title={m.adminOptions} />
          <Form.Field
            title={sharedMessages.state}
            name="state"
            component={Select}
            options={approvalStateOptions}
          />
          <Form.Field
            title={sharedMessages.stateDescription}
            name="state_description"
            component={Input}
            type="textarea"
            placeholder={m.userDescPlaceholder}
            description={m.stateDescriptionDesc}
          />
          <Form.Field
            label={m.skipAuthorization}
            name="skip_authorization"
            component={Checkbox}
            description={m.skipAuthorizationDesc}
            disabled={!isAdmin}
          />
          <Form.Field
            label={m.endorsed}
            name="endorsed"
            component={Checkbox}
            description={m.endorsedDesc}
            disabled={!isAdmin}
          />
        </>
      )}
      {update && mayEditBasicInformation && (
        <>
          <Form.SubTitle title={sharedMessages.contactInformation} className="mb-cs-s" />
          <CollaboratorSelect
            name="administrative_contact"
            title={sharedMessages.adminContact}
            placeholder={sharedMessages.contactFieldPlaceholder}
            entity={'client'}
            entityId={clientId}
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
            entity={'client'}
            entityId={clientId}
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
        </>
      )}
      {((isAdmin && update) || !update) && (
        <>
          <Form.SubTitle title={m.grantTypeAndRights} />
          <Form.Field
            title={m.grants}
            name="grants"
            encode={encodeGrants}
            decode={decodeGrants}
            component={Checkbox.Group}
            description={m.grantsDesc}
          >
            <Checkbox name="GRANT_AUTHORIZATION_CODE" label={m.grantAuthorizationLabel} />
            <Checkbox name="GRANT_REFRESH_TOKEN" label={m.grantRefreshTokenLabel} />
            {isAdmin && <Checkbox name="GRANT_PASSWORD" label={m.grantPasswordLabel} />}
          </Form.Field>
        </>
      )}
      <Notification small warning content={update ? m.updateWarning : m.rightsWarning} />
      <Form.Field
        name="rights"
        title={sharedMessages.rights}
        component={RightsGroup}
        rights={rights}
        pseudoRight={pseudoRights}
        entityTypeMessage={sharedMessages.client}
      />
      <SubmitBar>
        <Form.Submit
          message={update ? sharedMessages.saveChanges : m.createClient}
          component={SubmitButton}
        />
        {update && mayDelete && (
          <DeleteModalButton
            message={m.deleteClient}
            entityId={initialValues.ids.client_id}
            entityName={initialValues.name}
            title={m.deleteTitle}
            defaultMessage={m.deleteWarning}
            onApprove={handleDelete}
            mayPurge={isAdmin}
          />
        )}
      </SubmitBar>
    </Form>
  )
}

OAuthClientForm.propTypes = {
  error: PropTypes.string,
  initialValues: PropTypes.shape({
    owner_id: PropTypes.string,
    ids: PropTypes.shape({
      client_id: PropTypes.string,
    }).isRequired,
    name: PropTypes.string,
    description: PropTypes.string,
  }),
  isAdmin: PropTypes.bool.isRequired,
  isResctrictedUser: PropTypes.bool.isRequired,
  onDelete: PropTypes.func,
  onSubmit: PropTypes.func.isRequired,
  pseudoRights: PropTypes.rights.isRequired,
  readOnly: PropTypes.bool,
  rights: PropTypes.rights,
  update: PropTypes.bool,
  userId: PropTypes.string.isRequired,
}

OAuthClientForm.defaultProps = {
  initialValues: {
    ids: {
      client_id: '',
    },
    name: '',
    description: '',
    redirect_uris: [],
    logout_redirect_uris: [],
    skip_authorization: false,
    endorsed: false,
    grants: ['GRANT_AUTHORIZATION_CODE'],
    state: '',
    state_description: '',
  },
  update: false,
  readOnly: false,
  rights: undefined,
  error: undefined,
  onDelete: () => null,
}

export default OAuthClientForm
