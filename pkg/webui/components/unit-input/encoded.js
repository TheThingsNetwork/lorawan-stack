// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback } from 'react'

import PropTypes from '@ttn-lw/lib/prop-types'

import UnitInput from './unit-input'

const EncodedUnitInput = props => {
  const { onChange, encode, decode, value, ...rest } = props

  const decodedValue = decode(value)
  const { unit } = decodedValue

  const handleInputChange = useCallback(
    inputValue => {
      onChange(encode(inputValue, unit))
    },
    [unit, encode, onChange],
  )

  const handleUnitChange = useCallback(
    unit => {
      onChange(encode(decodedValue.value, unit), true)
    },
    [onChange, encode, decodedValue.value],
  )

  return (
    <UnitInput
      {...rest}
      onInputChange={handleInputChange}
      onUnitChange={handleUnitChange}
      value={decodedValue}
    />
  )
}

EncodedUnitInput.propTypes = {
  decode: PropTypes.func.isRequired,
  encode: PropTypes.func.isRequired,
  onChange: PropTypes.func.isRequired,
  units: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.message,
      value: PropTypes.string,
      factor: PropTypes.number,
    }),
  ).isRequired,
  value: PropTypes.string,
}

EncodedUnitInput.defaultProps = {
  value: undefined,
}

export default EncodedUnitInput
