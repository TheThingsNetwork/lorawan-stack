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
import humanFileSize from '@ttn-lw/lib/human-file-size'

import ButtonGroup from '../button/group'

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
    center: PropTypes.bool,
    changeMessage: PropTypes.message,
    /** `dataTransform` is a marshaler used to transform the raw field value into.
     * a value matching the field schema. */
    dataTransform: PropTypes.func,
    disabled: PropTypes.bool,
    id: PropTypes.string.isRequired,
    image: PropTypes.bool,
    imageClassName: PropTypes.string,
    largeFileWarningMessage: PropTypes.message,
    maxSize: PropTypes.number,
    mayRemove: PropTypes.bool,
    message: PropTypes.message,
    name: PropTypes.string.isRequired,
    onChange: PropTypes.func.isRequired,
    providedMessage: PropTypes.message,
    value: PropTypes.oneOfType([PropTypes.string, PropTypes.shape({})]),
    warningSize: PropTypes.number,
  }

  static defaultProps = {
    accept: undefined,
    center: false,
    dataTransform: defaultDataTransform,
    disabled: false,
    image: false,
    imageClassName: undefined,
    largeFileWarningMessage: undefined,
    maxSize: 16 * 1024 * 1024, // 16 MB
    mayRemove: true,
    message: m.selectAFile,
    changeMessage: m.changeFile,
    providedMessage: m.fileProvided,
    value: undefined,
    warningSize: undefined,
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

    if (files && files[0]) {
      if (files[0].size >= maxSize) {
        this.setState({ error: m.tooBig })
      } else {
        this.reader.readAsDataURL(files[0])
        this.setState({
          filename: files[0].name,
          error: undefined,
          isLarger: files[0].size >= warningSize,
        })
      }
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
    this.setState({ filename: '', error: undefined, isLarger: false })
    onChange(dataTransform(''), true)
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
              className="ml-cs-s"
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

    return <Message className="tc-subtle-gray" content={m.noFileSelected} />
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
      center,
    } = this.props
    const warningThreshold = humanFileSize(warningSize)

    return (
      <div className={classnames(style.container, { [style.center]: center })}>
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
          <ButtonGroup
            align={center ? 'center' : 'start'}
            className={classnames({ [style.buttonGroupCenter]: center })}
          >
            <Button
              type="button"
              aria-controls="fileupload"
              onClick={this.handleChooseClick}
              disabled={disabled}
              message={!value ? message : changeMessage}
              icon="attachment"
              className="mr-cs-s"
            />
            {this.statusMessage}
          </ButtonGroup>
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
