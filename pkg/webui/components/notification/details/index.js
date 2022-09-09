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

import React, { useCallback, useState } from 'react'
import { defineMessages } from 'react-intl'

import Button from '@ttn-lw/components/button'
import PortalledModal from '@ttn-lw/components/modal/portalled'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './details.styl'

const m = defineMessages({
  showDetails: 'Show details',
  details: 'Details',
  errorDetails: 'Error details',
  close: 'Close',
})

const Details = props => {
  const [showDetails, setShowDetails] = useState()
  const { details, isError } = props
  const content = typeof details === 'string' ? details : JSON.stringify(details, undefined, 2)

  const showDropdown = useCallback(() => {
    setShowDetails(true)
  }, [])

  const handleModalComplete = useCallback(() => {
    setShowDetails(false)
  }, [])

  return (
    <div className={style.details}>
      <Button
        className={style.detailsButton}
        naked
        onClick={showDropdown}
        message={m.showDetails}
        type="button"
      />
      <PortalledModal
        title={isError ? m.errorDetails : m.details}
        visible={showDetails}
        onComplete={handleModalComplete}
        approval={false}
        buttonMessage={m.close}
        approveButtonProps={{ primary: false, icon: undefined }}
        noTitleLine
      >
        <pre className={style.detailsCode}>{content}</pre>
      </PortalledModal>
    </div>
  )
}

Details.propTypes = {
  details: PropTypes.oneOfType([PropTypes.string, PropTypes.error]).isRequired,
  isError: PropTypes.bool,
}

Details.defaultProps = {
  isError: false,
}

export default Details
