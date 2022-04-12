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

import React, { useEffect, useState, useCallback, useRef } from 'react'
import classnames from 'classnames'
import { defineMessages, useIntl } from 'react-intl'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './status.styl'

const m = defineMessages({
  good: 'good',
  bad: 'bad',
  mediocre: 'mediocre',
  unknown: 'unknown',
})

const Status = React.forwardRef(
  (
    { className, status, label, pulse, pulseTrigger, labelValues, children, title, flipped },
    ref,
  ) => {
    const intl = useIntl()
    const [animate, setAnimate] = useState(false)
    const pulseArmed = useRef(false)
    useEffect(() => {
      if (pulseArmed.current) {
        setAnimate(true)
      } else {
        pulseArmed.current = true
      }
    }, [pulseTrigger])

    const handleAnimationEnd = useCallback(() => {
      setAnimate(false)
    }, [setAnimate])

    const cls = classnames(style.status, {
      [style.statusGood]: status === 'good',
      [style.statusBad]: status === 'bad',
      [style.statusMediocre]: status === 'mediocre',
      [style.statusUnknown]: status === 'unknown',
      [style[`${status}-pulse`]]: typeof pulse === 'boolean' ? pulse : status === 'good',
      [style.flipped]: flipped,
      [style[`triggered-${status}-pulse`]]: animate,
    })

    let statusLabel = null
    if (React.isValidElement(label)) {
      statusLabel = React.cloneElement(label, {
        ...label.props,
        className: classnames(label.props.className, style.statusLabel, {
          [style.flipped]: flipped,
        }),
      })
    } else {
      statusLabel = label && (
        <Message className={style.statusLabel} content={label} values={labelValues} />
      )
    }

    let translatedTitle

    if (title) {
      translatedTitle = typeof title === 'string' ? title : intl.formatMessage(title)
    } else if (label) {
      translatedTitle = typeof label === 'string' ? label : intl.formatMessage(label)
    } else {
      translatedTitle = intl.formatMessage(m[status])
    }

    return (
      <span
        className={classnames(className, style.container)}
        onAnimationEnd={handleAnimationEnd}
        ref={ref}
      >
        {flipped && <span className={classnames(cls)} title={translatedTitle} />}
        {statusLabel}
        {children}
        {!flipped && <span className={classnames(cls)} title={translatedTitle} />}
      </span>
    )
  },
)

Status.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  flipped: PropTypes.bool,
  label: PropTypes.message,
  labelValues: PropTypes.shape({}),
  pulse: PropTypes.bool,
  pulseTrigger: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.number,
    PropTypes.instanceOf(Date),
  ]),
  status: PropTypes.oneOf(['good', 'bad', 'mediocre', 'unknown']),
  title: PropTypes.message,
}

Status.defaultProps = {
  children: undefined,
  className: undefined,
  flipped: false,
  label: undefined,
  labelValues: undefined,
  pulse: undefined,
  pulseTrigger: undefined,
  status: 'unknown',
  title: undefined,
}

export default Status
