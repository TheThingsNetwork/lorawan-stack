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
import ReactSwitch from 'react-switch'

import COLORS from '@ttn-lw/constants/colors'

import Icon, { IconCheck, IconX } from '@ttn-lw/components/icon'

const Switch = props => (
  <ReactSwitch
    {...props}
    uncheckedIcon={
      <Icon
        icon={IconX}
        small
        style={{ color: 'white', marginLeft: '3px', marginTop: '4px', fontSize: '1rem' }}
      />
    }
    checkedIcon={
      <Icon
        small
        icon={IconCheck}
        style={{ color: 'white', marginLeft: '5px', marginTop: '4px', fontSize: '1rem' }}
      />
    }
    onColor={COLORS.C_TEXT_BRAND_NORMAL}
    offColor="#B4B7CB"
    activeBoxShadow={`"0 0 3px 5px ${COLORS.C_TEXT_BRAND_NORMAL}66, inset 0 0 3px 1px #0002"`}
    handleDiameter={18}
    height={24}
    width={44}
    data-test-id="switch"
  />
)

Switch.propTypes = ReactSwitch.propTypes
Switch.defaultProps = ReactSwitch.defaultProps

export default Switch
