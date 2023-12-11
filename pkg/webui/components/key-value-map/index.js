// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'
import classnames from 'classnames'

import Button from '@ttn-lw/components/button'
import Input from '@ttn-lw/components/input'

import PropTypes from '@ttn-lw/lib/prop-types'

import Entry from './entry'

import style from './key-value-map.styl'

const m = defineMessages({
  addEntry: 'Add entry',
})

const KeyValueMap = ({
  addMessage,
  removeMessage,
  additionalInputProps,
  className,
  disabled,
  indexAsKey,
  inputElement,
  isReadOnly,
  keyPlaceholder,
  name,
  onBlur,
  onChange,
  value,
  valuePlaceholder,
  distinctOptions,
  atLeastOneEntry,
  filterByTag,
}) => {
  const handleEntryChange = useCallback(
    (index, newValues) => {
      onChange(
        value.map((val, idx) => {
          if (index !== idx) {
            return val
          }

          return indexAsKey ? newValues.value : { ...val, ...newValues }
        }),
      )
    },
    [indexAsKey, onChange, value],
  )

  const removeEntry = useCallback(
    index => {
      onChange(value.filter((_, i) => i !== index) || [], true)
    },
    [onChange, value],
  )

  const addEmptyEntry = useCallback(() => {
    const entry = indexAsKey ? '' : { key: '', value: '' }

    onChange([...value, entry])
  }, [indexAsKey, onChange, value])

  return (
    <div data-test-id={'key-value-map'} className={classnames(className, style.container)}>
      <div>
        {value &&
          value.map((individualValue, index) => (
            <Entry
              key={`${name}[${index}]`}
              name={name}
              value={individualValue}
              fieldValue={value}
              keyPlaceholder={keyPlaceholder}
              valuePlaceholder={valuePlaceholder}
              index={index}
              onRemoveButtonClick={removeEntry}
              onChange={handleEntryChange}
              onBlur={onBlur}
              indexAsKey={indexAsKey}
              readOnly={isReadOnly(individualValue)}
              inputElement={inputElement}
              additionalInputProps={additionalInputProps}
              removeMessage={removeMessage}
              distinctOptions={distinctOptions}
              atLeastOneEntry={atLeastOneEntry}
              filterByTag={filterByTag}
            />
          ))}
      </div>
      <div>
        <Button
          name={`${name}.push`}
          type="button"
          message={addMessage}
          onClick={addEmptyEntry}
          disabled={disabled}
          icon="add"
        />
      </div>
    </div>
  )
}

KeyValueMap.propTypes = {
  addMessage: PropTypes.message,
  additionalInputProps: PropTypes.shape({}),
  atLeastOneEntry: PropTypes.bool,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  distinctOptions: PropTypes.bool,
  filterByTag: PropTypes.bool,
  indexAsKey: PropTypes.bool,
  inputElement: PropTypes.elementType,
  isReadOnly: PropTypes.func,
  keyPlaceholder: PropTypes.message,
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func,
  onChange: PropTypes.func,
  removeMessage: PropTypes.message,
  value: PropTypes.arrayOf(
    PropTypes.oneOfType([
      PropTypes.shape({
        key: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
        value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
      }),
      PropTypes.string,
    ]),
  ),
  valuePlaceholder: PropTypes.message.isRequired,
}

KeyValueMap.defaultProps = {
  additionalInputProps: {},
  className: undefined,
  onBlur: () => null,
  onChange: () => null,
  value: [],
  addMessage: m.addEntry,
  indexAsKey: false,
  keyPlaceholder: '',
  disabled: false,
  isReadOnly: () => null,
  inputElement: Input,
  removeMessage: undefined,
  distinctOptions: false,
  atLeastOneEntry: false,
  filterByTag: false,
}

export default KeyValueMap
