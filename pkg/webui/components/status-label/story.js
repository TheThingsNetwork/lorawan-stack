// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import StatusLabel from '.'

export default {
  title: 'Status Label',
  component: StatusLabel,
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/file/7pBLWK4tsjoAbyJq2viMAQ/console-redesign?type=design&node-id=1599-8145&mode=design&t=2KlaQGRV9FQm7Nv3-4',
    },
  },
}

export const Info = () => <StatusLabel type="info" content="Info" />
export const Warning = () => <StatusLabel type="warning" content="Warning" />
export const Error = () => <StatusLabel type="error" content="Error" />
export const Success = () => <StatusLabel type="success" content="Success" />
