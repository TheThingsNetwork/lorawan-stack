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
import bind from 'autobind-decorator'

const { Provider, Consumer } = React.createContext()

@bind
class SideNavigationProvider extends React.Component {
  state = {
    header: undefined,
    entries: undefined,
  }

  set(entries, header) {
    this.setState({ header, entries })
  }

  remove() {
    this.setState({ entries: undefined, header: undefined })
  }

  render() {
    const { children } = this.props
    const { header, entries } = this.state
    const value = {
      set: this.set,
      remove: this.remove,
      header,
      entries,
    }

    return <Provider value={value}>{children}</Provider>
  }
}

const withSideNavigation = selectData =>
  function(Component) {
    @bind
    class WithSideNavigation extends React.Component {
      constructor(props) {
        super(props)

        const { set } = props
        const { header, entries } = selectData(props)

        set(entries, header)
      }

      componentWillUnmount() {
        const { remove } = this.props

        remove()
      }

      render() {
        const { set, remove, ...rest } = this.props

        return <Component {...rest} />
      }
    }

    const withSideNavigationWrapper = props => (
      <Consumer>
        {({ set, remove }) => <WithSideNavigation {...props} set={set} remove={remove} />}
      </Consumer>
    )

    return withSideNavigationWrapper
  }

export { Consumer as SideNavigationConsumer, SideNavigationProvider, withSideNavigation }
