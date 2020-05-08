// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import Icon from '@ttn-lw/components/icon'

import PropTypes from '@ttn-lw/lib/prop-types'

import eventContext from '../context'

import style from './overview.styl'

const EventOverview = props => {
  const { className, children } = props
  const { isOpen, onDetailsOpen, expandable, widget } = React.useContext(eventContext)

  const canExpand = expandable && !widget

  const expandableProps = React.useMemo(() => {
    const res = {}
    if (canExpand) {
      res.role = 'button'
      res.onClick = onDetailsOpen
    }

    return res
  }, [canExpand, onDetailsOpen])

  const containerCls = classnames(className, {
    [style.container]: !widget,
    [style.overviewExpandable]: canExpand,
  })
  const overviewCls = classnames(className, style.overview, {
    [style.overviewWidget]: widget,
  })

  return (
    <div className={containerCls} {...expandableProps}>
      <div className={overviewCls}>
        {children}
        {canExpand && (
          <Icon className={style.icon} icon={isOpen ? 'arrow_drop_down' : 'arrow_drop_up'} />
        )}
      </div>
    </div>
  )
}

EventOverview.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
  className: PropTypes.string,
}

EventOverview.defaultProps = {
  className: undefined,
}

export default EventOverview
