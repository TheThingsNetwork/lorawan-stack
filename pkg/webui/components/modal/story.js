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
import { storiesOf } from '@storybook/react'
import { withInfo } from '@storybook/addon-info'
import Icon from '../icon'

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

storiesOf('Modal', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      text: 'The modal can be displayed inline or portalled via `<PortalledModal />`',
      propTables: [Modal],
    })(story)(context),
  )
  .add('Basic Modal', () => (
    <Modal title="Example Modal" message="This is something you need to know!" inline />
  ))
  .add('No Title', () => (
    <Modal message="This modal has no title. Might be useful in some situations." inline />
  ))
  .add('OAuth Authorize Example', () => (
    <Modal
      title="Request for Permission"
      subtitle="Console is requesting permission to do the following:"
      bottomLine={bottomLine}
      buttonMessage="Allow"
      approval
      logo
      inline
    >
      <div className={style.left}>
        <ul>
          <li>
            <Icon icon="check" className={style.icon} />
            View your profile
          </li>
          <li>
            <Icon icon="check" className={style.icon} />
            Make changes to your profile
          </li>
          <li>
            <Icon icon="check" className={style.icon} />
            Perform administrative action
          </li>
          <li>
            <Icon icon="check" className={style.icon} />
            List your applications
          </li>
          <li>
            <Icon icon="check" className={style.icon} />
            Degister new gateways in your account
          </li>
          <li>
            <Icon icon="check" className={style.icon} />
            Create and edit devices of your applications
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
  ))
  .add('As Overlay', () => (
    <Modal title="Example Modal" message="This is something you need to know!" />
  ))
