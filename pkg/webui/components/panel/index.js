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

import React from 'react'
import classnames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import Icon from '../icon'
import Link from '../link'

import Toggle from './toggle'

import styles from './panel.styl'

const Panel = ({
  children,
  title,
  icon,
  toggleOptions,
  activeToggle,
  onToggleClick,
  shortCutLinkTitle,
  shortCutLinkPath,
  className,
  messageDecorators,
  divider,
  shortCutLinkTarget,
}) => (
  <div className={classnames(styles.panel, className)}>
    <div className="d-flex j-between al-center mb-cs-m flex-wrap gap-cs-m">
      <div className="d-flex gap-cs-xs al-center">
        {icon && <Icon icon={icon} className={styles.panelHeaderIcon} />}
        <Message content={title} className={styles.panelHeaderTitle} />
        {messageDecorators}
      </div>
      {toggleOptions ? (
        <Toggle options={toggleOptions} active={activeToggle} onToggleChange={onToggleClick} />
      ) : (
        shortCutLinkTitle && (
          <Link primary to={shortCutLinkPath} className={styles.button} target={shortCutLinkTarget}>
            <Message content={shortCutLinkTitle} /> →
          </Link>
        )
      )}
    </div>
    {divider && <hr className={styles.panelDivider} />}
    {children}
  </div>
)

Panel.propTypes = {
  activeToggle: PropTypes.string,
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  divider: PropTypes.bool,
  icon: PropTypes.icon,
  messageDecorators: PropTypes.node,
  onToggleClick: PropTypes.func,
  shortCutLinkPath: PropTypes.string,
  shortCutLinkTarget: PropTypes.string,
  shortCutLinkTitle: PropTypes.message,
  title: PropTypes.message.isRequired,
  toggleOptions: PropTypes.arrayOf(PropTypes.shape({})),
}

Panel.defaultProps = {
  icon: undefined,
  toggleOptions: undefined,
  activeToggle: undefined,
  onToggleClick: () => null,
  className: undefined,
  messageDecorators: undefined,
  divider: false,
  shortCutLinkPath: undefined,
  shortCutLinkTitle: undefined,
  shortCutLinkTarget: undefined,
}

export default Panel