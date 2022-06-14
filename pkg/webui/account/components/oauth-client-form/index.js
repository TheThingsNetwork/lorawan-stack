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

import React, { useCallback } from 'react'
import { useIntl } from 'react-intl'

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

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import capitalizeMessage from '@ttn-lw/lib/capitalize-message'

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
  } = props
  const { formatMessage } = useIntl()

  const approvalStateOptions = approvalStates.map(state => ({
    value: state,
    label: capitalizeMessage(formatMessage({ id: `enum:${state}` })),
  }))

  const handleSubmit = useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values)
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
    ...(isAdmin
      ? {
          endorsed: false,
          skip_authorization: false,
          state: 'STATE_APPROVED',
          state_description: '',
        }
      : {}),
  }

  return (
    <Form
      error={error}
      onSubmit={handleSubmit}
      initialValues={initialValues}
      validationSchema={validationSchema}
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
      <Form.Field
        label={m.tieAccessToSession}
        name="tie_access_to_session"
        component={Checkbox}
        description={m.tieAccessToSessionDesc}
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
          />
          <Form.Field
            label={m.endorsed}
            name="endorsed"
            component={Checkbox}
            description={m.endorsedDesc}
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
        {update && (
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
    redirect_uris: PropTypes.arrayOf(PropTypes.string),
    logout_redirect_uris: PropTypes.arrayOf(PropTypes.string),
    skip_authorization: PropTypes.bool,
    endorsed: PropTypes.bool,
    grants: PropTypes.arrayOf(PropTypes.string),
    state: PropTypes.string,
    state_description: PropTypes.string,
    tie_access_to_session: PropTypes.bool,
  }),
  isAdmin: PropTypes.bool.isRequired,
  onDelete: PropTypes.func,
  onSubmit: PropTypes.func.isRequired,
  pseudoRights: PropTypes.rights.isRequired,
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
    tie_access_to_session: true,
  },
  update: false,
  rights: undefined,
  error: undefined,
  onDelete: () => null,
}

export default OAuthClientForm
