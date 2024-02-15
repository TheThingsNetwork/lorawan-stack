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

const PlusIcon = React.forwardRef(({ className }, ref) => (
  <svg
    fill="none"
    viewBox="0 0 20 20"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
    ref={ref}
  >
    <path
      d="m8.8009 11.199h-4.7716c-0.33621 0-0.62014-0.1153-0.85177-0.3457-0.23163-0.2305-0.34744-0.513-0.34744-0.8474 0-0.33451 0.11581-0.61901 0.34744-0.85353s0.51556-0.35178 0.85177-0.35178h4.7716v-4.7716c0-0.33621 0.11523-0.62014 0.34569-0.85177 0.23047-0.23163 0.51294-0.34744 0.8474-0.34744 0.33452 0 0.61902 0.11581 0.85352 0.34744s0.3518 0.51556 0.3518 0.85177v4.7716h4.7716c0.3362 0 0.6201 0.11523 0.8517 0.34569 0.2317 0.23047 0.3475 0.51294 0.3475 0.8474 0 0.33452-0.1158 0.61902-0.3475 0.85352-0.2316 0.2345-0.5155 0.3518-0.8517 0.3518h-4.7716v4.7716c0 0.3362-0.1153 0.6201-0.3457 0.8517-0.2305 0.2317-0.513 0.3475-0.8474 0.3475-0.33451 0-0.61901-0.1158-0.85353-0.3475-0.23452-0.2316-0.35178-0.5155-0.35178-0.8517v-4.7716z"
      fill="currentColor"
    />
  </svg>
))

PlusIcon.propTypes = {
  className: PropTypes.string,
}

PlusIcon.defaultProps = {
  className: undefined,
}

export default PlusIcon
