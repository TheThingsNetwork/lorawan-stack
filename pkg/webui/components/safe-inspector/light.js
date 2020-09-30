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
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'

import chunkArray from './chunk-array'

import style from './safe-inspector.styl'

const SafeInspectorLight = React.memo(({ small, isBytes, data, className }) => {
  const containerStyle = classnames(className, style.container, {
    [style.containerSmall]: small,
  })

  const formattedData = isBytes ? data.toUpperCase() : data
  let display = formattedData

  if (isBytes) {
    const chunks = chunkArray(data.toUpperCase().split(''), 2)
    display = chunks.map((chunk, index) => <span key={`${data}_chunk_${index}`}>{chunk}</span>)
  }

  return (
    <div className={containerStyle}>
      <div className={style.data}>{display}</div>
    </div>
  )
})

SafeInspectorLight.propTypes = {
  /** The classname to be applied. */
  className: PropTypes.string,
  /** The data to be displayed. */
  data: PropTypes.string.isRequired,
  /** Whether the data is in byte format. */
  isBytes: PropTypes.bool,
  /**
   * Whether a smaller style should be rendered (useful for display in
   * tables).
   */
  small: PropTypes.bool,
}

SafeInspectorLight.defaultProps = {
  className: undefined,
  isBytes: true,
  small: false,
}

export default SafeInspectorLight
