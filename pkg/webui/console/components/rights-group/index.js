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

import PropTypes from '../../../lib/prop-types'
import Checkbox from '../../../components/checkbox'

import style from './rights-group.styl'

const m = defineMessages({
  selectAll: 'Select All',
})

const computeState = function (values, rights) {
  const selectedCheckboxesCount = rights
    .reduce((count, val) => values[val] ? count + 1 : count, 0)
  const totalCheckboxesCount = rights.length

  return {
    allSelected: selectedCheckboxesCount === totalCheckboxesCount,
    indeterminate: selectedCheckboxesCount !== 0 && selectedCheckboxesCount !== totalCheckboxesCount,
  }
}

@bind
class RightsGroup extends React.Component {

  state = {}

  static getDerivedStateFromProps (props) {
    if ('value' in props) {
      const { value, rights } = props
      const { allSelected, indeterminate } = computeState(value, rights)

      return {
        allSelected,
        indeterminate,
        rights,
        value,
      }
    }

    return null
  }

  async handleChangeAll (event) {
    const { onChange, rights, universalRight } = this.props
    const { checked } = event.target

    const value = rights.reduce((values, right) => ({
      ...values,
      [right]: checked,
    }), {})

    const newValues = !('value' in this.props) ? { value } : {}

    await this.setState({
      allSelected: checked,
      indeterminate: false,
      ...newValues,
    })

    if (universalRight) {
      value[universalRight] = checked
    }

    onChange(value)
  }

  async handleChange (value) {
    const { onChange, rights, universalRight } = this.props
    const { allSelected, indeterminate } = computeState(value, rights)

    let newValues = {}
    if (!('value' in this.props)) {
      newValues = { value }
    }

    await this.setState({
      allSelected,
      indeterminate,
      ...newValues,
    })

    const result = universalRight
      ? { ...value, [universalRight]: allSelected }
      : value

    onChange(result)
  }

  render () {
    const {
      className,
      name,
      onBlur,
      universalRight,
      disabled,
    } = this.props

    const {
      indeterminate,
      value,
      allSelected,
      rights,
    } = this.state

    const cbs = rights
      .map(right => (
        <Checkbox
          className={style.rightLabel}
          key={right}
          name={right}
          label={{ id: `enum:${right}` }}
        />
      ))

    return (
      <div className={className}>
        <Checkbox
          className={classnames(style.selectAll, style.rightLabel)}
          name={universalRight || 'select-all'}
          label={universalRight ? { id: `enum:${universalRight}` } : m.selectAll}
          onChange={this.handleChangeAll}
          indeterminate={indeterminate}
          value={allSelected}
          disabled={disabled}
        />
        <Checkbox.Group
          className={style.group}
          horizontal
          name={name}
          value={value}
          onChange={this.handleChange}
          onBlur={onBlur}
          disabled={disabled}
        >
          {cbs}
        </Checkbox.Group>
      </div>
    )
  }
}

RightsGroup.propTypes = {
  className: PropTypes.string,
  name: PropTypes.string.isRequired,
  value: PropTypes.object,
  onChange: PropTypes.func,
  onBlur: PropTypes.func,
  universalRight: PropTypes.string,
  rights: PropTypes.arrayOf(PropTypes.string),
}

RightsGroup.defaultProps = {
  onChange: () => null,
  onBlur: () => null,
  rights: [],
}

export default RightsGroup
