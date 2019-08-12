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
import { RIGHT_ALL } from '../../lib/rights'

import Checkbox from '../../../components/checkbox'
import Notification from '../../../components/notification'

import style from './rights-group.styl'

const m = defineMessages({
  selectAll: 'Select All',
  outOfOwnScopeRights:
    'This entity possesses rights that are out of your scope of granted rights. These rights cannot be altered.',
  outOfOwnScopeRightsStrict:
    'This entity possesses rights that are out of your scope of granted rights. Modifying is hence prohibited.',
})

const computeState = function(values, rights, universalRight) {
  const selectedCheckboxesCount = values.length
  const totalCheckboxesCount = rights.length

  const universalRightIsChecked = values.includes(RIGHT_ALL) || values.includes(universalRight)

  return {
    allSelected: universalRightIsChecked || selectedCheckboxesCount === totalCheckboxesCount,
    indeterminate:
      !universalRightIsChecked &&
      selectedCheckboxesCount !== 0 &&
      selectedCheckboxesCount !== totalCheckboxesCount,
  }
}

@bind
class RightsGroup extends React.Component {
  state = {}

  static getDerivedStateFromProps(props) {
    if ('value' in props) {
      const { value, rights: grantableRights, universalRight: grantableUniversalRight } = props
      let universalRight = grantableUniversalRight
      const allGrantableRights = [...grantableRights, ...grantableUniversalRight]

      // Identify given rights that are out of the scope of the current user
      const outOfOwnScopeRights = value.filter(function(right) {
        if (!allGrantableRights.includes(right) && right !== RIGHT_ALL) {
          if (right.endsWith('_ALL')) {
            universalRight = right
          }
          return true
        }
        return false
      })

      // Compose rights list
      const rights = [...outOfOwnScopeRights, ...grantableRights]

      const { allSelected, indeterminate } = computeState(value, rights, universalRight)

      return {
        allSelected,
        indeterminate,
        outOfOwnScopeRights,
        rights,
        universalRight,
        value,
      }
    }

    return null
  }

  async handleChangeAll(event) {
    const { onChange } = this.props
    const { rights, outOfOwnScopeRights, universalRight } = this.state
    const { checked } = event.target

    let value

    // Determine new value based on universal rights and out of scope rights
    if (checked) {
      if (universalRight) {
        // Prefer universal right value, if present
        value = [universalRight]
      } else {
        // Else add rights individually
        value = [...rights]
      }
    } else {
      // On uncheck, leave out of scope rights checked, if present
      value = [...outOfOwnScopeRights]
    }

    onChange(value)
  }

  async handleChange(val) {
    const value = Object.keys(val).filter(right => val[right])
    const { onChange, rights } = this.props
    const { universalRight } = this.state
    const { allSelected } = computeState(value, rights, universalRight)

    // Set new right value and prefer universal right if applicable
    const result = universalRight && allSelected ? [universalRight] : [...value]

    onChange(result)
  }

  render() {
    const { className, name, onBlur, disabled, strict } = this.props

    const {
      indeterminate,
      outOfOwnScopeRights,
      value,
      allSelected,
      rights,
      universalRight,
    } = this.state

    const cbs = rights.map(right => (
      <Checkbox
        className={style.rightLabel}
        key={right}
        name={right}
        label={{ id: `enum:${right}` }}
        disabled={outOfOwnScopeRights.includes(right)}
      />
    ))

    const hasRightAll = Boolean(value.includes(RIGHT_ALL))
    const hasOutOfOwnScopeRights = Boolean(outOfOwnScopeRights.length)
    const allDisabled =
      disabled || outOfOwnScopeRights.includes(universalRight) || (strict && hasOutOfOwnScopeRights)

    // Marshal rights to key/value for checkbox group
    const rightsValues = rights.reduce(
      function(acc, right) {
        acc[right] = allSelected || value.includes(right)

        return acc
      },
      { [RIGHT_ALL]: hasRightAll },
    )

    return (
      <div className={className}>
        {hasOutOfOwnScopeRights && (
          <Notification small info={strict ? m.outOfOwnScopeRightsStrict : m.outOfOwnScopeRights} />
        )}
        <Checkbox
          className={classnames(style.selectAll, style.rightLabel)}
          name={universalRight || 'select-all'}
          label={universalRight ? { id: `enum:${universalRight}` } : m.selectAll}
          onChange={this.handleChangeAll}
          indeterminate={indeterminate}
          value={allSelected}
          disabled={allDisabled}
        />
        <Checkbox.Group
          className={style.group}
          horizontal
          name={name}
          value={rightsValues}
          onChange={this.handleChange}
          onBlur={onBlur}
          disabled={allDisabled}
        >
          {cbs}
        </Checkbox.Group>
      </div>
    )
  }
}

RightsGroup.propTypes = {
  /** The class to be added to the container */
  className: PropTypes.string,
  /** The name prop, used to connect to formik */
  name: PropTypes.string.isRequired,
  /** The rights value */
  value: PropTypes.array,
  /** Change event hook */
  onChange: PropTypes.func,
  /** Blur event hook */
  onBlur: PropTypes.func,
  /** The universal right literal comprising all other rights */
  universalRight: PropTypes.string,
  /** The list of rights options */
  rights: PropTypes.arrayOf(PropTypes.string),
  /** A flag identifying whether modifying rights is allowed when out of scope
   * rights are present. Can be used to prevent user error.
   */
  strict: PropTypes.bool,
}

RightsGroup.defaultProps = {
  onChange: () => null,
  onBlur: () => null,
  rights: [],
  value: [],
  strict: false,
}

export default RightsGroup
