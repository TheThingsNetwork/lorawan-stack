# The Things Stack for LoRaWAN JavaScript SDK
This is the official SDK for the The Things Stack for LoRaWAN.

## Installation

```bash
$ yarn add ttn
```

## Usage
### Basic Setup
```javascript
import TTN from 'ttn'

const token = 'paste-your-token-here'
const ttn = new TTN(token, {
  connectionType: 'http',
  baseURL: 'http://localhost:1885/api/v3',
  defaultUserId: 'testuser',
})
```

### Creating Applications
```javascript
const appData = {
  ids: {
    application_id: 'first-app',
  },
  name: 'Test App',
  description: 'Some description',
}

// Via Applications object
const firstApp = await ttn.Applications.create('testuser', appData)

// Via Application class
const secondApp = new ttn.Application(appData)
await secondApp.save()
```

## Development
### Building
```bash
$ yarn run build
```
This will transpile the source to `/dist`
### Watching Changes
```bash
$ yarn run watch
```
### Testing
```bash
$ yarn run jest
```

## Examples
There are some basic usage examples in `/src/examples`
