// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'

import Notification from '@ttn-lw/components/notification'
import Button from '@ttn-lw/components/button'

import RegistryTotals from './registry-totals'

const m = defineMessages({
  openSourceInfo:
    'You are currently using The Things Stack Open Source. More features can be unlocked by using The Things Stack Cloud.',
  plansButton: 'Get started with The Things Stack Cloud',
})

const NetworkInformationContainer = () => (
  <>
    <Notification
      info
      small
      content={m.openSourceInfo}
      children={
        <Button.AnchorLink
          primary
          href={'https://www.thethingsindustries.com/stack/plans/'}
          message={m.plansButton}
          target="_blank"
          external
        />
      }
      className="mt-cs-l"
    />
    <RegistryTotals />
  </>
)

export default NetworkInformationContainer
