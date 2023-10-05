// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useState, useEffect, useRef } from 'react'
import classnames from 'classnames'
import clipboard from 'clipboard'
import { defineMessages, useIntl } from 'react-intl'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './safe-inspector.styl'

const chunkArray = (array, chunkSize) =>
  Array.from({ length: Math.ceil(array.length / chunkSize) }, (_, index) =>
    array.slice(index * chunkSize, (index + 1) * chunkSize),
  )

const selectText = node => {
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
  toggleVisibility: 'Toggle visibility',
  arrayFormatting: 'Toggle array formatting',
  byteOrder: 'Switch byte order',
})

const MSB = 'msb'
const LSB = 'lsb'
const UINT32_T = 'uint32_t'
const representationRotateMap = {
  [MSB]: LSB,
  [LSB]: UINT32_T,
  [UINT32_T]: MSB,
}

const SafeInspector = ({
  data,
  hideable,
  initiallyVisible,
  enableUint32,
  className,
  isBytes,
  small,
  noCopyPopup,
  noCopy,
  noTransform,
  truncateAfter,
  disableResize,
}) => {
  const _timer = useRef(null)

  const [hidden, setHidden] = useState((hideable && !initiallyVisible) || false)
  const [byteStyle, setByteStyle] = useState(true)
  const [copied, setCopied] = useState(false)
  const [copyIcon, setCopyIcon] = useState('file_copy')
  const [representation, setRepresentation] = useState(MSB)
  const [truncated, setTruncated] = useState(false)

  const intl = useIntl()

  const containerElem = useRef(null)
  const displayElem = useRef(null)
  const buttonsElem = useRef(null)
  const copyElem = useRef(null)

  const getNextRepresentation = useCallback(
    current => {
      const next = representationRotateMap[current]

      return next === UINT32_T && !enableUint32 ? representationRotateMap[next] : next
    },
    [enableUint32],
  )

  const checkTruncateState = useCallback(() => {
    if (!containerElem.current) {
      return
    }

    const containerWidth = containerElem.current.offsetWidth
    const buttonsWidth = buttonsElem.current.offsetWidth
    const displayWidth = displayElem.current.offsetWidth
    const netContainerWidth = containerWidth - buttonsWidth - 14

    if (netContainerWidth < displayWidth && !truncated) {
      setTruncated(true)
    } else if (netContainerWidth > displayWidth && truncated) {
      setTruncated(false)
    }
  }, [truncated])

  const handleVisibiltyToggle = useCallback(() => {
    setHidden(prev => !prev)
    setByteStyle(prev => (!prev && !hidden ? true : prev))
    checkTruncateState()
  }, [checkTruncateState, hidden])

  const handleTransformToggle = useCallback(async () => {
    setByteStyle(prev => !prev)
    checkTruncateState()
  }, [checkTruncateState])

  const handleSwapToggle = useCallback(() => {
    setRepresentation(prev => getNextRepresentation(prev))
  }, [getNextRepresentation])

  const handleDataClick = useCallback(() => {
    if (!hidden) {
      selectText(displayElem.current)
    }
  }, [hidden])

  const handleCopyClick = useCallback(() => {
    if (copied) {
      return
    }

    setCopied(true)
    setCopyIcon('done')
    if (noCopyPopup) {
      _timer.current = setTimeout(() => {
        setCopied(false)
        setCopyIcon('file_copy')
      }, 2000)
    }
  }, [copied, noCopyPopup])

  const handleCopyAnimationEnd = useCallback(() => {
    setCopied(false)
    setCopyIcon('file_copy')
  }, [])

  useEffect(() => {
    if (copyElem && copyElem.current) {
      new clipboard(copyElem.current, { container: containerElem.current })
    }

    if (!disableResize) {
      const handleWindowResize = () => {
        // Your resize logic here
        checkTruncateState()
      }

      window.addEventListener('resize', handleWindowResize)
      checkTruncateState()

      return () => {
        window.removeEventListener('resize', handleWindowResize)
      }
    }

    return () => {
      clearTimeout(_timer.current)
    }
  }, [_timer, checkTruncateState, disableResize])

  const handleContainerClick = useCallback(e => {
    e.preventDefault()
    e.stopPropagation()
  }, [])

  let formattedData = isBytes ? data.toUpperCase() : data
  let display = formattedData

  if (isBytes) {
    let chunks = chunkArray(data.toUpperCase().split(''), 2)
    if (chunks.length > truncateAfter && !truncated) {
      setTruncated(true)
      chunks = chunks.slice(0, truncateAfter)
    }
    if (!byteStyle) {
      if (representation === UINT32_T) {
        formattedData = display = `0x${data}`
      } else {
        const orderedChunks = representation === MSB ? chunks : chunks.reverse()
        formattedData = display = orderedChunks.map(chunk => `0x${chunk.join('')}`).join(', ')
      }
    } else {
      display = chunks.map((chunk, index) => (
        <span key={`${data}_chunk_${index}`}>{hidden ? '••' : chunk}</span>
      ))
    }
  } else if (hidden) {
    display = '•'.repeat(Math.min(formattedData.length, truncateAfter))
  }

  if (truncated) {
    display = [...display, '…']
  }

  const containerStyle = classnames(className, style.container, {
    [style.containerSmall]: small,
    [style.containerHidden]: hidden,
  })

  const dataStyle = classnames(style.data, {
    [style.dataHidden]: hidden,
    [style.dataTruncated]: truncated,
  })

  const copyButtonStyle = classnames(style.buttonIcon, {
    [style.buttonIconCopied]: copied,
  })

  const renderButtonContainer = hideable || !noCopy || !noTransform

  return (
    <div ref={containerElem} className={containerStyle} onClick={handleContainerClick}>
      <div
        ref={displayElem}
        onClick={handleDataClick}
        className={dataStyle}
        title={truncated ? formattedData : undefined}
      >
        {display}
      </div>
      {renderButtonContainer && (
        <div ref={buttonsElem} className={style.buttons}>
          {!hidden && !byteStyle && isBytes && (
            <React.Fragment>
              <span>{representation}</span>
              <button
                title={intl.formatMessage(m.byteOrder)}
                className={style.buttonSwap}
                onClick={handleSwapToggle}
              >
                <Icon className={style.buttonIcon} small icon="swap_horiz" />
              </button>
            </React.Fragment>
          )}
          {!noTransform && !hidden && isBytes && (
            <button
              title={intl.formatMessage(m.arrayFormatting)}
              className={style.buttonTransform}
              onClick={handleTransformToggle}
            >
              <Icon className={style.buttonIcon} small icon="code" />
            </button>
          )}
          {!noCopy && (
            <button
              title={intl.formatMessage(sharedMessages.copyToClipboard)}
              className={style.buttonCopy}
              onClick={handleCopyClick}
              data-clipboard-text={formattedData}
              ref={copyElem}
              disabled={copied}
            >
              <Icon className={copyButtonStyle} onClick={handleCopyClick} small icon={copyIcon} />
              {copied && !noCopyPopup && (
                <Message
                  content={sharedMessages.copiedToClipboard}
                  onAnimationEnd={handleCopyAnimationEnd}
                  className={style.copyConfirm}
                />
              )}
            </button>
          )}
          {hideable && (
            <button
              title={intl.formatMessage(m.toggleVisibility)}
              className={style.buttonVisibility}
              onClick={handleVisibiltyToggle}
            >
              <Icon
                className={style.buttonIcon}
                small
                icon={hidden ? 'visibility' : 'visibility_off'}
              />
            </button>
          )}
        </div>
      )}
    </div>
  )
}

SafeInspector.propTypes = {
  /** The classname to be applied. */
  className: PropTypes.string,
  /** The data to be displayed. */
  data: PropTypes.string.isRequired,
  /** Whether the component should resize when its data is truncated. */
  disableResize: PropTypes.bool,
  /** Whether uint32_t notation should be enabled for byte representation. */
  enableUint32: PropTypes.bool,
  /** Whether the data can be hidden (like passwords). */
  hideable: PropTypes.bool,
  /** Whether the data is initially visible. */
  initiallyVisible: PropTypes.bool,
  /** Whether the data is in byte format. */
  isBytes: PropTypes.bool,
  /** Whether to hide the copy action. */
  noCopy: PropTypes.bool,
  /** Whether to hide the copy popup click and just display checkmark. */
  noCopyPopup: PropTypes.bool,
  /** Whether to hide the data transform action. */
  noTransform: PropTypes.bool,
  /**
   * Whether a smaller style should be rendered (useful for display in
   * tables).
   */
  small: PropTypes.bool,
  /** The input count (byte or characters, based on type) after which the
   * display is truncated.
   */
  truncateAfter: PropTypes.number,
}

SafeInspector.defaultProps = {
  className: undefined,
  noCopyPopup: false,
  disableResize: false,
  hideable: true,
  initiallyVisible: false,
  isBytes: true,
  small: false,
  noTransform: false,
  noCopy: false,
  enableUint32: false,
  truncateAfter: Infinity,
}

export default SafeInspector
