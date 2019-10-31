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

storiesOf('Notification/Notification', module)
  .add('Default', () => (
    <div>
      <Notification title="example message title" content="This is an example message" />
      <Notification content="This is an example message" />
      <Notification title="example message title" content="This is an example message" small />
      <Notification content="This is an example message" small />
    </div>
  ))
  .add('Info', () => (
    <div>
      <Notification title="example message title" info content="This message is good to know" />
      <Notification info content="This message is good to know" />
      <Notification
        title="example message title"
        info
        content="This message is good to know"
        small
      />
      <Notification info content="This message is good to know" small />
    </div>
  ))
  .add('Warning', () => (
    <div>
      <Notification
        title="example message title"
        warning
        content="This issue should be addressed!"
      />
      <Notification warning content="This issue should be addressed!" />
      <Notification
        title="example message title"
        warning
        content="This issue should be addressed!"
        small
      />
      <Notification warning content="This issue should be addressed!" small />
    </div>
  ))
  .add('Success', () => (
    <div>
      <Notification title="example message title" success content="Successful action!" />
      <Notification success content="Successful action!" />
      <Notification title="example message title" success content="Successful action!" small />
      <Notification success content="Successful action!" small />
    </div>
  ))
