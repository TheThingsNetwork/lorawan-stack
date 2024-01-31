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

const PlusIcon = ({ className }) => (
  <svg fill="none" viewBox="0 0 21 21" xmlns="http://www.w3.org/2000/svg" className={className}>
    <mask
      id="a"
      x="0"
      y="0"
      width="21"
      height="21"
      style={{ maskType: 'alpha' }}
      maskUnits="userSpaceOnUse"
    >
      <rect x=".66992" y=".97705" width="20" height="20" fill="#D9D9D9" />
    </mask>
    <g mask="url(#a)">
      <path
        d="m9.4708 12.176h-4.7716c-0.33621 0-0.62013-0.1152-0.85176-0.3457-0.23163-0.2304-0.34745-0.5129-0.34745-0.8474s0.11582-0.619 0.34745-0.8535 0.51555-0.35176 0.85176-0.35176h4.7716v-4.7716c0-0.33621 0.11522-0.62014 0.34568-0.85177 0.2305-0.23163 0.5129-0.34744 0.8474-0.34744s0.619 0.11581 0.8535 0.34744 0.3518 0.51556 0.3518 0.85177v4.7716h4.7716c0.3362 0 0.6201 0.11523 0.8518 0.34566 0.2316 0.2305 0.3474 0.513 0.3474 0.8474 0 0.3345-0.1158 0.619-0.3474 0.8535-0.2317 0.2346-0.5156 0.3518-0.8518 0.3518h-4.7716v4.7716c0 0.3362-0.1152 0.6202-0.3457 0.8518s-0.5129 0.3474-0.8474 0.3474-0.619-0.1158-0.8535-0.3474c-0.23452-0.2316-0.35178-0.5156-0.35178-0.8518v-4.7716z"
        fill="currentColor"
      />
    </g>
  </svg>
)

PlusIcon.propTypes = {
  className: PropTypes.string,
}

PlusIcon.defaultProps = {
  className: undefined,
}

export default PlusIcon
