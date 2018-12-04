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

import React, { Component } from 'react'
import { connect } from 'react-redux'
import bind from 'autobind-decorator'

import Modal from '../../components/modal'

import { removeModal } from '../../actions/modal'

/**
 * ConnectedModal is a modal component connected to the store. It will display
 * all messages fed into the `modal` subtree of the global store and remove them
 * upon completing the modal. This way, modal messages can be initiated from
 * anywhere by dispatching the `SET_MODAL` action.
 */
@connect(state => ({
  modal: state.modal,
}))
@bind
class ConnectedModal extends Component {

  handleComplete (result) {
    const { dispatch, modal } = this.props
    if (modal && modal.onComplete) {
      modal.onComplete(result)
    }

    dispatch(removeModal())
  }

  render () {
    const { dispatch, modal, ...rest } = this.props

    if (!modal) {
      return null
    }

    const { onComplete, ...modalRest } = modal
    const props = { ...rest, ...modalRest }

    return <Modal onComplete={this.handleComplete} {...props} />
  }
}

export default ConnectedModal
