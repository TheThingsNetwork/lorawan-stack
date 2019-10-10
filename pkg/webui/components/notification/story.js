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

import Notification from '.'
import ErrorNotification from '../error-notification'

storiesOf('Notification', module)
  .add('Default', () => (
    <div>
      <Notification title="example message title" message="This is an example message" />
      <Notification message="This is an example message" />
      <Notification title="example message title" message="This is an example message" small />
      <Notification message="This is an example message" small />
    </div>
  ))
  .add('Info', () => (
    <div>
      <Notification title="example message title" info="This message is good to know" />
      <Notification info="This message is good to know" />
      <Notification title="example message title" info="This message is good to know" small />
      <Notification info="This message is good to know" small />
    </div>
  ))
  .add('Warning', () => (
    <div>
      <Notification title="example message title" warning="This issue should be addressed!" />
      <Notification warning="This issue should be addressed!" />
      <Notification title="example message title" warning="This issue should be addressed!" small />
      <Notification warning="This issue should be addressed!" small />
    </div>
  ))
  .add('Error', () => (
    <div>
      <Notification title="example message title" error="We got a problem here!" />
      <ErrorNotification error="We got a problem here!" />
      <Notification title="example message title" error="We got a problem here!" small />
      <Notification error="We got a problem here!" small />
    </div>
  ))
  .add('Success', () => (
    <div>
      <Notification title="example message title" success="Successful action!" />
      <Notification success="Successful action!" />
      <Notification title="example message title" success="Successful action!" small />
      <Notification success="Successful action!" small />
    </div>
  ))
