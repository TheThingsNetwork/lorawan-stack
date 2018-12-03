// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import DOM from 'react-dom'
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import classnames from 'classnames'
import PropTypes from '../../lib/prop-types'

import sharedMessages from '../../lib/shared-messages'

import Message from '../../lib/components/message'
import Button from '../button'
import Logo from '../logo'

import { removeModal } from '../../actions/modal'

import style from './modal.styl'

@connect()
@bind
export default class Modal extends React.PureComponent {
  static propTypes = {
    title: PropTypes.message,
    children: PropTypes.oneOfType([
      PropTypes.arrayOf(PropTypes.element),
      PropTypes.element,
    ]),
    message: PropTypes.message,
    subtitle: PropTypes.message,
    bottomLine: PropTypes.oneOfType([
      PropTypes.element,
      PropTypes.message,
    ]),
    approval: PropTypes.bool,
    buttonMessage: PropTypes.message,
    cancelButtonMessage: PropTypes.message,
    method: PropTypes.string,
    buttonName: PropTypes.message,
    inline: PropTypes.bool,
    danger: PropTypes.bool,
  }

  handleApprove () {
    this.handleComplete(true)
  }

  handleCancel () {
    this.handleComplete(false)
  }

  handleComplete (result) {
    const { onComplete, dispatch } = this.props
    if (onComplete) {
      onComplete(result)
    }

    dispatch(removeModal())
  }

  render () {
    const {
      title,
      subtitle,
      children,
      message,
      logo,
      approval = false,
      formName,
      buttonMessage = this.props.approval ? sharedMessages.approve : sharedMessages.ok,
      cancelButtonMessage = sharedMessages.cancel,
      onComplete,
      bottomLine,
      inline = false,
      dispatch,
      danger,
      ...rest
    } = this.props

    const modalClassNames = classnames(style.modal, style.modal, {
      [inline]: inline,
    })

    const name = formName ? { name: formName } : {}
    const RootComponent = this.props.method ? 'form' : 'div'
    const messageElement = (<Message content={message} className={style.message} />)
    const bottomLineElement = <Message content={bottomLine} />

    let buttons = <div><Button message={buttonMessage} onClick={this.handleApprove} icon="check" /></div>

    if (approval) {
      buttons = (
        <div>
          <Button
            secondary
            message={cancelButtonMessage}
            onClick={this.handleCancel}
            name={formName}
            icon="clear"
            value="false"
            {...name}
          />
          <Button
            message={buttonMessage}
            onClick={this.handleApprove}
            name={formName}
            icon="check"
            value="true"
            danger={danger}
            {...name}
          />
        </div>
      )
    }

    const results = (
      <React.Fragment>
        {!inline && <div key="shadow" className={style.shadow} />}
        <RootComponent key="modal" className={modalClassNames} {...rest}>
          { title
            && <div className={style.titleSection}>
              <div>
                <h1><Message content={title} /></h1>
                { subtitle && (<Message content={subtitle} />) }
              </div>
              { logo && (<Logo className={style.logo} />)}
            </div>
          }
          { title && <div className={style.line} /> }
          <div className={style.body}>
            {children || messageElement}
          </div>
          <div className={style.controlBar}>
            <div>{ bottomLineElement }</div>
            {buttons}
          </div>
        </RootComponent>
      </React.Fragment>
    )

    const modalNode = document.getElementById('modal-container')

    if (modalNode) {
      return DOM.createPortal(
        results,
        modalNode
      )
    }

    return results
  }
}
