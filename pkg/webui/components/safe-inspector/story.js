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

import SafeInspector from '.'

storiesOf('Safe Inspector', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      propTables: [SafeInspector],
    })(story)(context),
  )
  .add('Default', () => <SafeInspector data="ab01f46d2f" />)
  .add('Multiple', () => (
    <React.Fragment>
      <SafeInspector data="ab01f46d2f" />
      <br />
      <SafeInspector data="ff0000" />
      <br />
      <SafeInspector data="f8a683c1d9b2" />
    </React.Fragment>
  ))
  .add('Small', () => <SafeInspector small data="ab01f46d2f" />)
  .add('Non Byte', () => (
    <SafeInspector data="somerandomhash.du9d8321-9dk2-edf2398efh8" isBytes={false} />
  ))
  .add('Initially Visible', () => <SafeInspector data="ab01f46d2f" initiallyVisible />)
  .add('Not hideable', () => <SafeInspector data="ab01f46d2f" hideable={false} />)
