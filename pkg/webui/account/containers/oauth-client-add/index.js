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
import { connect } from 'react-redux'
import { push } from 'connected-react-router'
import { defineMessages } from 'react-intl'

import toast from '@ttn-lw/components/toast'

import OAuthClientForm from '@account/components/oauth-client-form'

import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { createClient } from '@account/store/actions/clients'

const m = defineMessages({
  createSuccess: 'Client created',
  createFailure: 'There was an error creating this client',
})

const ClientAdd = props => {
  const { isAdmin, userId, rights, pseudoRights, navigateToOAuthClient, createOauthClient } = props

  const [error, setError] = useState()
  const handleSubmit = useCallback(
    async (values, setSubmitting) => {
      const { owner_id, ids } = values

      setError(undefined)

      try {
        await createOauthClient(
          owner_id,
          {
            ...values,
          },
          userId === owner_id,
        )

        navigateToOAuthClient(ids.client_id)
        toast({
          title: ids.client_id,
          message: m.createSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setSubmitting(false)
        setError(error)
        toast({
          title: ids.client_id,
          message: m.createFailure,
          type: toast.types.SUCCESS,
        })
      }
    },
    [userId, navigateToOAuthClient, createOauthClient],
  )

  return (
    <OAuthClientForm
      onSubmit={handleSubmit}
      error={error}
      userId={userId}
      isAdmin={isAdmin}
      rights={rights}
      pseudoRights={pseudoRights}
    />
  )
}

ClientAdd.propTypes = {
  createOauthClient: PropTypes.func.isRequired,
  isAdmin: PropTypes.bool.isRequired,
  navigateToOAuthClient: PropTypes.func.isRequired,
  pseudoRights: PropTypes.rights.isRequired,
  rights: PropTypes.rights,
  userId: PropTypes.string.isRequired,
}

ClientAdd.defaultProps = {
  rights: undefined,
}

export default connect(null, dispatch => ({
  navigateToOAuthClient: clientId => dispatch(push(`/oauth-clients/${clientId}`)),
  createOauthClient: (owner_id, client, userId) =>
    dispatch(attachPromise(createClient(owner_id, client, userId))),
}))(ClientAdd)
