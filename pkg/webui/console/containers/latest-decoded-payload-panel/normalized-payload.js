// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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
import classNames from 'classnames'
import { defineMessages } from 'react-intl'

import Icon from '@ttn-lw/components/icon'
import VerticalScrollFader from '@ttn-lw/components/vertical-scroll-fader'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import { getPayloadSections } from './utils'

import style from './latest-decoded-payload-panel.styl'

const m = defineMessages({
  battery: 'Battery',
  water: 'Water',
  metering: 'Metering',
  action: 'Action',
  position: 'Position',
  normalizedPayloadMessage: `We normalize data based on vendor's informations we gather in our <Link>Device repository</Link>.`,
})

const PropertyItem = ({ value }) => (
  <div className="p-cs-xs c-bg-neutral-semilight br-xs c-text-neutral-heavy fw-bold w-content lh-1">
    {value}
  </div>
)

PropertyItem.propTypes = {
  value: PropTypes.string.isRequired,
}

const PropertySection = ({ title, icon, groups, firstItems }) => (
  <div
    className={classNames(style.normalizedPayloadPropertySection, {
      [style.normalizedPayloadPropertySectionBorder]: !firstItems,
    })}
  >
    <div
      className={classNames('d-flex al-center gap-cs-xs c-text-neutral-semilight', {
        'pt-cs-s': !firstItems,
      })}
    >
      <Icon icon={icon} />
      <Message content={title} />
    </div>
    <div className="d-flex gap-cs-xs flex-wrap mt-cs-s">
      {groups.map((item, i) => (
        <PropertyItem key={i} value={item.value} />
      ))}
    </div>
  </div>
)

PropertySection.propTypes = {
  firstItems: PropTypes.bool.isRequired,
  groups: PropTypes.arrayOf(PropTypes.shape({})).isRequired,
  icon: PropTypes.element.isRequired,
  title: PropTypes.string.isRequired,
}

const NormalizedPayload = ({ payload }) => {
  const sections = getPayloadSections(payload, m)

  return (
    <>
      <VerticalScrollFader faderHeight="3rem" light className={style.normalizedPayloadScroll}>
        <div className="d-flex flex-wrap gap-cs-m p-cs-m">
          {sections.map((section, i) => (
            <PropertySection
              key={section.title}
              title={section.title}
              icon={section.icon}
              groups={section.groups}
              firstItems={i <= 1}
            />
          ))}
        </div>
      </VerticalScrollFader>
      <Message
        className="c-text-neutral-semilight"
        content={m.normalizedPayloadMessage}
        values={{
          Link: msg => (
            <Link.Anchor
              primary
              href="https://www.thethingsnetwork.org/device-repository"
              target="_blank"
            >
              {msg}
            </Link.Anchor>
          ),
        }}
      />
    </>
  )
}

NormalizedPayload.propTypes = {
  payload: PropTypes.shape({
    time: PropTypes.string,
    battery: PropTypes.number,
    soil: PropTypes.shape({
      depth: PropTypes.number,
      moisture: PropTypes.number,
      temperature: PropTypes.number,
      ec: PropTypes.number,
      pH: PropTypes.number,
      n: PropTypes.number,
      p: PropTypes.number,
      k: PropTypes.number,
    }),
    air: PropTypes.shape({
      location: PropTypes.string,
      temperature: PropTypes.number,
      relativeHumidity: PropTypes.number,
      pressure: PropTypes.number,
      co2: PropTypes.number,
      lightIntensity: PropTypes.number,
    }),
    wind: PropTypes.shape({
      speed: PropTypes.number,
      direction: PropTypes.number,
    }),
    water: PropTypes.shape({
      leak: PropTypes.bool,
      temperature: PropTypes.shape({
        min: PropTypes.number,
        max: PropTypes.number,
        avg: PropTypes.number,
        current: PropTypes.number,
      }),
    }),
    metering: PropTypes.shape({
      water: {
        total: PropTypes.number.isRequired,
      },
    }),
    action: PropTypes.shape({
      motion: {
        detected: PropTypes.bool.isRequired,
        count: PropTypes.number.isRequired,
      },
      contactState: PropTypes.string.isRequired,
    }),
    position: {
      latitude: PropTypes.string.isRequired,
      longitude: PropTypes.string.isRequired,
    },
  }).isRequired,
}

export default NormalizedPayload
