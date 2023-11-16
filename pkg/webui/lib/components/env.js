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

import React, { useContext, useEffect } from 'react'

import PropTypes from '@ttn-lw/lib/prop-types'
import { warn } from '@ttn-lw/lib/log'

const useEnv = () => {
  const context = useContext(EnvProviderContext)

  useEffect(() => {
    if (!context.env) {
      warn('No env in context')
    }
  }, [context.env])

  return { env: context.env }
}

export const EnvProviderContext = React.createContext()

export const EnvProvider = ({ children, env }) => (
  <EnvProviderContext.Provider value={{ env }}>{children}</EnvProviderContext.Provider>
)

EnvProvider.propTypes = {
  children: PropTypes.node.isRequired,
  env: PropTypes.env.isRequired,
}

export default useEnv
