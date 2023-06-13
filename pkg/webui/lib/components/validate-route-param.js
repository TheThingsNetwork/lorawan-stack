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

import React, { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import PropTypes from '../prop-types'

import GenericNotFound from './full-view-error/not-found'

const render = ({ Component, element }) => {
  if (Component) {
    return <Component />
  } else if (element) {
    return element
  }

  return null
}

const checkParams = (check, params) =>
  Object.keys(check).every(paramKey => {
    const singleCheck = check[paramKey]
    if (singleCheck instanceof RegExp) {
      return singleCheck.test(params[paramKey])
    } else if (typeof singleCheck === 'function') {
      return singleCheck(params)
    }
    return false
  })

const ValidateRouteParam = ({ check, Component, element, otherwise }) => {
  const params = useParams()
  const navigate = useNavigate()

  const shouldRender = checkParams(check, params)

  useEffect(() => {
    if (!shouldRender && otherwise.redirect) {
      navigate(otherwise.redirect)
    }
  }, [shouldRender, navigate, otherwise.redirect])

  if (shouldRender) {
    return render({ Component, element })
  }

  if (otherwise.render) {
    return otherwise.render()
  }
}

ValidateRouteParam.propTypes = {
  Component: PropTypes.func,
  check: PropTypes.objectOf(PropTypes.oneOfType([PropTypes.func, PropTypes.instanceOf(RegExp)]))
    .isRequired,
  element: PropTypes.node,
  otherwise: PropTypes.shape({
    render: PropTypes.func,
    redirect: PropTypes.string,
  }),
}

ValidateRouteParam.defaultProps = {
  Component: undefined,
  element: undefined,
  otherwise: {
    render: () => <GenericNotFound />,
  },
}

export default ValidateRouteParam
