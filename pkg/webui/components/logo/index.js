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
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './logo.styl'

const Logo = ({ className, logo, miniLogo, unlockSize }) => {
  const classname = classnames(style.container, className, {
    [style.unlockSize]: unlockSize,
  })

  return (
    <div className={classname}>
      <img {...logo} />
      <img {...miniLogo} />
    </div>
  )
}

const imgPropType = PropTypes.shape({
  src: PropTypes.string.isRequired,
  alt: PropTypes.string.isRequired,
})

Logo.propTypes = {
  brandLogo: imgPropType,
  className: PropTypes.string,
  logo: imgPropType.isRequired,
  miniLogo: imgPropType.isRequired,
  safe: PropTypes.bool,
  unlockSize: PropTypes.bool,
  vertical: PropTypes.bool,
}

Logo.defaultProps = {
  className: undefined,
  brandLogo: undefined,
  vertical: false,
  unlockSize: false,
  safe: false,
}

export default Logo
