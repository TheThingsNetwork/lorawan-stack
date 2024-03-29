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

import PropTypes from '@ttn-lw/lib/prop-types'

import DescriptionList from './shared/description-list'
import JSONPayload from './shared/json-payload'

const Value = React.memo(({ event }) => {
  const { data } = event
  return (
    <DescriptionList>
      {Array.isArray(data.value) || typeof data.value === 'object' ? (
        <JSONPayload data={data.value} />
      ) : (
        <DescriptionList.Item data={data.value} />
      )}
    </DescriptionList>
  )
})

Value.propTypes = {
  event: PropTypes.event.isRequired,
}

export default Value
