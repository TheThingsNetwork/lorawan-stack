// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import { connect } from 'react-redux'

import api from '../../../api'

import Modal from '../../../components/modal'
import Spinner from '../../../components/spinner'
import Icon from '../../../components/icon'
import Message from '../../../components/message'

import { getClient } from '../../../actions/client'

import style from './authorize.styl'

@connect(function (state, props) {
  const { client_id, redirect_uri } = Query.parse(props.location.search)

  return {
    user: state.user.user,
    client_id,
    redirectUri: redirect_uri,
    client: state.client[client_id] && state.client[client_id].client,
    fetching: state.client[client_id] && state.client[client_id].fetching,
  }
}
)
export default class Authorize extends PureComponent {

  componentDidMount () {
    const { dispatch, client_id } = this.props

    dispatch(getClient(client_id))
  }

  async handleLogout () {
    await api.oauth.logout()
    window.location = '/oauth/login'
  }

  render () {
    const {
      client,
      client_id,
      redirectUri,
      user,
    } = this.props

    const bottomLine = (
      <div>
        <span className={style.loginInfo}>You are logged in as {user.user_id}. <a href="#" onClick={this.handleLogout}>Logout</a></span>
        <span className={style.redirectInfo}>You will be redirected to <span>{redirectUri}</span></span>
      </div>
    )

    if (!client || client.fetching) {
      return <Spinner center children="Please wait…" />
    }

    return (
      <Modal
        title="Request for Permission"
        subtitle={`${capitalize(client_id)} is asking permission to do the following:`}
        bottomLine={bottomLine}
        buttonMessage="Allow"
        method="POST"
        formName="authorize"
        approval
        logo
      >
        <Fragment>
          <div className={style.left}>
            <ul>
              { client.rights.map(right => (
                <li key={right}><Icon icon="check" className={style.icon} /><Message content={right} /></li>
              )
              )}
            </ul>
          </div>
          <div className={style.right}>
            <h3>{capitalize(client_id)}</h3>
            <p>{client.description}</p>
          </div>
        </Fragment>
      </Modal>
    )
  }
}

// Capitalize the client_id until we have a display name field.
function capitalize (string) {
  return string.charAt(0).toUpperCase() + string.slice(1)
}
