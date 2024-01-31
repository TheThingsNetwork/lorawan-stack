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

import PropTypes from '@ttn-lw/lib/prop-types'

const StarIcon = ({ className }) => (
  <svg fill="none" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg" className={className}>
    <mask
      id="a"
      x="0"
      y="0"
      width="20"
      height="20"
      style={{ maskType: 'alpha' }}
      maskUnits="userSpaceOnUse"
    >
      <rect width="20" height="20" fill="#D9D9D9" />
    </mask>
    <g mask="url(#a)">
      <path
        d="m7.2736 14.313 2.7266-1.6152 2.7488 1.6152-0.7298-3.0589 2.367-2.0243-3.1304-0.27654-1.2556-2.9149-1.2555 2.9202-3.1305 0.27126 2.3892 2.019-0.72979 3.0642zm2.7266 0.52-3.638 2.175c-0.18661 0.1024-0.36954 0.1454-0.54879 0.1289-0.17926-0.0165-0.33658-0.0759-0.47199-0.1784-0.1354-0.1024-0.23694-0.2423-0.30463-0.4198-0.0677-0.1775-0.07595-0.3596-0.02474-0.5462l0.95499-4.0576-3.2183-2.7324c-0.15365-0.13539-0.24784-0.29184-0.28257-0.46934-0.03474-0.17748-0.02825-0.35041 0.01946-0.51878 0.04769-0.16837 0.14187-0.30921 0.28254-0.42253 0.14069-0.1133 0.3117-0.17731 0.51302-0.19204l4.2122-0.37546 1.667-3.8641c0.08419-0.19013 0.20223-0.33097 0.35411-0.42251 0.15187-0.09155 0.31376-0.13733 0.48563-0.13733 0.1719 0 0.3338 0.04578 0.4857 0.13733 0.1519 0.09154 0.2699 0.23238 0.3541 0.42251l1.667 3.8862 4.2122 0.35338c0.2013 0.01473 0.3724 0.08242 0.513 0.20308 0.1407 0.12068 0.2349 0.26521 0.2826 0.43358s0.0505 0.33762 0.0084 0.50774c-0.0421 0.17013-0.14 0.3229-0.2936 0.45829l-3.1962 2.7324 0.9549 4.0576c0.0512 0.1866 0.043 0.3687-0.0247 0.5462s-0.1692 0.3174-0.3046 0.4198c-0.1354 0.1025-0.2928 0.1619-0.472 0.1784-0.1793 0.0165-0.3622-0.0265-0.5488-0.1289l-3.638-2.175z"
        fill="currentColor"
      />
    </g>
  </svg>
)

StarIcon.propTypes = {
  className: PropTypes.string,
}

StarIcon.defaultProps = {
  className: undefined,
}

export default StarIcon
