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

import Spinner from '../../components/spinner'

/**
 * `withRequest` is a HOC that handles:
 *   * Requesting data on initial mount using the `loadData` prop.
 *   * Showing the loading spinner while the request in is progress using the `isFetchingTest` predicate.
 *   * Throwing an error received as the `error` prop.
 * @param {Function} mapPropsToRequest - Selects the `request` given the wrapped component props.
 * @param {Function} mapPropsToFetching - Selects the `fetching` value given the wrapped component props.
 * If evaluates to `true`, then the loading spinner is rendered, otherwise renders the wrapped component.
 * @param {Function} mapPropsToError - Selects the `error` value given the wrapped component props.
 * @returns {Function} - An instance of the `withRequest` HOC.
 */
const withRequest = (
  mapPropsToRequest,
  mapPropsToFetching = ({ fetching } = {}) => fetching,
  mapPropsToError = ({ error } = {}) => error,
) => Component =>
  class WithRequest extends React.Component {
    state = { initialFetching: true }
    componentDidMount() {
      mapPropsToRequest(this.props)
    }

    componentDidUpdate(prevProps) {
      const prevFetching = mapPropsToFetching(prevProps)
      const fetching = mapPropsToFetching(this.props)
      const { initialFetching } = this.state

      // Avoid initial render with old data (when request has been performed
      // before, thus fetching calculation being initially true, before the new
      // request is being made).
      if (initialFetching && prevFetching !== !fetching) {
        // Remove internal fetching state as soon as the fetching calculation
        // has switched.
        this.setState({ initialFetching: false })
      }

      const error = mapPropsToError(this.props)
      const prevError = mapPropsToError(prevProps)

      // Check for errors only after component mounts and makes the request.
      if (Boolean(error) && prevError !== error) {
        throw error
      }
    }

    render() {
      const { initialFetching } = this.state
      if (initialFetching || mapPropsToFetching(this.props)) {
        return <Spinner center />
      }

      return <Component {...this.props} />
    }
  }

export default withRequest
