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

import style from './story.styl'
import Icon from '.'

const icons = [
  'devices',
  'integration',
  'settings',
  'lock',
  'lock_open',
  'close',
  'menu',
  'dashboard',
  'transform',
  'data',
  'sort',
  'overview',
  'application',
  'gateway',
  'organization',
]

const doc = `Icons can be used using \`display: {flex|inline-block}\`.
\`inline-block\` is used by default. To use \`flex\` instead, overwrite the
display value of the wrapping \`<span />\`in your local scoped css. The
positioning will differ slightly, so the nudge props can be used to fine-tune
the appearance.`

storiesOf('Icon', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      propTables: [Icon],
      text: doc,
    })(story)(context),
  )
  .add('Icons', () =>
    icons.map(function(icon) {
      return (
        <div className={style.wrapper} key={icon}>
          <Icon icon={icon} />
          {icon}
        </div>
      )
    }),
  )
  .add('Usage', () => (
    <div className={style.wrapper}>
      <div className={style.block}>
        <Icon icon="devices" />
        <span>{'display: inline-block'}</span>
      </div>
      <br />
      <div className={style.flex}>
        <Icon icon="devices" />
        <span>{'display: flex'}</span>
      </div>
    </div>
  ))
