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

import React, { PureComponent, Fragment } from 'react'
import Query from 'query-string'
import { defineMessages } from 'react-intl'
import { connect } from 'react-redux'
import { replace } from 'connected-react-router'
import bind from 'autobind-decorator'

import api from '@oauth/api'

import Modal from '@ttn-lw/components/modal'
import Icon from '@ttn-lw/components/icon'

import ErrorMessage from '@ttn-lw/lib/components/error-message'
import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import { withEnv } from '@ttn-lw/lib/components/env'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import getCookieValue from '@ttn-lw/lib/cookie'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './authorize.styl'

const m = defineMessages({
  modalTitle: 'Request for permission',
  modalSubtitle: '{clientName} is requesting to be granted the following rights:',
  loginInfo: 'You are logged in as {userId}.',
  redirectInfo: 'You will be redirected to {redirectUri}',
  authorize: 'Authorize',
  noDescription: 'This client does not provide a description',
})

@connect(
  undefined,
  dispatch => ({
    redirectToLogin: () => dispatch(replace('/login')),
  }),
)
@withEnv
export default class Authorize extends PureComponent {
  static propTypes = {
    env: PropTypes.env,
    location: PropTypes.location.isRequired,
    redirectToLogin: PropTypes.func.isRequired,
  }

  static defaultProps = {
    env: undefined,
  }

  @bind
  async handleLogout() {
    const { redirectToLogin } = this.props
    await api.oauth.logout()
    redirectToLogin()
  }

  render() {
    const {
      env: {
        pageData: { client, user, error },
      },
      location,
    } = this.props
    const { redirect_uri } = Query.parse(location.search)

    if (error) {
      return <ErrorMessage content={error} />
    }

    const redirectUri = redirect_uri || client.redirect_uris[0]
    const clientName = client.name || capitalize(client.ids.client_id)

    const bottomLine = (
      <div>
        <span>
          <Message
            className={style.loginInfo}
            content={m.loginInfo}
            values={{ userId: user.ids.user_id }}
          />{' '}
          <Message
            content={sharedMessages.logout}
            component="a"
            href="#"
            onClick={this.handleLogout}
          />
        </span>
        <Message content={m.redirectInfo} values={{ redirectUri }} />
      </div>
    )

    return (
      <Fragment>
        <IntlHelmet title={m.authorize} />
        <Modal
          title={m.modalTitle}
          subtitle={{ ...m.modalSubtitle, values: { clientName } }}
          bottomLine={bottomLine}
          buttonMessage={m.authorize}
          method="POST"
          formName="authorize"
          approval
          logo
        >
          <Fragment>
            <input type="hidden" name="csrf" value={getCookieValue('_oauth_csrf')} />
            <div className={style.left}>
              <ul>
                {client.rights.map(right => (
                  <li key={right}>
                    <Icon icon="check" className={style.icon} />
                    <Message content={{ id: `enum:${right}` }} firstToUpper />
                  </li>
                ))}
              </ul>
            </div>
            <div className={style.right}>
              <h3>{clientName}</h3>
              <p>
                {Boolean(client.description) ? (
                  client.description
                ) : (
                  <Message className={style.noDescription} content={m.noDescription} />
                )}
              </p>
            </div>
          </Fragment>
        </Modal>
      </Fragment>
    )
  }
}

// Capitalize the client_id until we have a display name field.
function capitalize(string) {
  return string.charAt(0).toUpperCase() + string.slice(1)
}
