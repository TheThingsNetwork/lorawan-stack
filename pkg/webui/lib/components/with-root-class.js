// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import PropTypes from '@ttn-lw/lib/prop-types'

import useRootClass from '../hooks/use-root-class'

const WithRootClass = ({ children, className, id }) => {
  useRootClass(className, id)

  return children
}

WithRootClass.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  id: PropTypes.string,
}

WithRootClass.defaultProps = {
  className: undefined,
  id: 'app',
}

export default WithRootClass
