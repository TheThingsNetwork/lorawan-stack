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
 * `withComputedProps` is a HOC that can be used to pass-in props that are
 * derived from the other props of the components. This is useful to ensure that
 * expensive prop computations only need to be done once upon prop changes.
 * @param {Function} computeProps - The function that returns the computed props.
 * @returns {Function} - An instance of the `computeProps` HOC.
 */
export default computeProps => Component =>
  class withComputedProps extends React.Component {
    render() {
      return <Component {...computeProps(this.props)} />
    }
  }
