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

import Input from '@ttn-lw/components/input'
import Button from '@ttn-lw/components/button'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './key-value-map.styl'

const m = defineMessages({
  deleteEntry: 'Delete entry',
})

class Entry extends React.Component {
  static propTypes = {
    readOnly: PropTypes.bool,
  }
  static defaultProps = { readOnly: false }

  _getKeyInputName() {
    const { name, index } = this.props

    return `${name}[${index}].key`
  }

  _getValueInputName() {
    const { name, index } = this.props

    return `${name}[${index}].value`
  }

  @bind
  handleRemoveButtonClicked(event) {
    const { onRemoveButtonClick, index } = this.props

    onRemoveButtonClick(index, event)
  }

  @bind
  handleKeyChanged(newKey) {
    const { onChange, index } = this.props

    onChange(index, { key: newKey })
  }

  @bind
  handleValueChanged(newValue) {
    const { onChange, index } = this.props

    onChange(index, { value: newValue })
  }

  @bind
  handleBlur(event) {
    const { name, onBlur, value } = this.props

    const { relatedTarget } = event
    const nextTarget = relatedTarget || {}

    if (
      nextTarget.name !== this._getKeyInputName() &&
      nextTarget.name !== this._getValueInputName()
    ) {
      onBlur({
        target: {
          name,
          value,
        },
      })
    }
  }

  render() {
    const { keyPlaceholder, valuePlaceholder, value, indexAsKey, readOnly } = this.props

    return (
      <div className={style.entriesRow}>
        {!indexAsKey && (
          <Input
            data-test-id={this._getKeyInputName()}
            className={style.input}
            name={this._getKeyInputName()}
            placeholder={keyPlaceholder}
            type="text"
            onChange={this.handleKeyChanged}
            onBlur={this.handleBlur}
            value={value.key}
            readOnly={readOnly}
            code
          />
        )}
        <Input
          data-test-id={this._getValueInputName()}
          className={classnames(style.input, { [style.inputIndexAsKey]: indexAsKey })}
          name={this._getValueInputName()}
          placeholder={valuePlaceholder}
          type="text"
          onChange={this.handleValueChanged}
          onBlur={this.handleBlur}
          value={indexAsKey ? value : value.value}
          readOnly={readOnly}
          code
        />
        <Button
          type="button"
          onClick={this.handleRemoveButtonClicked}
          icon="delete"
          title={m.deleteEntry}
          danger
        />
      </div>
    )
  }
}

Entry.propTypes = {
  index: PropTypes.number.isRequired,
  indexAsKey: PropTypes.bool.isRequired,
  keyPlaceholder: PropTypes.message.isRequired,
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func.isRequired,
  onChange: PropTypes.func.isRequired,
  onRemoveButtonClick: PropTypes.func.isRequired,
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
}

export default Entry
