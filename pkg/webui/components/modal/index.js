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
import PropTypes from 'prop-types'

import Button from '../button'
import Logo from '../logo'

import style from './modal.styl'

@bind
export default class Modal extends React.PureComponent {
  static propTypes = {
    title: PropTypes.string,
    children: PropTypes.oneOfType([
      PropTypes.arrayOf(PropTypes.element),
      PropTypes.element,
    ]),
    message: PropTypes.string,
    subtitle: PropTypes.string,
    bottomLine: PropTypes.oneOfType([
      PropTypes.element,
      PropTypes.string,
    ]),
    approval: PropTypes.bool,
    buttonMessage: PropTypes.string,
    cancelButtonMessage: PropTypes.string,
    method: PropTypes.string,
    buttonName: PropTypes.string,
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
      bottomLine,
      logo,
      approval = false,
      formName,
      buttonMessage = this.props.approval ? 'Approve' : 'Ok',
      cancelButtonMessage = 'Cancel',
      onComplete,
      ...rest
    } = this.props

    const name = formName ? { name: formName } : {}
    const RootComponent = this.props.method ? 'form' : 'div'
    const messageElement = (<span className={style.message}>{message}</span>)

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
              <h1>{title}</h1>
              { subtitle && (<span>{subtitle}</span>) }
            </div>
            { logo && (<Logo className={style.logo} />)}
          </div>
        }
        { title && <div className={style.line} /> }
        <div className={style.body}>
          {children || messageElement}
        </div>
        <div className={style.controlBar}>
          <div><span>{ bottomLine }</span></div>
          {buttons}
        </div>
      </RootComponent>,
    ]
  }
}
