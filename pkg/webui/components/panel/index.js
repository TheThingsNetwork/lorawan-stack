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

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import Icon from '../icon'
import Button from '../button'

import Toggle from './toggle'

import styles from './panel.styl'

const Panel = ({
  children,
  title,
  icon,
  toggleOptions,
  activeToggle,
  onToggleClick,
  buttonTitle,
}) => (
  <div className={styles.panel}>
    <div className="d-flex j-between mb-cs-xl">
      <div className="d-flex gap-cs-xs al-center">
        {icon && <Icon icon={icon} className={styles.icon} />}
        <Message content={title} className={styles.title} />
      </div>
      {toggleOptions ? (
        <Toggle options={toggleOptions} active={activeToggle} onToggleChange={onToggleClick} />
      ) : (
        <Button message={buttonTitle} unstyled className={styles.button} />
      )}
    </div>
    {children}
  </div>
)

Panel.propTypes = {
  activeToggle: PropTypes.string,
  buttonTitle: PropTypes.string,
  children: PropTypes.node.isRequired,
  icon: PropTypes.string,
  onToggleClick: PropTypes.func,
  title: PropTypes.message.isRequired,
  toggleOptions: PropTypes.arrayOf(PropTypes.shape({})),
}

Panel.defaultProps = {
  buttonTitle: undefined,
  icon: undefined,
  toggleOptions: undefined,
  activeToggle: undefined,
  onToggleClick: () => null,
}

export default Panel
