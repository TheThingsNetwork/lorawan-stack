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
import { defineMessages } from 'react-intl'

import api from '@console/api'

import Select from '@ttn-lw/components/select'
import toast from '@ttn-lw/components/toast'

import PropTypes from '@ttn-lw/lib/prop-types'

import { LORAWAN_PHY_VERSIONS } from '@console/lib/device-utils'

const m = defineMessages({
  phyVersionError: 'Failed to fetch phy versions',
})

const PhyVersionInput = props => {
  const { onChange, disabled, value, frequencyPlan, ...rest } = props

  const [phyVersions, setPhyVersions] = React.useState([])
  const [options, setOptions] = React.useState(LORAWAN_PHY_VERSIONS)

  React.useEffect(() => {
    const currentPhyVersions = phyVersions.filter(({ band_id }) => frequencyPlan.includes(band_id))

    if (currentPhyVersions.length > 0) {
      const versions = currentPhyVersions[0].phy_versions
      if (versions.lengh === 1) {
        onChange(versions[0])
      }

      const options = versions.map(phyVersion =>
        LORAWAN_PHY_VERSIONS.find(({ value }) => value === phyVersion),
      )
      setOptions(options.length === 0 ? LORAWAN_PHY_VERSIONS : options)
    }
  }, [frequencyPlan, onChange, phyVersions])

  React.useEffect(() => {
    const fetchPhyVersions = async () => {
      try {
        const { version_info } = await api.configuration.getPhyVersions()
        setPhyVersions(version_info)
      } catch (err) {
        toast({
          type: toast.types.ERROR,
          message: m.phyVersionError,
        })
      }
    }

    fetchPhyVersions()
  }, [])

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
  onChange: PropTypes.func.isRequired,
  value: PropTypes.string,
}

PhyVersionInput.defaultProps = {
  disabled: false,
  value: undefined,
  frequencyPlan: '',
}

export default PhyVersionInput
