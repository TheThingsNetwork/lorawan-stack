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
import { defineMessages } from 'react-intl'
import classnames from 'classnames'

import Button from '@ttn-lw/components/button'

import PropTypes from '@ttn-lw/lib/prop-types'

import { hexToBase64, base64ToHex } from '@console/lib/bytes'

import Entry from './entry'

import style from './key-value-map.styl'

const m = defineMessages({
  addEntry: 'Add entry',
})

class KeyValueMap extends React.PureComponent {
  static propTypes = {
    addMessage: PropTypes.message,
    className: PropTypes.string,
    disabled: PropTypes.bool,
    indexAsKey: PropTypes.bool,
    keyPlaceholder: PropTypes.message,
    name: PropTypes.string.isRequired,
    onBlur: PropTypes.func,
    onChange: PropTypes.func,
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

  static defaultProps = {
    className: undefined,
    onBlur: () => null,
    onChange: () => null,
    value: [],
    addMessage: m.addEntry,
    indexAsKey: false,
    keyPlaceholder: '',
    disabled: false,
  }

  @bind
  handlePasswordUsernameChange(newValues) {
    const { onChange, value } = this.props
    const encodedNewValues = newValues

    // eslint-disable-next-line no-console
    console.log(encodedNewValues)

    onChange([...value, ...encodedNewValues])
  }

  @bind
  handleEntryChange(index, newValues) {
    const { onChange, value, indexAsKey } = this.props

    onChange(
      value.map((val, idx) => {
        if (index !== idx) {
          return val
        }

        return indexAsKey ? newValues.value : { ...val, ...newValues }
      }),
    )
  }

  @bind
  removeEntry(index) {
    const { onChange, value } = this.props

    onChange(value.filter((_, i) => i !== index) || [], true)
  }

  @bind
  addEmptyEntry() {
    const { onChange, value, indexAsKey } = this.props
    const entry = indexAsKey ? undefined : { key: '', value: undefined }

    onChange([...value, entry])
  }

  @bind
  decodeValue(value) {
    if (value.key === 'Authorization' && value.value.startsWith('Basic')) {
      const [username, ...password] = hexToBase64(value.value).split(':')
      const decodedValue = {
        key: username,
        value: password.join(':'),
      }
      return decodedValue
    }

    return value
  }

  render() {
    const {
      className,
      name,
      value,
      keyPlaceholder,
      valuePlaceholder,
      addMessage,
      onBlur,
      indexAsKey,
      disabled,
    } = this.props

    const mayShowAddBasicAuthButton = !(
      value.some(header => header.key === 'Authorization') &&
      value.some(header => header.value?.startsWith('Basic'))
    )

    return (
      <div data-test-id={'key-value-map'} className={classnames(className, style.container)}>
        <div>
          {value &&
            value.map((value, index) => (
              <Entry
                key={`${name}[${index}]`}
                name={name}
                value={this.decodeValue(value)}
                keyPlaceholder={keyPlaceholder}
                valuePlaceholder={valuePlaceholder}
                index={index}
                onRemoveButtonClick={this.removeEntry}
                onChange={this.handleEntryChange}
                onBlur={onBlur}
                indexAsKey={indexAsKey}
              />
            ))}
        </div>
        <div>
          <Button
            name={`${name}.push`}
            type="button"
            message={addMessage}
            onClick={this.addEmptyEntry}
            disabled={disabled}
            icon="add"
          />
        </div>
        <div>
          {mayShowAddBasicAuthButton && (
            <Button
              name={`${name}.push`}
              type="button"
              message={'add basic auth'}
              onClick={this.addEmptyEntry}
              disabled={disabled}
              icon="add"
              secondary
            />
          )}
        </div>
      </div>
    )
  }
}

export default KeyValueMap
