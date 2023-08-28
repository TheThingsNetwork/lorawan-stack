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

import React, { useCallback, useContext, useRef } from 'react'
import classnames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import { RadioGroupContext } from './group'

import style from './radio-button.styl'

const RadioButton = ({
  className,
  name,
  label,
  disabled,
  readOnly,
  autoFocus,
  onBlur,
  onFocus,
  value,
  checked,
  id,
  onChange,
}) => {
  const input = useRef()
  const context = useContext(RadioGroupContext)

  const handleChange = useCallback(
    event => {
      if (context) {
        const { onChange: groupOnChange } = context
        groupOnChange(event)
      }

      onChange(event)
    },
    [onChange, context],
  )

  const focus = useCallback(
    val => {
      if (input && input.current) {
        input.current.focus()
      }

      onFocus(val)
    },
    [onFocus],
  )

  const blur = useCallback(
    val => {
      if (input && input.current) {
        input.current.blur()
      }

      onBlur(val)
    },
    [onBlur],
  )

  const radioProps = {}
  let groupCls
  if (context) {
    radioProps.name = context.name
    radioProps.disabled = disabled || context.disabled
    radioProps.checked = value === context.value
    groupCls = context.className
  } else {
    radioProps.name = name
    radioProps.disabled = disabled
    radioProps.checked = checked
    radioProps.value = value
  }

  const cls = classnames(className, style.wrapper, groupCls, {
    [style.disabled]: radioProps.disabled,
  })

  return (
    <label className={cls}>
      <span className={style.radio}>
        <input
          type="radio"
          ref={input}
          readOnly={readOnly}
          autoFocus={autoFocus}
          onBlur={blur}
          onFocus={focus}
          onChange={handleChange}
          value={value}
          id={id}
          {...radioProps}
        />
        <span className={style.dot} />
      </span>
      {label && <Message className={style.label} content={label} />}
    </label>
  )
}

RadioButton.propTypes = {
  autoFocus: PropTypes.bool,
  checked: PropTypes.bool,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  id: PropTypes.string,
  label: PropTypes.message,
  name: PropTypes.string,
  onBlur: PropTypes.func,
  onChange: PropTypes.func,
  onFocus: PropTypes.func,
  readOnly: PropTypes.bool,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.bool]),
}

RadioButton.defaultProps = {
  className: undefined,
  checked: false,
  disabled: false,
  label: undefined,
  name: undefined,
  readOnly: false,
  value: undefined,
  autoFocus: false,
  id: undefined,
  onChange: () => null,
  onBlur: () => null,
  onFocus: () => null,
}

export default RadioButton
