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
  PHY_V1_0,
  PHY_V1_0_1,
  PHY_V1_0_2_REV_A,
  PHY_V1_0_2_REV_B,
  PHY_V1_0_3_REV_A,
  PHY_V1_1_REV_A,
  PHY_V1_1_REV_B,
  MAC_V1_0,
  MAC_V1_0_1,
  MAC_V1_0_2,
  MAC_V1_0_3,
  MAC_V1_0_4,
  MAC_V1_1,
  LORAWAN_VERSIONS,
} from '@console/lib/device-utils'

const phyVersionsMap = {
  [PHY_V1_0.value]: [MAC_V1_0, MAC_V1_0_4],
  [PHY_V1_0_1.value]: [MAC_V1_0_1, MAC_V1_0_4],
  [PHY_V1_0_2_REV_A.value]: [MAC_V1_0_2, MAC_V1_0_4],
  [PHY_V1_0_2_REV_B.value]: [MAC_V1_0_2, MAC_V1_0_4],
  [PHY_V1_0_3_REV_A.value]: [MAC_V1_0_3, MAC_V1_0_4],
  [PHY_V1_1_REV_A.value]: [MAC_V1_1, MAC_V1_0_4],
  [PHY_V1_1_REV_B.value]: [MAC_V1_1, MAC_V1_0_4],
}

const LorawanVersionInput = props => {
  const { phyVersion, onChange, value, ...rest } = props

  const [options, setOptions] = React.useState(LORAWAN_VERSIONS)
  React.useEffect(() => {
    if (phyVersion) {
      setOptions(phyVersionsMap[phyVersion])
    } else {
      setOptions(LORAWAN_VERSIONS)
    }
  }, [phyVersion])

  return <Select onChange={onChange} value={value} options={options} {...rest} />
}

LorawanVersionInput.propTypes = {
  onChange: PropTypes.func.isRequired,
  phyVersion: PropTypes.string,
  value: PropTypes.string,
}

LorawanVersionInput.defaultProps = {
  phyVersion: undefined,
  value: undefined,
}

export default LorawanVersionInput
