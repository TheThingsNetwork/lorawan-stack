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

import Message from '../../lib/components/message'
import PropTypes from '../../lib/prop-types'
import { RadioGroupContext } from './group'

import style from './radio-button.styl'

@bind
class RadioButton extends React.PureComponent {
  static contextType = RadioGroupContext

  static propTypes = {
    autoFocus: PropTypes.bool,
    checked: PropTypes.bool,
    className: PropTypes.string,
    disabled: PropTypes.bool,
    label: PropTypes.message,
    name: PropTypes.string,
    onBlur: PropTypes.func,
    onChange: PropTypes.func,
    onFocus: PropTypes.func,
    readOnly: PropTypes.bool,
    value: PropTypes.string,
  }

  static defaultProps = {
    className: undefined,
    checked: false,
    disabled: false,
    label: undefined,
    name: undefined,
    readOnly: false,
    value: undefined,
    autoFocus: false,
    onChange: () => null,
    onBlur: () => null,
    onFocus: () => null,
  }

  constructor(props) {
    super(props)

    this.input = React.createRef()
  }

  handleChange(event) {
    const { onChange } = this.props

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
      value,
      checked,
    } = this.props

    const radioProps = {}
    let groupCls
    if (this.context) {
      radioProps.name = this.context.name
      radioProps.disabled = disabled || this.context.disabled
      radioProps.checked = value === this.context.value
      groupCls = this.context.className
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
            ref={this.input}
            readOnly={readOnly}
            autoFocus={autoFocus}
            onBlur={onBlur}
            onFocus={onFocus}
            onChange={this.handleChange}
            value={value}
            {...radioProps}
          />
          <span className={style.dot} />
        </span>
        {label && <Message className={style.label} content={label} />}
      </label>
    )
  }
}

export default RadioButton
