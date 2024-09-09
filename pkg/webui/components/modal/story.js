// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import Logo from '@assets/static/logo.svg'

import Icon, { IconCheck } from '@ttn-lw/components/icon'

import style from './story.styl'

import Modal from '.'

const bottomLine = (
  <div>
    <span className={style.loginInfo}>
      You are logged in as John Doe. <a href="#">Logout</a>
    </span>
    <span className={style.redirectInfo}>
      You will be redirected to <span>/test/test</span>
    </span>
  </div>
)

export default {
  title: 'Modal',
  component: Modal,
  parameters: {
    docs: {
      description: {
        component: `The modal can be displayed inline or portalled via \`<PortalledModal />\``,
      },
    },
  },
}

export const BasicModal = () => (
  <Modal title="Example Modal" message="This is something you need to know!" inline />
)

export const NoTitle = () => (
  <Modal message="This modal has no title. Might be useful in some situations." inline />
)

export const OAuthAuthorizeExample = () => (
  <Modal
    title="Request for Permission"
    subtitle="Console is requesting permission to do the following:"
    bottomLine={bottomLine}
    buttonMessage="Allow"
    approval
    logo={Logo}
    inline
  >
    <div className={style.left}>
      <ul>
        <li>
          <Icon icon={IconCheck} className={style.icon} />
          View your profile
        </li>
        <li>
          <Icon icon={IconCheck} className={style.icon} />
          Make changes to your profile
        </li>
        <li>
          <Icon icon={IconCheck} className={style.icon} />
          Perform administrative action
        </li>
        <li>
          <Icon icon={IconCheck} className={style.icon} />
          List your applications
        </li>
        <li>
          <Icon icon={IconCheck} className={style.icon} />
          Degister new gateways in your account
        </li>
        <li>
          <Icon icon={IconCheck} className={style.icon} />
          Create and edit end devices of your applications
        </li>
      </ul>
    </div>
    <div className={style.right}>
      <h3>
        Console <span title="This application is an official application">Official</span>
      </h3>
      <p>The Console is The Things Stack's official web application.</p>
    </div>
  </Modal>
)

OAuthAuthorizeExample.story = {
  name: 'OAuth Authorize Example',
}

export const AsOverlay = () => (
  <Modal title="Example Modal" message="This is something you need to know!" />
)
