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
import { defineMessages, injectIntl } from 'react-intl'
import classnames from 'classnames'

import PropTypes from '../../../lib/prop-types'
import { RIGHT_ALL } from '../../lib/rights'
import withComputedProps from '../../../lib/components/with-computed-props'

import Checkbox from '../../../components/checkbox'
import Notification from '../../../components/notification'
import Radio from '../../../components/radio-button'

import style from './rights-group.styl'

const m = defineMessages({
  selectAll: 'Select All',
  outOfOwnScopeRights:
    'This {entityType} has more rights than you have. These rights can not be modified.',
  outOfOwnScopePseudoRight:
    "This {entityType} has a wildcard right that you don't have. The {entityType} can therefore only be removed entirely.",
  grantType: 'Grant type',
  allCurrentAndFutureRights: 'Grant all current and future rights',
  selectIndividualRights: 'Grant individual rights',
})

const computeProps = function(props) {
  const { value, pseudoRight: grantablePseudoRight, rights: grantableRights } = props

  // Extract the pseudo right from own rights or granted rights
  let derivedPseudoRight = []
  if (grantablePseudoRight && !Array.isArray(grantablePseudoRight)) {
    derivedPseudoRight = [grantablePseudoRight]
  } else if (Boolean(grantablePseudoRight) && Array.isArray(grantablePseudoRight)) {
    derivedPseudoRight = value.filter(right => right !== RIGHT_ALL && right.endsWith('_ALL'))
  }

  // Filter out rights that the entity has but may not be granted by the user
  const outOfOwnScopeRights = !Boolean(grantablePseudoRight)
    ? value.filter(right => !grantableRights.includes(right))
    : []

  // Extract all rights by combining granted and grantable rights
  const derivedRights = [...grantableRights, ...outOfOwnScopeRights].sort()

  // Store whether out of scope pseudo rights are present
  const hasOutOfOwnScopePseudoRight =
    outOfOwnScopeRights.filter(right => right.endsWith('_ALL')).length !== 0

  // Store granted individual rights
  const grantedIndividualRights = value.filter(right => !derivedPseudoRight.includes(right))

  // Store out of own scope individual rights
  const outOfOwnScopeIndividualRights = !Boolean(grantablePseudoRight)
    ? grantedIndividualRights.filter(right => !grantableRights.includes(right))
    : []

  // Determine whether a pseudo right is granted
  const hasPseudoRightGranted =
    value.includes(RIGHT_ALL) ||
    derivedPseudoRight.some(derivedRight => value.includes(derivedRight))

  // Determine the current grant type
  const grantType = hasPseudoRightGranted ? 'pseudo' : 'individual'

  return {
    outOfOwnScopeIndividualRights,
    hasOutOfOwnScopePseudoRight,
    derivedPseudoRight,
    derivedRights,
    hasPseudoRightGranted,
    grantType,
    ...props,
  }
}

@withComputedProps(computeProps)
@injectIntl
@bind
class RightsGroup extends React.Component {
  static propTypes = {
    /** The class to be added to the container */
    className: PropTypes.string,
    /** The rights derived from the granted and grantable rights **/
    derivedPseudoRight: PropTypes.oneOfType(PropTypes.right, PropTypes.arrayOf(PropTypes.right)),
    /** The pseudo right derived from the current entity or user **/
    derivedRights: PropTypes.rights.isRequired,
    /** A flag indicating whether the whole component should be disabled **/
    disabled: PropTypes.bool,
    /** The message depicting the type of entity this component is setting the
     * rights for.
     */
    entityTypeMessage: PropTypes.message.isRequired,
    /** The right grant type **/
    grantType: PropTypes.oneOf(['pseudo', 'individual']).isRequired,
    /** Whether the entity has a pseudo right that the current use does not have **/
    hasOutOfOwnScopePseudoRight: PropTypes.bool.isRequired,
    /** Whether the entity has a pseudo right granted **/
    hasPseudoRightGranted: PropTypes.bool.isRequired,
    /** The intl object provided by injectIntl of react-intl, used to translate
     * messages
     */
    intl: PropTypes.shape({
      formatMessage: PropTypes.func.isRequired,
    }).isRequired,
    /** Blur event hook */
    onBlur: PropTypes.func,
    /** Change event hook */
    onChange: PropTypes.func,
    /** A list of rights that are outside the scope of the current user **/
    outOfOwnScopeIndividualRights: PropTypes.rights.isRequired,
    /** The pseudo right literal comprising all other rights */
    pseudoRight: PropTypes.string,
    /** The rights value */
    value: PropTypes.rights.isRequired,
  }

  static defaultProps = {
    className: undefined,
    disabled: false,
    onBlur: () => null,
    onChange: () => null,
    pseudoRight: undefined,
    derivedPseudoRight: undefined,
  }

  state = {
    individualRightValue: [],
  }

  static getDerivedStateFromProps(props, state) {
    const { individualRightValue: oldIndividualRightValue } = state
    const { value, hasPseudoRightGranted } = props

    // Store the individual right values when the grant type is changed to
    // pseudo right
    const individualRightValue = !hasPseudoRightGranted ? value : oldIndividualRightValue

    return { individualRightValue }
  }

  handleChangeAll(event) {
    const { onChange, outOfOwnScopeIndividualRights, derivedRights } = this.props
    const { checked } = event.target

    let value

    if (checked) {
      // Fill up with individual rights
      value = [...derivedRights]
    } else {
      // On uncheck, leave out of scope rights checked, if present
      value = [...outOfOwnScopeIndividualRights]
    }

    onChange(value)
  }

  handleChange(val) {
    const { onChange } = this.props
    const value = Object.keys(val).filter(right => val[right])

    onChange(value)
  }

  handleGrantTypeChange(val) {
    const { onChange, pseudoRight } = this.props
    const { individualRightValue } = this.state

    if (val === 'pseudo') {
      onChange([pseudoRight])
    } else {
      onChange(individualRightValue)
    }
  }

  render() {
    const {
      intl,
      className,
      disabled,
      entityTypeMessage,
      onBlur,
      outOfOwnScopeIndividualRights,
      hasOutOfOwnScopePseudoRight,
      grantType,
      derivedPseudoRight,
      derivedRights,
    } = this.props
    const { individualRightValue } = this.state

    const selectedCheckboxesCount = individualRightValue.filter(right => !right.endsWith('_ALL'))
      .length
    const totalCheckboxesCount = derivedRights.length
    const allSelected = selectedCheckboxesCount === totalCheckboxesCount
    const indeterminate =
      selectedCheckboxesCount !== 0 && selectedCheckboxesCount !== totalCheckboxesCount
    const allDisabled = grantType === 'pseudo' || disabled || hasOutOfOwnScopePseudoRight

    let selectAllName = 'select-all'
    let selectAllTitle = m.selectAll
    if (Boolean(derivedPseudoRight) && !Array.isArray(derivedPseudoRight)) {
      selectAllName = derivedPseudoRight
      selectAllTitle = { id: `enum:${derivedPseudoRight}` }
    }

    // Marshal rights to key/value for checkbox group
    const rightsValues = derivedRights.reduce(function(acc, right) {
      acc[right] = allSelected || individualRightValue.includes(right)

      return acc
    }, {})

    const cbs = derivedRights.map(right => (
      <Checkbox
        className={style.rightLabel}
        key={right}
        name={right}
        disabled={outOfOwnScopeIndividualRights.includes(right)}
        label={{ id: `enum:${right}` }}
      />
    ))

    return (
      <div className={className}>
        {hasOutOfOwnScopePseudoRight && (
          <Notification
            small
            warning={m.outOfOwnScopePseudoRight}
            messageValues={{ entityType: intl.formatMessage(entityTypeMessage).toLowerCase() }}
          />
        )}
        <Radio.Group
          className={style.grantType}
          name="grant_type"
          value={grantType}
          onChange={this.handleGrantTypeChange}
          disabled={!Boolean(derivedPseudoRight)}
        >
          <Radio label={m.allCurrentAndFutureRights} value="pseudo" />
          <Radio label={m.selectIndividualRights} value="individual" />
        </Radio.Group>
        <Checkbox
          className={classnames(style.selectAll, style.rightLabel)}
          name={selectAllName}
          label={selectAllTitle}
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

export default RightsGroup
