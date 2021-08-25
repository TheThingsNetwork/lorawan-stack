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

import React from 'react'
import { injectIntl } from 'react-intl'
import classnames from 'classnames'
import bind from 'autobind-decorator'

import Icon from '@ttn-lw/components/icon'
import Spinner from '@ttn-lw/components/spinner'
import Button from '@ttn-lw/components/button'

import { isSafariUserAgent } from '@ttn-lw/lib/navigator'
import combineRefs from '@ttn-lw/lib/combine-refs'
import PropTypes from '@ttn-lw/lib/prop-types'

import ByteInput from './byte'
import Toggled from './toggled'
import Generate from './generate'

import style from './input.styl'

class Input extends React.Component {
  static propTypes = {
    action: PropTypes.shape({
      ...Button.propTypes,
    }),
    actionDisable: PropTypes.bool,
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
    icon: PropTypes.string,
    inputRef: PropTypes.shape({ current: PropTypes.shape({}) }),
    inputWidth: PropTypes.inputWidth,
    intl: PropTypes.shape({
      formatMessage: PropTypes.func,
    }).isRequired,
    label: PropTypes.string,
    loading: PropTypes.bool,
    max: PropTypes.number,
    onBlur: PropTypes.func,
    onChange: PropTypes.func,
    onEnter: PropTypes.func,
    onFocus: PropTypes.func,
    placeholder: PropTypes.message,
    readOnly: PropTypes.bool,
    showPerChar: PropTypes.bool,
    title: PropTypes.message,
    type: PropTypes.string,
    valid: PropTypes.bool,
    value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
    warning: PropTypes.bool,
  }

  static defaultProps = {
    action: undefined,
    actionDisable: false,
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
    showPerChar: false,
    title: undefined,
    type: 'text',
    valid: false,
    value: '',
    warning: false,
    inputRef: null,
    forwardedRef: null,
  }

  state = {
    focus: false,
  }

  input = React.createRef(null)

  _computeByteInputWidth() {
    const { showPerChar, max } = this.props
    const isSafari = isSafariUserAgent()

    const maxValue = showPerChar ? Math.ceil(max / 2) : max
    const multiplier = isSafari ? 2.1 : 1.8

    let width
    if (maxValue === 16) {
      width = isSafari ? '34rem' : '30rem'
    } else {
      width = `${maxValue * multiplier + 0.65}rem`
    }

    return width
  }

  focus() {
    if (this.input.current) {
      this.input.current.focus()
    }

    this.setState({ focus: true })
  }

  blur() {
    if (this.input.current) {
      this.input.current.blur()
    }

    this.setState({ focus: false })
  }

  render() {
    const {
      icon,
      value,
      error,
      warning,
      valid,
      placeholder,
      readOnly,
      type,
      disabled,
      onChange,
      onFocus,
      onBlur,
      onEnter,
      className,
      label,
      component,
      loading,
      title,
      inputWidth,
      intl,
      code,
      action,
      autoComplete,
      showPerChar,
      forwardedRef,
      inputRef,
      actionDisable,
      children,
      ...rest
    } = this.props

    const { focus } = this.state
    const inputWidthValue = inputWidth || (type === 'byte' ? undefined : 'm')

    let Component = component
    let inputStyle
    if (type === 'byte') {
      Component = ByteInput
      const { max } = this.props
      if (!inputWidthValue && max) {
        inputStyle = { maxWidth: this._computeByteInputWidth() }
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

    const v = valid && (Component.validate ? Component.validate(value, this.props) : true)
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

    const passedProps = {
      ...rest,
      ...(type === 'byte' ? { showPerChar } : {}),
      ref: inputRef ? combineRefs([this.input, inputRef]) : this.input,
    }

    return (
      <div className={classnames(className, style.container)}>
        <div className={inputCls} style={inputStyle}>
          {icon && <Icon className={style.icon} icon={icon} />}
          <Component
            key="i"
            className={style.input}
            type={type}
            value={value}
            onFocus={this.onFocus}
            onBlur={this.onBlur}
            onChange={this.onChange}
            onKeyDown={this.onKeyDown}
            placeholder={inputPlaceholder}
            disabled={disabled}
            readOnly={readOnly}
            title={inputTitle}
            autoComplete={autoComplete}
            {...passedProps}
          />
          {v && <Valid show={v} />}
          {loading && <Spinner className={style.spinner} small />}
        </div>
        {hasAction && (
          <div className={style.actions}>
            <Button className={style.button} {...action} disabled={disabled || actionDisable} />
          </div>
        )}
        {children}
      </div>
    )
  }

  @bind
  onFocus(evt) {
    const { onFocus } = this.props

    this.setState({ focus: true })
    onFocus(evt)
  }

  @bind
  onBlur(evt) {
    const { onBlur } = this.props

    this.setState({ focus: false })
    onBlur(evt)
  }

  @bind
  onChange(evt) {
    const { onChange } = this.props
    const { value } = evt.target

    onChange(value)
  }

  @bind
  onKeyDown(evt) {
    if (evt.key === 'Enter') {
      this.props.onEnter(evt.target.value)
    }
  }
}

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

Input.Toggled = Toggled
Input.Generate = Generate

export default injectIntl(Input, { forwardRef: true })
