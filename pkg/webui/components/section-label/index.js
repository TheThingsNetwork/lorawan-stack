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

import React from 'react'
import classnames from 'classnames'

import Button from '@ttn-lw/components/button'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './section-label.styl'

const SectionLabel = ({
  label,
  icon,
  className,
  onClick,
  buttonDisabled,
  'data-test-id': dataTestId,
}) => (
  <div className={classnames(className, style.sectionLabel)} data-test-id={dataTestId}>
    {label}
    <Button naked icon={icon} disabled={buttonDisabled} onClick={onClick} />
  </div>
)

SectionLabel.propTypes = {
  buttonDisabled: PropTypes.bool,
  className: PropTypes.string,
  'data-test-id': PropTypes.string,
  icon: PropTypes.string.isRequired,
  label: PropTypes.oneOfType([PropTypes.node, PropTypes.string]).isRequired,
  onClick: PropTypes.func,
}

SectionLabel.defaultProps = {
  onClick: () => null,
  buttonDisabled: false,
  className: undefined,
  'data-test-id': 'section-label',
}

export default SectionLabel
