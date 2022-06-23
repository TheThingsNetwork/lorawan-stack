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

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import from from '@ttn-lw/lib/from'

import { useFormContext } from '..'

import Tooltip from './tooltip'

import style from './field.styl'

const InfoField = props => {
  const { children, className, title, disabled: fieldDisabled, tooltipId } = props
  const { disabled: formDisabled } = useFormContext()
  const disabled = formDisabled || fieldDisabled
  const cls = classnames(className, style.field, from(style, { disabled }))

  return (
    <div className={cls}>
      {title && (
        <div className={style.label}>
          <Message content={title} className={style.title} />
          {tooltipId && <Tooltip id={tooltipId} glossaryTerm={title} />}
        </div>
      )}
      <div className={classnames(style.componentArea, style.infoArea)}>{children}</div>
    </div>
  )
}

InfoField.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  title: PropTypes.message,
  tooltipId: PropTypes.string,
}

InfoField.defaultProps = {
  className: undefined,
  title: undefined,
  disabled: false,
  tooltipId: undefined,
}

export default InfoField
