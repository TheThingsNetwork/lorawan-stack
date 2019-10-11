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
import classnames from 'classnames'

import Message from '../../lib/components/message'
import PropTypes from '../../lib/prop-types'

import style from './status.styl'

const Status = function({ className, status, label, pulse, labelValues, children }) {
  const cls = classnames(style.status, {
    [style.statusGood]: status === 'good',
    [style.statusBad]: status === 'bad',
    [style.statusMediocre]: status === 'mediocre',
    [style.statusUnknown]: status === 'unknown',
    [style[`${status}-pulse`]]: pulse,
  })

  let statusLabel = null
  if (React.isValidElement(label)) {
    statusLabel = React.cloneElement(label, {
      ...label.props,
      className: classnames(label.props.className, style.statusLabel),
    })
  } else {
    statusLabel = <Message className={style.statusLabel} content={label} values={labelValues} />
  }

  return (
    <span className={classnames(className, style.container)}>
      {statusLabel}
      <span className={classnames(cls)} />
      {children}
    </span>
  )
}

Status.propTypes = {
  className: PropTypes.string,
  label: PropTypes.message,
  labelValues: PropTypes.shape({}),
  pulse: PropTypes.bool,
  status: PropTypes.oneOf(['good', 'bad', 'mediocre', 'unknown']),
}

Status.defaultProps = {
  className: undefined,
  pulse: true,
  label: undefined,
  labelValues: undefined,
  status: 'unknown',
}

export default Status
