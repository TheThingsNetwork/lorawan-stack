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

import React, { useCallback } from 'react'
import { useSelector } from 'react-redux'
import { Navigate } from 'react-router-dom'

import toast from '@ttn-lw/components/toast'

import PropTypes from '../../../lib/prop-types'

const Require = ({ children, featureCheck, condition, otherwise }) => {
  const rights = useSelector(state => featureCheck?.rightsSelector(state))
  const newCondition = condition || (Boolean(featureCheck) && featureCheck.check(rights))

  const alternativeRender = useCallback(() => {
    if (typeof otherwise === 'object') {
      const { render, redirect, message } = otherwise

      if (message) {
        toast({
          type: toast.types.WARNING,
          message,
        })
      }

      if (typeof redirect === 'string') {
        return <Navigate to={redirect} />
      } else if (typeof render === 'function') {
        return render()
      }
    }

    return null
  }, [otherwise])

  if (!newCondition) {
    return alternativeRender()
  }

  return children
}

Require.propTypes = {
  children: PropTypes.node.isRequired,
  otherwise: PropTypes.shape({
    redirect: PropTypes.oneOfType([PropTypes.string, PropTypes.func]),
    render: PropTypes.func,
    message: PropTypes.message,
  }),
}
Require.defaultProps = {
  otherwise: undefined,
  condition: false,
}

export default Require
