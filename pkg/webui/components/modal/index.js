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

import React from 'react'
import classnames from 'classnames'
import FocusLock from 'react-focus-lock'
import { RemoveScroll } from 'react-remove-scroll'

import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './modal.styl'

const Modal = props => {
  const {
    buttonName,
    buttonMessage,
    title,
    subtitle,
    children,
    message,
    logo,
    approval,
    formName,
    cancelButtonMessage,
    onComplete,
    bottomLine,
    inline,
    danger,
    approveButtonProps,
    cancelButtonProps,
    ...rest
  } = props

  const approveButtonRef = React.useRef(null)
  const approvalAllowed = !Boolean(approveButtonProps.disabled)

  const modalClassNames = classnames(style.modal, style.modal, {
    [style.modalInline]: inline,
    [style.modalAbsolute]: !Boolean(inline),
  })

  const handleModalActivate = React.useCallback(() => {
    if (approveButtonRef.current !== null) {
      approveButtonRef.current.focus()
    }
  }, [])
  const handleComplete = React.useCallback(
    result => {
      onComplete(result)
    },
    [onComplete],
  )
  const handleApprove = React.useCallback(() => {
    handleComplete(true)
  }, [handleComplete])
  const handleCancel = React.useCallback(() => {
    handleComplete(false)
  }, [handleComplete])
  const handleKeyDown = React.useCallback(
    evt => {
      if (approval && evt.key === 'Escape') {
        evt.stopPropagation()
        handleCancel()

        return
      }

      if (approvalAllowed && evt.key === 'Enter') {
        evt.stopPropagation()
        handleApprove()

        return
      }
    },
    [approval, approvalAllowed, handleApprove, handleCancel],
  )

  const name = formName ? { name: formName } : {}
  const RootComponent = props.method ? 'form' : 'div'
  const messageElement = message && <Message content={message} className={style.message} />
  const bottomLineElement =
    typeof bottomLine === 'object' && Boolean(bottomLine.id) ? (
      <Message content={bottomLine} />
    ) : (
      bottomLine
    )

  const approveButtonMessage =
    buttonMessage !== undefined
      ? buttonMessage
      : approval
      ? sharedMessages.approve
      : sharedMessages.ok
  let buttons = (
    <div>
      <Button
        message={approveButtonMessage}
        onClick={handleApprove}
        icon="check"
        ref={approveButtonRef}
        {...approveButtonProps}
      />
    </div>
  )

  if (approval) {
    buttons = (
      <div>
        <Button
          secondary
          message={cancelButtonMessage}
          onClick={handleCancel}
          name={formName}
          icon="clear"
          value="false"
          {...name}
          {...cancelButtonProps}
        />
        <Button
          message={approveButtonMessage}
          onClick={handleApprove}
          name={formName}
          icon="check"
          value="true"
          danger={danger}
          ref={approveButtonRef}
          {...name}
          {...approveButtonProps}
        />
      </div>
    )
  }

  return (
    <FocusLock autoFocus returnFocus onActivation={handleModalActivate}>
      <RemoveScroll>
        {!inline && <div key="shadow" className={style.shadow} />}
        <RootComponent
          data-test-id="modal-window"
          key="modal"
          className={modalClassNames}
          onKeyDown={handleKeyDown}
          aria-modal="true"
          role="dialog"
          tabIndex={-1}
          {...rest}
        >
          {title && (
            <div className={style.titleSection}>
              <div>
                <Message className={style.title} content={title} component="h1" />
                {subtitle && <Message component="span" content={subtitle} />}
              </div>
              {logo}
            </div>
          )}
          {title && <div className={style.line} />}
          <div className={style.body}>{children || messageElement}</div>
          <div className={style.controlBar}>
            <div>{bottomLineElement}</div>
            {buttons}
          </div>
        </RootComponent>
      </RemoveScroll>
    </FocusLock>
  )
}

Modal.propTypes = {
  approval: PropTypes.bool,
  approveButtonProps: PropTypes.shape({
    disabled: PropTypes.bool,
  }),
  bottomLine: PropTypes.oneOfType([PropTypes.element, PropTypes.message]),
  buttonMessage: PropTypes.message,
  buttonName: PropTypes.message,
  cancelButtonMessage: PropTypes.message,
  cancelButtonProps: PropTypes.shape({}),
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.element), PropTypes.element]),
  danger: PropTypes.bool,
  formName: PropTypes.string,
  inline: PropTypes.bool,
  logo: PropTypes.node,
  message: PropTypes.message,
  method: PropTypes.string,
  name: PropTypes.string,
  onComplete: PropTypes.func,
  subtitle: PropTypes.message,
  title: PropTypes.message,
}

Modal.defaultProps = {
  bottomLine: undefined,
  buttonMessage: undefined,
  buttonName: undefined,
  cancelButtonMessage: sharedMessages.cancel,
  children: undefined,
  danger: false,
  formName: undefined,
  logo: undefined,
  message: undefined,
  method: undefined,
  onComplete: () => null,
  inline: false,
  approval: true,
  subtitle: undefined,
  title: undefined,
  name: undefined,
  approveButtonProps: {},
  cancelButtonProps: {},
}

export default Modal
