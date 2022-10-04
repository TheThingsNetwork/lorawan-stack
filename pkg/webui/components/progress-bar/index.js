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

import React, { PureComponent } from 'react'
import classnames from 'classnames'
import { defineMessages } from 'react-intl'

import Message from '@ttn-lw/lib/components/message'
import RelativeDateTime from '@ttn-lw/lib/components/date-time/relative'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './progress-bar.styl'

const m = defineMessages({
  estimatedCompletion: 'Estimated completion {eta}',
  progress: '{current} of {target}',
  percentage: '{percentage}% finished',
})

export default class ProgressBar extends PureComponent {
  static propTypes = {
    children: PropTypes.node,
    /* The class to be attached to the outer container. */
    className: PropTypes.string,
    /* The current progress value, used in conjunction with the `target` value. */
    current: PropTypes.number,
    /* Current percentage. */
    percentage: PropTypes.number,
    /* Decimals to be shown for the percentage value. */
    percentageDecimals: PropTypes.number,
    /* Flag indicating whether an ETA estimation is shown. */
    showEstimation: PropTypes.bool,
    /* Flag indicating whether a status text is shown (percentage value). */
    showStatus: PropTypes.bool,
    /* The target value, used in conjunction with the `current` value. */
    target: PropTypes.number,
  }

  static defaultProps = {
    children: undefined,
    className: undefined,
    current: 0,
    percentage: undefined,
    percentageDecimals: 2,
    showEstimation: true,
    showStatus: false,
    target: 1,
  }

  state = {
    estimatedDuration: Infinity,
    startTime: undefined,
    elapsedTime: undefined,
    estimations: 0,
  }

  static getDerivedStateFromProps(props, state) {
    const { current, target, showEstimation } = props
    const { percentage = (current / target) * 100 } = props
    let { estimatedDuration, startTime, elapsedTime, estimations } = state

    if (!showEstimation) {
      return { estimatedDuration, startTime, elapsedTime, estimations }
    }

    if (percentage === 0) {
      startTime = Date.now()
      return { estimatedDuration: Infinity, startTime, elapsedTime, estimations: 0 }
    }

    elapsedTime = Date.now() - startTime
    estimatedDuration = Math.max(0, elapsedTime * (100 / percentage))
    estimations++

    return { estimatedDuration, startTime, elapsedTime, estimations }
  }

  render() {
    const { current, target, showStatus, percentageDecimals, showEstimation, className, children } =
      this.props
    const { percentage = (current / target) * 100 } = this.props
    const { estimatedDuration, startTime, estimations } = this.state
    const displayPercentage = (Math.max(0, Math.min(100, percentage)) || 0).toFixed(
      percentageDecimals,
    )
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

    return (
      <div className={classnames(className, style.container)} data-test-id="progress-bar">
        <div className={style.bar}>
          <div style={{ width: `${displayPercentage}%` }} className={style.filler} />
        </div>
        {showStatus && (
          <div className={style.status}>
            {this.props.percentage === undefined && (
              <div>
                <Message content={m.progress} values={{ current, target }} /> (
                <Message content={m.percentage} values={{ percentage: displayPercentage }} />)
              </div>
            )}
            {children}
            {this.props.percentage !== undefined && (
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
}
