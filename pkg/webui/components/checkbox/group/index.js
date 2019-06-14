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

import PropTypes from '../../../lib/prop-types'

import style from './group.styl'

export const CheckboxGroupContext = React.createContext()

@bind
class CheckboxGroup extends React.Component {

  constructor (props) {
    super(props)

    let value
    if ('value' in props) {
      value = props.value
    } else if ('initialValue' in props) {
      value = props.initialValue
    } else {
      value = {}
    }

    this.state = { value }
  }

  static getDerivedStateFromProps (props) {
    if ('value' in props) {
      return { value: props.value || {}}
    }

    return null
  }

  async handleCheckboxChange (event) {
    const { onChange } = this.props
    const { target } = event

    const value = { ...this.state.value, [target.name]: target.checked }

    if (!('value' in this.props)) {
      await this.setState({ value })
    }

    onChange(value)
  }

  getCheckboxValue (name) {
    const { value } = this.state

    return value[name] || false
  }

  render () {
    const {
      className,
      disabled,
      onFocus,
      onBlur,
      horizontal,
      children,
    } = this.props

    const ctx = {
      className: style.groupCheckbox,
      onChange: this.handleCheckboxChange,
      getValue: this.getCheckboxValue,
      onBlur,
      onFocus,
      disabled,
    }

    const cls = classnames(className, style.group, {
      [style.horizontal]: horizontal,
    })

    return (
      <div className={cls}>
        <CheckboxGroupContext.Provider value={ctx}>
          {children}
        </CheckboxGroupContext.Provider>
      </div>
    )
  }
}

CheckboxGroup.propTypes = {
  className: PropTypes.string,
  name: PropTypes.string.isRequired,
  disabled: PropTypes.bool,
  horizontal: PropTypes.bool,
  value: PropTypes.object,
  initialValue: PropTypes.object,
  onChange: PropTypes.func,
  onBlur: PropTypes.func,
  onFocus: PropTypes.func,
  children: PropTypes.oneOfType([
    PropTypes.arrayOf(PropTypes.node),
    PropTypes.node,
  ]).isRequired,
}

CheckboxGroup.defaultProps = {
  disabled: false,
  horizontal: false,
  onChange: () => null,
  onBlur: () => null,
  onFocus: () => null,
}

export default CheckboxGroup
