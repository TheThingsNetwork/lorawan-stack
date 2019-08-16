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
class BreadcrumbsProvider extends React.Component {
  state = {
    breadcrumbs: [],
  }

  add(id, breadcrumb) {
    this.setState(prev => ({
      breadcrumbs: [...prev.breadcrumbs, { id, breadcrumb }],
    }))
  }

  remove(id) {
    this.setState(prev => ({
      breadcrumbs: prev.breadcrumbs.filter(b => b.id !== id),
    }))
  }

  render() {
    const { children } = this.props
    const value = {
      add: this.add,
      remove: this.remove,
      breadcrumbs: this.state.breadcrumbs.map(b => b.breadcrumb),
    }

    return <Provider value={value}>{children}</Provider>
  }
}

const withBreadcrumb = (id, element) =>
  function(Component) {
    class BreadcrumbsConsumer extends React.Component {
      add() {
        const { add, breadcrumb } = this.props

        add(id, breadcrumb)
      }

      remove() {
        const { remove } = this.props

        remove(id)
      }

      componentDidMount() {
        this.add()
      }

      componentWillUnmount() {
        this.remove()
      }

      render() {
        const { add, remove, breadcrumb, ...rest } = this.props

        return <Component {...rest} />
      }
    }

    const BreadcrumbsConsumerContainer = props => (
      <Consumer>
        {({ add, remove }) => (
          <BreadcrumbsConsumer {...props} add={add} remove={remove} breadcrumb={element(props)} />
        )}
      </Consumer>
    )

    return BreadcrumbsConsumerContainer
  }

withBreadcrumb.displayName = 'withBreadcrumb'

export { Consumer as BreadcrumbsConsumer, BreadcrumbsProvider, withBreadcrumb }
