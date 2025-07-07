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

import React, { useCallback, useImperativeHandle, useRef, useState } from 'react'
import { defineMessages, useIntl } from 'react-intl'
import classnames from 'classnames'

import Icon, { IconEye, IconEyeOff } from '@ttn-lw/components/icon'
import Spinner from '@ttn-lw/components/spinner'
import Button from '@ttn-lw/components/button'
import Tooltip from '@ttn-lw/components/tooltip'

import Message from '@ttn-lw/lib/components/message'

import { isSafariUserAgent } from '@ttn-lw/lib/navigator'
import combineRefs from '@ttn-lw/lib/combine-refs'
import PropTypes from '@ttn-lw/lib/prop-types'

import ByteInput from './byte'
import Toggled from './toggled'
import Generate from './generate'

import style from './input.styl'

const m = defineMessages({
  showValue: 'Show value',
  hideValue: 'Hide value',
})

const Input = React.forwardRef((props, ref) => {
  const {
    action,
    actionDisable,
    append,
    autoComplete,
    children,
    className,
    code,
    component,
    disabled,
    error,
    forwardedRef,
    icon,
    inputRef,
    inputWidth,
    label,
    loading,
    onBlur,
    onChange,
    onEnter,
    onFocus,
    placeholder,
    readOnly,
    sensitive,
    showPerChar,
    title,
    type,
    valid,
    value,
    warning,
    max,
    ...rest
  } = props
  const [focus, setFocus] = useState(false)
  const [hidden, setHidden] = useState(sensitive)
  const input = useRef(null)
  const intl = useIntl()

  const computeByteInputWidth = useCallback(() => {
    const isSafari = isSafariUserAgent()
    const maxValue = showPerChar ? Math.ceil(max / 2) : max
    const multiplier = isSafari ? 2.1 : 1.8

    let width
    if (maxValue === 16) {
      width = isSafari ? 34 : 30
    } else {
      width = maxValue * multiplier + 0.65
    }

    if (sensitive) {
      width += 2.3
    }

    return `${width}rem`
  }, [sensitive, showPerChar, max])

  const handleHideToggleClick = useCallback(() => {
    setHidden(prevHidden => !prevHidden)
  }, [])

  const focusInput = useCallback(() => {
    if (input.current) {
      input.current.focus()
    }

    setFocus(true)
  }, [])

  const blurInput = useCallback(() => {
    if (input.current) {
      input.current.blur()
    }

    setFocus(false)
  }, [])

  // Expose the 'focus' and 'blur' methods to the parent component
  useImperativeHandle(ref, () => ({
    focus: focusInput,
    blur: blurInput,
  }))

  const onFocusCallback = useCallback(
    evt => {
      setFocus(true)
      onFocus(evt)
    },
    [onFocus],
  )

  const onBlurCallback = useCallback(
    evt => {
      setFocus(false)
      onBlur(evt)
    },
    [onBlur],
  )

  const onChangeCallback = useCallback(
    evt => {
      const { value } = evt.target
      onChange(value)
    },
    [onChange],
  )

  const onKeyDownCallback = useCallback(
    evt => {
      if (evt.key === 'Enter') {
        onEnter(evt.target.value)
      }
    },
    [onEnter],
  )

  const inputWidthValue = inputWidth || (type === 'byte' ? undefined : 'm')

  let Component = component
  let inputStyle
  if (type === 'byte') {
    Component = ByteInput
    if (!inputWidthValue && props.max) {
      inputStyle = { maxWidth: computeByteInputWidth() }
    }
  } else if (type === 'textarea') {
    Component = 'textarea'
  }

  let inputPlaceholder = placeholder
  if (typeof placeholder === 'object') {
    inputPlaceholder = intl.formatMessage(placeholder, placeholder.values)
  }

  let inputTitle = title
  if (typeof title === 'object') {
    inputTitle = intl.formatMessage(title, title.values)
  }

  const v = valid && (Component.validate ? Component.validate(value, props) : true)
  const hasAction = Boolean(action)

  const inputCls = classnames(style.inputBox, {
    [style[`input-width-${inputWidthValue}`]]: inputWidthValue,
    [style.focus]: focus,
    [style.error]: error,
    [style.readOnly]: readOnly,
    [style.warn]: !error && warning,
    [style.disabled]: disabled,
    [style.code]: code,
    [style.actionable]: hasAction,
    [style.textarea]: type === 'textarea',
  })
  const inputElemCls = classnames(style.input, { [style.hidden]: hidden })

  const passedProps = {
    max,
    ...rest,
    ...(type === 'byte' ? { showPerChar } : {}),
    ref: inputRef ? combineRefs([input, inputRef]) : input,
  }

  return (
    <div className={classnames(className, style.container)}>
      <div className={inputCls} style={inputStyle}>
        {icon && <Icon className={style.icon} icon={icon} />}
        <Component
          key="i"
          className={inputElemCls}
          type={type}
          value={value}
          onFocus={onFocusCallback}
          onBlur={onBlurCallback}
          onChange={onChangeCallback}
          onKeyDown={onKeyDownCallback}
          placeholder={inputPlaceholder}
          disabled={disabled}
          readOnly={readOnly}
          title={inputTitle}
          autoComplete={autoComplete}
          data-1p-ignore={!sensitive}
          data-lpignore={!sensitive}
          {...passedProps}
        />
        {v && <Valid show={v} />}
        {loading && <Spinner className={style.spinner} small />}
        {sensitive && value.length !== 0 && (
          <Tooltip
            delay={[1250, 200]}
            hideOnClick={false}
            content={<Message content={hidden ? m.showValue : m.hideValue} />}
            trigger="mouseenter"
            small
          >
            <Button
              icon={hidden ? IconEye : IconEyeOff}
              className={style.hideToggle}
              onClick={handleHideToggleClick}
              naked
              type="button"
            />
          </Tooltip>
        )}
        {append && <div className={style.append}>{append}</div>}
      </div>
      {hasAction && (
        <div className={style.actions}>
          <Button
            className={style.button}
            secondary
            {...action}
            disabled={disabled || actionDisable}
          />
        </div>
      )}
      {children}
    </div>
  )
})

const Valid = props => {
  const classname = classnames(style.valid, {
    [style.show]: props.show,
  })

  return (
    <svg viewBox="0 0 512 512" className={classname}>
      <path d="M256 32a224 224 0 1 0 0 448 224 224 0 0 0 0-448zm115 149L232 360c-1 1-3 3-5 3-3 0-4-1-5-3l-79-76-2-1-1-3 1-4 1-1 25-25c1-2 2-3 4-3 3 0 5 2 6 3l45 43 111-143 4-1 3 1 31 24 1 4-1 3z" />
    </svg>
  )
}

Valid.propTypes = {
  show: PropTypes.bool,
}

Valid.defaultProps = {
  show: false,
}

Input.propTypes = {
  action: PropTypes.shape({
    ...Button.propTypes,
  }),
  actionDisable: PropTypes.bool,
  append: PropTypes.node,
  autoComplete: PropTypes.oneOf([
    'current-password',
    'email',
    'name',
    'new-password',
    'off',
    'on',
    'url',
    'username',
  ]),
  children: PropTypes.node,
  className: PropTypes.string,
  code: PropTypes.bool,
  component: PropTypes.oneOfType([PropTypes.string, PropTypes.node]),
  disabled: PropTypes.bool,
  error: PropTypes.bool,
  forwardedRef: PropTypes.shape({ current: PropTypes.shape({}) }),
  icon: PropTypes.icon,
  inputRef: PropTypes.shape({ current: PropTypes.shape({}) }),
  inputWidth: PropTypes.inputWidth,
  label: PropTypes.string,
  loading: PropTypes.bool,
  max: PropTypes.number,
  onBlur: PropTypes.func,
  onChange: PropTypes.func,
  onEnter: PropTypes.func,
  onFocus: PropTypes.func,
  placeholder: PropTypes.message,
  readOnly: PropTypes.bool,
  sensitive: PropTypes.bool,
  showPerChar: PropTypes.bool,
  title: PropTypes.message,
  type: PropTypes.string,
  valid: PropTypes.bool,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
  warning: PropTypes.bool,
}

Input.defaultProps = {
  action: undefined,
  actionDisable: false,
  append: null,
  autoComplete: 'off',
  children: undefined,
  className: undefined,
  code: false,
  component: 'input',
  disabled: false,
  error: false,
  /** Default `inputWidth` value is set programmatically based on input type. */
  inputWidth: undefined,
  icon: undefined,
  label: undefined,
  loading: false,
  max: undefined,
  onFocus: () => null,
  onBlur: () => null,
  onChange: () => null,
  onEnter: () => null,
  placeholder: undefined,
  readOnly: false,
  sensitive: false,
  showPerChar: false,
  title: undefined,
  type: 'text',
  valid: false,
  value: '',
  warning: false,
  inputRef: null,
  forwardedRef: null,
}

Input.Toggled = Toggled
Input.Generate = Generate

export default Input
