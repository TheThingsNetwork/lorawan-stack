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

import Button from '../button'
import SubmitBar from '.'

storiesOf('SubmitBar', module)
  .add('Only Submit', () => (
    <SubmitBar>
      <Button message="Save Changes" icon="done" />
    </SubmitBar>
  ))
  .add('Submit and Reset', () => (
    <SubmitBar>
      <Button message="Save Changes" icon="done" />
      <Button message="Delete" icon="delete" naked danger />
    </SubmitBar>
  ))
  .add('Submit and Text', () => (
    <SubmitBar align="start">
      <Button message="Save Changes" icon="done" />
      <SubmitBar.Message content="Note: Device level message payload formats take precedence over application level message payload formats." />
    </SubmitBar>
  ))
