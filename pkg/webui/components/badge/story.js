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

import { IconArrowDown, IconCheck } from '@ttn-lw/components/icon'

import Badge from '.'

export default {
  title: 'Badge',
  component: Badge,
}

export const Success = () => <Badge status="success">Success</Badge>

export const Error = () => <Badge status="error">Error</Badge>

export const Info = () => <Badge status="info">Info</Badge>

export const Warning = () => <Badge status="warning">Warning</Badge>

export const WithStartIcon = () => (
  <Badge status="error" startIcon={IconArrowDown}>
    -2
  </Badge>
)

export const WithEndIcon = () => (
  <Badge status="success" endIcon={IconCheck}>
    Done
  </Badge>
)
