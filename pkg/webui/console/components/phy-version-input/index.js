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

import Select from '@ttn-lw/components/select'

import PropTypes from '@ttn-lw/lib/prop-types'

import {
  LORAWAN_PHY_VERSIONS,
  parseLorawanMacVersion,
  LORAWAN_VERSION_PAIRS,
} from '@console/lib/device-utils'

const PhyVersionInput = props => {
  const { lorawanVersion, onChange, disabled, value, ...rest } = props

  const lorawanVersionRef = React.useRef(lorawanVersion)
  const [options, setOptions] = React.useState(LORAWAN_PHY_VERSIONS)

  React.useEffect(() => {
    const options =
      LORAWAN_VERSION_PAIRS[parseLorawanMacVersion(lorawanVersion)] || LORAWAN_PHY_VERSIONS
    if (options.length === 1) {
      onChange(options[0].value)
    } else if (lorawanVersion !== lorawanVersionRef.current) {
      lorawanVersionRef.current = lorawanVersion
      onChange('')
    }

    setOptions(options)
  }, [lorawanVersion, onChange])

  return (
    <Select
      options={options}
      onChange={onChange}
      disabled={options.length <= 1 || disabled}
      value={value}
      {...rest}
    />
  )
}

PhyVersionInput.propTypes = {
  disabled: PropTypes.bool,
  frequencyPlan: PropTypes.string,
  lorawanVersion: PropTypes.string,
  onChange: PropTypes.func.isRequired,
  value: PropTypes.string,
}

PhyVersionInput.defaultProps = {
  disabled: false,
  value: undefined,
  frequencyPlan: '',
  lorawanVersion: '',
}

export default PhyVersionInput
