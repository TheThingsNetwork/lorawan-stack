// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useContext, useEffect, useState } from 'react'

import PropTypes from '@ttn-lw/lib/prop-types'

const BreadcrumbsContext = React.createContext({
  add: () => {},
  remove: () => {},
  breadcrumbs: [],
})
const { Provider, Consumer } = BreadcrumbsContext

const BreadcrumbsProvider = ({ children }) => {
  const [breadcrumbs, setBreadcrumbs] = useState([])

  const add = (id, breadcrumbs) => {
    setBreadcrumbs(prev => {
      const index = prev.findIndex(({ id: breadcrumbId }) => breadcrumbId === id)
      if (index === -1) {
        return [...prev, ...breadcrumbs].sort((a, b) => (a.id < b.id ? -1 : 1))
      }

      // Replace breadcrumbs with existing id.
      return [...prev.slice(0, index), ...breadcrumbs, ...prev.slice(index + breadcrumbs.length)]
    })
  }

  const remove = id => {
    setBreadcrumbs(prev => prev.filter(b => b.id !== id))
  }

  const value = {
    add,
    remove,
    breadcrumbs,
  }

  return <Provider value={value}>{children}</Provider>
}

BreadcrumbsProvider.propTypes = {
  children: PropTypes.node.isRequired,
}

const useBreadcrumbs = (id, breadcrumbs) => {
  const context = useContext(BreadcrumbsContext)

  useEffect(() => {
    context.add(
      id,
      breadcrumbs.map(b => ({ ...b, id })),
    )
    return () => {
      context.remove(id)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])
}

export { Consumer as BreadcrumbsConsumer, BreadcrumbsProvider, BreadcrumbsContext, useBreadcrumbs }
