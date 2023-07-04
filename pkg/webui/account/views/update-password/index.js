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
import { Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { defineMessages } from 'react-intl'

import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import ChangePasswordForm from '@account/containers/change-password-form'

import style from '@account/views/front/front.styl'

import useRequest from '@ttn-lw/lib/hooks/use-request'
import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getIsConfiguration } from '@account/store/actions/identity-server'

const m = defineMessages({
  sessionRevoked: 'Your password was changed and all active sessions were revoked',
})

const siteName = selectApplicationSiteName()

const UpdatePassword = () => {
  const [fetching, error] = useRequest(getIsConfiguration())
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  if (Boolean(error)) {
    throw error
  }

  const handleSubmitSuccess = useCallback(
    revokeSession => {
      navigate('/login', {
        state: {
          info: revokeSession ? m.sessionRevoked : sharedMessages.passwordChanged,
        },
      })
    },
    [navigate],
  )

  const userParam = searchParams.get('user')
  const currentParam = searchParams.get('current')

  if (!Boolean(userParam) || !Boolean(currentParam)) {
    return <Navigate to={{ pathname: '/' }} />
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

export default UpdatePassword
