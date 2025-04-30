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
import classnames from 'classnames'

import Radio from '@ttn-lw/components/radio-button'
import VerticalScrollFader from '@ttn-lw/components/vertical-scroll-fader'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './rights-group-modal.styl'

const RightsGroupModal = ({ className, rights, value, onChange, name }) => (
  <VerticalScrollFader className={style.radioGroup}>
    <div className={className}>
      <Radio.Group className="flex-column gap-cs-s" name={name} value={value} onChange={onChange}>
        {rights.map((r, idx) => (
          <div
            key={idx}
            className="p-vert-cs-xl p-sides-cs-l c-bg-neutral-min br-xl border-regular cursor-pointer"
            onClick={() => onChange(r.grants)}
          >
            <Radio
              label={r.title}
              value={r.grants}
              className="mb-0"
              onClick={e => e.stopPropagation()}
            />
            {r.descriptionItems.map((di, idx2) => (
              <Message
                component="p"
                className={classnames(style.descriptionItem, 'mt-0 mb-0 c-text-neutral-light')}
                content={di}
                key={idx2}
              />
            ))}
          </div>
        ))}
      </Radio.Group>
    </div>
  </VerticalScrollFader>
)

RightsGroupModal.propTypes = {
  className: PropTypes.string,
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func.isRequired,
  rights: PropTypes.rightsRadioItems.isRequired,
  value: PropTypes.arrayOf(PropTypes.string).isRequired,
}

RightsGroupModal.defaultProps = {
  className: undefined,
}

export default RightsGroupModal
