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
import bind from 'autobind-decorator'
import { injectIntl } from 'react-intl'

import Link from '@ttn-lw/components/link'
import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import ButtonIcon from './button-icon'

import style from './button.styl'

function assembleClassnames({
  message,
  danger,
  warning,
  secondary,
  naked,
  unstyled,
  icon,
  busy,
  large,
  className,
  error,
  raw,
  disabled,
}) {
  if (unstyled) {
    return className
  }
  return classnames(style.button, className, {
    [style.danger]: danger,
    [style.warning]: warning,
    [style.secondary]: secondary,
    [style.naked]: naked,
    [style.busy]: busy,
    [style.withIcon]: icon !== undefined && message,
    [style.onlyIcon]: icon !== undefined && !message,
    [style.error]: error && !busy,
    [style.large]: large,
    [style.raw]: raw,
    [style.disabled]: disabled,
  })
}

const buttonChildren = props => {
  const { icon, busy, message, children } = props

  const content = Boolean(children) ? (
    children
  ) : (
    <>
      {icon ? <ButtonIcon icon={icon} type="left" /> : null}
      {message ? <Message content={message} /> : null}
    </>
  )

  return (
    <div className={style.content}>
      {busy ? <Spinner className={style.spinner} small after={200} /> : null}
      {content}
    </div>
  )
}

@injectIntl
class Button extends React.PureComponent {
  @bind
  handleClick(evt) {
    const { busy, disabled, onClick } = this.props

    if (busy || disabled) {
      return
    }

    onClick(evt)
  }

  render() {
    const {
      autoFocus,
      disabled,
      name,
      type,
      value,
      title: rawTitle,
      intl,
      busy,
      onBlur,
    } = this.props

    let title = rawTitle
    if (typeof rawTitle === 'object' && rawTitle.id && rawTitle.defaultMessage) {
      title = intl.formatMessage(title)
    }

    const htmlProps = { autoFocus, name, type, value, title, onBlur }
    const buttonClassNames = assembleClassnames(this.props)
    return (
      <button
        className={buttonClassNames}
        onClick={this.handleClick}
        children={buttonChildren(this.props)}
        disabled={busy || disabled}
        {...htmlProps}
      />
    )
  }
}

Button.defaultProps = {
  onClick: () => null,
  onBlur: undefined,
}

Button.Link = function (props) {
  const { disabled, titleMessage } = props
  const buttonClassNames = assembleClassnames(props)
  const { to } = props
  return (
    <Link
      className={buttonClassNames}
      to={to}
      disabled={disabled}
      title={titleMessage}
      children={buttonChildren(props)}
    />
  )
}
Button.Link.displayName = 'Button.Link'

Button.AnchorLink = function (props) {
  const { target, title, name } = props
  const htmlProps = { target, title, name }
  const buttonClassNames = assembleClassnames(props)
  return (
    <Link.Anchor
      className={buttonClassNames}
      href={props.href}
      children={buttonChildren(props)}
      {...htmlProps}
    />
  )
}
Button.AnchorLink.displayName = 'Button.AnchorLink'

Button.Icon = ButtonIcon
Button.Icon.displayName = 'Button.Icon'

const commonPropTypes = {
  /** The message to be displayed within the button. */
  message: PropTypes.message,
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
   * A flag specifying whether the `secodnary` styling should applied to the
   * button.
   */
  secondary: PropTypes.bool,
  /**
   * A flag specifying whether the `naked` styling should applied to the
   * button.
   */
  naked: PropTypes.bool,
  /**
   * A flag specifying whether the `raw` styling should applied to the button.
   */
  raw: PropTypes.bool,
  /**
   * A flag specifying whether the `large` styling should applied to the button.
   */
  large: PropTypes.bool,
  /**
   * A flag specifying whether the `error` styling should applied to the button.
   */
  error: PropTypes.bool,
  /** The name of an icon to be displayed within the button. */
  icon: PropTypes.string,
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
  busy: commonPropTypes.busy,
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  icon: commonPropTypes.icon,
  message: commonPropTypes.message,
}

buttonChildren.defaultProps = {
  busy: undefined,
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

Button.Link.propTypes = {
  ...commonPropTypes,
  ...Link.propTypes,
}

Button.AnchorLink.propTypes = {
  ...commonPropTypes,
  ...Link.Anchor.propTypes,
}

export default Button
