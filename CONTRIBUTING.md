# Contributing to The Things Network Stack

## Branching

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

## Commit Messages

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

Changes that affect multiple components can be comma separated.

Good commit messages:
- `ns: Fix MIC check`
- `make: Set version from git tag, closes #123`
- `ns,as,gs: Fix TLS check`

Make sure that commits are scoped to something meaningful and could, potentially, be merged individually.

### Body

The body may contain a more detailed description of the commit, explaining what it changes and why. The "how" is less relevant, as this should be obvious from the diff.
