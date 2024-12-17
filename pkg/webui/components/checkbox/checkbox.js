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

import React, { useState, useEffect, useContext, useCallback } from 'react'
import classnames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { CheckboxGroupContext } from './group'

import style from './checkbox.styl'

const Checkbox = props => {
  const {
    autoFocus,
    checked: propChecked,
    children,
    className,
    disabled,
    id,
    indeterminate,
    label,
    labelAsTitle,
    name,
    onBlur,
    onChange,
    onFocus,
    readOnly,
    value,
    ...rest
  } = props
  const hasValue = 'value' in props
  const context = useContext(CheckboxGroupContext)
  const composedValue = value && context ? value[name] : false
  const [checkedState, setChecked] = useState(composedValue)

  useEffect(() => {
    if (hasValue && value !== checkedState) {
      setChecked(value)
    }

    return null
  }, [value, hasValue, checkedState])

  const handleChange = useCallback(
    event => {
      const { checked } = event.target

      if (!hasValue && !context) {
        setChecked(checked)
      }

      if (context) {
        const { onChange: groupOnChange } = context
        groupOnChange(event)
      }

      onChange(event)
    },
    [context, onChange, hasValue],
  )

  const inputRef = React.useRef()

  const checkboxProps = {}
  let groupCls

  if (context) {
    checkboxProps.onBlur = context.onBlur
    checkboxProps.onFocus = context.onFocus
    checkboxProps.disabled = disabled || context.disabled
    checkboxProps.checked = context.getValue(name)
    groupCls = context.className
  } else {
    checkboxProps.onBlur = onBlur
    checkboxProps.onFocus = onFocus
    checkboxProps.disabled = disabled
    checkboxProps.checked = checkedState
  }
  checkboxProps.value = checkboxProps.checked

  const cls = classnames(className, style.wrapper, groupCls, {
    [style.disabled]: checkboxProps.disabled,
    [style.indeterminate]: indeterminate,
  })

  const labelCls = classnames(style.label, {
    [style.labelAsTitle]: labelAsTitle,
  })

  return (
    <label className={cls}>
      <div>
        <span className={style.checkbox}>
          <input
            type="checkbox"
            ref={inputRef}
            name={name}
            readOnly={readOnly}
            autoFocus={autoFocus}
            onChange={handleChange}
            id={id}
            aria-describedby={rest['aria-describedby']}
            aria-invalid={rest['aria-invalid']}
            {...checkboxProps}
          />
          <span className={style.checkmark} />
        </span>
        {label && <Message className={labelCls} content={label} />}
      </div>
      {children}
    </label>
  )
}

Checkbox.propTypes = {
  autoFocus: PropTypes.bool,
  checked: PropTypes.bool,
  children: PropTypes.node,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  id: PropTypes.string,
  indeterminate: PropTypes.bool,
  label: PropTypes.message,
  labelAsTitle: PropTypes.bool,
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func,
  onChange: PropTypes.func,
  onFocus: PropTypes.func,
  readOnly: PropTypes.bool,
  value: PropTypes.oneOfType([PropTypes.bool, PropTypes.object]),
}

Checkbox.defaultProps = {
  checked: false,
  children: null,
  className: undefined,
  label: sharedMessages.enabled,
  labelAsTitle: false,
  disabled: false,
  id: undefined,
  readOnly: false,
  autoFocus: false,
  onChange: () => null,
  onBlur: () => null,
  onFocus: () => null,
  indeterminate: false,
  value: false,
}

export default Checkbox
