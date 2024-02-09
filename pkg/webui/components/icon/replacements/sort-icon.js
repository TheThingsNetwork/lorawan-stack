// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

const SortIcon = React.forwardRef(({ className }, ref) => (
  <svg
    fill="none"
    viewBox="0 0 20 20"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
    ref={ref}
  >
    <path
      d="m14.453 6.7123-3.6558-3.6557c-0.1152-0.11514-0.2352-0.19671-0.3598-0.24468-0.1248-0.04797-0.2735-0.07196-0.44621-0.07196-0.13438 0-0.2687 0.02399-0.40303 0.07196s-0.25907 0.12954-0.37422 0.24468l-3.6558 3.6557c-0.36463 0.36462-0.45098 0.782-0.25907 1.2522 0.1919 0.47016 0.53734 0.70524 1.0363 0.70524h7.3117c0.5181 0 0.8731-0.23508 1.065-0.70524 0.1919-0.47015 0.1056-0.88753-0.259-1.2522zm-8.8947 6.5793 3.6558 3.6557c0.11515 0.1151 0.23509 0.1967 0.35982 0.2447 0.12474 0.048 0.27346 0.0719 0.44624 0.0719 0.1343 0 0.2686-0.0239 0.4029-0.0719 0.1344-0.048 0.2591-0.1296 0.3742-0.2447l3.6559-3.6557c0.3646-0.3647 0.451-0.782 0.2591-1.2522-0.1919-0.4701-0.5374-0.7052-1.0363-0.7052h-7.3116c-0.51815 0-0.87318 0.2351-1.065 0.7052-0.1919 0.4702-0.10562 0.8875 0.25901 1.2522z"
      clipRule="evenodd"
      fill="currentColor"
      fillRule="evenodd"
    />
  </svg>
))

SortIcon.propTypes = {
  className: PropTypes.string,
}

SortIcon.defaultProps = {
  className: undefined,
}

export default SortIcon
