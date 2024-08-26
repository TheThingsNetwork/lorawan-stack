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
import bind from 'autobind-decorator'

import Button from '@ttn-lw/components/button'

import toast, { ToastContainer } from '.'

const TOAST_TITLE = 'Toast title'
const TOAST_CONTENT = 'Toast content'

class Example extends React.Component {
  @bind
  showToast(evt) {
    let type = null
    switch (evt.currentTarget.name) {
      case 'success':
        type = toast.types.SUCCESS
        break
      case 'info':
        type = toast.types.INFO
        break
      case 'warning':
        type = toast.types.WARNING
        break
      case 'error':
        type = toast.types.ERROR
        break
      default:
        type = toast.types.DEFAULT
    }

    toast({
      type,
      title: TOAST_TITLE,
      message: TOAST_CONTENT,
    })
  }

  render() {
    return (
      <div style={{ width: '100%', height: '100%' }}>
        <ToastContainer />
        <Button onClick={this.showToast} message="Show Default Toast" name="default" primary />
        <Button onClick={this.showToast} message="Show Success Toast" name="success" primary />
        <Button onClick={this.showToast} message="Show Error Toast" name="error" primary />
        <Button onClick={this.showToast} message="Show Warning Toast" name="warning" primary />
        <Button onClick={this.showToast} message="Show Info Toast" name="info" primary />
      </div>
    )
  }
}

export default {
  title: 'Toast',
}

export const Default = () => <Example />
