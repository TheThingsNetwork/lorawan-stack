// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import Icon, { IconInfoCircle } from '@ttn-lw/components/icon'

import Tooltip from '.'

const containerStyles = {
  width: '300px',
  margin: '0 auto',
}

const splitterStyles = {
  marginTop: '40px',
}

export default {
  title: 'Tooltip',
}

export const TextContent = () => (
  <div style={containerStyles}>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with top placement" placement="top">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with top-start placement" placement="top-start">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with top-end placement" placement="top-end">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with bottom placement" placement="bottom">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with bottom-start placement" placement="bottom-start">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with bottom-end placement" placement="bottom-end">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with right placement" placement="right">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with right-start placement" placement="right-start">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with right-end placement" placement="right-end">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with left placement" placement="left">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with left-start placement" placement="left-start">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
    <div style={splitterStyles} />
    <Tooltip content="Tooltip with left-end placement" placement="left-end">
      <Icon icon={IconInfoCircle} />
    </Tooltip>
  </div>
)

TextContent.story = {
  name: 'Text content',
}
