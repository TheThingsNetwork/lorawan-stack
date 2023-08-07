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

import PropTypes from '@ttn-lw/lib/prop-types'
import useDerivedRightProps from '@ttn-lw/lib/hooks/use-derived-rights-props'

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

const RightsGroup = ({
  className,
  disabled,
  entityTypeMessage,
  onBlur,
  onChange,
  pseudoRight,
  rights,
  value,
}) => {
  const {
    outOfOwnScopeIndividualRights,
    hasOutOfOwnScopePseudoRight,
    derivedPseudoRight,
    derivedRights,
    hasPseudoRightGranted,
    grantType,
  } = useDerivedRightProps({ value, pseudoRight, rights })
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

RightsGroup.propTypes = {
  /** The class to be added to the container. */
  className: PropTypes.string,
  /** A flag indicating whether the whole component should be disabled. */
  disabled: PropTypes.bool,
  /**
   * The message depicting the type of entity this component is setting the
   * rights for.
   */
  entityTypeMessage: PropTypes.message.isRequired,
  /** The Blur event hook. */
  onBlur: PropTypes.func,
  /** The Change event hook. */
  onChange: PropTypes.func,
  /** The pseudo right literal comprising all other rights. */
  pseudoRight: PropTypes.oneOfType([PropTypes.string, PropTypes.arrayOf(PropTypes.string)]),
  /** The pseudo right derived from the current entity or user. */
  rights: PropTypes.rights.isRequired,
  /** The rights value. */
  value: PropTypes.rights.isRequired,
}

RightsGroup.defaultProps = {
  className: undefined,
  disabled: false,
  onBlur: () => null,
  onChange: () => null,
  pseudoRight: undefined,
}

export default RightsGroup
