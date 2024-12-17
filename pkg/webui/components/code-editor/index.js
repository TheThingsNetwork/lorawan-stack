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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import ReactAce from 'react-ace'
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'
import combineRefs from '@ttn-lw/lib/combine-refs'

import 'brace/mode/javascript'
import 'brace/mode/json'
import './ttn-theme'

import style from './code-editor.styl'

const CodeEditor = ({
  className,
  language,
  name,
  value,
  placeholder,
  readOnly,
  editorOptions,
  height,
  showGutter,
  minLines,
  maxLines,
  commands,
  editorRef,
  onBlur,
  onChange,
  onFocus,
  scrollToBottom,
  darkTheme,
}) => {
  const [focus, setFocus] = useState(false)
  const aceRef = useRef()
  const oldValue = useRef(value)

  const handleFocus = useCallback(
    evt => {
      setFocus(true)
      onFocus(evt)
    },
    [onFocus],
  )

  const handleBlur = useCallback(
    evt => {
      setFocus(false)
      onBlur(evt)
    },
    [onBlur],
  )

  const handleChange = useCallback(
    evt => {
      onChange(evt)
    },
    [onChange],
  )

  const empty = !value || value === ''
  const currentValue = empty && !focus ? placeholder : value

  useEffect(() => {
    if (scrollToBottom && value !== oldValue.current) {
      const row = aceRef.current.editor.session.getLength()
      aceRef.current.editor.gotoLine(row)
      oldValue.current = value
    }
  }, [scrollToBottom, value])

  const editorCls = classnames(className, style.wrapper, {
    [style.focus]: focus,
    [style.readOnly]: readOnly,
  })

  const options = {
    tabSize: 2,
    useSoftTabs: true,
    fontFamily: '"IBM Plex Mono", Consolas, Monaco, "Andale Mono", "Ubuntu Mono", monospace',
    fontSize: '13px',
    highlightSelectedWord: true,
    displayIndentGuides: true,
    showFoldWidgets: false,
    behavioursEnabled: !(readOnly || empty),
    ...editorOptions,
  }

  return (
    <div className={editorCls} data-test-id={`code-editor-${name}`}>
      <ReactAce
        // Rendered options.
        theme={darkTheme ? 'ttn-dark' : 'ttn-light'}
        minLines={minLines}
        maxLines={maxLines}
        // Session options.
        mode={language}
        // Editor options.
        readOnly={readOnly}
        highlightActiveLine
        showGutter={showGutter}
        // Other props.
        name={name}
        onChange={handleChange}
        value={currentValue}
        defaultValue={placeholder}
        setOptions={options}
        width="100%"
        height={height}
        onFocus={handleFocus}
        onBlur={handleBlur}
        editorProps={{ $blockScrolling: Infinity }}
        commands={commands}
        ref={editorRef ? combineRefs([aceRef, editorRef]) : aceRef}
        wrapEnabled
      />
    </div>
  )
}

CodeEditor.propTypes = {
  className: PropTypes.string,
  /** New commands to add to the editor, see official docs. */
  commands: PropTypes.arrayOf(PropTypes.shape({})),
  /** A flag indicating whether the editor should use a dark theme. */
  darkTheme: PropTypes.bool,
  /** See `https://github.com/ajaxorg/ace/wiki/Configuring-Ace`. */
  editorOptions: PropTypes.shape({}),
  editorRef: PropTypes.shape({ current: PropTypes.shape({}) }),
  /** The height of the editor. */
  height: PropTypes.string,
  /** The language to highlight. */
  language: PropTypes.oneOf(['javascript', 'json']),
  /** Maximum lines of code allowed. */
  maxLines: PropTypes.number,
  /** Minimum lines of code allowed. */
  minLines: PropTypes.number,
  /** The name of the editor (should be unique). */
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func,
  onChange: PropTypes.func,
  onFocus: PropTypes.func,
  /** The default value of the editor. */
  placeholder: PropTypes.string,
  /** A flag identifying whether the editor is editable. */
  readOnly: PropTypes.bool,
  /** A flag indicating whether the editor should scroll to the bottom when
   * the value has been updated, useful for logging use cases.
   */
  scrollToBottom: PropTypes.bool,
  showGutter: PropTypes.bool,
  /** The current value of the editor. */
  value: PropTypes.string,
}

CodeEditor.defaultProps = {
  className: undefined,
  commands: undefined,
  darkTheme: false,
  editorOptions: undefined,
  height: '30rem',
  language: 'javascript',
  maxLines: Infinity,
  minLines: 1,
  onBlur: () => null,
  onChange: () => null,
  onFocus: () => null,
  placeholder: '',
  readOnly: false,
  scrollToBottom: false,
  showGutter: true,
  value: '',
  editorRef: null,
}

export default CodeEditor
