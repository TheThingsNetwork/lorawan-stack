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

import Icon from '../icon'
import Spinner from '../spinner'
import PropTypes from '../../lib/prop-types'
import ByteInput from './byte'
import Toggled from './toggled'

import style from './input.styl'

@injectIntl
@bind
class Input extends React.Component {
  static propTypes = {
    icon: PropTypes.string,
    value: PropTypes.oneOfType([
      PropTypes.string,
      PropTypes.number,
    ]),
    onFocus: PropTypes.func,
    onBlur: PropTypes.func,
    onChange: PropTypes.func,
    onEnter: PropTypes.func,
    placeholder: PropTypes.message,
    error: PropTypes.bool,
    warning: PropTypes.bool,
    valid: PropTypes.bool,
    disabled: PropTypes.bool,
    readOnly: PropTypes.bool,
    type: PropTypes.string.isRequired,
    label: PropTypes.string,
    loading: PropTypes.bool,
    title: PropTypes.message,
  }

  static defaultProps = {
    onFocus: () => null,
    onBlur: () => null,
    onChange: () => null,
    onEnter: () => null,
    type: 'text',
  }

  state = {
    focus: false,
  }

  input = React.createRef()

  focus () {
    if (this.input.current) {
      this.input.current.focus()
    }

    this.setState({ focus: true })
  }

  blur () {
    if (this.input.current) {
      this.input.current.blur()
    }

    this.setState({ focus: false })
  }

  render () {
    const {
      icon,
      value = '',
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
      component = 'input',
      loading,
      title,
      intl,
      horizontal,
      ...rest
    } = this.props

    const {
      focus,
    } = this.state

    let Component = component
    if (type === 'byte') {
      Component = ByteInput
    } else if (type === 'textarea') {
      Component = 'textarea'
    }

    let inputPlaceholder = placeholder
    if (typeof placeholder === 'object') {
      inputPlaceholder = intl.formatMessage(placeholder)
    }

    let inputTitle = title
    if (typeof title === 'object') {
      inputTitle = intl.formatMessage(title)
    }

    const v = valid && (Component.validate ? Component.validate(value, this.props) : true)

    const classname = classnames(style.inputBox, className, {
      [style.focus]: focus,
      [style.error]: error,
      [style.readOnly]: readOnly,
      [style.warn]: !error && warning,
      [style.disabled]: disabled,
    })

    return (
      <div className={classname}>
        {icon && <Icon className={style.icon} icon={icon} />}
        <Component
          ref={this.input}
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
          {...rest}
        />
        { v && <Valid show={v} /> }
        { loading && <Spinner className={style.spinner} small /> }
      </div>
    )
  }

  onFocus (evt) {
    const { onFocus } = this.props

    this.setState({ focus: true })
    onFocus(evt)
  }

  onBlur (evt) {
    const { onBlur } = this.props

    this.setState({ focus: false })
    onBlur(evt)
  }

  onChange (evt) {
    const { onChange } = this.props
    const { value } = evt.target

    onChange(value)
  }

  onKeyDown (evt) {
    if (evt.key === 'Enter') {
      this.props.onEnter(evt.target.value)
    }
  }
}

const Valid = function (props) {
  const classname = classnames(style.valid, {
    [style.show]: props.show,
  })

  return (
    <svg viewBox="0 0 512 512" className={classname}>
      <path d="M256 32a224 224 0 1 0 0 448 224 224 0 0 0 0-448zm115 149L232 360c-1 1-3 3-5 3-3 0-4-1-5-3l-79-76-2-1-1-3 1-4 1-1 25-25c1-2 2-3 4-3 3 0 5 2 6 3l45 43 111-143 4-1 3 1 31 24 1 4-1 3z" />
    </svg>
  )
}

Input.Toggled = Toggled

export default Input
