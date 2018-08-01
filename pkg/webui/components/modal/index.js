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
import bind from 'autobind-decorator'
import PropTypes from '../../lib/prop-types'

import Message from '../message'
import Button from '../button'
import Logo from '../logo'

import style from './modal.styl'

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
  }

  handleApprove () {
    const { onComplete } = this.props
    if (onComplete) {
      onComplete(true)
    }
  }

  handleCancel () {
    const { onComplete } = this.props
    if (onComplete) {
      onComplete(false)
    }
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
      buttonMessage = this.props.approval ? 'Approve' : 'Ok',
      cancelButtonMessage = 'Cancel',
      onComplete,
      bottomLine,
      ...rest
    } = this.props

    const name = formName ? { name: formName } : {}
    const RootComponent = this.props.method ? 'form' : 'div'
    const messageElement = (<span className={style.message}>{message}</span>)
    const bottomLineElement = bottomLine === 'object'
      ? <Message content={bottomLine} />
      : bottomLine


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
            {...name}
          />
        </div>
      )
    }

    return [
      <div key="shadow" className={style.shadow} />,
      <RootComponent key="modal" className={style.modal} {...rest}>
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
          <div><span>{ bottomLineElement }</span></div>
          {buttons}
        </div>
      </RootComponent>,
    ]
  }
}
