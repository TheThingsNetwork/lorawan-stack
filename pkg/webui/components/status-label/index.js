// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import React, { useMemo } from 'react'
import classnames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import from from '@ttn-lw/lib/from'

import Icon, {
  IconCircleCheckFilled,
  IconAlertTriangleFilled,
  IconAlertCircleFilled,
  IconInfoCircleFilled,
} from '../icon'

import style from './status-label.styl'

const StatusLabel = ({ className, icon, type, content, onClick, contentValues }) => {
  const statusIcon = useMemo(() => {
    if (icon) {
      return icon
    }
    if (type === 'success') {
      return IconCircleCheckFilled
    }
    if (type === 'warning') {
      return IconAlertTriangleFilled
    }
    if (type === 'error') {
      return IconAlertCircleFilled
    }
    return IconInfoCircleFilled
  }, [icon, type])
  return (
    <div
      className={classnames(
        style.statusLabel,
        className,
        style[type],
        ...from(style, {
          noContent: !content,
        }),
      )}
      onClick={onClick}
    >
      <Icon icon={statusIcon} />

      {!!content && <Message className={style.message} content={content} values={contentValues} />}
    </div>
  )
}

StatusLabel.propTypes = {
  className: PropTypes.string,
  content: PropTypes.message,
  contentValues: PropTypes.object,
  icon: PropTypes.icon,
  onClick: PropTypes.func,
  type: PropTypes.oneOf(['success', 'warning', 'error', 'info']).isRequired,
}

StatusLabel.defaultProps = {
  className: false,
  icon: undefined,
  onClick: () => {},
  content: undefined,
  contentValues: undefined,
}

export default StatusLabel
