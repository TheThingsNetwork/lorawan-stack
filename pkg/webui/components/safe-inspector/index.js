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
import bind from 'autobind-decorator'
import classnames from 'classnames'
import clipboard from 'clipboard'
import { defineMessages, injectIntl } from 'react-intl'

import Icon from '../icon'
import Message from '../../lib/components/message'
import PropTypes from '../../lib/prop-types'

import style from './safe-inspector.styl'

function chunkArray(array, chunkSize) {
  return Array.from({ length: Math.ceil(array.length / chunkSize) }, (_, index) =>
    array.slice(index * chunkSize, (index + 1) * chunkSize),
  )
}

function selectText(node) {
  if (document.body.createTextRange) {
    const range = document.body.createTextRange()
    range.moveToElementText(node)
    range.select()
  } else if (window.getSelection) {
    const selection = window.getSelection()
    const range = document.createRange()
    range.selectNodeContents(node)
    selection.removeAllRanges()
    selection.addRange(range)
  }
}

const m = defineMessages({
  copied: 'Copied to clipboard!',
  toggleVisibility: 'Toggle visibility',
  copyClipboard: 'Copy to clipboard',
  arrayFormatting: 'Toggle array formatting',
  byteOrder: 'Switch byte order',
})

@injectIntl
@bind
export class SafeInspector extends Component {
  static defaultProps = {
    className: undefined,
    noCopyPopup: false,
    disableResize: false,
    hideable: true,
    initiallyVisible: false,
    isBytes: true,
    small: false,
  }

  static propTypes = {
    /** The classname to be applied **/
    className: PropTypes.string,
    /** Whether to hide the copy popup click and just display checkmark */
    data: PropTypes.string.isRequired,
    /** The data to be displayed */
    disableResize: PropTypes.bool,
    /** Whether the component should resize when its data is truncated */
    hideable: PropTypes.bool,
    /** Whether the data can be hidden (like passwords) */
    initiallyVisible: PropTypes.bool,
    /** Whether the data is initially visible */
    intl: PropTypes.shape({
      formatMessage: PropTypes.func,
    }).isRequired,
    /** Utility functions passed via react-intl hoc **/
    isBytes: PropTypes.bool,
    /** Whether the data is in byte format */
    noCopyPopup: PropTypes.bool,
    /** Whether a smaller style should be rendered (useful for display in tables) */
    small: PropTypes.bool,
  }

  constructor(props) {
    super(props)

    this._timer = null

    this.state = {
      hidden: (props.hideable && !props.initiallyVisible) || false,
      byteStyle: true,
      copied: false,
      copyIcon: 'file_copy',
      msb: true,
      truncated: false,
    }

    this.containerElem = React.createRef()
    this.displayElem = React.createRef()
    this.buttonsElem = React.createRef()
    this.copyElem = React.createRef()
  }

  handleVisibiltyToggle() {
    this.setState(prev => ({
      byteStyle: !prev.hidden ? true : prev.byteStyle,
      hidden: !prev.hidden,
    }))
  }

  async handleTransformToggle() {
    await this.setState(prev => ({ byteStyle: !prev.byteStyle }))
    this.checkTruncateState()
  }

  handleSwapToggle() {
    this.setState(prev => ({ msb: !prev.msb }))
  }

  handleDataClick(e) {
    if (!this.state.hidden) {
      selectText(this.displayElem.current)
    }
    e.stopPropagation()
  }

  handleCopyClick(e) {
    const { noCopyPopup } = this.props
    this.setState({ copied: true, copyIcon: 'done' })
    if (noCopyPopup) {
      this._timer = setTimeout(() => {
        this.setState({ copied: false, copyIcon: 'file_copy' })
      }, 2000)
    }
    e.stopPropagation()
  }

  handleCopyAnimationEnd() {
    this.setState({ copied: false, copyIcon: 'file_copy' })
  }

  componentDidMount() {
    const { disableResize } = this.props
    new clipboard(this.copyElem.current)
    if (!disableResize) {
      window.addEventListener('resize', this.handleWindowResize)
      this.checkTruncateState()
    }
  }

  componentWillUnmount() {
    const { disableResize } = this.props
    if (!disableResize) {
      window.removeEventListener('resize', this.handleWindowResize)
    }
    clearTimeout(this._timer)
  }

  handleWindowResize() {
    this.checkTruncateState()
  }

  checkTruncateState() {
    if (!this.containerElem.current) {
      return
    }

    const containerWidth = this.containerElem.current.offsetWidth
    const buttonsWidth = this.buttonsElem.current.offsetWidth
    const displayWidth = this.displayElem.current.offsetWidth
    const netContainerWidth = containerWidth - buttonsWidth - 14
    if (netContainerWidth < displayWidth && !this.state.truncated) {
      this.setState({ truncated: true })
    } else if (netContainerWidth > displayWidth && this.state.truncated) {
      this.setState({ truncated: false })
    }
  }

  render() {
    const { hidden, byteStyle, msb, copied, copyIcon } = this.state

    const { className, data, isBytes, hideable, small, intl, noCopyPopup } = this.props

    let formattedData = isBytes ? data.toUpperCase() : data
    let display = formattedData

    if (isBytes) {
      const chunks = chunkArray(data.toUpperCase().split(''), 2)
      if (!byteStyle) {
        const orderedChunks = msb ? chunks : chunks.reverse()
        formattedData = display = orderedChunks.map(chunk => `0x${chunk.join('')}`).join(', ')
      } else {
        display = chunks.map((chunk, index) => (
          <span key={`${data}_chunk_${index}`}>{hidden ? '••' : chunk}</span>
        ))
      }
    } else if (hidden) {
      display = '•'.repeat(formattedData.length)
    }

    const containerStyle = classnames(className, style.container, {
      [style.containerSmall]: small,
      [style.containerHidden]: hidden,
    })

    const dataStyle = classnames(style.data, {
      [style.dataHidden]: hidden,
      [style.dataTruncated]: this.state.truncated,
    })

    const copyButtonStyle = classnames(style.buttonIcon, {
      [style.buttonIconCopied]: copied,
    })

    return (
      <div ref={this.containerElem} className={containerStyle}>
        <div ref={this.displayElem} onClick={this.handleDataClick} className={dataStyle}>
          {display}
        </div>
        <div ref={this.buttonsElem} className={style.buttons}>
          {!hidden && !byteStyle && isBytes && (
            <React.Fragment>
              <span>{msb ? 'msb' : 'lsb'}</span>
              <button
                title={intl.formatMessage(m.byteOrder)}
                className={style.buttonSwap}
                onClick={this.handleSwapToggle}
              >
                <Icon className={style.buttonIcon} small icon="swap_horiz" />
              </button>
            </React.Fragment>
          )}
          {!hidden && isBytes && (
            <button
              title={intl.formatMessage(m.arrayFormatting)}
              className={style.buttonTransform}
              onClick={this.handleTransformToggle}
            >
              <Icon className={style.buttonIcon} small icon="code" />
            </button>
          )}
          <button
            title={intl.formatMessage(m.copyClipboard)}
            className={style.buttonCopy}
            onClick={this.handleCopyClick}
            data-clipboard-text={formattedData}
            ref={this.copyElem}
          >
            <Icon
              className={copyButtonStyle}
              onClick={this.handleCopyClick}
              small
              icon={copyIcon}
            />
            {copied && !noCopyPopup && (
              <Message
                content={m.copied}
                onAnimationEnd={this.handleCopyAnimationEnd}
                className={style.copyConfirm}
              />
            )}
          </button>
          {hideable && (
            <button
              title={intl.formatMessage(m.toggleVisibility)}
              className={style.buttonVisibility}
              onClick={this.handleVisibiltyToggle}
            >
              <Icon
                className={style.buttonIcon}
                small
                icon={hidden ? 'visibility' : 'visibility_off'}
              />
            </button>
          )}
        </div>
      </div>
    )
  }
}

export default SafeInspector
