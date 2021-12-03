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
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import Notification from '@ttn-lw/components/notification'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './file-input.styl'

const m = defineMessages({
  selectAFile: 'Select a file…',
  changeFile: 'Change file…',
  noFileSelected: 'No file selected',
  fileProvided: 'A file has been provided',
  tooBig: 'The selected file is too large',
  remove: 'Remove',
})

const defaultDataTransform = content => content.replace(/^.*;base64,/, '')

export default class FileInput extends Component {
  static propTypes = {
    accept: PropTypes.oneOfType([PropTypes.string, PropTypes.array]),
    changeMessage: PropTypes.message,
    /** `dataTransform` is a marshaler used to transform the raw field value into.
     * a value matching the field schema. */
    dataTransform: PropTypes.func,
    disabled: PropTypes.bool,
    id: PropTypes.string.isRequired,
    image: PropTypes.bool,
    imageClassName: PropTypes.string,
    largeFileWarningMessage: PropTypes.string.isRequired,
    maxSize: PropTypes.number,
    mayRemove: PropTypes.bool,
    message: PropTypes.message,
    name: PropTypes.string.isRequired,
    onChange: PropTypes.func.isRequired,
    providedMessage: PropTypes.message,
    value: PropTypes.oneOfType([PropTypes.string, PropTypes.shape({})]),
    warningSize: PropTypes.number.isRequired,
  }

  static defaultProps = {
    accept: undefined,
    dataTransform: defaultDataTransform,
    disabled: false,
    image: false,
    imageClassName: undefined,
    maxSize: 16 * 1024 * 1024, // 16 MB
    mayRemove: true,
    message: m.selectAFile,
    changeMessage: m.changeFile,
    providedMessage: m.fileProvided,
    value: undefined,
  }

  constructor(props) {
    super(props)

    this.reader = new FileReader()
    this.reader.onload = this.handleFileRead
    this.fileInputRef = React.createRef()
    this.imageRef = React.createRef()

    this.state = {
      filename: '',
      isLarger: false,
    }
  }

  @bind
  handleFileRead(event) {
    const { onChange, dataTransform, image } = this.props
    const { result: content } = event.target

    if (image && Boolean(this.imageRef.current)) {
      this.imageRef.current.style.display = 'block'
    }

    const data = dataTransform(content)
    onChange(data, true)
  }

  @bind
  handleChange(event) {
    const { maxSize, warningSize } = this.props
    const { files } = event.target

    if (files && files[0] && files[0].size <= maxSize) {
      if (typeof warningSize === 'number' && files && files[0] && files[0].size >= warningSize) {
        this.setState({ isLarger: true })
      }
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
    const { onChange, dataTransform } = this.props

    this.fileInputRef.current.value = null
    this.setState({ filename: '', error: undefined })
    onChange(dataTransform(''), true)
  }

  @bind
  humanFileSize(size) {
    const i = size === 0 ? 0 : Math.floor(Math.log(size) / Math.log(1024))
    return `${(size / Math.pow(1024, i)).toFixed(2) * 1} ${['B', 'kB', 'MB', 'GB', 'TB'][i]}`
  }

  handleImageError(error) {
    error.target.style.display = 'none'
  }

  get statusMessage() {
    const { value, providedMessage, mayRemove } = this.props
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
          {mayRemove && (
            <Button
              className={style.removeButton}
              message={m.remove}
              onClick={this.handleRemoveClick}
              type="button"
              icon="delete"
              danger
              naked
            />
          )}
        </React.Fragment>
      )
    }

    return <Message className={style.noFile} content={m.noFileSelected} />
  }

  render() {
    const {
      message,
      changeMessage,
      name,
      id,
      accept,
      value,
      disabled,
      image,
      imageClassName,
      largeFileWarningMessage,
      warningSize,
    } = this.props
    const warningThreshold = this.humanFileSize(warningSize)

    return (
      <div className={style.container}>
        {this.state.isLarger && (
          <Notification
            className={style.notification}
            content={largeFileWarningMessage}
            messageValues={{ warningThreshold }}
            small
            warning
          />
        )}
        <div>
          {image && Boolean(value) && (
            <img
              className={classnames(style.image, imageClassName)}
              alt="Current image"
              src={value}
              onError={this.handleImageError}
              ref={this.imageRef}
            />
          )}
          <Button
            type="button"
            aria-controls="fileupload"
            onClick={this.handleChooseClick}
            disabled={disabled}
            message={!value ? message : changeMessage}
            icon="attachment"
            secondary
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
      </div>
    )
  }
}
