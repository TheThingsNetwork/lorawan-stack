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

import React, { Component } from 'react'
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'

import PropTypes from '../../lib/prop-types'
import Message from '../../lib/components/message'
import Icon from '../../components/icon'
import Button from '../../components/button'

import style from './file-input.styl'

const m = defineMessages({
  selectAFile: 'Select a file…',
  changeFile: 'Change file…',
  noFileSelected: 'No file selected',
  fileProvided: 'A file has been provided',
  tooBig: 'The selected file is too large.',
  remove: 'Remove',
})

const dataTransform = function(content) {
  return content.replace(/^.*;base64,/, '')
}

export default class FileInput extends Component {
  constructor(props) {
    super(props)

    this.reader = new FileReader()
    this.reader.onload = this.handleFileRead
    this.fileInputRef = React.createRef()

    this.state = {
      filename: '',
    }
  }

  static propTypes = {
    accept: PropTypes.string,
    dataTransform: PropTypes.func,
    disabled: PropTypes.bool,
    maxSize: PropTypes.number,
    message: PropTypes.message,
    name: PropTypes.string.isRequired,
    onChange: PropTypes.func.isRequired,
    providedMessage: PropTypes.message,
    value: PropTypes.string,
  }

  static defaultProps = {
    accept: undefined,
    disabled: false,
    message: m.selectAFile,
    dataTransform,
    maxSize: 10 * 1024 * 1024, // 10 MB
    providedMessage: m.fileProvided,
    value: undefined,
  }

  @bind
  handleFileRead(event) {
    const { onChange, dataTransform } = this.props
    const { result: content } = event.target

    const data = dataTransform(content)
    onChange(data, true)
  }

  @bind
  handleChange(event) {
    const { maxSize } = this.props
    const { files } = event.target

    if (files && files[0] && files[0].size <= maxSize) {
      this.setState({ filename: files[0].name, error: undefined })
      this.reader.readAsDataURL(files[0])
    } else {
      this.setState({ error: m.tooBig })
    }
  }

  @bind
  handleChooseClick() {
    this.fileInputRef.current.click()
  }

  @bind
  handleRemoveClick() {
    const { onChange } = this.props

    this.setState({ filename: '', error: undefined })
    onChange('', true)
  }

  get statusMessage() {
    const { value, providedMessage } = this.props
    const { filename, error } = this.state
    const hasInitialValue = value && !filename
    const hasError = Boolean(error)

    if (hasError) {
      return (
        <React.Fragment>
          <Icon className={style.errorIcon} icon="error" />
          <Message className={style.error} content={error} />
        </React.Fragment>
      )
    } else if (hasInitialValue || Boolean(filename)) {
      return (
        <React.Fragment>
          {hasInitialValue ? <Message content={providedMessage} /> : filename}
          <Button
            className={style.removeButton}
            message={m.remove}
            onClick={this.handleRemoveClick}
            icon="delete"
            secondary
            naked
          />
        </React.Fragment>
      )
    }

    return <Message className={style.noFile} content={m.noFileSelected} />
  }

  render() {
    const { message, name, accept, value, disabled } = this.props
    const id = `file_input_${name}`

    return (
      <div className={style.container}>
        <Button
          type="button"
          aria-controls="fileupload"
          onClick={this.handleChooseClick}
          disabled={disabled}
          message={!value ? message : m.changeFile}
          icon="attachment"
        />
        <span className={style.status}>{this.statusMessage}</span>
        <input
          name={name}
          id={id}
          className={style.input}
          type="file"
          onChange={this.handleChange}
          ref={this.fileInputRef}
          accept={accept}
          disabled={disabled}
          tabIndex="-1"
        />
      </div>
    )
  }
}
