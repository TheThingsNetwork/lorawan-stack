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

import NavigationLink from '../link'
import Message from '../../../lib/components/message'
import Icon from '../../icon'
import PropTypes from '../../../lib/prop-types'

import style from './bar.styl'

const NavigationBar = function({ className, children }) {
  return <nav className={className}>{children}</nav>
}

NavigationBar.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
}

NavigationBar.defaultProps = {
  className: undefined,
}

const NavigationBarItem = ({ icon, title, className, ...rest }) => (
  <NavigationLink
    {...rest}
    className={classnames(style.link, className)}
    activeClassName={style.linkActive}
  >
    {icon && <Icon icon={icon} className={style.icon} />}
    <Message content={title} />
  </NavigationLink>
)

NavigationBarItem.propTypes = {
  className: PropTypes.string,
  icon: PropTypes.string.isRequired,
  title: PropTypes.message.isRequired,
}

NavigationBarItem.defaultProps = {
  className: undefined,
}

NavigationBar.Item = NavigationBarItem

export default NavigationBar
