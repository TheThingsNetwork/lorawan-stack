# Contributing to The Things Network Stack

Thank you for being willing to contribute to the Things Network Stack! Our core team appreciates and values external contributions, even for tiny fixes.

We usually welcome these kinds of contributions:
* **Bug fixes and feature implementation**: If you notice that a component of the Things Network Stack is not behaving as it should, or if you'd like to see a feature implemented in the Stack, you're welcome to submit modifications. Our core team might already be working on it - make sure you communicate about it in a related [GitHub issue](https://github.com/TheThingsNetwork/ttn/issues).
* **Documentation**: Writing documentation on internal packages and external APIs helps users and fellow contributors alike to understand better how to work with our stack, and is a great way to contribute without submitting code. You can either submit documentation in this repo, to document how the Stack is implemented, or in [our `docs` repo](https://github.com/TheThingsNetwork/docs), that is linked to our [official documentation](https://www.thethingsnetwork.org/docs).

If you'd like to contribute by writing code, you'll find [here](DEVELOPMENT.md) how to set up your development environment. We also have code and Git workflow conventions that will help you make better contributions.

+ [Git branching workflow](#branching)
+ [Commit conventions](#commit)
+ [Coding conventions](#code)

## <a name="branching"></a>Branching

### Naming

All branches shall have one of these names.

- `master`: the default branch. This is a clean branch where reviewed, approved and CI passed pull requests are merged into. Merging to this branch is restricted to project maintainers
- `fix/#-short-name` or `fix/short-name`: refers to a fix, preferably with issue number. The short name describes the bug or issue
- `feature/#-short-name` or `feature/short-name`: (main) feature branch, preferably with issue number. The short name describes the feature
  - `feature/#-short-name-part`: a sub scope of the feature in a separate branch, that is intended to merge into the main feature branch before the main feature branch is merged into `master`
- `issue/#-short-name`: anything else that refers to an issue but is not clearly a fix nor a feature

### Scope

A fix, feature or issue branch should be **small and focused** and should be scoped to a **single specific task**. Do not combine new features and refactoring of existing code.

### Rebasing and Merging

Before feature branches are merged, they shall rebased on top of their target branch. Do not rebase a branch if others are still working on a derived branch.

Interactive rebase (git rebase -i) may be used to rewrite commit messages that do not follow these contribution guidelines

## <a name="commit"></a>Commit Messages

The first line of a commit message is the subject. The commit message may contain a body, separated from the subject by an empty line.

### Subject

The subject contains the concerning component or topic and a concise message in [the imperative mood](https://chris.beams.io/posts/git-commit/#imperative), starting with a capital. The subject may also contain references to issues or other resources.

The component or topic is typically a few characters long and should always be present. Component names are:

- `api`: API, typically protos
- `gs`: Gateway Server
- `ns`: Network Server
- `as`: Application Server
- `is`: Identity Server
- `webui`: Console
- `util`: utilities
- `ci`: CI instructions, e.g. Travis file
- `doc`: documentation
- `dev`: other non-functional development changes, e.g. Makefile, .gitignore, editor config
- `all`: changes affecting all code, e.g. primitive types

Changes that affect multiple components can be comma separated.

Good commit messages:
- `ns: Fix MIC check`
- `make: Set version from git tag, closes #123`
- `ns,as,gs: Fix TLS check`

Make sure that commits are scoped to something meaningful and could, potentially, be merged individually.

### Body

The body may contain a more detailed description of the commit, explaining what it changes and why. The "how" is less relevant, as this should be obvious from the diff.

## <a name="code"></a>Code

### Formatting

We want our code to be consistent across our projects, so we'll have to agree on a number of formatting rules. These rules should usually usually be applied by your editor. Make sure to install the [editorconfig](https://editorconfig.org) plugin for your editor.

Go code can be automatically formatted using the [`gofmt`](https://golang.org/cmd/gofmt/) tool. There are many editor plugins that call `gofmt` when you save your files.

#### General

We use **utf-8**, **LF** line endings, a **final newline** and we **trim whitespace** from the end of the line (except in Markdown).

#### Tabs vs Spaces

Many developers have strong opinions about using tabs vs spaces. We apply the following rules:

- All `.go` files are indented using **tabs**
- The `Makefile` and all `.make` files are indented using **tabs**
- All other files are indented using **two spaces**

#### Line length

- If a line is longer than 80 columns, try to find a "natural" break
- If a line is longer than 120 columns, insert a line break
- In very special cases, longer lines are tolerated

### Linting

We use [`golint`](github.com/golang/lint/golint) to lint `.go` files and [`eslint`](https://eslint.org) to lint `.js` and `.vue` files. These tools should automatically be installed when initializing your development environment.

### Variable naming

Variable names should be short and concise.

We follow the [official go guidelines](https://github.com/golang/go/wiki/CodeReviewComments#variable-names) and try to be consistent with Go standard library as much as possible, everything not defined in the tables below should follow Go standard library naming scheme.

#### Single-word entities

| entity               | name    | example type                                                  |
| :------------------: | :-----: | :-----------------------------------------------------------: |
| context              | ctx     | context.Context                                               |
| mutex                | mu      | sync.Mutex                                                    |
| configuration        | conf    | github.com/TheThingsNetwork/ttn/pkg/config.Config             |
| logger               | log     | github.com/TheThingsNetwork/ttn/pkg/ttnlog.Logger             |
| message              | msg     | github.com/TheThingsNetwork/ttn/api/gateway.UplinkMessage     |
| status               | st      | github.com/TheThingsNetwork/ttn/api/gateway.Status            |
| server               | srv     | github.com/TheThingsNetwork/ttn/pkg/network-server.Server     |
| EUI                  | eui     | github.com/TheThingsNetwork/ttn/pkg/types.DevEUI              |
| ID                   | id      | string                                                        |
| counter              | cnt     | int                                                           |
| gateway              | gtw     |                                                               |
| application          | app     |                                                               |
| end device           | dev     |                                                               |
| user                 | user    |                                                               |


#### 2-word entities
In case both of the words have an implementation-specific meaning, the variable name is the combination of first letter of each word.

| entity                                                  | name    |
| :-----------------------------------------------------: | :-----: |
| wait group                                              | wg      |
| gateway server                                          | gs      |
| network server                                          | ns      |
| join server                                             | js      |
| application server                                      | as      |
| identity server                                         | is      |

In case one of the words specifies the meaning of the variable in a specific language construct context, the variable name is the combination of abbrevations of the words. 

#### Well-known variable names

These are the names of variables that occur often in the code. Be consistent in naming them, even when their
meaning is obvious from the context.

| entity                          | name    |
| :-----------------------------: | :-----: |
| gateway id                      | gtwID   |
| gateway EUI                     | gtwEUI  |
| application id                  | appID   |
| application EUI                 | appEUI  |
| device id                       | devID   |
| user id                         | userID  |
