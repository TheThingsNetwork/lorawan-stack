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
import bind from 'autobind-decorator'
import classnames from 'classnames'

import PropTypes from '../../lib/prop-types'
import Message from '../../lib/components/message'
import sharedMessages from '../../lib/shared-messages'
import { CheckboxGroupContext } from './group'

import style from './checkbox.styl'

@bind
class Checkbox extends React.PureComponent {
  static contextType = CheckboxGroupContext

  static propTypes = {
    autoFocus: PropTypes.bool,
    checked: PropTypes.bool,
    className: PropTypes.string,
    disabled: PropTypes.bool,
    indeterminate: PropTypes.bool,
    label: PropTypes.message,
    name: PropTypes.string.isRequired,
    onBlur: PropTypes.func,
    onChange: PropTypes.func,
    onFocus: PropTypes.func,
    readOnly: PropTypes.bool,
    value: PropTypes.oneOfType([PropTypes.bool, PropTypes.object]),
  }

  static defaultProps = {
    checked: false,
    className: undefined,
    label: sharedMessages.enabled,
    disabled: false,
    readOnly: false,
    autoFocus: false,
    onChange: () => null,
    onBlur: () => null,
    onFocus: () => null,
    indeterminate: false,
    value: false,
  }

  constructor(props) {
    super(props)

    this.input = React.createRef()
    let value
    if ('value' in props && this.context) {
      value = props.value[name]
    } else {
      value = false
    }

    this.state = {
      checked: value,
    }
  }

  static getDerivedStateFromProps(props, state) {
    const { value } = props

    if ('value' in props && value !== state.checked) {
      return { checked: value }
    }

    return null
  }

  handleChange(event) {
    const { onChange } = this.props
    const { checked } = event.target

    if (!('value' in this.props) && !this.context) {
      this.setState({ checked })
    }

    if (this.context) {
      const { onChange: groupOnChange } = this.context
      groupOnChange(event)
    }

    onChange(event)
  }

  focus() {
    if (this.input && this.input.current) {
      this.input.current.focus()
    }
  }

  blur() {
    if (this.input && this.input.current) {
      this.input.current.blur()
    }
  }

  render() {
    const {
      className,
      name,
      label,
      disabled,
      readOnly,
      autoFocus,
      onBlur,
      onFocus,
      indeterminate,
    } = this.props
    const { checked } = this.state

    const checkboxProps = {}
    let groupCls
    if (this.context) {
      checkboxProps.onBlur = this.context.onBlur
      checkboxProps.onFocus = this.context.onFocus
      checkboxProps.disabled = disabled || this.context.disabled
      checkboxProps.checked = this.context.getValue(name)
      groupCls = this.context.className
    } else {
      checkboxProps.onBlur = onBlur
      checkboxProps.onFocus = onFocus
      checkboxProps.disabled = disabled
      checkboxProps.checked = checked
    }

    const cls = classnames(className, style.wrapper, groupCls, {
      [style.disabled]: checkboxProps.disabled,
      [style.indeterminate]: indeterminate,
    })

    return (
      <label className={cls}>
        <span className={style.checkbox}>
          <input
            type="checkbox"
            ref={this.input}
            name={name}
            readOnly={readOnly}
            autoFocus={autoFocus}
            onChange={this.handleChange}
            {...checkboxProps}
          />
          <span className={style.checkmark} />
        </span>
        {label && <Message className={style.label} content={label} />}
      </label>
    )
  }
}

export default Checkbox
