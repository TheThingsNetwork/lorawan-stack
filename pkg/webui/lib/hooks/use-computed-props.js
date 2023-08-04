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

/**
 * `useComputedProps` is a hook that can be used to pass-in props that are
 * derived from the other props of the components. This is useful to ensure that
 * expensive prop computations only need to be done once upon prop changes.
 *
 * @param {Function} computeProps - The function that returns the computed props.
 * @param {object} props - The props that are used to compute the computed props.
 * @returns {object} - The computed props.
 */
const useComputedProps = (computeProps, props) => {
  const computedProps = React.useMemo(() => computeProps(props), [computeProps, props])
  return computedProps
}

export default useComputedProps
