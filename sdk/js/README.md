# The Things Network Stack for LoRaWAN JavaScript SDK
This is the official SDK for the The Things Network Stack for LoRaWAN.

## Installation

```
$ yarn add ttn
```

## Usage
### Basic Setup
```javascript
import Ttn from 'ttn'

const token = 'paste-your-token-here'
const ttn = new Ttn(token, {
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
const firstApp = await Ttn.Applications.create('testuser', appData)

// Via Application class
const secondApp = new Ttn.Application(appData)
await secondApp.save()
```

## Development
### Building
```
$ yarn run build
```
This will transpile the source to `/dist`
### Watching Changes
```
$ yarn run watch
```
### Testing
```
$ yarn run jest
```

## Examples
There are some basic usage examples in `/src/examples`
