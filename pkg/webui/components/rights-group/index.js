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

import React, { useCallback, useEffect } from 'react'
import { defineMessages, useIntl } from 'react-intl'
import classnames from 'classnames'

import Checkbox from '@ttn-lw/components/checkbox'
import Notification from '@ttn-lw/components/notification'
import Radio from '@ttn-lw/components/radio-button'

import Message from '@ttn-lw/lib/components/message'

import useComputedProps from '@ttn-lw/lib/hooks/use-computed-props'

import { RIGHT_ALL } from '@console/lib/rights'

import style from './rights-group.styl'

const m = defineMessages({
  selectAll: 'Select all',
  outOfOwnScopeRights:
    'This {entityType} has more rights than you have. These rights can not be modified.',
  outOfOwnScopePseudoRight:
    "This {entityType} has a wildcard right that you don't have. The {entityType} can therefore only be removed entirely.",
  grantType: 'Grant type',
  allCurrentAndFutureRights: 'Grant all current and future rights',
  selectIndividualRights: 'Grant individual rights',
  RIGHT_APPLICATION_LINK_DESCRIPTION:
    'This implicitly includes the rights to view application information, read application traffic and write downlinks',
})

const computeProps = props => {
  const { value, pseudoRight: grantablePseudoRight, rights: grantableRights } = props

  // Extract the pseudo right from own rights or granted rights.
  let derivedPseudoRight = []
  if (grantablePseudoRight && !Array.isArray(grantablePseudoRight)) {
    derivedPseudoRight = [grantablePseudoRight]
  } else if (grantablePseudoRight && Array.isArray(grantablePseudoRight)) {
    derivedPseudoRight = grantablePseudoRight
  } else {
    derivedPseudoRight = value.filter(right => right !== RIGHT_ALL && right.endsWith('_ALL'))
  }
  // Filter out rights that the entity has but may not be granted by the user.
  const outOfOwnScopeRights = !Boolean(grantablePseudoRight)
    ? value.filter(right => !grantableRights.includes(right))
    : []

  // Extract all rights by combining granted and grantable rights.
  const derivedRights = [...grantableRights, ...outOfOwnScopeRights].sort()

  // Store whether out of scope pseudo rights are present.
  const hasOutOfOwnScopePseudoRight =
    outOfOwnScopeRights.filter(right => right.endsWith('_ALL')).length !== 0

  // Store granted individual rights.
  const grantedIndividualRights = value.filter(right => !derivedPseudoRight.includes(right))

  // Store out of own scope individual rights.
  const outOfOwnScopeIndividualRights = !Boolean(grantablePseudoRight)
    ? grantedIndividualRights.filter(right => !grantableRights.includes(right))
    : []

  // Determine whether a pseudo right is granted.
  const hasPseudoRightGranted =
    value.includes(RIGHT_ALL) ||
    derivedPseudoRight.some(derivedRight => value.includes(derivedRight))

  // Determine the current grant type.
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

const RightsGroup = props => {
  const {
    className,
    derivedPseudoRight,
    derivedRights,
    disabled,
    entityTypeMessage,
    grantType,
    hasOutOfOwnScopePseudoRight,
    hasPseudoRightGranted,
    onBlur,
    onChange,
    outOfOwnScopeIndividualRights,
    pseudoRight,
    value,
  } = useComputedProps(computeProps, props)
  const intl = useIntl()
  const { formatMessage } = intl
  const [individualRightValue, setIndividualRightValue] = React.useState([])

  useEffect(() => {
    const newIndividualRightValue = !hasPseudoRightGranted ? value : individualRightValue
    setIndividualRightValue(newIndividualRightValue)
  }, [value, hasPseudoRightGranted, individualRightValue])

  const handleChangeAll = useCallback(
    event => {
      const { checked } = event.target

      let value

      if (checked) {
        // Fill up with individual rights.
        value = [...derivedRights]
      } else {
        // On uncheck, leave out of scope rights checked, if present.
        value = [...outOfOwnScopeIndividualRights]
      }

      onChange(value)
    },
    [onChange, derivedRights, outOfOwnScopeIndividualRights],
  )

  const handleChange = useCallback(
    val => {
      const value = Object.keys(val).filter(right => val[right])

      onChange(value)
    },
    [onChange],
  )

  const handleGrantTypeChange = useCallback(
    val => {
      if (val === 'pseudo') {
        onChange(Array.isArray(pseudoRight) ? pseudoRight : [pseudoRight])
      } else {
        onChange(individualRightValue)
      }
    },
    [onChange, individualRightValue, pseudoRight],
  )

  const selectedCheckboxesCount = individualRightValue.filter(
    right => !right.endsWith('_ALL'),
  ).length
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

  // Marshal rights to key/value for checkbox group.
  const rightsValues = derivedRights.reduce((acc, right) => {
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
      children={
        Boolean(m[`${right}_DESCRIPTION`]) && (
          <Message
            className={style.description}
            component="div"
            content={m[`${right}_DESCRIPTION`]}
          />
        )
      }
    />
  ))

  const noRights = derivedPseudoRight.length === 0 || derivedRights.length === 0

  return (
    <div className={className}>
      {hasOutOfOwnScopePseudoRight && (
        <Notification
          small
          warning
          content={m.outOfOwnScopePseudoRight}
          messageValues={{ entityType: formatMessage(entityTypeMessage).toLowerCase() }}
        />
      )}
      <Radio.Group
        className={style.grantType}
        name="grant_type"
        value={grantType}
        onChange={handleGrantTypeChange}
        disabled={noRights}
      >
        <Radio label={m.allCurrentAndFutureRights} value="pseudo" />
        <Radio label={m.selectIndividualRights} value="individual" />
      </Radio.Group>
      <Checkbox
        className={classnames(style.selectAll, style.rightLabel)}
        name={selectAllName}
        label={selectAllTitle}
        onChange={handleChangeAll}
        indeterminate={indeterminate}
        value={allSelected}
        disabled={allDisabled}
      />
      <Checkbox.Group
        className={style.group}
        name={name}
        value={rightsValues}
        onChange={handleChange}
        onBlur={onBlur}
        disabled={allDisabled}
      >
        {cbs}
      </Checkbox.Group>
    </div>
  )
}

export default RightsGroup
