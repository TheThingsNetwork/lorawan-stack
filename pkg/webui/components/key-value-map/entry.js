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

import React, { useCallback, useMemo, useState } from 'react'
import { defineMessages } from 'react-intl'
import classnames from 'classnames'

import { IconTrash } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './key-value-map.styl'

const m = defineMessages({
  deleteEntry: 'Delete entry',
})

const Entry = ({
  readOnly,
  name,
  value,
  fieldValue,
  index,
  onRemoveButtonClick,
  onChange,
  onBlur,
  inputElement: InputElement,
  indexAsKey,
  valuePlaceholder,
  keyPlaceholder,
  additionalInputProps,
  removeMessage,
  distinctOptions,
  atLeastOneEntry,
  filterByTag,
}) => {
  const [currentValue, setCurrentValue] = useState(value)
  const [newOptions, setNewOptions] = useState(undefined)
  const { options, ...additionalInputPropsRest } = additionalInputProps
  const _getKeyInputName = useMemo(() => `${name}[${index}].key`, [index, name])

  const _getValueInputName = useMemo(() => `${name}[${index}].value`, [index, name])

  const handleRemoveButtonClick = useCallback(
    event => {
      onRemoveButtonClick(index, event)
    },
    [index, onRemoveButtonClick],
  )

  const handleKeyChanged = useCallback(
    newKey => {
      onChange(index, { key: newKey })
    },
    [index, onChange],
  )

  const handleValueChanged = useCallback(
    newValue => {
      setCurrentValue(newValue)
      onChange(index, { value: newValue })
    },
    [index, onChange],
  )

  const handleBlur = useCallback(
    event => {
      const { relatedTarget } = event
      const nextTarget = relatedTarget || {}

      if (nextTarget.name !== _getKeyInputName && nextTarget.name !== _getValueInputName) {
        onBlur({
          target: {
            name,
            value,
          },
        })
      }
    },
    [onBlur, name, value, _getKeyInputName, _getValueInputName],
  )

  const handleOptionComposition = useCallback(() => {
    let newOptions
    if (currentValue) {
      newOptions = options.filter(v => !fieldValue.includes(v.value) || v.value === currentValue)
    } else {
      newOptions = options.filter(v => !fieldValue.includes(v.value))
    }

    let taggedOptions = newOptions
    if (fieldValue.length >= 2 && filterByTag) {
      const selectedOption = options.find(v => v.value === fieldValue[0])
      taggedOptions = newOptions.filter(v => selectedOption.tag === v.tag)
    }

    setNewOptions(taggedOptions)
  }, [currentValue, options, fieldValue, filterByTag])

  const showRemoveButton = atLeastOneEntry ? index !== 0 : true

  return (
    <div className={classnames(style.inputContainer, 'd-flex j-start mb-cs-s')}>
      {!indexAsKey && (
        <InputElement
          data-test-id={_getKeyInputName}
          className="flex-grow mr-cs-s"
          name={_getKeyInputName}
          placeholder={keyPlaceholder}
          type="text"
          onChange={handleKeyChanged}
          onBlur={handleBlur}
          value={value.key}
          readOnly={readOnly}
          code
          {...additionalInputProps}
        />
      )}
      <InputElement
        data-test-id={_getValueInputName}
        className={classnames(style.input, { [style.inputIndexAsKey]: indexAsKey })}
        name={_getValueInputName}
        placeholder={valuePlaceholder}
        type="text"
        onChange={handleValueChanged}
        onBlur={handleBlur}
        onFocus={distinctOptions && options ? handleOptionComposition : undefined}
        value={indexAsKey ? value : value.value}
        readOnly={readOnly}
        code
        options={options ? (newOptions ?? options) : undefined}
        {...additionalInputPropsRest}
      />
      {showRemoveButton && (
        <Button
          type="button"
          onClick={handleRemoveButtonClick}
          icon={IconTrash}
          title={m.deleteEntry}
          message={removeMessage}
          disabled={readOnly}
          secondary
          danger
        />
      )}
    </div>
  )
}

Entry.propTypes = {
  additionalInputProps: PropTypes.shape({
    options: PropTypes.array,
  }).isRequired,
  atLeastOneEntry: PropTypes.bool,
  distinctOptions: PropTypes.bool,
  fieldValue: PropTypes.any,
  filterByTag: PropTypes.bool,
  index: PropTypes.number.isRequired,
  indexAsKey: PropTypes.bool.isRequired,
  inputElement: PropTypes.elementType.isRequired,
  keyPlaceholder: PropTypes.message.isRequired,
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func.isRequired,
  onChange: PropTypes.func.isRequired,
  onRemoveButtonClick: PropTypes.func.isRequired,
  readOnly: PropTypes.bool,
  removeMessage: PropTypes.message,
  value: PropTypes.oneOfType([
    PropTypes.shape({
      key: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
      value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
    }),
    PropTypes.string,
  ]),
  valuePlaceholder: PropTypes.message.isRequired,
}

Entry.defaultProps = {
  value: undefined,
  readOnly: false,
  removeMessage: sharedMessages.remove,
  distinctOptions: false,
  fieldValue: undefined,
  atLeastOneEntry: false,
  filterByTag: false,
}

export default Entry
