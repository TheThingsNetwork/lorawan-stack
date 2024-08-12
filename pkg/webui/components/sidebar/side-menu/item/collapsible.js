// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useRef } from 'react'
import classnames from 'classnames'

import Icon, { IconChevronDown } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import SideNavigationList from '../list'

import style from './item.styl'

const CollapsibleItem = ({ children, onClick, isExpanded, title, icon, depth }) => {
  const ref = useRef()

  return (
    <div className={style.container} ref={ref}>
      <Button className={style.link} onClick={onClick}>
        {icon && <Icon icon={icon} className={style.icon} />}
        <Message content={title} className={style.title} />
        <Icon
          icon={IconChevronDown}
          size={14}
          className={classnames(style.expandIcon, {
            [style.expandIconOpen]: isExpanded,
          })}
        />
      </Button>
      <SideNavigationList depth={depth + 1} className={style.subItemList} isExpanded={isExpanded}>
        {children}
      </SideNavigationList>
    </div>
  )
}

CollapsibleItem.propTypes = {
  children: PropTypes.node,
  depth: PropTypes.number.isRequired,
  icon: PropTypes.icon,
  isExpanded: PropTypes.bool.isRequired,
  onClick: PropTypes.func.isRequired,
  title: PropTypes.message.isRequired,
}

CollapsibleItem.defaultProps = {
  children: undefined,
  icon: undefined,
}

export default CollapsibleItem
