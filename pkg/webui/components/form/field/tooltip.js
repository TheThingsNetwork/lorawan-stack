// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'

import Icon, { IconHelp } from '@ttn-lw/components/icon'
import Tooltip from '@ttn-lw/components/tooltip'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import { descriptions, links } from '@ttn-lw/lib/field-description-messages'

import style from './field.styl'

const m = defineMessages({
  descriptionTitle: 'What is this?',
  locationTitle: 'What should I enter here?',
  absenceTitle: 'What if I cannot find the correct value?',
  viewGlossaryPage: 'View glossary page',
  readMore: 'Read more',
})

const Content = props => {
  const { tooltipDescription, glossaryTerm, children } = props
  const { description, location, absence, glossaryId } = tooltipDescription

  const hasLocation = Boolean(location)
  const hasAbsence = Boolean(absence)
  const hasChildren = Boolean(children)
  const hasGlossary = Boolean(glossaryId)

  return (
    <div className={style.tooltipContent}>
      <Message className={style.tooltipTitle} content={m.descriptionTitle} component="h4" />
      <Message
        className={style.tooltipDescription}
        content={description}
        component="p"
        convertBackticks
      />
      {hasLocation && (
        <>
          <Message className={style.tooltipTitle} content={m.locationTitle} component="h4" />
          <Message
            className={style.tooltipDescription}
            content={location}
            component="p"
            convertBackticks
          />
        </>
      )}
      {hasAbsence && (
        <>
          <Message className={style.tooltipTitle} content={m.absenceTitle} component="h4" />
          <Message
            className={style.tooltipDescription}
            content={absence}
            component="p"
            convertBackticks
          />
        </>
      )}
      {(hasChildren || hasGlossary) && (
        <div className={style.tooltipLinks}>
          {children}
          {hasGlossary && (
            <Link.GlossaryLink
              term={glossaryTerm}
              glossaryId={glossaryId}
              title={m.viewGlossaryPage}
              primary
            />
          )}
        </div>
      )}
    </div>
  )
}

Content.propTypes = {
  children: PropTypes.node,
  glossaryTerm: PropTypes.message,
  tooltipDescription: PropTypes.shape({
    description: PropTypes.message.isRequired,
    location: PropTypes.message,
    absence: PropTypes.message,
    glossaryId: PropTypes.string,
  }).isRequired,
}

Content.defaultProps = {
  children: null,
  glossaryTerm: undefined,
}

const FieldTooltip = React.memo(props => {
  const { id, glossaryTerm } = props

  const tooltipDescription = descriptions[id]
  if (!tooltipDescription) {
    return null
  }

  const tooltipAdditionalLink = links[id]
  let link = null
  if (tooltipAdditionalLink) {
    const { documentationPath, externalUrl } = tooltipAdditionalLink
    if (documentationPath) {
      link = (
        <Link.DocLink primary path={documentationPath}>
          <Message content={m.readMore} />
        </Link.DocLink>
      )
    } else if (externalUrl) {
      link = (
        <Link.Anchor primary href={externalUrl} target="_blank">
          <Message content={m.readMore} />
        </Link.Anchor>
      )
    }
  }

  return (
    <Tooltip
      className={style.tooltip}
      placement="bottom-start"
      interactive
      small
      content={
        <Content
          glossaryTerm={glossaryTerm}
          tooltipDescription={tooltipDescription}
          children={link}
        />
      }
    >
      <Icon className={style.tooltipIcon} icon={IconHelp} />
    </Tooltip>
  )
})

FieldTooltip.propTypes = {
  glossaryTerm: PropTypes.message,
  id: PropTypes.string.isRequired,
}

FieldTooltip.defaultProps = {
  glossaryTerm: undefined,
}

export default FieldTooltip
