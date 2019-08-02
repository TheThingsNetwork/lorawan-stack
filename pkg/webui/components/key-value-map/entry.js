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

import PropTypes from '../../lib/prop-types'
import Input from '../input'
import Button from '../button'

import style from './key-value-map.styl'

const m = defineMessages({
  deleteEntry: 'Delete Entry',
})

@bind
class Entry extends React.Component {

  handleRemoveButtonClicked (event) {
    const { onRemoveButtonClick, index } = this.props
    onRemoveButtonClick(index, event)
  }

  handleKeyChanged (newKey) {
    const { onChange, index } = this.props
    onChange(index, { key: newKey })
  }

  handleValueChanged (newValue) {
    const { onChange, index } = this.props
    onChange(index, { value: newValue })
  }

  render () {
    const {
      name,
      index,
      keyPlaceholder,
      valuePlaceholder,
      value,
      onBlur,
    } = this.props

    return (
      <div className={style.entriesRow}>
        <Input
          className={style.input}
          name={`${name}[${index}].key`}
          placeholder={keyPlaceholder}
          type="text"
          onChange={this.handleKeyChanged}
          value={value.key}
        />
        <Input
          className={style.input}
          name={`${name}[${index}].value`}
          placeholder={valuePlaceholder}
          type="text"
          onChange={this.handleValueChanged}
          onBlur={onBlur}
          value={value.value}
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
  className: PropTypes.string,
  name: PropTypes.string.isRequired,
  value: PropTypes.object.isRequired,
  keyPlaceholder: PropTypes.message.isRequired,
  valuePlaceholder: PropTypes.message.isRequired,
  index: PropTypes.number.isRequired,
  onRemoveButtonClick: PropTypes.func.isRequired,
  onChange: PropTypes.func.isRequired,
  onBlur: PropTypes.func.isRequired,
}

export default Entry
