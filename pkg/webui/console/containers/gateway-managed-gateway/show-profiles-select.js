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
import { defineMessages } from 'react-intl'
import PropTypes from 'prop-types'

import Select from '@ttn-lw/components/select'
import Form from '@ttn-lw/components/form'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

const m = defineMessages({
  showProfilesOf: 'Show profiles of',
  yourself: 'Yourself',
})

const profileOptions = [
  { value: false, label: m.yourself },
  { value: true, label: 'TTI' },
]
const ShowProfilesSelect = ({ name }) => (
  <Form.Field
    name={name}
    title={m.showProfilesOf}
    component={Select}
    options={profileOptions}
    tooltipId={tooltipIds.GATEWAY_SHOW_PROFILES}
  />
)

ShowProfilesSelect.propTypes = {
  name: PropTypes.string.isRequired,
}

export default ShowProfilesSelect
