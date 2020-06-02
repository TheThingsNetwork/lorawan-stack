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
import { action } from '@storybook/addon-actions'

import Link from '@ttn-lw/components/link'

import Prompt from '.'

const shouldBlockNavigation = location => location.pathname.endsWith('should-block')

const linkContainerStyle = {
  margin: '10px 0',
}

storiesOf('Prompt', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      source: true,
      propTables: [Prompt],
    })(story)(context),
  )
  .add('Default', () => (
    <div>
      Navigate using the links below to trigger the `Prompt` component to appear.
      <div style={linkContainerStyle}>
        <Link to="/should-not-block" title="should not block">
          should not block
        </Link>
      </div>
      <div style={linkContainerStyle}>
        <Link to="/should-block" title="should block">
          should block
        </Link>
      </div>
      <div style={linkContainerStyle}>
        <Link to="/should-block" title="should also block">
          should block
        </Link>
      </div>
      <Prompt
        when
        onApprove={action('onApprove')}
        onCancel={action('onCancel')}
        shouldBlockNavigation={shouldBlockNavigation}
        modal={{
          title: 'Block modal title',
        }}
      >
        <p>
          This modal prompts the user to confirm current route change. By pressing `Cancel` the user
          blocks navigation, while pressing `Approve` allows navigation to happen.
        </p>
      </Prompt>
    </div>
  ))
