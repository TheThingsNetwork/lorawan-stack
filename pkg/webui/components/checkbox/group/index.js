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

import React, { useCallback } from 'react'
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './group.styl'

export const CheckboxGroupContext = React.createContext()

const CheckboxGroup = props => {
  const {
    children,
    className,
    horizontal,
    onChange,
    disabled,
    onFocus,
    onBlur,
    value: valueProp,
    ...rest
  } = props
  const hasValue = Boolean(valueProp)
  const [value, setValue] = React.useState(hasValue ? valueProp : rest.initialValue || {})

  React.useEffect(() => {
    if (valueProp) {
      setValue(valueProp || {})
    }
  }, [valueProp])

  const handleCheckboxChange = useCallback(
    async event => {
      const { target } = event

      const newValue = { ...value, [target.name]: target.checked }

      if (!hasValue) {
        setValue(newValue)
      }

      onChange(newValue)
    },
    [onChange, hasValue, value],
  )

  const getCheckboxValue = useCallback(name => value[name] || false, [value])

  const ctx = {
    className: style.groupCheckbox,
    onChange: handleCheckboxChange,
    getValue: getCheckboxValue,
    onBlur,
    onFocus,
    disabled,
  }

  const cls = classnames(className, style.group, {
    [style.horizontal]: horizontal,
  })

  return (
    <div className={cls}>
      <CheckboxGroupContext.Provider value={ctx}>{children}</CheckboxGroupContext.Provider>
    </div>
  )
}

CheckboxGroup.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  horizontal: PropTypes.bool,
  initialValue: PropTypes.shape({}),
  onBlur: PropTypes.func,
  onChange: PropTypes.func,
  onFocus: PropTypes.func,
  value: PropTypes.shape({}),
}

CheckboxGroup.defaultProps = {
  className: undefined,
  disabled: false,
  initialValue: undefined,
  value: {},
  horizontal: false,
  onChange: () => null,
  onBlur: () => null,
  onFocus: () => null,
}

export default CheckboxGroup
