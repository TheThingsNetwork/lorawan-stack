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

import React, { useCallback } from 'react'
import classNames from 'classnames'

import Link from '@ttn-lw/components/link'
import Icon from '@ttn-lw/components/icon'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './search-panel.styl'

const PanelItem = ({ icon, title, path, subtitle, isFocused, index, onClick, onMouseEnter }) => {
  const handleHover = useCallback(() => {
    onMouseEnter(index)
  }, [index, onMouseEnter])
  return (
    <Link
      id={`search-item-${index}`}
      className={classNames(style.resultItem, { [style.resultItemFocus]: isFocused })}
      to={path}
      data-index={index}
      onClick={onClick}
      onMouseEnter={handleHover}
    >
      <Icon className="c-icon-neutral-normal" icon={icon} />
      <div className="d-flex flex-column gap-cs-xxs">
        <div className="c-text-neutral-heavy fw-bold">{title}</div>
        <div className="c-text-neutral-light">{subtitle}</div>
      </div>
    </Link>
  )
}

PanelItem.propTypes = {
  icon: PropTypes.icon.isRequired,
  index: PropTypes.number.isRequired,
  isFocused: PropTypes.bool.isRequired,
  onClick: PropTypes.func.isRequired,
  onMouseEnter: PropTypes.func.isRequired,
  path: PropTypes.string.isRequired,
  subtitle: PropTypes.node,
  title: PropTypes.node.isRequired,
}

PanelItem.defaultProps = {
  subtitle: undefined,
}

PanelItem.defaultProps = {
  subtitle: undefined,
}

export default PanelItem
