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
import Message from '../../lib/components/message'

import sharedMessages from '../../lib/shared-messages'

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
    constructor(props) {
      super(props)
      // Avoid render of old content by setting an initial fetching state if
      // the component is mounted with fetching prop evaluating to false.
      // This way we can close the "fetching gap" between the initial render
      // and the next render after the request action has been dispatched.
      this.state = {
        initialFetching: mapPropsToFetching(props) === false,
      }
    }
    componentDidMount() {
      const { initialFetching } = this.state
      mapPropsToRequest(this.props)

      if (initialFetching) {
        this.setState({ initialFetching: false })
      }
    }

    componentDidUpdate(prevProps) {
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
        return (
          <Spinner center>
            <Message content={sharedMessages.fetching} />
          </Spinner>
        )
      }

      return <Component {...this.props} />
    }
  }

export default withRequest
