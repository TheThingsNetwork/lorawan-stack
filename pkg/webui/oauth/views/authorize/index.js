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

import api from '../../api'
import sharedMessages from '../../../lib/shared-messages'

import ErrorMessage from '../../../lib/components/error-message'
import Modal from '../../../components/modal'
import Icon from '../../../components/icon'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { withEnv } from '../../../lib/components/env'
import getCookieValue from '../../../lib/cookie'

import style from './authorize.styl'

const m = defineMessages({
  modalTitle: 'Request for Permission',
  modalSubtitle: '{clientName} is requesting permissions to do the following:',
  loginInfo: 'You are logged in as {userId}.',
  redirectInfo: 'You will be redirected to {redirectUri}',
  authorize: 'Authorize',
})

@connect(undefined, dispatch => ({
  redirectToLogin: () => dispatch(replace('/login')),
}))
@withEnv
@bind
export default class Authorize extends PureComponent {

  async handleLogout () {
    const { redirectToLogin } = this.props
    await api.oauth.logout()
    redirectToLogin()
  }

  render () {
    const { env: { page_data: { client, user, error }}, location } = this.props
    const { redirect_uri } = Query.parse(location.search)

    if (error) {
      return <ErrorMessage content={error} />
    }

    const redirectUri = redirect_uri || client.redirect_uris[0]
    const clientName = capitalize(client.ids.client_id)

    const bottomLine = (
      <div>
        <span>
          <Message
            className={style.loginInfo}
            content={m.loginInfo}
            values={{ userId: user.ids.user_id }}
          />
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
          subtitle={{ ...m.modalSubtitle, values: { clientName }}}
          bottomLine={bottomLine}
          buttonMessage={m.authorize}
          method="POST"
          formName="authorize"
          approval
          logo
        >
          <Fragment>
            <input type="hidden" name="csrf" value={getCookieValue('_csrf')} />
            <div className={style.left}>
              <ul>
                { client.rights.map(right => (
                  <li key={right}>
                    <Icon icon="check" className={style.icon} />
                    <Message content={{ id: `enum:${right}` }} />
                  </li>
                )
                )}
              </ul>
            </div>
            <div className={style.right}>
              <h3>{capitalize(client.ids.client_id)}</h3>
              <p>{client.description}</p>
            </div>
          </Fragment>
        </Modal>
      </Fragment>
    )
  }
}

// Capitalize the client_id until we have a display name field.
function capitalize (string) {
  return string.charAt(0).toUpperCase() + string.slice(1)
}
