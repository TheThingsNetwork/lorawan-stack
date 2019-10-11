# The Things Stack for LoRaWAN Development

The Things Stack components are primarily built in Go, while we use Node for web front-ends. It is assumed that you have decent knowledge and experience with these technologies. If you want to get more familiar with Go, we strongly recommend to take [A Tour of Go](https://tour.golang.org/).

## Development Environment

The Things Network's development environment heavily relies on [`make`](https://www.gnu.org/software/make/). Under the hood, `make` calls other tools such as `git`, `go`, `yarn` etc. Recent versions are supported; Node v10.x and Go v1.12.x. Let's first make sure you have `go`, `node` and `yarn`:

On macOS using [Homebrew](https://brew.sh):

```sh
brew install go node yarn
```

On Ubuntu (or on Windows [using the Windows Subsystem for Linux](https://www.microsoft.com/nl-NL/store/p/ubuntu/9nblggh4msv6?rtc=1)):

```sh
curl -sL https://deb.nodesource.com/setup_10.x | sudo -E bash -
sudo apt-get install -y build-essential nodejs

curl -sSL https://dl.google.com/go/go1.12.3.linux-amd64.tar.gz | sudo tar -xz -C /usr/local
sudo ln -s /usr/local/go/bin/* /usr/local/bin
```

### Cloning the repository

If you are unfamiliar with forking projects on GitHub or cloning them locally, please [see the GitHub documentation](https://help.github.com/articles/fork-a-repo/).

### Getting started

As most of the tasks will be managed by `make` and `mage` we will first initialize the tooling. You may want to run this commands from time to time:

```sh
make init
```

For convenience, you can initialize the development databases with some defaults.

>Note: this requires [Docker Desktop](https://www.docker.com/products/docker-desktop).

```sh
make dev.stack.init
```

This starts a CockroachDB and Redis database in Docker containers, creates a database, migrates tables and creates a user `admin` with password `admin`.

### Managing the development databases

You can also use the following commands to start, stop and erase databases.

```bash
make dev.databases.start # Starts all databases in a Docker container
make dev.databases.stop  # Stops all databases

# The contents of the databases will be saved in .dev/data.

make dev.databases.erase # Stop all databases and erase storage.
```

#### CockroachDB

CockroachDB is a distributed SQL database that we use in the Identity Server.

You can use `make dev.databases.sql` to enter an SQL shell.

#### Redis

Redis is an in-memory data store that we use as a database for "hot" data.

You can use `make dev.databases.redis-cli` to enter a Redis-CLI shell.

### Testing

```sh
./mage go:test js:test jsSDK:test
```

### Building

There is a single binary for the server, `ttn-lw-stack`, as well as a binary for the command-line interface `ttn-lw-cli`. The single binary contains all components start one or multiple components. This allows you to run The Things Stack with one command in simple deployment scenarios, as well as distributing micro-services for more advanced scenarios.

We provide binary releases for all supported platforms, including packages for various package managers at https://github.com/TheThingsNetwork/lorawan-stack/releases. We suggest you use the compiled packages we provide in production scenarios.

For development/testing purposes we suggest either running required binaries via `go run` (e.g. `go run ./cmd/ttn-lw-cli` from repository root for CLI), or using `go build` directly. Note, that the frontend (if used) needs to be built via `./mage js:build` before `go build` or `go run` commands are run.

### Releasing

If you want, you can build a release snapshot with `./mage buildSnapshot`.

>Note: You will at least need to have [`rpm`](http://rpm5.org/) and [`snapcraft`](https://snapcraft.io/) in your `PATH`.

This will compile binaries for all supported platforms, `deb`, `rpm` and Snapcraft packages, release archives in `dist`, as well as Docker images.

>Note: The operating system and architecture represent the name of the directory in `dist` in which the binaries are placed.
>For example, the binaries for Darwin x64 (macOS) will be located at `dist/darwin_amd64`.

Releasing a new version consists of the following steps:

1. Updating the `CHANGELOG.md` file:
  - Change the **Unreleased** section to the new version
  - Check if we didn't forget anything important
  - Remove empty subsections
  - Update the list of links in the bottom of the file
2. Bumping the version
3. Writing the version files
4. Updating the `SECURITY.md` file with the supported versions
5. Creating the version bump commit
6. Creating the version tag
7. Building the release and pushing to package managers (this is done by CI)

Our development tooling helps with this process. The `mage` command has the following commands for version bumps:

```
version:bumpMajor      bumps a major version (from 3.4.5 -> 4.0.0).
version:bumpMinor      bumps a minor version (from 3.4.5 -> 3.5.0).
version:bumpPatch      bumps a patch version (from 3.4.5 -> 3.4.6).
version:bumpRC         bumps a release candidate version (from 3.4.5-rc1 -> 3.4.5-rc2).
version:bumpRelease    bumps a pre-release to a release version (from 3.4.5-rc1 -> 3.4.5).
```

These bumps can be combined (i.e. `version:bumpMinor version:bumpRC` bumps 3.4.5 -> 3.5.0-rc1). Apart from these bump commands, we have commands for writing version files (`version:files`), creating the bump commit (`version:commitBump`) and the version tag (`version:tag`).

A typical release process is executed directly on the `master` branch and looks like this:

```sh
git add CHANGELOG.md
./mage version:bumpPatch version:files version:commitBump version:tag // bump, write files, commit and tag.
git push origin $(mage version:current) // push the tag
git push origin master // push the master branch
```

Note that you must have sufficient repository rights to push to `master`.

After pushing the tag, our CI system will start building the release. When this is done, you'll find a new release on the [releases page](https://github.com/TheThingsNetwork/lorawan-stack/releases). After this is done, you'll need to edit the release notes. The release process will do its best to generate release notes for us, but they typically require a bit of editing.

#### API

> Note: If you don't work on changes in the API you can skip this section.

Our APIs are defined in `.proto` files in the `api` folder. These files describe the messages and interfaces of the different components of The Things Stack. If this is the first time you hear the term "protocol buffers" you should probably read the [protocol buffers documentation](https://developers.google.com/protocol-buffers/docs/proto3) before you continue.

From the `.proto` files, we generate code using the `protoc` compiler. As we plan to compile to a number of different languages, we decided to put the compiler and its dependencies in a Docker image, so make sure you have [Docker](https://www.docker.com/) installed before you try to compile them.

The actual commands for compilation are handled by our Makefile, so the only thing you have to execute, is:

```sh
./mage proto:clean proto:all
```

#### Folder Structure

```
.
├── .editorconfig       configuration for your editor, see editorconfig.org
├── CODEOWNERS          maintainers of folders who are required to approve pull requests
├── CONTRIBUTING.md     guidelines for contributing: branching, commits, code style, etc.
├── DEVELOPMENT.md      guide for setting up your development environment
├── docker-compose.yml  deployment file (including databases) for Docker Compose
├── Dockerfile          formula for building Docker images
├── LICENSE             the license that explains what you're allowed to do with this code
├── Makefile            dev/test/build tooling
├── README.md           general information about this project
│   ...
├── api                 contains the protocol buffer definitions for our API
├── cmd                 contains the different binaries that form The Things Stack for LoRaWAN
│   ├── internal        contains internal files shared between the different binaries
│   │   ...
│   ├── ttn-lw-cli      the command-line-interface for The Things Stack for LoRaWAN
│   └── ttn-lw-stack    bundles the server binaries that form The Things Stack for LoRaWAN
├── config              configuration for our JavaScript SDK and frontend
├── doc                 detailed documentation on the workings of The Things Stack for LoRaWAN
├── pkg                 contains all libraries used in The Things Stack for LoRaWAN
│   ├── component       contains the base component; all other components extend this component
│   ├── config          package for configuration using config files, environment and CLI flags
│   ├── console         package that provides the web server for the console
│   ├── errors          package for rich errors that include metadata and cross API boundaries
│   ├── log             package for logging
│   ├── messages        contains non-proto messages (such as the messages that are sent over MQTT)
│   ├── metrics         package for metrics collection
│   ├── ttnpb           contains generated code from our protocol buffer definitions and some helper functions
│   ├── types           contains primitive types
│   ├── webui           contains js code for the console and oauth provider
│   └── ...
├── public              frontend code will be compiled to this folder - not added to git
├── release             binaries will be compiled to this folder - not added to git
└── sdk                 source code for our SDKs
    └── js              source code for our JavaScript SDK
```


## Frontend
### Introduction
The Things Stack for LoRaWAN includes two frontend applications: the **Console** and **OAuth Provider**. Both applications use [React](https://reactjs.org/) as frontend framework. The `console` and `oauth` packages of the backend expose their respective web servers and handle all logic that cannot be done in the browser. Otherwise both applications are single page applications (SPA) that run entirely in the browser.

#### Console
The Console is the official management application of The Things Stack. It can be used to register applications, end devices or gateways, monitor network traffic, or configure network related options, among other things. The console uses an OAuth access token to communicate with The Things Stack.

#### OAuth
The OAuth app provides the necessary frontend for the OAuth provider of The Things Stack. It is used e.g. to display the authorization screen that users get prompted with when they want to authorize a third-party app to access The Things Stack.

### Building the frontend

#### Prerequisites
In order to build the frontend, you'll need the following:
* Node JS, version 10.x (see [Development Environment](#development-environment))
* NPM (node package manager), version >= 6.4.1 (⟶ `npm install -g npm` or `npm update npm -g` should give you the latest version)

#### Build process
You can control whether to build the frontend for production or development by setting the `$NODE_ENV` environment variable to either `development` or `production`. The frontend can then be built using:

```sh
mage js:build
```

This will initiate the following actions:

* Install a version-fixed binary of `yarn`
* Install JS SDK node dependencies via `yarn`
* Build the JS SDK
* Extract backend locale messages
* Install frontend node dependencies via `yarn`
* Build the frontend (using `webpack`) and output into `/public`

The difference of a development build includes:

* Including source maps
* Using DLL bundle for modules to reduce build time
* A couple of special build options to improve usage with `webpack-dev-server`

After successfully running the build command, The Things Stack has all necessary files to run the Console and OAuth provider applications.

### Development
#### Serving the frontend for development
For development purposes, the frontend can be run using `webpack-dev-server`. After following the [Getting Started](#getting-started) section to initialize The Things Stack and doing an initial build of the frontend via `mage js:build`, it can be served using:

```sh
export NODE_ENV=development
mage js:serve
```

The development server runs on `http://localhost:8080` and will proxy all api calls to port `1885`. The serve command watches any changes inside `pkg/webui` and refreshes automatically. 
In order to set up The Things Stack to support running the frontend via `webpack-dev-server`, the following environment setup is needed:

```
NODE_ENV=development
TTN_LW_LOG_LEVEL=debug
TTN_LW_IS_OAUTH_UI_JS_FILE="libs.bundle.js oauth.js"
TTN_LW_CONSOLE_UI_JS_FILE="libs.bundle.js console.js"
TTN_LW_CONSOLE_UI_CANONICAL_URL=http://localhost:8080/console
TTN_LW_CONSOLE_OAUTH_AUTHORIZE_URL=http://localhost:8080/oauth/authorize
TTN_LW_CONSOLE_OAUTH_TOKEN_URL=http://localhost:8080/oauth/token
TTN_LW_IS_OAUTH_UI_CANONICAL_URL=http://localhost:8080/oauth
TTN_LW_IS_EMAIL_NETWORK_IDENTITY_SERVER_URL=http://localhost:8080/oauth.js
TTN_LW_CONSOLE_UI_ASSETS_BASE_URL=http://localhost:8080/assets
```
*Note: We recommend using an environment switcher like [`direnv`](https://direnv.net/) to help you setting up environments for different tasks easily.*  
All of the configuration options above can also be set using configuration files or runtime flags. For more info in this regard, [see this guide](https://github.com/TheThingsNetwork/lorawan-stack/blob/master/doc/config.md).

#### Testing
For frontend testing, we use [`jest`](https://jestjs.io/en/). We currently don't enforce any coverage minimum, but consider testing for complex logic. We use both snapshot testing for React Components with [`enzyme`](https://airbnb.io/enzyme/) and plain `jest` testing for arbitrary logic.

To run the frontend tests, use:

```sh
mage js:test
```

### Internationalization (i18n)
The Things Stack for LoRaWAN employs i18n to provide usage experience in different languages. As such, also the frontend uses translatable messages. For this purpose we use [`react-intl`](https://github.com/yahoo/react-intl), which helps us greatly to define text messages used in the frontend.
The workflow for defining messages is as follows:

1. Add a `react-intl` message using `intl.defineMessages({…})`
  * This can be done either inline, or by adding it to `pkg/webui/lib/shared-messages.js`
2. Use the message in components (e.g. `sharedMessage.myMessage`)

After adding messages this way, it needs to be added the locales file `pkg/webui/locales/*.js` by using:

```sh
mage js:translations
```
*Note: When using `mage js:serve`, this command will be run automatically after any change*

The message definitions in `pkg/webui/locales` can be used to provide translations in other languages (e.g. `fr.js`). Keep in mind that locale files are checked in and committed, any discrepancy in the locales file with the defined messages will lead to a CI failure.

### Frontend Folder Structure
⟶ `pkg/webui`

```
.
├── assets            assets (eg. vectors, images) used by the frontend
├── components        react components shared throughout the frontend
├── console           root of the console application
│   ├── api           api definitions to communicate with The Things Stack
│   ├── containers    container components
│   ├── lib           utility classes and functions
│   ├── store         redux actions, reducers and logic middlewares
│   ├── views         whole view components of the console (~pages)
├── containers        global react container components
├── lib               global utility classes and functions
├── locales           frontend and backend locale jsons used for i18n
├── oauth             root of the oauth application
│   ├── api           api definitions to communicate with The Things Stack
│   ├── store         redux actions, reducers and logic middlewares
│   ├── views         whole view components of the oauth provider (~pages)
├── styles            global stylus (~css) styles and mixins
├── console.js        entry point of the console app
├── oauth.js          entry point of the oauth app
├── manifest.go       generated manifest of the frontend, containing file hashes
├── template.go       go template module used to render the frontend HTML
```

## Documentation

The documentation site for The Things Stack is built from the `doc` folder. 
All content is stored as Markdown files in `doc/content`.

In order to build the documentation site with the right theme, you need to run
`./mage docs:deps` from time to time. 

You can start a development server with live reloading by running
`./mage docs:server`. This command will print the address of the server.

The documentation site can be built by running `./mage docs:build`. This will 
output the site to `docs/public`.

For more details on how our documentation site is written, see the [Hugo docs](https://gohugo.io/documentation/).
