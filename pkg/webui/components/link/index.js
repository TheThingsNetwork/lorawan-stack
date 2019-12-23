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
import { Link as RouterLink } from 'react-router-dom'
import classnames from 'classnames'
import { injectIntl } from 'react-intl'

import PropTypes from '../../lib/prop-types'
import { withEnv } from '../../lib/components/env'
import { url as urlPattern } from '../../lib/regexp'

import style from './link.styl'

const formatTitle = function(content, values, formatter) {
  if (typeof content === 'object' && content.id && content.defaultMessage) {
    return formatter(content, values)
  }

  return content
}

const Link = function(props) {
  const {
    className,
    disabled,
    title,
    titleValues,
    id,
    children,
    to,
    replace,
    target,
    showVisited,
    intl,
  } = props

  const formattedTitle = formatTitle(title, titleValues, intl.formatMessage)

  return (
    <RouterLink
      className={
        className
          ? classnames(className, { [style.disabled]: disabled })
          : classnames(style.link, { [style.linkVisited]: showVisited, [style.disabled]: disabled })
      }
      id={id}
      title={formattedTitle}
      replace={replace}
      to={to}
      target={target}
    >
      {children}
    </RouterLink>
  )
}

Link.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  id: PropTypes.string,
  intl: PropTypes.shape({
    formatMessage: PropTypes.func,
  }).isRequired,
  replace: PropTypes.bool,
  showVisited: PropTypes.bool,
  target: PropTypes.string,
  title: PropTypes.message,
  titleValues: PropTypes.shape({}),
  to: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.shape({
      pathname: PropTypes.string,
      search: PropTypes.string,
      hash: PropTypes.string,
      state: PropTypes.object,
    }),
  ]).isRequired,
}

Link.defaultProps = {
  children: undefined,
  className: undefined,
  disabled: false,
  id: undefined,
  showVisited: false,
  replace: false,
  target: undefined,
  title: undefined,
  titleValues: undefined,
}

const AnchorLink = function(props) {
  const {
    className,
    name,
    title,
    titleValues,
    id,
    href,
    target,
    children,
    showVisited,
    intl,
  } = props

  const formattedTitle = formatTitle(title, titleValues, intl.formatMessage)

  return (
    <a
      className={
        className ? className : classnames(style.link, { [style.linkVisited]: showVisited })
      }
      title={formattedTitle}
      id={id}
      href={href}
      target={target}
      name={name}
    >
      {children}
    </a>
  )
}

AnchorLink.propTypes = {
  href: PropTypes.string.isRequired,
  id: PropTypes.string,
  name: PropTypes.string,
  showVisited: PropTypes.bool,
  target: PropTypes.string,
  title: PropTypes.string,
}

AnchorLink.defaultProps = {
  showVisited: false,
}

Link.Anchor = injectIntl(AnchorLink)

const BaseAnchorLink = function({ env, href, ...rest }) {
  const { appRoot } = env

  // Prevent prefixing proper URLs
  const path = href.match(urlPattern) ? href : appRoot + href

  return <Link.Anchor href={path} {...rest} />
}

BaseAnchorLink.propTypes = AnchorLink.propTypes
BaseAnchorLink.defaultProps = AnchorLink.defaultProps

Link.BaseAnchor = withEnv(BaseAnchorLink)

export default injectIntl(Link)
