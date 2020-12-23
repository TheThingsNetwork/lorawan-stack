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

import React, { useState, useCallback, useEffect } from 'react'
import { defineMessages } from 'react-intl'
import queryString from 'query-string'

import tts from '@account/api/tts'

import Spinner from '@ttn-lw/components/spinner'
import ErrorNotification from '@ttn-lw/components/error-notification'
import Notification from '@ttn-lw/components/notification'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import style from '@account/views/front/front.styl'

import PropTypes from '@ttn-lw/lib/prop-types'
import { isNotFoundError, createFrontendError } from '@ttn-lw/lib/errors/utils'
import { selectApplicationSiteName, selectApplicationSiteTitle } from '@ttn-lw/lib/selectors/env'

const m = defineMessages({
  backToAccount: 'Back to Account',
  contactInfoValidation: 'Contact info validation',
  validateSuccess: 'Validation successful',
  validateFail: 'There was an error and the contact info could not be validated',
  validatingAccount: 'Validating account…',
  tokenNotFoundTitle: 'Token not found',
  tokenNotFoundMessage:
    'The validation token was not found. This could mean that the contact info has already been validated. Otherwise, please contact an administrator.',
})

const siteName = selectApplicationSiteName()
const siteTitle = selectApplicationSiteTitle()

const Validate = ({ location }) => {
  const [error, setError] = useState(undefined)
  const [success, setSuccess] = useState(undefined)
  const [fetching, setFetching] = useState(true)

  const handleError = useCallback(error => {
    setError(error)
    setFetching(false)
    setSuccess(undefined)
  }, [])

  const handleSuccess = useCallback(() => {
    setError(undefined)
    setFetching(false)
    setSuccess(m.validateSuccess)
  }, [])

  useEffect(() => {
    const makeRequest = async () => {
      const validationData = queryString.parse(location.search)
      try {
        await tts.ContactInfo.validate({
          token: validationData.token,
          id: validationData.reference,
        })
        handleSuccess()
      } catch (error) {
        if (isNotFoundError(error)) {
          handleError(createFrontendError(m.tokenNotFoundTitle, m.tokenNotFoundMessage))
        } else {
          handleError(error)
        }
      }
    }
    makeRequest()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className={style.form}>
      <IntlHelmet title={m.contactInfoValidation} />
      <h1 className={style.title}>
        {siteName}
        <br />
        <Message component="strong" content={m.contactInfoValidation} />
      </h1>
      <hr className={style.hRule} />
      {fetching ? (
        <Spinner after={0} faded className={style.spinner}>
          <Message content={m.validatingAccount} />
        </Spinner>
      ) : (
        <>
          {error && <ErrorNotification small content={error} title={m.validateFail} />}
          {success && <Notification small success content={success} title={m.validateSuccess} />}
        </>
      )}
      <Button.Link
        to="/"
        icon="keyboard_arrow_left"
        message={{ ...m.backToAccount, values: { siteTitle } }}
      />
    </div>
  )
}

Validate.propTypes = {
  location: PropTypes.location.isRequired,
}

export default Validate
