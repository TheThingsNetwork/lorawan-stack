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
import { withRouter } from 'react-router-dom'

import PropTypes from '../../lib/prop-types'

@withRouter
class ErrorView extends React.Component {
  static propTypes = {
    ErrorComponent: PropTypes.func.isRequired,
    children: PropTypes.node.isRequired,
    history: PropTypes.history.isRequired,
  }

  state = {
    error: undefined,
    hasCaught: false,
  }

  unlisten = () => null

  componentWillUnmount() {
    this.unlisten()
  }

  componentDidCatch(error) {
    this.setState({
      hasCaught: true,
      error,
    })

    // Clear the error when the route changes (e.g. user clicking a link)
    const { history } = this.props
    this.unlisten = history.listen((location, action) => {
      if (this.state.hasCaught) {
        this.setState({ hasCaught: false, error: undefined })
        this.unlisten()
      }
    })
  }

  render() {
    const { children, ErrorComponent } = this.props
    const { hasCaught, error } = this.state

    if (hasCaught) {
      return <ErrorComponent error={error} />
    }

    return React.Children.only(children)
  }
}

export default ErrorView
