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
import classnames from 'classnames'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import Select from '../../../components/select'
import Input from '../../../components/input'

import PropTypes from '../../../lib/prop-types'
import computePrefix from './compute-prefix'

import style from './join-eui-prefixes-input.styl'

const m = defineMessages({
  empty: 'No prefix',
})

const getOptions = prefixes =>
  prefixes
    .filter(prefix => Boolean(prefix.length))
    .map(function({ join_eui, length }) {
      const computedPrefix = computePrefix(join_eui, length).toUpperCase()

      return { label: computedPrefix, value: computedPrefix }
    })

const emptyOption = { label: m.empty, value: '' }

@bind
class JoinEUIPrefixesInput extends React.PureComponent {
  inputRef = React.createRef()

  state = {
    prefix: emptyOption.value,
  }

  _getPrefixSelectName() {
    const { name } = this.props

    return `${name}-prefix`
  }

  async handleChange(value) {
    const { onChange } = this.props
    const { prefix } = this.state

    if (!Boolean(prefix)) {
      await this.setState({ prefix: emptyOption.value })
      onChange(value)
    } else {
      onChange(`${prefix}${value}`)
    }
  }

  async handlePrefixChange(prefix) {
    const { onChange } = this.props

    await this.setState({ prefix })
    onChange(prefix)
    if (this.inputRef.current) {
      const instance = this.inputRef.current

      instance.focus()
    }
  }

  handleBlur(event) {
    const { name, onBlur } = this.props
    const { target, relatedTarget } = event

    const nextTarget = Boolean(relatedTarget) ? relatedTarget : {}
    const selectName = this._getPrefixSelectName()

    if (target.name === name && nextTarget.name !== selectName) {
      const { prefix } = this.state
      const { value } = target

      target.value = `${prefix}${value}`
      onBlur(event)
    } else if (target.name === selectName && nextTarget.name !== name) {
      const { prefix } = this.state
      const { value } = this.props

      target.value = `${prefix}${value}`
      onBlur(event)
    }
  }

  render() {
    const { className, name, disabled, value, error, prefixes, fetching, showPrefixes } = this.props
    const { prefix } = this.state

    let selectComponent = null
    if (showPrefixes) {
      const selectOptions = getOptions(prefixes)
      selectOptions.unshift(emptyOption)

      selectComponent = (
        <Select
          className={style.select}
          name={this._getPrefixSelectName()}
          disabled={disabled}
          options={selectOptions}
          onChange={this.handlePrefixChange}
          onBlur={this.handleBlur}
          error={error}
          isLoading={fetching}
          value={prefix}
        />
      )
    }

    let inputValue = value
    let charsLeft = 16
    if (Boolean(prefix) && Boolean(value)) {
      inputValue = value.slice(prefix.length, value.length)
      charsLeft -= prefix.length
    }

    return (
      <div className={classnames(className, style.container)}>
        {selectComponent}
        <Input
          showPerChar
          ref={this.inputRef}
          className={style.byte}
          value={inputValue}
          defaultValue={inputValue}
          name={name}
          disabled={disabled}
          min={charsLeft}
          max={charsLeft}
          type="byte"
          onChange={this.handleChange}
          onBlur={this.handleBlur}
          error={error}
        />
      </div>
    )
  }
}

JoinEUIPrefixesInput.propTypes = {
  className: PropTypes.string,
  disabled: PropTypes.bool,
  fetching: PropTypes.bool,
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func,
  onChange: PropTypes.func.isRequired,
  prefixes: PropTypes.arrayOf(
    PropTypes.shape({
      join_eui: PropTypes.string,
      length: PropTypes.number,
    }),
  ),
  showPrefixes: PropTypes.bool,
  value: PropTypes.string,
}

JoinEUIPrefixesInput.defaultProps = {
  className: undefined,
  disabled: false,
  onBlur: () => null,
  fetching: false,
  prefixes: [],
  showPrefixes: true,
}

export default JoinEUIPrefixesInput
