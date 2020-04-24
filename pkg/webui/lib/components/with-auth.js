// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
import { withRouter, Redirect } from 'react-router-dom'
import { defineMessages } from 'react-intl'

import Spinner from '@ttn-lw/components/spinner'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectApplicationRootPath } from '@ttn-lw/lib/selectors/env'
import { createFrontendError } from '@ttn-lw/lib/errors/utils'

import Message from './message'

// Define a minimum set of rights, without which it makes no sense to access the
// console.
const minimumRights = ['RIGHT_APPLICATION', 'RIGHT_GATEWAY', 'RIGHT_ORGANIZATION']

const m = defineMessages({
  errTooFewRights: 'Your account does not possess sufficient rights to use the console.',
  errTooFewRightsTitle: 'Insufficient rights',
  errStateRequested:
    'Your account still needs to be approved by an administrator. You will receive a confirmation email once your account is approved.',
  errStateRequestedTitle: 'Account unapproved',
  errStateRejected: 'Your account has been rejected by an administrator.',
  errStateRejectedTitle: 'Account rejected',
  errStateSuspended:
    'Your account has been suspended by an administrator. Please contact support for further information about your account status.',
  errStateSuspendedTitle: 'Account suspended',
  errEmailValidation: 'Your account is restricted until your email address has been validated.',
  errEmailValidationTitle: 'Email validation pending',
})

/**
 * `Auth` is a component that wraps a tree that requires the user to be
 * authenticated. If no user is authenticated, it redirects to the OAuth login.
 */
@withRouter
class Auth extends React.PureComponent {
  static propTypes = {
    children: PropTypes.node.isRequired,
    errorComponent: PropTypes.elementType.isRequired,
    fetching: PropTypes.bool.isRequired,
    isAdmin: PropTypes.bool,
    location: PropTypes.location.isRequired,
    rights: PropTypes.rights,
    user: PropTypes.user,
    userError: PropTypes.error,
  }
  static defaultProps = {
    user: undefined,
    isAdmin: undefined,
    rights: undefined,
    userError: undefined,
  }

  render() {
    const {
      user,
      fetching,
      userError,
      errorComponent,
      children,
      location,
      rights,
      isAdmin,
    } = this.props

    if (fetching) {
      return (
        <Spinner center>
          <Message content={sharedMessages.fetching} />
        </Spinner>
      )
    }

    let error

    if (userError) {
      error = userError
    } else if (
      // Check whether the user has at least basic rights, without which it
      // makes no sense to access the console.
      Boolean(user) &&
      !isAdmin &&
      !rights.some(r => minimumRights.some(mr => r.startsWith(mr)))
    ) {
      // Provide relevant error messages if possible.
      if (user.state === 'STATE_REQUESTED') {
        error = createFrontendError(m.errStateRequestedTitle, m.errStateRequested)
      } else if (user.state === 'STATE_REJECTED') {
        error = createFrontendError(m.errStateRejectedTitle, m.errStateRejected)
      } else if (user.state === 'STATE_SUSPENDED') {
        error = createFrontendError(m.errStateSuspendedTitle, m.errStateSuspended)
      } else if (!user.primary_email_address_validated_at) {
        error = createFrontendError(m.errEmailValidationTitle, m.errEmailValidation)
      } else {
        error = createFrontendError(m.errTooFewRightsTitle, m.errTooFewRights)
      }
    }

    if (error) {
      // Redirect to root to prevent side effects.
      if (location.pathname !== '/') {
        return <Redirect to="" />
      }

      const Component = errorComponent
      return <Component error={error} />
    }

    if (!Boolean(user)) {
      // If the user is logged out, redirect to the login endpoint and show a
      // loading spinner.
      window.location = `${selectApplicationRootPath()}/login/ttn-stack?next=${location.pathname}`
      return (
        <Spinner after={0} center>
          <Message content={sharedMessages.redirecting} />
        </Spinner>
      )
    }

    return children
  }
}

export default Auth
