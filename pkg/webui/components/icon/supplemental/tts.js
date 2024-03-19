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

const TtsIcon = React.forwardRef(({ className }, ref) => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 20 20"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
    ref={ref}
  >
    <path
      fillRule="evenodd"
      clipRule="evenodd"
      d="M11.688 7.18194C11.148 7.44535 10.9237 8.09669 11.1871 8.63674L13.718 13.8259C13.9814 14.3659 14.6327 14.5902 15.1728 14.3268L19.0843 12.419C19.6243 12.1556 19.8486 11.5043 19.5852 10.9642L17.0543 5.77509C16.7909 5.23504 16.1396 5.01077 15.5995 5.27418L11.688 7.18194ZM8.14838 7.18194C7.60833 7.44535 7.38405 8.09669 7.64745 8.63674L10.1783 13.8259C10.4418 14.3659 11.0931 14.5902 11.6331 14.3268L12.0646 14.1163C12.5784 13.8658 12.7918 13.2461 12.5412 12.7324L10.6673 8.8904C10.2638 8.06309 10.6074 7.0653 11.4347 6.66178L13.0716 5.86343C13.2787 5.76244 13.3531 5.49613 13.1648 5.36334C12.851 5.14203 12.4294 5.09395 12.0599 5.27418L8.14838 7.18194ZM4.08013 8.63674C3.81674 8.09669 4.04101 7.44535 4.58108 7.18195L8.49258 5.27418C8.86767 5.09124 9.29642 5.14353 9.61161 5.37347C9.7963 5.5082 9.72072 5.77166 9.51525 5.87188L7.89376 6.66273C7.06642 7.06625 6.72286 8.06402 7.12637 8.89135L8.99478 12.7222C9.24536 13.2359 9.03201 13.8556 8.51824 14.1061L8.06585 14.3268C7.52578 14.5902 6.87445 14.3659 6.61105 13.8259L4.08013 8.63674ZM1.02779 7.18194C0.487733 7.44535 0.263456 8.09669 0.526856 8.63674L3.05777 13.8259C3.32117 14.3659 3.9725 14.5902 4.51257 14.3268L4.95532 14.1108C5.46909 13.8603 5.68244 13.2406 5.43186 12.7269L3.56024 8.88945C3.15673 8.06214 3.50029 7.06435 4.32762 6.66083L5.95551 5.86686C6.16192 5.76619 6.23682 5.50105 6.05002 5.36746C5.73563 5.14263 5.3111 5.09284 4.9393 5.27418L1.02779 7.18194Z"
      fill="currentColor"
    />
  </svg>
))

TtsIcon.propTypes = {
  className: PropTypes.string,
}

TtsIcon.defaultProps = {
  className: undefined,
}

export default TtsIcon
