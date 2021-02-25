// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { Redirect } from 'react-router-dom'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'
import { useDispatch } from 'react-redux'
import queryString from 'query-string'

import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import ChangePasswordForm from '@account/containers/change-password-form'

import style from '@account/views/front/front.styl'

import useRequest from '@ttn-lw/lib/hooks/use-request'
import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getIsConfiguration } from '@account/store/actions/identity-server'

const m = defineMessages({
  sessionRevoked: 'Your password was changed and all active sessions were revoked',
})

const siteName = selectApplicationSiteName()

const UpdatePassword = ({ location }) => {
  const [fetching, error] = useRequest(getIsConfiguration())

  if (Boolean(error)) {
    throw error
  }

  const dispatch = useDispatch()
  const handleSubmitSuccess = useCallback(
    revokeSession => {
      dispatch(
        push('/login', { info: revokeSession ? m.sessionRevoked : sharedMessages.passwordChanged }),
      )
    },
    [dispatch],
  )

  const { user: userParam, current: currentParam } = queryString.parse(location.search)
  if (!Boolean(userParam) || !Boolean(currentParam)) {
    return <Redirect to={{ pathname: '/' }} />
  }

  if (fetching) {
    return (
      <Spinner center>
        <Message content={sharedMessages.fetching} />
      </Spinner>
    )
  }

  return (
    <div className={style.form}>
      <IntlHelmet title={m.forgotPassword} />
      <h1 className={style.title}>
        {siteName}
        <br />
        <Message component="strong" content={sharedMessages.changePassword} />
      </h1>
      <hr className={style.hRule} />
      <ChangePasswordForm
        userId={userParam}
        old={currentParam}
        cancelRoute="/login"
        onSubmitSuccess={handleSubmitSuccess}
      />
    </div>
  )
}

UpdatePassword.propTypes = {
  location: PropTypes.location.isRequired,
}

export default UpdatePassword
