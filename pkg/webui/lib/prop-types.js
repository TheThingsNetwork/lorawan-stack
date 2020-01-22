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
import { components } from '../constants/components'

const PropTypes = { ...originalPropTypes }

PropTypes.formatters = PropTypes.shape({
  up_formatter: PropTypes.string,
  up_formatter_parameter: PropTypes.string,
  down_formatter: PropTypes.string,
  down_formatter_parameter: PropTypes.string,
})

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
  PropTypes.oneOfType([
    PropTypes.shape({
      details: PropTypes.array.isRequired,
      message: PropTypes.string.isRequired,
      code: PropTypes.number.isRequired,
    }),
    PropTypes.shape({
      details: PropTypes.array.isRequired,
      message: PropTypes.string.isRequired,
      grpc_code: PropTypes.number.isRequired,
    }),
  ]),
  PropTypes.message,
  PropTypes.string,
  PropTypes.shape({
    message: PropTypes.string,
    stack: PropTypes.object,
  }),
  PropTypes.instanceOf(Error),
])

PropTypes.link = PropTypes.shape({
  title: PropTypes.message.isRequired,
  icon: PropTypes.string,
  path: PropTypes.string.isRequired,
  exact: PropTypes.bool,
  hidden: PropTypes.bool,
})

// Entities and entity-related prop-types

PropTypes.event = PropTypes.shape({
  name: PropTypes.string.isRequired,
  time: PropTypes.string.isRequired,
  identifiers: PropTypes.array.isRequired,
  data: PropTypes.object,
})
PropTypes.events = PropTypes.arrayOf(PropTypes.event)

PropTypes.gateway = PropTypes.shape({
  antennas: PropTypes.Array,
  ids: PropTypes.shape({
    gateway_id: PropTypes.string.isRequired,
  }).isRequired,
  name: PropTypes.string,
  description: PropTypes.string,
  created_at: PropTypes.string,
  updated_at: PropTypes.string,
  frequency_plan_id: PropTypes.string.isRequired,
  gateway_server_address: PropTypes.string,
})

PropTypes.gatewayStats = PropTypes.shape({
  connected_at: PropTypes.string.isRequired,
  last_uplink_received_at: PropTypes.string,
  protocol: PropTypes.string,
  uplink_count: PropTypes.string,
  downlink_count: PropTypes.string,
  round_trip_times: PropTypes.shape({}),
  last_status_received_at: PropTypes.oneOfType([PropTypes.string, PropTypes.instanceOf(Date)]),
})

PropTypes.application = PropTypes.shape({
  created_at: PropTypes.string.isRequired,
  description: PropTypes.string,
  ids: PropTypes.shape({
    application_id: PropTypes.string.isRequired,
  }).isRequired,
  name: PropTypes.string,
  updated_at: PropTypes.string.isRequired,
})

PropTypes.pubsub = PropTypes.shape({
  ids: PropTypes.shape({
    pub_sub_id: PropTypes.string.isRequired,
    application_ids: PropTypes.shape({
      application_id: PropTypes.string,
    }),
  }).isRequired,
  created_at: PropTypes.string.isRequired,
  updated_at: PropTypes.string.isRequired,
  format: PropTypes.string,
  base_topic: PropTypes.string,
  nats: PropTypes.shape({
    server_url: PropTypes.string,
  }),
  mqtt: PropTypes.shape({
    server_url: PropTypes.string,
    client_id: PropTypes.string,
    username: PropTypes.string,
    password: PropTypes.string,
    subscribe_qos: PropTypes.string,
    publish_qos: PropTypes.string,
    use_tls: PropTypes.bool,
    tls_ca: PropTypes.string,
    tls_client_cert: PropTypes.string,
    tls_client_key: PropTypes.string,
  }),
})

PropTypes.user = PropTypes.shape({
  ids: PropTypes.shape({
    user_id: PropTypes.string.isRequired,
  }).isRequired,
})

PropTypes.stackComponent = PropTypes.shape({
  enabled: PropTypes.bool.isRequired,
  base_url: PropTypes.string,
})

PropTypes.env = PropTypes.shape({
  appRoot: PropTypes.string.isRequired,
  assetsRoot: PropTypes.string.isRequired,
  siteName: PropTypes.string.isRequired,
  siteTitle: PropTypes.string.isRequired,
  siteSubTitle: PropTypes.string,
  pageData: PropTypes.shape({}),
  config: PropTypes.shape({
    language: PropTypes.string,
    supportLink: PropTypes.string,
    stack: PropTypes.shape({
      is: PropTypes.stackComponent,
      as: PropTypes.stackComponent,
      ns: PropTypes.stackComponent,
      js: PropTypes.stackComponent,
      gs: PropTypes.stackComponent,
    }),
  }).isRequired,
})

PropTypes.device = PropTypes.shape({
  ids: PropTypes.shape({
    device_id: PropTypes.string.isRequired,
    application_ids: PropTypes.shape({
      application_id: PropTypes.string.isRequired,
    }),
  }).isRequired,
  name: PropTypes.string,
  created_at: PropTypes.string,
  updated_at: PropTypes.string,
  description: PropTypes.string,
  locations: PropTypes.shape({
    // User is an object containing latitude and longitude property of number.
    user: PropTypes.shape({
      latitude: PropTypes.number,
      longitude: PropTypes.number,
    }),
  }),
  lorawan_phy_version: PropTypes.string,
  lorawan_version: PropTypes.string,
  supports_join: PropTypes.bool,
  frequency_plan_id: PropTypes.string,
})

PropTypes.organization = PropTypes.shape({
  ids: PropTypes.shape({
    organization_id: PropTypes.string.isRequired,
  }),
  name: PropTypes.string,
  description: PropTypes.string,
  created_at: PropTypes.string,
  updated_at: PropTypes.string,
})

PropTypes.match = PropTypes.shape({
  path: PropTypes.string.isRequired,
  url: PropTypes.string.isRequired,
})

PropTypes.location = PropTypes.shape({
  hash: PropTypes.string,
  key: PropTypes.string,
  pathname: PropTypes.string.isRequired,
  search: PropTypes.string,
  state: PropTypes.shape({
    info: PropTypes.message,
  }),
})

PropTypes.history = PropTypes.shape({
  listen: PropTypes.func,
})

PropTypes.collaborator = PropTypes.shape({
  rights: PropTypes.rights,
})

PropTypes.apiKey = PropTypes.shape({
  id: PropTypes.string.isRequired,
  rights: PropTypes.rights,
})

PropTypes.right = PropTypes.string
PropTypes.rights = PropTypes.arrayOf(PropTypes.right)

PropTypes.component = PropTypes.oneOf(components)
PropTypes.components = PropTypes.arrayOf(PropTypes.component)

PropTypes.webhook = PropTypes.shape({
  base_url: PropTypes.string.isRequired,
  created_at: PropTypes.string.isRequired,
  format: PropTypes.oneOf(['json', 'protobuf']).isRequired,
  ids: PropTypes.shape({
    application_ids: PropTypes.shape({
      application_id: PropTypes.string,
    }).isRequired,
    webhook_id: PropTypes.string.isRequired,
  }).isRequired,
  updated_at: PropTypes.string,
})
PropTypes.webhooks = PropTypes.arrayOf(PropTypes.webhook)

export default PropTypes
