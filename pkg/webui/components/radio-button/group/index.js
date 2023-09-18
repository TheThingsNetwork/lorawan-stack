// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useEffect, useState } from 'react'
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './group.styl'

export const RadioGroupContext = React.createContext()

const findCheckedRadio = children => {
  let value
  let matched = false

  React.Children.forEach(children, radio => {
    if (radio && radio.props && !matched && radio.props.checked) {
      value = radio.props.value
      matched = true
    }
  })

  return value
}

const RadioGroup = props => {
  const { className, name, disabled, horizontal, onChange, children } = props
  const [value, setValue] = useState(() => {
    if ('value' in props) {
      return props.value
    } else if ('initialValue' in props) {
      return props.initialValue
    }

    return findCheckedRadio(children)
  })

  useEffect(() => {
    if ('value' in props) {
      setValue(props.value)
    } else {
      const foundValue = findCheckedRadio(children)
      if (foundValue && foundValue !== value) {
        setValue(foundValue)
      }
    }
  }, [props, children, value])

  const handleRadioChange = useCallback(
    event => {
      const { target } = event

      // Retain boolean type if the value was initially provided as boolean.
      const value = typeof props.value === 'boolean' ? target.value === 'true' : target.value

      if (!('value' in props)) {
        setValue(value)
      }

      onChange(value)
    },
    [onChange, props],
  )

  const ctx = {
    className: style.groupRadio,
    onChange: handleRadioChange,
    disabled,
    value,
    name,
  }

  const cls = classnames(className, style.group, {
    [style.horizontal]: horizontal,
  })

  return (
    <div className={cls}>
      <RadioGroupContext.Provider value={ctx}>{children}</RadioGroupContext.Provider>
    </div>
  )
}

RadioGroup.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  horizontal: PropTypes.bool,
  initialValue: PropTypes.string,
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.bool]),
}

RadioGroup.defaultProps = {
  className: undefined,
  disabled: false,
  initialValue: undefined,
  value: undefined,
  horizontal: false,
  onChange: () => null,
}

export default RadioGroup
