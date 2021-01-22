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

import { CreateFrequencyPlansSelect } from '@console/containers/freq-plans-select'

import PropTypes from '@ttn-lw/lib/prop-types'

const FreqPlansSelect = React.memo(props => {
  const { bandId, ...rest } = props

  return React.createElement(
    CreateFrequencyPlansSelect('ns', {
      optionsFormatter: plans => {
        const region = bandId.split('_')[0]
        if (!Boolean(region)) {
          return plans.map(plan => ({ value: plan.id, label: plan.name }))
        }

        return plans
          .filter(plan => plan.id.startsWith(region))
          .map(plan => ({ value: plan.id, label: plan.name }))
      },
    }),
    rest,
  )
})

FreqPlansSelect.propTypes = {
  bandId: PropTypes.string.isRequired,
}

export default FreqPlansSelect
