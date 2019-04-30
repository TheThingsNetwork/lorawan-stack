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

import originalPropTypes from 'prop-types'

const PropTypes = { ...originalPropTypes }

PropTypes.message = PropTypes.oneOfType([
  PropTypes.shape({
    id: PropTypes.string.isRequired,
    value: PropTypes.object,
    defaultMessage: PropTypes.string,
  }),
  PropTypes.string,
  PropTypes.element,
])

PropTypes.error = PropTypes.oneOfType([
  PropTypes.shape({
    details: PropTypes.array.isRequired,
    message: PropTypes.string.isRequired,
    code: PropTypes.number.isRequired,
  }),
  PropTypes.message,
  PropTypes.string,
])

PropTypes.link = PropTypes.shape({
  title: PropTypes.message.isRequired,
  icon: PropTypes.string,
  path: PropTypes.string.isRequired,
  exact: PropTypes.bool,
})

PropTypes.event = PropTypes.shape({
  name: PropTypes.string.isRequired,
  time: PropTypes.string.isRequired,
  identifiers: PropTypes.array.isRequired,
  data: PropTypes.object,
})

export default PropTypes
