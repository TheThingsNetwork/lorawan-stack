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

import React, { useCallback, forwardRef, useMemo, useRef } from 'react'
import classnames from 'classnames'
import { useIntl } from 'react-intl'
import { isPlainObject } from 'lodash'

import Link from '@ttn-lw/components/link'
import Spinner from '@ttn-lw/components/spinner'
import Icon, { IconChevronDown } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import combineRefs from '@ttn-lw/lib/combine-refs'
import PropTypes from '@ttn-lw/lib/prop-types'
import filterDataProps from '@ttn-lw/lib/filter-data-props'

import Dropdown from '../dropdown'
import Tooltip from '../tooltip'

import style from './button.styl'

const assembleClassnames = ({
  message,
  danger,
  warning,
  primary,
  secondary,
  tertiary,
  naked,
  icon,
  small,
  busy,
  dropdownItems,
  className,
  error,
  withAlert,
  noDropdownIcon,
}) =>
  classnames(style.button, {
    [className]: !Boolean(dropdownItems), // If there are dropdown items, the button is wrapped in a div with the className.
    [style.danger]: danger,
    [style.warning]: warning,
    [style.primary]: primary,
    [style.secondary]: secondary,
    [style.tertiary]: tertiary,
    [style.naked]: naked,
    [style.busy]: busy,
    [style.small]: small,
    [style.withIcon]: icon !== undefined && message,
    [style.noDropdownIcon]: noDropdownIcon,
    [style.onlyIcon]: icon !== undefined && !message,
    [style.withDropdown]: Boolean(dropdownItems),
    [style.error]: error && !busy,
    [style.withAlert]: withAlert,
  })

const buttonChildren = props => {
  const {
    dropdownItems,
    icon,
    busy,
    message,
    messageValues,
    noDropdownIcon,
    children,
    small,
    large,
  } = props

  const content = (
    <>
      {icon && <Icon className={style.icon} icon={icon} size={small ? 16 : large ? 24 : 18} />}
      {message && (
        <Message content={message} values={messageValues} className={style.linkButtonMessage} />
      )}
      {children}
      {dropdownItems && (
        <>
          {!noDropdownIcon && (
            <Icon
              className={style.expandIcon}
              icon={IconChevronDown}
              size={small ? 12 : large ? 22 : 18}
            />
          )}
        </>
      )}
    </>
  )

  return (
    <>
      {content}
      {busy && <Spinner className={style.spinner} small after={200} />}
    </>
  )
}

const Button = forwardRef((props, ref) => {
  const {
    autoFocus,
    disabled,
    dropdownItems,
    dropdownClassName,
    dropdownPosition,
    name,
    type,
    value,
    title: rawTitle,
    busy,
    onBlur,
    onClick,
    form,
    className,
    portalledDropdown,
    tooltip,
    tooltipPlacement,
    onMouseEnter,
    onMouseLeave,
    ...rest
  } = props
  const innerRef = useRef()
  const combinedRef = combineRefs([ref, innerRef])

  const dataProps = useMemo(() => filterDataProps(rest), [rest])

  const handleClick = useCallback(
    evt => {
      if (busy || disabled) {
        return
      }

      // Passing a value to the onClick handler is useful for components that
      // are rendered multiple times, e.g. in a list. The value can be used to
      // identify the component that was clicked.
      onClick(evt, value)
    },
    [busy, disabled, onClick, value],
  )

  const intl = useIntl()

  let title = rawTitle
  if (typeof rawTitle === 'object' && rawTitle.id && rawTitle.defaultMessage) {
    title = intl.formatMessage(title)
  }

  const htmlProps = { autoFocus, name, type, value, title, onBlur, form, ...dataProps }
  const buttonClassNames = assembleClassnames(props)

  const buttonElement = (
    <button
      className={buttonClassNames}
      onClick={handleClick}
      children={buttonChildren({ ...props })}
      disabled={busy || disabled}
      ref={combinedRef}
      onMouseEnter={onMouseEnter}
      onMouseLeave={onMouseLeave}
      {...htmlProps}
    />
  )

  const wrappedButtonElement = tooltip ? (
    <Tooltip
      content={isPlainObject(tooltip) ? <Message content={tooltip} /> : tooltip}
      placement={tooltipPlacement}
      delay={0}
      noOffset
      small
    >
      {buttonElement}
    </Tooltip>
  ) : (
    buttonElement
  )

  if (dropdownItems) {
    return (
      <div className={classnames(className, 'pos-relative')}>
        {wrappedButtonElement}
        <Dropdown.Attached
          className={dropdownClassName}
          attachedRef={innerRef}
          position={dropdownPosition}
          portalled={portalledDropdown}
        >
          {dropdownItems}
        </Dropdown.Attached>
      </div>
    )
  }

  return wrappedButtonElement
})

Button.defaultProps = {
  onClick: () => null,
  onMouseEnter: () => null,
  onBlur: undefined,
}

const LinkButton = props => {
  const { disabled, titleMessage, onClick, value, tooltip, tooltipPlacement, dataTestId, ...rest } =
    props
  const buttonClassNames = assembleClassnames(props)
  const dataProps = useMemo(() => filterDataProps(rest), [rest])
  const { to } = props

  const handleClick = useCallback(
    evt => {
      // Passing a value to the onClick handler is useful for components that
      // are rendered multiple times, e.g. in a list. The value can be used to
      // identify the component that was clicked.
      onClick(evt, value)
    },
    [onClick, value],
  )

  const buttonElement = (
    <Link
      className={buttonClassNames}
      to={to}
      disabled={disabled}
      title={titleMessage}
      children={buttonChildren(props)}
      onClick={handleClick}
      dataTestId={dataTestId}
      {...dataProps}
    />
  )

  const wrappedButtonElement = tooltip ? (
    <Tooltip
      content={isPlainObject(tooltip) ? <Message content={tooltip} /> : tooltip}
      placement={tooltipPlacement}
      delay={0}
      noOffset
      small
    >
      {buttonElement}
    </Tooltip>
  ) : (
    buttonElement
  )

  return wrappedButtonElement
}

const AnchorLinkButton = props => {
  const { target, title, name, href, external, ...rest } = props
  const dataProps = useMemo(() => filterDataProps(rest), [rest])
  const htmlProps = { target, title, name, ...dataProps }
  const buttonClassNames = assembleClassnames(props)
  return (
    <Link.Anchor
      className={buttonClassNames}
      href={href}
      children={buttonChildren(props)}
      external={external}
      {...htmlProps}
    />
  )
}

const commonPropTypes = {
  /** The message to be displayed within the button. */
  message: PropTypes.message,
  /** The message values. */
  messageValues: PropTypes.object,
  /**
   * A flag specifying whether the `danger` styling should applied to the
   * button.
   */
  danger: PropTypes.bool,
  /**
   * A flag specifying whether the `warning` styling should applied to the
   * button.
   */
  warning: PropTypes.bool,
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
  /**
   * A flag specifying whether the `error` styling should applied to the button.
   */
  error: PropTypes.bool,
  /** The name of an icon to be displayed within the button. */
  icon: PropTypes.icon,
  /**
   * A flag specifying whether the button in the `busy` state and the
   * appropriate styling should be applied.
   */
  busy: PropTypes.bool,
  /**
   * A flag specifying whether the button in the `disabled` state and the
   * appropriate styling should be applied. Also passes the `disabled` html prop
   * to the button element.
   */
  disabled: PropTypes.bool,
  /** The html `name` prop passed to the <button /> element. */
  name: PropTypes.string,
  /** The html `type` prop passed to the <button /> element. */
  type: PropTypes.string,
  /** The html `value` prop passed to the <button /> element. */
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
  /** The html `autofocus` prop passed to the <button /> element. */
  autoFocus: PropTypes.bool,
  /** A message to be evaluated and passed to the <button /> element. */
  title: PropTypes.message,
  /** The tooltip message to be displayed when hovering over the button. */
  tooltip: PropTypes.oneOfType([PropTypes.message, PropTypes.node]),
  /** Dropdown items of the button. */
  dropdownItems: PropTypes.node,
  /** A flag specifying whether the small styling should applied to the button. */
  small: PropTypes.bool,
  /** A flag specifying whether the large styling should applied to the button. */
  large: PropTypes.bool,
}

buttonChildren.propTypes = {
  /**
   * Possible children components of the button:
   * Spinner, Icon, and/or Message.
   */
  busy: commonPropTypes.busy,
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  icon: commonPropTypes.icon,
  message: commonPropTypes.message,
  messageValues: commonPropTypes.messageValues,
}

buttonChildren.defaultProps = {
  busy: undefined,
  icon: undefined,
  message: undefined,
  messageValues: undefined,
  children: null,
  small: false,
  large: false,
}

Button.propTypes = {
  onBlur: PropTypes.func,
  /**
   * A click listener to be called when the button is pressed.
   * Not called if the button is in the `busy` or `disabled` state.
   */
  onClick: PropTypes.func,
  onMouseEnter: PropTypes.func,
  portalledDropdown: PropTypes.bool,
  ...commonPropTypes,
}

Button.defaultProps = {
  onClick: () => null,
  onMouseEnter: () => null,
  portalledDropdown: false,
}

LinkButton.propTypes = {
  onClick: PropTypes.func,
  ...commonPropTypes,
  ...Link.propTypes,
}

LinkButton.defaultProps = {
  onClick: () => null,
}

Button.Link = LinkButton
Button.Link.displayName = 'Button.Link'

AnchorLinkButton.propTypes = {
  ...commonPropTypes,
  ...Link.Anchor.propTypes,
}

Button.AnchorLink = AnchorLinkButton
Button.AnchorLink.displayName = 'Button.AnchorLink'

export default Button
