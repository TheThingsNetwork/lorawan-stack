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

import ApiKeyModal from '../api-key-modal'
import PropTypes from '../../lib/prop-types'
import sharedMessages from '../../lib/shared-messages'
import SubmitBar from '../submit-bar'
import Field from '../field'
import FieldGroup from '../field/group'
import Button from '../button'
import Message from '../../lib/components/message'
import ApiKeyForm from './form'
import validationSchema from './validation-schema'

import style from './api-key-form.styl'

@bind
class CreateForm extends React.Component {

  state = {
    modal: null,
  }

  async handleModalApprove () {
    const { onCreateSuccess } = this.props
    const { key } = this.state

    await this.setState({ modal: null })
    await onCreateSuccess(key)
  }

  async handleCreate (values) {
    const { rights, name } = values
    const { onCreate } = this.props

    const key = {
      name,
      rights: Object.keys(rights).filter(r => rights[r]),
    }

    return await onCreate(key)
  }

  async handleCreateSuccess (key) {
    await this.setState({
      modal: {
        secret: key.key,
        rights: key.rights,
        onComplete: this.handleModalApprove,
        approval: false,
      },
      key,
    })
  }

  render () {
    const {
      rights,
      onCreateFailure,
    } = this.props
    const { modal } = this.state

    const modalProps = modal ? modal : {}
    const modalVisible = Boolean(modal)
    const { rightsItems, rightsValues } = rights.reduce(
      function (acc, right) {
        acc.rightsItems.push(
          <Field
            className={style.rightLabel}
            key={right}
            name={right}
            type="checkbox"
            title={{ id: `enum:${right}` }}
            form
          />
        )
        acc.rightsValues[right] = false

        return acc
      },
      {
        rightsItems: [],
        rightsValues: {},
      }
    )
    const initialValues = {
      name: '',
      rights: rightsValues,
    }

    return (
      <React.Fragment>
        <ApiKeyModal
          {...modalProps}
          visible={modalVisible}
          approval={false}
        />
        <ApiKeyForm
          rights={rights}
          onSubmit={this.handleCreate}
          onSubmitSuccess={this.handleCreateSuccess}
          onSubmitFailure={onCreateFailure}
          validationSchema={validationSchema}
          initialValues={initialValues}
        >
          <Message
            component="h4"
            content={sharedMessages.generalInformation}
          />
          <Field
            title={sharedMessages.name}
            name="name"
            type="text"
            autoFocus
          />
          <FieldGroup
            name="rights"
            title={sharedMessages.rights}
          >
            {rightsItems}
          </FieldGroup>
          <SubmitBar>
            <Button
              type="submit"
              message={sharedMessages.createApiKey}
            />
          </SubmitBar>
        </ApiKeyForm>
      </React.Fragment>
    )
  }
}

CreateForm.propTypes = {
  /** The list of rights */
  rights: PropTypes.arrayOf(PropTypes.string),
  /**
   * Called after successful creation of the API key.
   * Receives the key object as an argument.
   */
  onCreateSuccess: PropTypes.func,
  /**
   * Called after unsuccessful creation of the API key.
   * Receives the error object as an argument.
   */
  onCreateFailure: PropTypes.func,
  /**
   * Called on form submission.
   * Receives the key object as an argument.
   */
  onCreate: PropTypes.func.isRequired,
}

CreateForm.defaultProps = {
  rights: [],
  onCreateSuccess: () => null,
  onCreateFailure: () => null,
}

export default CreateForm
