// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback } from 'react'
import classnames from 'classnames'
import MaskedInput from 'react-text-mask'

import PropTypes from '@ttn-lw/lib/prop-types'
import { warn } from '@ttn-lw/lib/log'

import style from './input.styl'

const PLACEHOLDER_CHAR = '·'

const hex = /[0-9a-f]/i
const voidChars = RegExp(`[ ${PLACEHOLDER_CHAR}]`, 'g')

const masks = {}
const mask = (min, max, showPerChar = false) => {
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

const upper = str => str.toUpperCase()

const clean = str => (typeof str === 'string' ? str.replace(voidChars, '') : str)

const ByteInput = ({
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
}) => {
  const onCopy = useCallback(evt => {
    const input = evt.target
    const selectedValue = input.value.substr(
      input.selectionStart,
      input.selectionEnd - input.selectionStart,
    )
    evt.clipboardData.setData('text/plain', clean(selectedValue))
    evt.preventDefault()
  }, [])

  const onPaste = useCallback(
    evt => {
      // Ignore empty pastes.
      if (evt?.clipboardData?.getData('text/plain')?.length === 0) {
        return
      }
      const val = evt.target.value
      const cleanedSelection = clean(
        val.substr(
          evt.target.selectionStart,
          Math.max(1, evt.target.selectionEnd - evt.target.selectionStart),
        ),
      )

      // To avoid the masked input from cutting off characters when the cursor
      // is placed in the mask placeholders, the placeholder chars are removed before
      // the paste is applied, unless the user made a selection to paste into.
      // This will ensure a consistent pasting experience.
      if (!unbounded && cleanedSelection === '') {
        evt.target.value = val.replace(voidChars, '')
      }
    },
    [unbounded],
  )

  const onChangeCallback = useCallback(
    evt => {
      const { value: oldValue, unbounded } = rest
      const data = evt?.nativeEvent?.data

      // Due to the way that react-text-mask works, it is not possible to
      // store the cleaned value, since it would create ambiguity between
      // values like `AA` and `AA `. This causes backspaces to not work
      // if it targets the space character, since the deleted space would
      // be re-added right away. Hence, unbounded inputs need to remove
      // the space paddings manually.
      let value = unbounded ? evt.target.value : clean(evt.target.value)

      // Make sure values entered at the end of the input (with placeholders)
      // are added as expected. `selectionStart` cannot be used due to
      // inconsistent behavior on Android phones.
      if (
        evt.target.value.endsWith(PLACEHOLDER_CHAR) &&
        data &&
        hex.test(data) &&
        oldValue === value
      ) {
        value += data
      }

      onChange({
        target: {
          name: evt.target.name,
          value,
        },
      })
    },
    [onChange, rest],
  )

  const onBlurCallback = useCallback(
    evt => {
      onBlur({
        relatedTarget: evt.relatedTarget,
        target: {
          name: evt.target.name,
          value: clean(evt.target.value),
        },
      })
    },
    [onBlur],
  )

  const onCut = useCallback(evt => {
    evt.preventDefault()
    // Recreate cut action by deleting and reusing copy handler.
    document.execCommand('copy')
    document.execCommand('delete')
  }, [])

  // Instead of calculating the max width dynamically, which leads to various issues
  // with pasting, it's better to use a high max value for unbounded inputs instead.
  const calculatedMax = max || 4096

  if (!unbounded && typeof max !== 'number') {
    warn(
      'Byte input has been setup without `max` prop. Always use a max prop unless using `unbounded`',
    )
  }

  return (
    <MaskedInput
      key="input"
      className={classnames(className, style.byte)}
      value={value}
      mask={mask(min, calculatedMax, showPerChar)}
      placeholderChar={PLACEHOLDER_CHAR}
      keepCharPositions={false}
      pipe={upper}
      onChange={onChangeCallback}
      placeholder={placeholder}
      onCopy={onCopy}
      onCut={onCut}
      onBlur={onBlurCallback}
      onPaste={onPaste}
      showMask={!placeholder && !unbounded}
      guide={!unbounded}
      {...rest}
      type="text"
    />
  )
}

ByteInput.propTypes = {
  className: PropTypes.string,
  max: PropTypes.number,
  min: PropTypes.number,
  onBlur: PropTypes.func,
  onChange: PropTypes.func.isRequired,
  onFocus: PropTypes.func,
  placeholder: PropTypes.message,
  showPerChar: PropTypes.bool,
  unbounded: PropTypes.bool,
  value: PropTypes.string.isRequired,
}

ByteInput.defaultProps = {
  className: undefined,
  min: 0,
  max: undefined,
  placeholder: undefined,
  showPerChar: false,
  onBlur: () => null,
  onFocus: () => null,
  unbounded: false,
}

export default ByteInput
