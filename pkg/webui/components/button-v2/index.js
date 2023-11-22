// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, forwardRef, useMemo } from 'react'
import classnames from 'classnames'
import { useIntl } from 'react-intl'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './button.styl'

const filterDataProps = props =>
  Object.keys(props)
    .filter(key => key.startsWith('data-'))
    .reduce((acc, key) => {
      acc[key] = props[key]
      return acc
    }, {})

const assembleClassnames = ({ message, primary, naked, icon, className }) =>
  classnames(style.button, className, {
    [style.primary]: primary,
    [style.naked]: naked,
    [style.withIcon]: icon !== undefined && message,
    [style.onlyIcon]: icon !== undefined && !message,
  })

const buttonChildren = props => {
  const { icon, message, children } = props

  const content = Boolean(children) ? (
    children
  ) : (
    <>
      {icon ? <Icon className={style.icon} icon={icon} /> : null}
      {message ? <Message content={message} className={style.linkButtonMessage} /> : null}
    </>
  )

  return <>{content}</>
}

const Button = forwardRef((props, ref) => {
  const {
    autoFocus,
    withDowpdown,
    name,
    type,
    value,
    title: rawTitle,
    onBlur,
    onClick,
    form,
    ...rest
  } = props

  const dataProps = useMemo(() => filterDataProps(rest), [rest])

  const handleClick = useCallback(
    evt => {
      // Passing a value to the onClick handler is useful for components that
      // are rendered multiple times, e.g. in a list. The value can be used to
      // identify the component that was clicked.
      onClick(evt, value)
    },
    [onClick, value],
  )

  const intl = useIntl()

  let title = rawTitle
  if (typeof rawTitle === 'object' && rawTitle.id && rawTitle.defaultMessage) {
    title = intl.formatMessage(title)
  }

  const htmlProps = { autoFocus, name, type, value, title, onBlur, form, ...dataProps }
  const buttonClassNames = assembleClassnames(props)
  return (
    <button
      className={buttonClassNames}
      onClick={handleClick}
      children={buttonChildren(props)}
      ref={ref}
      {...htmlProps}
    />
  )
})

Button.defaultProps = {
  onClick: () => null,
  onBlur: undefined,
}

const commonPropTypes = {
  /** The message to be displayed within the button. */
  message: PropTypes.message,
  /**
   * A flag specifying whether the `primary` styling should applied to the
   * button.
   */
  primary: PropTypes.bool,
  /**
   * A flag specifying whether the `naked` styling should applied to the
   * button.
   */
  naked: PropTypes.bool,
  /** The name of an icon to be displayed within the button. */
  icon: PropTypes.string,
  /** The html `name` prop passed to the <button /> element. */
  name: PropTypes.string,
  /** The html `type` prop passed to the <button /> element. */
  type: PropTypes.string,
  /** A flag specifying whether no additional styles should be
   * attached to the button. This can helpful to achieve individual stylings.
   */
  unstyled: PropTypes.bool,
  /** The html `value` prop passed to the <button /> element. */
  value: PropTypes.string,
  /** The html `autofocus` prop passed to the <button /> element. */
  autoFocus: PropTypes.bool,
  /** A message to be evaluated and passed to the <button /> element. */
  title: PropTypes.message,
}

buttonChildren.propTypes = {
  /**
   * Possible children components of the button:
   * Spinner, Icon, and/or Message.
   */
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  icon: commonPropTypes.icon,
  message: commonPropTypes.message,
}

buttonChildren.defaultProps = {
  icon: undefined,
  message: undefined,
  children: null,
}

Button.propTypes = {
  onBlur: PropTypes.func,
  /**
   * A click listener to be called when the button is pressed.
   * Not called if the button is in the `busy` or `disabled` state.
   */
  onClick: PropTypes.func,
  ...commonPropTypes,
}

export default Button
