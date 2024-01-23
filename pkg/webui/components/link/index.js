// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { Link as RouterLink } from 'react-router-dom'
import classnames from 'classnames'
import { defineMessages, useIntl } from 'react-intl'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import { url as urlPattern } from '@ttn-lw/lib/regexp'
import { selectDocumentationUrlConfig, selectApplicationRootPath } from '@ttn-lw/lib/selectors/env'
import filterDataProps from '@ttn-lw/lib/filter-data-props'

import style from './link.styl'

const m = defineMessages({
  glossaryTitle: 'See "{term}" in the glossary',
  defaultGlossaryTitle: 'See in the glossary',
})

const appRoot = selectApplicationRootPath()
const docBaseUrl = selectDocumentationUrlConfig()

const formatTitle = (content, values, formatter) => {
  if (typeof content === 'object' && content.id && content.defaultMessage) {
    return formatter(content, values)
  }

  return content
}

const Link = props => {
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
    onClick,
    secondary,
    primary,
    tabIndex,
    role,
    state,
  } = props

  const { formatMessage } = useIntl()
  const formattedTitle = formatTitle(title, titleValues, formatMessage)
  const classNames = classnames(style.link, className, {
    [style.linkVisited]: showVisited,
    [style.primary]: primary,
    [style.secondary]: secondary,
  })

  if (disabled) {
    return (
      <span className={classnames(classNames, style.disabled)} role={role}>
        {children}
      </span>
    )
  }

  return (
    <RouterLink
      className={classNames}
      id={id}
      title={formattedTitle}
      replace={replace}
      to={to}
      target={target}
      onClick={onClick}
      tabIndex={tabIndex}
      role={role}
      state={state}
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
  onClick: PropTypes.func,
  primary: PropTypes.bool,
  replace: PropTypes.bool,
  role: PropTypes.string,
  secondary: PropTypes.bool,
  showVisited: PropTypes.bool,
  state: PropTypes.shape({}),
  tabIndex: PropTypes.string,
  target: PropTypes.string,
  title: PropTypes.message,
  titleValues: PropTypes.shape({}),
  to: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.shape({
      pathname: PropTypes.string,
      search: PropTypes.string,
      hash: PropTypes.string,
      state: PropTypes.shape({}),
    }),
  ]).isRequired,
}

Link.defaultProps = {
  children: undefined,
  className: undefined,
  disabled: false,
  id: undefined,
  onClick: () => null,
  primary: false,
  showVisited: false,
  replace: false,
  role: undefined,
  secondary: false,
  tabIndex: undefined,
  target: undefined,
  title: undefined,
  titleValues: undefined,
  state: undefined,
}

const DocLink = props => {
  const {
    className,
    name,
    title,
    titleValues,
    id,
    path,
    children,
    showVisited,
    secondary,
    primary,
    disabled,
    tabIndex,
    raw,
    onClick,
  } = props

  const { formatMessage } = useIntl()
  const classNames = classnames(style.link, style.docLink, className, {
    [style.linkVisited]: showVisited,
    [style.primary]: primary,
    [style.secondary]: secondary,
  })
  if (disabled) {
    return <span className={classnames(classNames, style.disabled)}>{children}</span>
  }
  const link = docBaseUrl.concat(path)
  const formattedTitle = formatTitle(title, titleValues, formatMessage)

  return (
    <a
      className={classNames}
      title={formattedTitle}
      id={id}
      href={link}
      target="_blank"
      name={name}
      onClick={onClick}
      tabIndex={tabIndex}
    >
      {!raw && <Icon className={style.docIcon} icon="book" />}
      {children}
      {!raw && <Icon className={style.icon} icon="launch" />}
    </a>
  )
}

DocLink.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  id: PropTypes.string,
  name: PropTypes.string,
  onClick: PropTypes.func,
  path: PropTypes.string.isRequired,
  primary: PropTypes.bool,
  raw: PropTypes.bool,
  secondary: PropTypes.bool,
  showVisited: PropTypes.bool,
  tabIndex: PropTypes.number,
  title: PropTypes.message,
  titleValues: PropTypes.shape({}),
}

DocLink.defaultProps = {
  children: undefined,
  className: undefined,
  disabled: false,
  id: undefined,
  primary: false,
  showVisited: false,
  secondary: false,
  tabIndex: undefined,
  title: undefined,
  titleValues: undefined,
  name: undefined,
  raw: false,
  onClick: () => null,
}

Link.DocLink = DocLink

const GlossaryLink = ({ title, glossaryId, term, primary, secondary, className }) => {
  const { formatMessage } = useIntl()
  const hasTerm = Boolean(term)

  return (
    <Link.DocLink
      primary={primary}
      secondary={secondary}
      className={className}
      path={`/reference/glossary#${glossaryId}`}
      title={hasTerm ? m.glossaryTitle : m.defaultGlossaryTitle}
      titleValues={hasTerm ? { term: formatTitle(term, undefined, formatMessage) } : undefined}
      tabIndex={-1}
      external
    >
      <Message content={title} />
    </Link.DocLink>
  )
}

GlossaryLink.propTypes = {
  className: PropTypes.string,
  glossaryId: PropTypes.string.isRequired,
  primary: PropTypes.bool,
  secondary: PropTypes.bool,
  term: PropTypes.message,
  title: PropTypes.message.isRequired,
}

GlossaryLink.defaultProps = {
  className: '',
  primary: false,
  secondary: false,
  term: undefined,
}

Link.GlossaryLink = GlossaryLink

const AnchorLink = props => {
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
    secondary,
    primary,
    disabled,
    external,
    tabIndex,
    ...rest
  } = props

  const { formatMessage } = useIntl()
  const formattedTitle = formatTitle(title, titleValues, formatMessage)
  const classNames = classnames(style.link, className, {
    [style.linkVisited]: showVisited,
    [style.primary]: primary,
    [style.secondary]: secondary,
  })
  const dataProps = useMemo(() => filterDataProps(rest), [rest])

  if (disabled) {
    return <span className={classnames(classNames, style.disabled)}>{children}</span>
  }

  return (
    <a
      className={classNames}
      title={formattedTitle}
      id={id}
      href={href}
      target={external ? '_blank' : target}
      name={name}
      tabIndex={tabIndex}
      {...dataProps}
    >
      {children}
      {external ? <Icon className={style.icon} icon="launch" /> : null}
    </a>
  )
}

AnchorLink.propTypes = {
  href: PropTypes.string.isRequired,
  id: PropTypes.string,
  name: PropTypes.string,
  showVisited: PropTypes.bool,
  tabIndex: PropTypes.string,
  target: PropTypes.string,
  title: PropTypes.string,
}

AnchorLink.defaultProps = {
  showVisited: false,
}

Link.Anchor = AnchorLink

const BaseAnchorLink = ({ href, ...rest }) => {
  // Prevent prefixing proper URLs.
  const path = href.match(urlPattern) ? href : appRoot + href

  return <Link.Anchor href={path} {...rest} />
}

BaseAnchorLink.propTypes = AnchorLink.propTypes
BaseAnchorLink.defaultProps = AnchorLink.defaultProps

Link.BaseAnchor = BaseAnchorLink

export default Link
