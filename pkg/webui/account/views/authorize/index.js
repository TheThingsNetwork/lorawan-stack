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
import { defineMessages } from 'react-intl'
import { useDispatch } from 'react-redux'
import { useSearchParams } from 'react-router-dom'

import Modal from '@ttn-lw/components/modal'
import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import ErrorMessage from '@ttn-lw/lib/components/error-message'
import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import Logo from '@account/containers/logo'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import {
  selectCSRFToken,
  selectPageData,
  selectApplicationRootPath,
} from '@ttn-lw/lib/selectors/env'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { logout } from '@account/store/actions/user'

import style from './authorize.styl'

const m = defineMessages({
  modalTitle: 'Request for permission',
  modalSubtitle: '{clientName} is requesting to be granted the following rights:',
  loginInfo: 'You are logged in as {userId}.',
  redirectInfo: 'You will be redirected to {redirectUri}',
  authorize: 'Authorize {clientName}',
  noDescription: 'This client does not provide a description',
  allRights:
    'This client is requesting <b>all possible current and future rights</b>. This includes reading, writing and deletion of gateways, end devices and applications, as well as their network traffic.',
})

const capitalize = string => string.charAt(0).toUpperCase() + string.slice(1)

const pageData = selectPageData()
const csrfToken = selectCSRFToken()

const Authorize = () => {
  const dispatch = useDispatch()
  const handleLogout = useCallback(async () => {
    await dispatch(attachPromise(logout()))
    window.location = `${selectApplicationRootPath()}/login`
  }, [dispatch])

  const { client, user, error } = pageData
  const [searchParams] = useSearchParams()

  if (error) {
    return <ErrorMessage content={error} />
  }

  const redirectUri = searchParams.get('redirect_uri') || client.redirect_uris[0]
  const clientName = client.name || capitalize(client.ids.client_id)

  const bottomLine = (
    <div>
      <span>
        <Message
          className={style.loginInfo}
          content={m.loginInfo}
          values={{ userId: user.name || user.ids.user_id }}
        />{' '}
        <Button
          message={sharedMessages.logout}
          type="button"
          onClick={handleLogout}
          className={style.logoutButton}
          unstyled
        />
      </span>
      <Message content={m.redirectInfo} values={{ redirectUri }} />
    </div>
  )

  return (
    <div className={style.container}>
      <IntlHelmet title={m.authorize} values={{ clientName }} />
      <Modal
        title={m.modalTitle}
        subtitle={{ ...m.modalSubtitle, values: { clientName } }}
        bottomLine={bottomLine}
        buttonMessage={{ ...m.authorize, values: { clientName } }}
        method="POST"
        formName="authorize"
        approval
        logo={<Logo />}
      >
        <>
          <input type="hidden" name="_csrf" value={csrfToken} />
          <div className={style.left}>
            <ul>
              {client.rights.map(right => (
                <li key={right}>
                  <Icon icon="check" className={style.icon} />
                  <Message content={{ id: `enum:${right}` }} firstToUpper />
                </li>
              ))}
              {client.rights.length === 1 && client.rights[0] === 'RIGHT_ALL' && (
                <Message
                  className={style.noteText}
                  values={{ b: str => <b key="bold">{str}</b> }}
                  content={m.allRights}
                />
              )}
            </ul>
          </div>
          <div className={style.right}>
            <h3>{clientName}</h3>
            <p>
              {Boolean(client.description) ? (
                client.description
              ) : (
                <Message className={style.noteText} content={m.noDescription} />
              )}
            </p>
          </div>
        </>
      </Modal>
    </div>
  )
}

export default Authorize
