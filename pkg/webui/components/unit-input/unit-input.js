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

import React from 'react'
import classnames from 'classnames'

import Select from '@ttn-lw/components/select'
import Input from '@ttn-lw/components/input'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './unit-input.styl'

const UnitInput = props => {
  const {
    className,
    defaultUnit,
    name,
    units,
    unitSelector,
    value: { value, unit },
    onBlur,
    required,
    disabled,
    error,
    selectWidth,
    inputWidth,
    onInputChange,
    onUnitChange,
  } = props
  const maskedUnits = unitSelector ? units.filter(u => unitSelector.includes(u.value)) : units
  return (
    <React.Fragment>
      <div className={classnames(className, style.container)}>
        <Input
          data-test-id={name}
          className={style.number}
          type="number"
          step="any"
          name={name}
          id={name}
          onBlur={onBlur}
          value={value}
          onChange={onInputChange}
          required={required}
          disabled={disabled}
          error={error}
          inputWidth={inputWidth}
        />
        <Select
          className={style.select}
          name={`${name}-select`}
          options={maskedUnits}
          onChange={onUnitChange}
          onBlur={onBlur}
          value={unit || defaultUnit || maskedUnits[0].value}
          disabled={disabled}
          inputWidth={selectWidth}
        />
      </div>
    </React.Fragment>
  )
}

UnitInput.propTypes = {
  className: PropTypes.string,
  defaultUnit: PropTypes.string,
  disabled: PropTypes.bool,
  error: PropTypes.bool,
  inputWidth: PropTypes.inputWidth,
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func.isRequired,
  onInputChange: PropTypes.func.isRequired,
  onUnitChange: PropTypes.func.isRequired,
  required: PropTypes.bool,
  selectWidth: PropTypes.inputWidth,
  unitSelector: PropTypes.arrayOf(PropTypes.string),
  units: PropTypes.arrayOf(
    PropTypes.shape({
      value: PropTypes.string.isRequired,
      label: PropTypes.message.isRequired,
    }),
  ).isRequired,
  value: PropTypes.shape({
    unit: PropTypes.string,
    value: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  }),
}

UnitInput.defaultProps = {
  className: undefined,
  defaultUnit: undefined,
  inputWidth: 'm',
  disabled: false,
  required: false,
  value: undefined,
  error: false,
  selectWidth: 'full',
  unitSelector: undefined,
}

export default UnitInput
