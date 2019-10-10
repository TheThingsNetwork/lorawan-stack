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
import ReactAce from 'react-ace'
import classnames from 'classnames'
import bind from 'autobind-decorator'

import PropTypes from '../../lib/prop-types'

import 'brace/mode/javascript'
import 'brace/mode/json'
import './ttn-theme'

import style from './code-editor.styl'

@bind
class CodeEditor extends React.Component {
  constructor(props) {
    super(props)

    this.state = { focus: false }
    this.aceRef = React.createRef()
  }

  onFocus(evt) {
    const { onFocus } = this.props

    this.setState({ focus: true }, function() {
      if (onFocus) {
        onFocus(evt)
      }
    })
  }

  onBlur(evt) {
    const { onBlur } = this.props

    this.setState({ focus: false }, function() {
      if (onBlur) {
        onBlur(evt)
      }
    })
  }

  componentDidUpdate({ value }) {
    const { value: oldValue, scrollToBottom } = this.props

    if (scrollToBottom && value !== oldValue) {
      const row = this.aceRef.current.editor.session.getLength()
      this.aceRef.current.editor.gotoLine(row)
    }
  }

  render() {
    const {
      className,
      language = 'javascript',
      name,
      onChange,
      value,
      placeholder,
      readOnly = false,
      editorOptions,
      height = '30rem',
      showGutter = true,
      minLines = 1,
      maxLines = Infinity,
      commands,
    } = this.props

    const { focus } = this.state

    const empty = !value || value === ''
    const currentValue = empty && !focus ? placeholder : value

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
      <div className={editorCls}>
        <ReactAce
          // rendered options
          theme="ttn"
          minLines={minLines}
          maxLines={maxLines}
          // session options
          mode={language}
          // editor options
          readOnly={readOnly}
          highlightActiveLine
          showGutter={showGutter}
          // other
          name={name}
          onChange={onChange}
          value={currentValue}
          defaultValue={placeholder}
          setOptions={options}
          width="100%"
          height={height}
          onFocus={this.onFocus}
          onBlur={this.onBlur}
          editorProps={{ $blockScrolling: Infinity }}
          commands={commands}
          ref={this.aceRef}
        />
      </div>
    )
  }
}

CodeEditor.propTypes = {
  /** The language to highlight */
  language: PropTypes.oneOf(['javascript', 'json']),
  /** The name of the editor (should be unique) */
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  onFocus: PropTypes.func,
  onBlur: PropTypes.func,
  /** The current value of the editor */
  value: PropTypes.string,
  /** The default value of the editor */
  placeholder: PropTypes.string,
  /** A flag identifying whether the editor is editable */
  readOnly: PropTypes.bool,
  /** see https://github.com/ajaxorg/ace/wiki/Configuring-Ace */
  editorOptions: PropTypes.object,
  /** The height of the editor */
  height: PropTypes.string,
  /** A flag indicating whether the editor should scroll to the bottom when the
   * value has been updated, useful for logging use cases.
   */
  scrollToBottom: PropTypes.bool,
  /** A flag identifying whether  */
  showGutter: PropTypes.bool,
  /** Minimum lines of code allowed */
  minLines: PropTypes.number,
  /** Maximum lines of code allowed */
  maxLines: PropTypes.number,
  /** New commands to add to the editor, see official docs */
  commands: PropTypes.array,
}

export default CodeEditor
