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

import Button from '@ttn-lw/components/button'

import SubmitBar from '.'

export default {
  title: 'SubmitBar',
}

export const OnlySubmit = () => (
  <SubmitBar>
    <Button message="Save Changes" icon="done" />
  </SubmitBar>
)

export const SubmitAndReset = () => (
  <SubmitBar>
    <Button message="Save Changes" icon="done" />
    <Button message="Delete" icon="delete" naked danger />
  </SubmitBar>
)

SubmitAndReset.story = {
  name: 'Submit and Reset',
}

export const SubmitAndText = () => (
  <SubmitBar align="start">
    <Button message="Save Changes" icon="done" />
    <SubmitBar.Message content="Note: End device level message payload formats take precedence over application level message payload formats" />
  </SubmitBar>
)

SubmitAndText.story = {
  name: 'Submit and Text',
}
