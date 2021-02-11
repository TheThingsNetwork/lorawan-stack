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
import MaskedInput from 'react-text-mask'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './input.styl'

const PLACEHOLDER_CHAR = '·'

const hex = /[0-9a-f]/i

const masks = {}
const mask = function (min, max, showPerChar = false) {
  const key = `${min}-${max}`
  if (masks[key]) {
    return masks[key]
  }

  const wordSize = showPerChar ? 2 : 1

  let length = 3 * Math.floor(max / wordSize) - 1
  if (showPerChar && max % wordSize !== 0) {
    // Account for the space and the extra character.
    length += wordSize
  }

  const r = new Array(length).fill(hex)
  for (let i = 0; i < r.length; i++) {
    if ((i + 1) % 3 === 0) {
      r[i] = ' '
    }
  }

  return r
}

const upper = function (str) {
  return str.toUpperCase()
}

const clean = function (str) {
  return str.replace(new RegExp(`[ ${PLACEHOLDER_CHAR}]`, 'g'), '')
}

export default class ByteInput extends React.Component {
  static propTypes = {
    className: PropTypes.string,
    max: PropTypes.number,
    min: PropTypes.number,
    onBlur: PropTypes.func,
    onChange: PropTypes.func.isRequired,
    placeholder: PropTypes.message,
    showPerChar: PropTypes.bool,
    unbounded: PropTypes.bool,
    value: PropTypes.string.isRequired,
  }

  static defaultProps = {
    className: undefined,
    min: 0,
    max: undefined,
    placeholder: undefined,
    showPerChar: false,
    onBlur: () => null,
    unbounded: false,
  }

  input = React.createRef()

  static validate(value, props) {
    const { min = 0, max = 256 } = props
    const len = Math.floor(value.length / 2)
    return min <= len && len <= max
  }

  render() {
    const {
      onBlur,
      value,
      className,
      min,
      max,
      onChange,
      placeholder,
      showPerChar,
      unbounded,
      ...rest
    } = this.props

    const valueLength = clean(value).length || 0
    const calculatedMax = max || Math.max(Math.floor(valueLength / 2) + 1, 1)

    return (
      <MaskedInput
        ref={this.input}
        key="input"
        className={classnames(className, style.byte)}
        value={value}
        mask={mask(min, calculatedMax, showPerChar)}
        placeholderChar={PLACEHOLDER_CHAR}
        keepCharPositions={false}
        pipe={upper}
        onChange={this.onChange}
        placeholder={placeholder}
        onCopy={this.onCopy}
        onCut={this.onCut}
        onBlur={this.onBlur}
        onPaste={this.onPaste}
        showMask={!placeholder && !unbounded}
        guide={!unbounded}
        {...rest}
      />
    )
  }

  focus() {
    if (this.input.current && this.input.current.inputElement) {
      const { inputElement } = this.input.current

      let i = inputElement.value.indexOf(PLACEHOLDER_CHAR)
      if (i === -1) {
        i = inputElement.value.length
      }

      setTimeout(function () {
        inputElement.focus()
        inputElement.setSelectionRange(i, i)
      }, 0)
    }
  }

  @bind
  onChange(evt) {
    const { max, showPerChar } = this.props
    const { data } = evt.nativeEvent

    let value = clean(event.target.value)
    const normalizedMax = showPerChar ? Math.ceil(max / 2) : max

    // Check if the value already has length equal to `max`.
    const isValueMaxLength = value.length === normalizedMax * 2
    // Check if the cursor is placed after the last placeholder chararcter.
    const isCursorAfterMask = evt.target.selectionStart === normalizedMax * 3 - 1

    if (!isValueMaxLength && isCursorAfterMask && data !== null) {
      const hexMatch = data.match(hex)
      if (hexMatch !== null) {
        value += hexMatch[0]
      }
    }

    this.props.onChange({
      target: {
        name: evt.target.name,
        value,
      },
    })
  }

  @bind
  onBlur(evt) {
    this.props.onBlur({
      relatedTarget: evt.relatedTarget,
      target: {
        name: evt.target.name,
        value: clean(evt.target.value),
      },
    })
  }

  @bind
  onCopy(evt) {
    const input = evt.target
    const value = input.value.substr(
      input.selectionStart,
      input.selectionEnd - input.selectionStart,
    )
    evt.clipboardData.setData('text/plain', clean(value))
    evt.preventDefault()
  }

  @bind
  onPaste(evt) {
    const { min, showPerChar, unbounded } = this.props
    if (unbounded) {
      evt.preventDefault()
      this.input.current.inputElement.value = evt.clipboardData.getData('text/plain')
      mask(min, evt.clipboardData.getData('text/plain').length, showPerChar)
      this.onChange(evt)
    }
  }

  @bind
  onCut(evt) {
    const input = evt.target
    const value = input.value.substr(
      input.selectionStart,
      input.selectionEnd - input.selectionStart,
    )
    evt.clipboardData.setData('text/plain', clean(value))
    evt.preventDefault()

    // Emit the cut value.
    const cut = input.value.substr(0, input.selectionStart) + input.value.substr(input.selectionEnd)
    evt.target.value = cut
    this.onChange({
      target: {
        value: cut,
      },
    })
  }
}
