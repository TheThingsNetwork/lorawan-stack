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

import React, { useMemo, useState } from 'react'
import classnames from 'classnames'
import { defineMessages } from 'react-intl'

import Message from '@ttn-lw/lib/components/message'
import RelativeDateTime from '@ttn-lw/lib/components/date-time/relative'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './progress-bar.styl'

const m = defineMessages({
  estimatedCompletion: 'Estimated completion {eta}',
  progress: '{current, number} of {target, number}',
  percentage: '{percentage, number, percent} finished',
})

const ProgressBar = props => {
  const {
    barClassName,
    children,
    className,
    current,
    headerTargetMessage,
    itemName,
    showEstimation,
    showHeader,
    showStatus,
    target,
    warn,
  } = props
  const { percentage = (current / target) * 100 } = props
  const startTime = useMemo(() => (percentage === 0 ? Date.now() : null), [percentage])
  const [estimations, setEstimations] = useState(0)
  const estimatedDuration = useMemo(() => {
    const newElapsedTime = Date.now() - startTime
    const newEstimatedDuration = Math.max(0, newElapsedTime * (100 / percentage))
    setEstimations(estimations => estimations + 1)
    return newEstimatedDuration
  }, [percentage, startTime])

  const fraction = Math.max(0, Math.min(1, percentage / 100))
  const displayPercentage = (fraction || 0) * 100
  let displayEstimation = null

  if (showEstimation && percentage < 100) {
    const now = Date.now()
    let eta = new Date(startTime + estimatedDuration)
    if (eta <= now) {
      // Avoid estimations in the past.
      eta = new Date(now + 1000)
    }
    displayEstimation =
      !showEstimation ||
      estimations < 3 || // Avoid inaccurate early estimations.
      estimatedDuration === Infinity ||
      !startTime ? null : (
        <div>
          <span>
            Estimated completion <RelativeDateTime value={eta} />
          </span>
        </div>
      )
  }

  const fillerCls = classnames(style.filler, {
    [style.warn]: warn >= 80,
    [style.limit]: warn >= 100,
  })

  return (
    <div className={classnames(className, style.container)} data-test-id="progress-bar">
      {showHeader && (
        <div className={style.progressBarValues}>
          <p className="m-0">
            <b>
              {current} {itemName}
            </b>
          </p>
          {headerTargetMessage}
        </div>
      )}
      <div className={classnames(style.bar, barClassName)}>
        <div style={{ width: `${displayPercentage}%` }} className={fillerCls} />
      </div>
      {showStatus && (
        <div className={style.status}>
          {props.percentage === undefined && !showHeader && (
            <div>
              <Message content={m.progress} values={{ current, target }} /> (
              <Message content={m.percentage} values={{ percentage: fraction }} />)
            </div>
          )}
          {children}
          {props.percentage !== undefined && (
            <Message
              component="div"
              content={m.percentage}
              values={{ percentage: displayPercentage }}
            />
          )}
          {displayEstimation}
        </div>
      )}
    </div>
  )
}

ProgressBar.propTypes = {
  /* The class to be attached to the bar. */
  barClassName: PropTypes.string,
  children: PropTypes.node,
  /* The class to be attached to the outer container. */
  className: PropTypes.string,
  /* The current progress value, used in conjunction with the `target` value. */
  current: PropTypes.number,
  headerTargetMessage: PropTypes.message,
  itemName: PropTypes.message,
  /* Current percentage. */
  percentage: PropTypes.number,
  /* Flag indicating whether an ETA estimation is shown. */
  showEstimation: PropTypes.bool,
  /* Flag indicating whether a header with current and target is shown is shown. */
  showHeader: PropTypes.bool,
  /* Flag indicating whether a status text is shown (percentage value). */
  showStatus: PropTypes.bool,
  /* The target value, used in conjunction with the `current` value. */
  target: PropTypes.number,
  warn: PropTypes.number,
}

ProgressBar.defaultProps = {
  barClassName: undefined,
  children: undefined,
  className: undefined,
  current: 0,
  percentage: undefined,
  showEstimation: true,
  showStatus: false,
  showHeader: false,
  target: 1,
  headerTargetMessage: undefined,
  itemName: undefined,
  warn: undefined,
}

export default ProgressBar
