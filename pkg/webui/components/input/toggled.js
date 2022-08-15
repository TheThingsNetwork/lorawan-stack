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

import React, { Component } from 'react'
import bind from 'autobind-decorator'
import classnames from 'classnames'

import Checkbox from '@ttn-lw/components/checkbox'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './toggled.styl'

import Input from '.'

class Toggled extends Component {
  @bind
  handleCheckboxChange(event) {
    const enabled = event.target.checked
    const { value } = this.props.value

    this.props.onChange({ value, enabled }, true)
  }

  @bind
  handleInputChange(value) {
    const { enabled } = this.props.value

    this.props.onChange({ value, enabled })
  }

  render() {
    const { value, type, enabledMessage, className, children, ...rest } = this.props

    const isEnabled = value.enabled || false
    const checkboxId = `${rest.id}_checkbox`

    return (
      <div className={classnames(className, style.container)}>
        <div className={style.checkboxContainer}>
          <label className={style.checkbox} htmlFor={checkboxId}>
            <Checkbox
              name={`${rest.name}.enable`}
              onChange={this.handleCheckboxChange}
              value={isEnabled}
              id={checkboxId}
              label={enabledMessage}
              labelAsTitle
            />
          </label>
          {children}
        </div>
        {isEnabled && (
          <Input
            {...rest}
            className={style.input}
            type="text"
            value={value.value || ''}
            onChange={this.handleInputChange}
          />
        )}
      </div>
    )
  }
}

Toggled.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  enabledMessage: PropTypes.message,
  error: PropTypes.bool,
  icon: PropTypes.string,
  label: PropTypes.string,
  loading: PropTypes.bool,
  onChange: PropTypes.func.isRequired,
  placeholder: PropTypes.string,
  readOnly: PropTypes.bool,
  type: PropTypes.string,
  valid: PropTypes.bool,
  value: PropTypes.shape({
    value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
    enabled: PropTypes.bool,
  }),
  warning: PropTypes.bool,
}

Toggled.defaultProps = {
  className: undefined,
  children: null,
  disabled: false,
  enabledMessage: sharedMessages.enabled,
  error: false,
  icon: undefined,
  label: undefined,
  loading: false,
  placeholder: undefined,
  readOnly: false,
  valid: false,
  value: undefined,
  warning: false,
  type: 'text',
}

export default Toggled
