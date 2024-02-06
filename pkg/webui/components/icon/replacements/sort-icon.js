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
    viewBox="0 0 21 21"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
    ref={ref}
  >
    <path
      d="m14.919 6.796-3.8386-3.8385c-0.1209-0.1209-0.2469-0.20654-0.3778-0.25691-0.131-0.05037-0.2872-0.07556-0.4685-0.07556-0.1411 0-0.28213 0.02519-0.42318 0.07556s-0.27203 0.13601-0.39293 0.25691l-3.8386 3.8385c-0.38286 0.38285-0.47353 0.8211-0.27203 1.3148 0.2015 0.49367 0.56421 0.7405 1.0881 0.7405h7.6772c0.544 0 0.9168-0.24683 1.1183-0.7405 0.2015-0.49366 0.1108-0.93191-0.272-1.3148zm-9.3394 6.9082 3.8386 3.8385c0.1209 0.1209 0.24684 0.2065 0.37781 0.2569 0.13098 0.0504 0.28713 0.0755 0.46853 0.0755 0.141 0 0.2821-0.0251 0.4231-0.0755 0.1411-0.0504 0.272-0.136 0.3929-0.2569l3.8387-3.8385c0.3828-0.3829 0.4735-0.8211 0.272-1.3148-0.2015-0.4936-0.5642-0.7405-1.0881-0.7405h-7.6772c-0.54406 0-0.91684 0.2469-1.1183 0.7405-0.2015 0.4937-0.11083 0.9319 0.27203 1.3148z"
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
