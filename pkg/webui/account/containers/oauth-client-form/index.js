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

import React, { useState, useCallback } from 'react'
import { defineMessages, useIntl } from 'react-intl'
import { connect } from 'react-redux'
import { push } from 'connected-react-router'

import tts from '@account/api/tts'

import Checkbox from '@ttn-lw/components/checkbox'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import Select from '@ttn-lw/components/select'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'

import RightsGroup from '@console/components/rights-group'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
/* import { getApplicationId } from '@ttn-lw/lib/selectors/id'
import { id as applicationIdRegexp } from '@ttn-lw/lib/regexp' */

const m = defineMessages({
  clientName: 'OAuth Client name',
  clientIdPlaceholder: 'my-new-oauth-client',
  clientNamePlaceholder: 'My new OAuth Client',
  clientDescPlaceholder: 'Description for my new OAuth Client',
  clientDescDescription:
    'Optional OAuth Client description; can also be used to save notes about the client',
  createClient: 'Create OAuth Client',
  deleteTitle: 'Are you sure you want to delete this account?',
  deleteWarning:
    "This will <strong>PERMANENTLY DELETE THIS ACCOUNT</strong> and <strong>LOCK THE USER ID AND EMAIL FOR RE-REGISTRATION</strong>. Associated entities (e.g. gateways, applications and end devices) owned by this user that do not have any other collaborators will become <strong>UNACCESSIBLE</strong> and it will <strong>NOT BE POSSIBLE TO REGISTER ENTITIES WITH THE SAME ID OR EUI's AGAIN</strong>. Make sure you assign new collaborators to such entities if you plan to continue using them.",
  purgeWarning:
    "This will <strong>PERMANENTLY DELETE THIS ACCOUNT</strong>. Associated entities (e.g. gateways, applications and end devices) owned by this user that do not have any other collaborators will become <strong>UNACCESSIBLE</strong> and it will <strong>NOT BE POSSIBLE TO REGISTER ENTITIES WITH THE SAME ID OR EUI's AGAIN</strong>. Make sure you assign new collaborators to such entities if you plan to continue using them.",
  deleteConfirmMessage: "Please type in this user's user ID to confirm.",
})

const capitalize = str => str.charAt(0).toUpperCase() + str.slice(1)

const approvalStates = [
  'STATE_REQUESTED',
  'STATE_APPROVED',
  'STATE_REJECTED',
  'STATE_FLAGGED',
  'STATE_SUSPENDED',
]

const encodeGrants = value => {
  const grants = Object.keys(value).map(grant => {
    if (value[grant]) {
      return grant
    }

    return null
  })

  return grants
}

const decodeGrants = value => {
  const grants = value.reduce((g, i) => {
    g[i] = true
    return g
  }, {})

  return grants
}

const validationSchema = Yup.object().shape({
  owner_id: Yup.string().required(sharedMessages.validateRequired),
  ids: Yup.object().shape({
    client_id: Yup.string()
      .min(2, Yup.passValues(sharedMessages.validateTooShort))
      .max(36, Yup.passValues(sharedMessages.validateTooLong))
      /* .matches(userIdRegexp, Yup.passValues(sharedMessages.validateIdFormat)) */
      .required(sharedMessages.validateRequired),
  }),
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(2000, Yup.passValues(sharedMessages.validateTooLong)),
  description: Yup.string(),
  redirect_uris: Yup.array().max(10, Yup.passValues(sharedMessages.attributesValidateTooMany)),
  logout_redirect_uris: Yup.array().max(
    10,
    Yup.passValues(sharedMessages.attributesValidateTooMany),
  ),
  skip_authorization: Yup.bool(),
  endorsed: Yup.bool(),
  grants: Yup.array().max(3, Yup.passValues(sharedMessages.attributesValidateTooMany)),
  state: Yup.string()
    .oneOf(approvalStates, sharedMessages.validateRequired)
    .required(sharedMessages.validateRequired),
  state_description: Yup.string(),
  rights: Yup.array().min(1, sharedMessages.validateRights),
})

const OAuthClientForm = props => {
  const {
    isAdmin,
    userId,
    rights,
    pseudoRights,
    navigateToOAuthClient,
    update,
    initialValues: values,
  } = props
  const { formatMessage } = useIntl()
  console.log(values)

  const approvalStateOptions = approvalStates.map(state => ({
    value: state,
    label: capitalize(formatMessage({ id: `enum:${state}` })),
  }))

  const [error, setError] = useState()
  const handleSubmit = useCallback(
    async (values, { setSubmitting }) => {
      const { owner_id, client_id } = values

      setError(undefined)

      try {
        await tts.Clients.create(
          owner_id,
          {
            ...values,
          },
          userId === owner_id,
        )
        navigateToOAuthClient(client_id)
      } catch (error) {
        setSubmitting(false)
        setError(error)
      }
    },
    [userId, navigateToOAuthClient],
  )

  const handleDelete = useCallback(() => {
    console.log('deleted')
  }, [])

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
    >
      <Form.Field
        title={'OAuth Client'}
        name="ids.client_id"
        placeholder={m.clientIdPlaceholder}
        component={Input}
        disabled={update}
        autoFocus={!update}
        required
      />
      <Form.Field
        title={m.clientName}
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
        title={'Redirect URLs'}
        valuePlaceholder={'https://example.com/'}
        addMessage={'Add redirect URL'}
        component={KeyValueMap}
        indexAsKey
        description={'The allowed redirect URIs against which authorization requests are checked'}
      />
      <Form.Field
        name="logout_redirect_uris"
        title={'Logout redirect URLs'}
        valuePlaceholder={'https://example.com/'}
        addMessage={'Add logout redirect URL'}
        component={KeyValueMap}
        indexAsKey
        description={
          'The allowed logout redirect URIs against which client initiated logout requests are checked'
        }
      />
      {isAdmin && (
        <>
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
          />
          <Form.Field
            title={'Skip Authorization'}
            name="skip_authorization"
            component={Checkbox}
            description={'If set, the authorization page will be skipped'}
          />
          <Form.Field
            title={'Endorsed'}
            name="endorsed"
            component={Checkbox}
            description={'If set, the authorization page will show endorsement'}
          />
        </>
      )}
      <Form.Field
        title={'Grants'}
        name="grants"
        encode={encodeGrants}
        decode={decodeGrants}
        component={Checkbox.Group}
        description={'OAuth flows that can be used for the client to get a token'}
      >
        <Checkbox name="GRANT_AUTHORIZATION_CODE" label={'Grant authorization code'} />
        <Checkbox name="GRANT_PASSWORD" label={'Grant refresh token'} />
        {isAdmin && <Checkbox name="GRANT_REFRESH_TOKEN" label={'Grant password'} />}
      </Form.Field>
      <Form.Field
        name="rights"
        title={sharedMessages.rights}
        component={RightsGroup}
        rights={rights}
        pseudoRight={pseudoRights}
        entityTypeMessage={'OAuth Client'}
      />
      <SubmitBar>
        <Form.Submit
          message={update ? sharedMessages.saveChanges : m.createClient}
          component={SubmitButton}
        />
        {update && (
          <DeleteModalButton
            message={'Delete OAuth Client'}
            entityId={initialValues.ids.client_id}
            entityName={initialValues.name}
            title={m.deleteTitle}
            confirmMessage={m.deleteConfirmMessage}
            defaultMessage={m.deleteWarning}
            onApprove={handleDelete}
            shouldConfirm
            mayPurge
          />
        )}
      </SubmitBar>
    </Form>
  )
}

OAuthClientForm.propTypes = {
  initialValues: PropTypes.shape({
    ids: PropTypes.shape({
      client_id: PropTypes.string,
    }).isRequired,
    name: PropTypes.string,
    description: PropTypes.string,
  }),
  isAdmin: PropTypes.bool.isRequired,
  navigateToOAuthClient: PropTypes.func.isRequired,
  pseudoRights: PropTypes.rights.isRequired,
  rights: PropTypes.rights.isRequired,
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
}

export default connect(dispatch => ({
  navigateToOAuthClient: clientId => dispatch(push(`/oauth-clients/${clientId}`)),
}))(OAuthClientForm)
