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

import React, { useContext, useEffect } from 'react'
import bind from 'autobind-decorator'

import PropTypes from '@ttn-lw/lib/prop-types'

const BreadcrumbsContext = React.createContext()
const { Provider, Consumer } = BreadcrumbsContext

class BreadcrumbsProvider extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
  }

  state = {
    breadcrumbs: [],
  }

  @bind
  add(id, breadcrumb) {
    this.setState(prev => {
      const index = prev.breadcrumbs.findIndex(({ id: breadcrumbId }) => breadcrumbId === id)
      if (index === -1) {
        return {
          breadcrumbs: [...prev.breadcrumbs, { id, breadcrumb }].sort((a, b) =>
            a.id < b.id ? -1 : 1,
          ),
        }
      }

      // Replace breadcrumb with existing id.
      return {
        breadcrumbs: [
          ...prev.breadcrumbs.slice(0, index),
          { id, breadcrumb },
          ...prev.breadcrumbs.slice(index + 1),
        ],
      }
    })
  }

  @bind
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

const withBreadcrumb = (id, element) => Component => {
  class BreadcrumbsConsumer extends React.Component {
    static propTypes = {
      add: PropTypes.func.isRequired,
      breadcrumb: PropTypes.oneOfType([PropTypes.func, PropTypes.element]).isRequired,
      remove: PropTypes.func.isRequired,
    }

    constructor(props) {
      super(props)

      this.add()
    }

    add() {
      const { add, breadcrumb } = this.props

      add(id, breadcrumb)
    }

    remove() {
      const { remove } = this.props

      remove(id)
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

const useBreadcrumbs = (id, element) => {
  const context = useContext(BreadcrumbsContext)

  useEffect(() => {
    context.add(id, element)
    return () => {
      context.remove(id)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])
}

export {
  Consumer as BreadcrumbsConsumer,
  BreadcrumbsProvider,
  withBreadcrumb,
  BreadcrumbsContext,
  useBreadcrumbs,
}
