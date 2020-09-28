// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
  PHY_V1_0,
  PHY_V1_0_1,
  PHY_V1_0_2_REV_A,
  PHY_V1_0_2_REV_B,
  PHY_V1_0_3_REV_A,
  PHY_V1_1_REV_A,
  PHY_V1_1_REV_B,
  LORAWAN_PHY_VERSIONS,
  parseLorawanMacVersion,
} from '@console/lib/device-utils'

const lorawanVersionPairs = {
  100: [PHY_V1_0],
  101: [PHY_V1_0_1],
  102: [PHY_V1_0_2_REV_A, PHY_V1_0_2_REV_B],
  103: [PHY_V1_0_3_REV_A],
  104: LORAWAN_PHY_VERSIONS,
  110: [PHY_V1_1_REV_A, PHY_V1_1_REV_B],
  0: LORAWAN_PHY_VERSIONS,
}

const getOptions = lwVersion => lorawanVersionPairs[parseLorawanMacVersion(lwVersion)]

const PhyVersionInput = props => {
  const { lorawanVersion, onChange, disabled, value, ...rest } = props

  const [phyVersions, setPhyVersions] = React.useState(getOptions(lorawanVersion))

  const lorawanVersionRef = React.useRef(lorawanVersion)
  React.useEffect(() => {
    const options = getOptions(lorawanVersion)
    setPhyVersions(options)

    if (!value && options.length <= 1) {
      onChange(options[0].value)
    } else if (lorawanVersion !== lorawanVersionRef.current) {
      onChange(options[0].value)
      lorawanVersionRef.current = lorawanVersion
    }
  }, [lorawanVersion, onChange, value])

  return (
    <Select
      options={phyVersions}
      onChange={onChange}
      disabled={phyVersions.length <= 1 || disabled}
      value={value}
      {...rest}
    />
  )
}

PhyVersionInput.propTypes = {
  disabled: PropTypes.bool,
  lorawanVersion: PropTypes.string,
  onChange: PropTypes.func.isRequired,
  value: PropTypes.string,
}

PhyVersionInput.defaultProps = {
  lorawanVersion: undefined,
  disabled: false,
  value: undefined,
}

export default PhyVersionInput
