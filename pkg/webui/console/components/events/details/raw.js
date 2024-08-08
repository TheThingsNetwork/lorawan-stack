// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import classNames from 'classnames'

import CodeEditor from '@ttn-lw/components/code-editor'

import PropTypes from '@ttn-lw/lib/prop-types'

const RawEventDetails = React.memo(props => {
  const { className, id, details, darkTheme } = props

  const formattedDetails = JSON.stringify(details, null, 2)

  return (
    <CodeEditor
      className={classNames(className, 'br-0 border-none')}
      readOnly
      name={id}
      language="json"
      value={formattedDetails}
      darkTheme={darkTheme}
    />
  )
})

RawEventDetails.propTypes = {
  className: PropTypes.string,
  darkTheme: PropTypes.bool,
  details: PropTypes.shape({}).isRequired,
  id: PropTypes.string.isRequired,
}

RawEventDetails.defaultProps = {
  className: undefined,
  darkTheme: false,
}

export default RawEventDetails
