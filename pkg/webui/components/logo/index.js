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

import PropTypes from '../../lib/prop-types'
import Link from '../link'

import style from './logo.styl'

const Logo = function({ className, logo, secondaryLogo, vertical }) {
  const classname = classnames(style.container, className, {
    [style.vertical]: vertical,
    [style.customBranding]: Boolean(secondaryLogo),
  })

  return (
    <div className={classname}>
      <div className={style.logo}>
        <Link className={style.logoContainer} to="/">
          <img {...logo} />
        </Link>
      </div>
      {Boolean(secondaryLogo) && (
        <div className={style.secondaryLogo}>
          <div id="secondary-logo" className={style.secondaryLogoContainer}>
            <img {...secondaryLogo} />
          </div>
        </div>
      )}
    </div>
  )
}

const imgPropType = PropTypes.shape({
  src: PropTypes.string.isRequired,
  alt: PropTypes.string.isRequired,
})

Logo.propTypes = {
  className: PropTypes.string,
  logo: imgPropType.isRequired,
  secondaryLogo: imgPropType,
  vertical: PropTypes.bool,
}

Logo.defaultProps = {
  className: undefined,
  secondaryLogo: undefined,
  vertical: false,
}

export default Logo
