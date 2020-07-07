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

import React from 'react'
import { defineMessages } from 'react-intl'
import queryString from 'query-string'

import api from '@account/api'

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

export default class Validate extends React.PureComponent {
  static propTypes = {
    location: PropTypes.location.isRequired,
  }

  state = {
    error: undefined,
    success: undefined,
    fetching: true,
  }

  handleError(error) {
    this.setState({ error, fetching: false, success: undefined })
  }

  handleSuccess() {
    this.setState({ success: m.validateSuccess, fetching: false, error: undefined })
  }

  async componentDidMount() {
    const validationData = queryString.parse(this.props.location.search)
    try {
      await api.users.validate({
        token: validationData.token,
        id: validationData.reference,
      })
      this.handleSuccess()
    } catch (error) {
      if (isNotFoundError(error)) {
        this.handleError(createFrontendError(m.tokenNotFoundTitle, m.tokenNotFoundMessage))
      } else {
        this.handleError(error)
      }
    }
  }

  render() {
    const { fetching, error, success } = this.state

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
}
