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

import originalPropTypes from 'prop-types'

import ONLINE_STATUS from '@ttn-lw/constants/online-status'
import { components } from '@ttn-lw/constants/components'
import { ORGANIZATION, USER, APPLICATION, GATEWAY, CLIENT } from '@console/constants/entities'

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
    value: PropTypes.shape({}),
    defaultMessage: PropTypes.string,
  }),
  PropTypes.string,
  PropTypes.element,
])

PropTypes.icon = PropTypes.shape({})

PropTypes.error = PropTypes.oneOfType([
  PropTypes.oneOfType([
    PropTypes.shape({
      details: PropTypes.arrayOf(PropTypes.shape({})),
      message: PropTypes.string.isRequired,
      code: PropTypes.number.isRequired,
    }),
    PropTypes.shape({
      details: PropTypes.arrayOf(PropTypes.shape({})),
      message: PropTypes.string.isRequired,
      grpc_code: PropTypes.number.isRequired,
    }),
  ]),
  PropTypes.message,
  PropTypes.string,
  PropTypes.shape({
    message: PropTypes.string,
    stack: PropTypes.shape({}),
  }),
  PropTypes.instanceOf(Error),
])

PropTypes.link = PropTypes.shape({
  title: PropTypes.message.isRequired,
  icon: PropTypes.icon,
  path: PropTypes.string.isRequired,
  exact: PropTypes.bool,
  hidden: PropTypes.bool,
})

PropTypes.inputWidth = PropTypes.oneOf(['3xs', 'xxs', 'xs', 's', 'm', 'l', 'full'])

PropTypes.onlineStatus = PropTypes.oneOf(Object.values(ONLINE_STATUS))

// Entities and entity-related prop-types.
//
PropTypes.entityLocation = PropTypes.shape({
  latitude: PropTypes.number,
  longitude: PropTypes.number,
  altitude: PropTypes.number,
  source: PropTypes.string,
})

PropTypes.entityLocations = PropTypes.shape({
  user: PropTypes.entityLocation,
  'frm-payload': PropTypes.entityLocation,
  'lora-cloud-device-management-v1-gnss': PropTypes.entityLocation,
  'lora-cloud-device-management-v1-wifi': PropTypes.entityLocation,
  'lora-cloud-device-management-v1-unknown': PropTypes.entityLocation,
  'lora-cloud-geolocation-v3-gnss': PropTypes.entityLocation,
  'lora-cloud-geolocation-v3-rssi': PropTypes.entityLocation,
  'lora-cloud-geolocation-v3-tdoa': PropTypes.entityLocation,
  'lora-cloud-geolocation-v3-rssitdoacombined': PropTypes.entityLocation,
})

PropTypes.event = PropTypes.shape({
  name: PropTypes.string.isRequired,
  time: PropTypes.string.isRequired,
  identifiers: PropTypes.arrayOf(PropTypes.shape({})),
  data: PropTypes.shape({}),
})
PropTypes.events = PropTypes.arrayOf(PropTypes.event)

PropTypes.eventFilter = PropTypes.shape({
  id: PropTypes.string,
  filter: PropTypes.arrayOf(PropTypes.string),
  filterRegExp: PropTypes.arrayOf(PropTypes.string),
})
PropTypes.eventFilters = PropTypes.arrayOf(PropTypes.eventFilter)

PropTypes.location = PropTypes.shape({
  hash: PropTypes.string,
  key: PropTypes.string,
  pathname: PropTypes.string.isRequired,
  search: PropTypes.string,
  state: PropTypes.shape({
    info: PropTypes.message,
  }),
})

PropTypes.gateway = PropTypes.shape({
  antennas: PropTypes.arrayOf(
    PropTypes.shape({
      location: PropTypes.entityLocation,
    }),
  ),
  ids: PropTypes.shape({
    gateway_id: PropTypes.string,
  }).isRequired,
  name: PropTypes.string,
  description: PropTypes.string,
  created_at: PropTypes.string,
  updated_at: PropTypes.string,
  frequency_plan_id: PropTypes.string,
  gateway_server_address: PropTypes.string,
  schedule_anytime_delay: PropTypes.string,
})

PropTypes.gatewayStats = PropTypes.shape({
  connected_at: PropTypes.string,
  disconnected_at: PropTypes.string,
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

PropTypes.profilePicture = PropTypes.shape({
  sizes: PropTypes.shape({
    0: PropTypes.string,
  }),
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
  csrfToken: PropTypes.string.isRequired,
  sentryDsn: PropTypes.string,
  pageData: PropTypes.shape({}),
  config: PropTypes.shape({
    language: PropTypes.string,
    supportLink: PropTypes.string,
    documentationBaseUrl: PropTypes.string,
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
  locations: PropTypes.entityLocations,
  lorawan_phy_version: PropTypes.string,
  lorawan_version: PropTypes.string,
  supports_join: PropTypes.bool,
  frequency_plan_id: PropTypes.string,
})

PropTypes.deviceTemplate = PropTypes.shape({
  end_device: PropTypes.shape({
    supports_join: PropTypes.bool,
    multicast: PropTypes.bool,
    lorawan_version: PropTypes.string.isRequired,
    lorawan_phy_version: PropTypes.string.isRequired,
  }),
  field_mask: PropTypes.shape({
    paths: PropTypes.arrayOf(PropTypes.string),
  }),
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

PropTypes.marker = PropTypes.shape({
  position: PropTypes.shape({
    latitude: PropTypes.number,
    longitude: PropTypes.number,
  }).isRequired,
  accuracy: PropTypes.number,
  children: PropTypes.node,
})

PropTypes.markers = PropTypes.arrayOf(PropTypes.marker)

PropTypes.history = PropTypes.shape({
  listen: PropTypes.func,
})

PropTypes.right = PropTypes.string
PropTypes.rights = PropTypes.arrayOf(PropTypes.right)

PropTypes.collaborator = PropTypes.shape({
  rights: PropTypes.rights,
})

PropTypes.apiKey = PropTypes.shape({
  id: PropTypes.string.isRequired,
  rights: PropTypes.rights,
})

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
PropTypes.webhookTemplate = PropTypes.shape({
  ids: PropTypes.shape({
    template_id: PropTypes.string.isRequired,
  }).isRequired,
})
PropTypes.webhookTemplates = PropTypes.arrayOf(PropTypes.webhookTemplate)

PropTypes.euiPrefix = PropTypes.shape({
  join_eui: PropTypes.string,
  length: PropTypes.number,
})

PropTypes.passwordRequirements = PropTypes.shape({
  min_length: PropTypes.number,
  max_length: PropTypes.number,
  min_uppercase: PropTypes.number,
  min_digits: PropTypes.number,
})

PropTypes.euiPrefixes = PropTypes.arrayOf(PropTypes.euiPrefix)

PropTypes.routingPolicy = PropTypes.shape({
  uplink: PropTypes.shape({
    application_data: PropTypes.bool,
    join_request: PropTypes.bool,
    localization: PropTypes.bool,
    mac_data: PropTypes.bool,
    signal_quality: PropTypes.bool,
  }),
  downlink: PropTypes.shape({
    application_data: PropTypes.bool,
    join_accept: PropTypes.bool,
    mac_data: PropTypes.bool,
  }),
})

PropTypes.entity = PropTypes.oneOf([APPLICATION, GATEWAY, ORGANIZATION, USER, CLIENT])

PropTypes.unifiedEntity = PropTypes.shape({
  entity: PropTypes.shape({
    name: PropTypes.string,
  }),
  id: PropTypes.string.isRequired,
  path: PropTypes.string.isRequired,
  type: PropTypes.string.isRequired,
})
PropTypes.unifiedEntities = PropTypes.arrayOf(PropTypes.unifiedEntity)

PropTypes.notificationData = PropTypes.shape({
  data: PropTypes.oneOfType([
    PropTypes.shape({
      state: PropTypes.string,
    }),

    PropTypes.shape({
      create_client_request: PropTypes.shape({
        client: PropTypes.shape({
          ids: PropTypes.shape({
            client_id: PropTypes.string,
          }),
        }),
        collaborator: PropTypes.oneOfType([
          PropTypes.shape({
            organization_ids: PropTypes.shape({
              organization_id: PropTypes.string,
            }),
          }),
          PropTypes.shape({
            user_ids: PropTypes.shape({
              user_id: PropTypes.string,
            }),
          }),
        ]),
      }),
      client: PropTypes.shape({
        ids: PropTypes.shape({
          client_id: PropTypes.string,
        }),
      }),
      collaborator: PropTypes.oneOfType([
        PropTypes.shape({
          organization_ids: PropTypes.shape({
            organization_id: PropTypes.string,
          }),
        }),
        PropTypes.shape({
          user_ids: PropTypes.shape({
            user_id: PropTypes.string,
          }),
        }),
      ]),
    }),

    PropTypes.shape({
      id: PropTypes.string.isRequired,
      rights: PropTypes.arrayOf(PropTypes.string),
    }),

    PropTypes.shape({
      data: PropTypes.shape({
        ids: PropTypes.oneOfType([
          PropTypes.shape({
            organization_ids: PropTypes.shape({
              organization_id: PropTypes.string.isRequired,
            }),
          }),
          PropTypes.shape({
            user_ids: PropTypes.shape({
              user_id: PropTypes.string.isRequired,
            }),
          }),
        ]).isRequired,
      }).isRequired,
    }),

    PropTypes.shape({
      user: PropTypes.shape({
        ids: PropTypes.shape({
          user_id: PropTypes.string,
        }),
        name: PropTypes.string,
        description: PropTypes.string,
        primary_email_address: PropTypes.string,
      }).isRequired,
    }),
  ]).isRequired,
  entity_ids: PropTypes.shape({}).isRequired,
})

export default PropTypes
