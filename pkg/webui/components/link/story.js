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

import React from 'react'
import { storiesOf } from '@storybook/react'

import Link from '.'

const divStyle = { width: '100px', height: '100px', backgroundColor: 'lightblue' }

const titleStyle = { marginRight: '20px' }
storiesOf('Link', module)
  .add('Default (Unstyled)', () => (
    <div>
      <div>
        <span style={titleStyle}>link:</span>
        <Link to="/">Show more</Link>
      </div>
      <div>
        <span style={titleStyle}>anchor link:</span>
        <Link.Anchor href="/">Show more</Link.Anchor>
      </div>
      <div>
        <span style={titleStyle}>external anchor link:</span>
        <Link.Anchor external href="/">
          More information
        </Link.Anchor>
      </div>
      <div>
        <span style={titleStyle}>doc link:</span>
        <Link.DocLink path="/getting-started">Getting Started</Link.DocLink>
      </div>
    </div>
  ))
  .add('As wrapper', () => (
    <div>
      <div>
        <span style={titleStyle}>linked block element:</span>
        <Link to="/">
          <div style={divStyle} />
        </Link>
      </div>
      <div>
        <span style={titleStyle}>anchor linked block element:</span>
        <Link.Anchor href="/">
          <div style={divStyle} />
        </Link.Anchor>
      </div>
    </div>
  ))
  .add('Primary', () => (
    <div>
      <div>
        <span style={titleStyle}>link:</span>
        <Link primary to="/">
          Show more
        </Link>
      </div>
      <div>
        <span style={titleStyle}>anchor link:</span>
        <Link.Anchor primary href="/">
          Show more
        </Link.Anchor>
      </div>
      <div>
        <span style={titleStyle}>external anchor link:</span>
        <Link.Anchor primary external href="/">
          More information
        </Link.Anchor>
      </div>
      <div>
        <span style={titleStyle}>doc link:</span>
        <Link.DocLink primary path="/getting-started">
          Getting Started
        </Link.DocLink>
      </div>
    </div>
  ))
  .add('Secondary', () => (
    <div>
      <div>
        <span style={titleStyle}>link:</span>
        <Link secondary to="/">
          Show more
        </Link>
      </div>
      <div>
        <span style={titleStyle}>anchor link:</span>
        <Link.Anchor secondary href="/">
          Show more
        </Link.Anchor>
      </div>
      <div>
        <span style={titleStyle}>external anchor link:</span>
        <Link.Anchor secondary external href="/">
          More information
        </Link.Anchor>
      </div>
      <div>
        <span style={titleStyle}>doc link:</span>
        <Link.DocLink secondary path="/getting-started">
          Getting Started
        </Link.DocLink>
      </div>
    </div>
  ))
  .add('Show Visited', () => (
    <div>
      <div>
        <span style={titleStyle}>link:</span>
        <Link primary showVisited to="/">
          Show more
        </Link>
      </div>
      <div>
        <span style={titleStyle}>anchor link:</span>
        <Link.Anchor primary showVisited href="/">
          Show more
        </Link.Anchor>
      </div>
      <div>
        <span style={titleStyle}>external anchor link:</span>
        <Link.Anchor external primary showVisited href="/">
          More information
        </Link.Anchor>
      </div>
      <div>
        <span style={titleStyle}>doc link:</span>
        <Link.DocLink primary showVisited path="/getting-started">
          Getting Started
        </Link.DocLink>
      </div>
    </div>
  ))
  .add('Disabled', () => (
    <div>
      <div>
        <span style={titleStyle}>link:</span>
        <Link primary disabled to="/">
          Show more
        </Link>
      </div>
      <div>
        <span style={titleStyle}>anchor link:</span>
        <Link.Anchor primary disabled href="/">
          Show more
        </Link.Anchor>
      </div>
    </div>
  ))
