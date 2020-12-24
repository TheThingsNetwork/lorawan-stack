# The Things Stack for LoRaWAN JavaScript SDK
This is the official SDK for the The Things Stack for LoRaWAN.

## Usage

### Basic Setup

```javascript
import TTS from 'tts'

const token = 'paste-your-token-here'
const tts = new TTS({
  authorization: {
    mode: 'key', // Currently allows 'key' and 'session'.
    key: token,
  },
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

// Via Applications object.
const firstApp = await tts.Applications.create('testuser', appData)

// Via Application class.
const secondApp = new tts.Application(appData)
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
