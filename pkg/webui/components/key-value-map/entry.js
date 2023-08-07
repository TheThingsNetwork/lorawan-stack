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

import React, { useCallback, useMemo } from 'react'
import { defineMessages } from 'react-intl'
import classnames from 'classnames'

import Button from '@ttn-lw/components/button'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './key-value-map.styl'

const m = defineMessages({
  deleteEntry: 'Delete entry',
})

const Entry = ({
  readOnly,
  name,
  value,
  index,
  onRemoveButtonClick,
  onChange,
  onBlur,
  inputElement: InputElement,
  indexAsKey,
  valuePlaceholder,
  keyPlaceholder,
  additionalInputProps,
}) => {
  const _getKeyInputName = useMemo(() => `${name}[${index}].key`, [index, name])

  const _getValueInputName = useMemo(() => `${name}[${index}].value`, [index, name])

  const handleRemoveButtonClicked = useCallback(
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
      onChange(index, { value: newValue })
    },
    [index, onChange],
  )

  const handleBlur = useCallback(
    event => {
      const { relatedTarget } = event
      const nextTarget = relatedTarget || {}

      if (nextTarget.name !== _getKeyInputName() && nextTarget.name !== _getValueInputName()) {
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

  return (
    <div className={style.entriesRow}>
      {!indexAsKey && (
        <InputElement
          data-test-id={_getKeyInputName()}
          className={style.input}
          name={_getKeyInputName()}
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
        data-test-id={_getValueInputName()}
        className={classnames(style.input, { [style.inputIndexAsKey]: indexAsKey })}
        name={_getValueInputName()}
        placeholder={valuePlaceholder}
        type="text"
        onChange={handleValueChanged}
        onBlur={handleBlur}
        value={indexAsKey ? value : value.value}
        readOnly={readOnly}
        code
        {...additionalInputProps}
      />
      <Button
        type="button"
        onClick={handleRemoveButtonClicked}
        icon="delete"
        title={m.deleteEntry}
        disabled={readOnly}
        danger
      />
    </div>
  )
}

Entry.propTypes = {
  additionalInputProps: PropTypes.shape({}).isRequired,
  index: PropTypes.number.isRequired,
  indexAsKey: PropTypes.bool.isRequired,
  inputElement: PropTypes.elementType.isRequired,
  keyPlaceholder: PropTypes.message.isRequired,
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func.isRequired,
  onChange: PropTypes.func.isRequired,
  onRemoveButtonClick: PropTypes.func.isRequired,
  readOnly: PropTypes.bool,
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
}

export default Entry
