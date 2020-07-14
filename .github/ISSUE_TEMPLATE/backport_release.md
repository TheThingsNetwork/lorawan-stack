---
name: Backport Release
about: Checklist for Backport releases

---

<!--
Please check items along as you follow the release process.
-->

#### Overview

This is a checklist for backport releases. This is filled in by both the releaser and the reviewer where necessary.

#### Preparation

- [ ] Create a release/<version> branch off the backport/<minor> branch
```bash
$ git checkout backport/<minor>
$ git checkout -b release/${version}
```
- [ ] Cherry pick the necessary commits.
```bash
$ git cherrypick <commit>
```
- [ ] Update the `CHANGELOG.md` file as explained in the [Changelog Update](https://github.com/TheThingsNetwork/lorawan-stack/blob/master/DEVELOPMENT.md#changelog-update) section.
Once complete, you can add the file to staging
```bash
$ git add CHANGELOG.md
```
- [ ] Bump version as explained in the section [Version Bump](https://github.com/TheThingsNetwork/lorawan-stack/blob/master/DEVELOPMENT.md#version-bump).
- [ ] Create a pull request targeting `backport/<minor>`.

#### Check 1 (for reviewers)
- [ ] The correct base branch is used.
- [ ] The Changelog is complete i.e., contains only the changes that are in the release (not more/less).
- [ ] The version files are correctly updated.

#### Release
- [ ] Once this PR is approved and merged, checkout the latest  `backport/<minor>` branch locally.
- [ ] Create a version tag as explained in the section [Version Tag](https://github.com/TheThingsNetwork/lorawan-stack/blob/master/DEVELOPMENT.md#version-tag).
- [ ] Push the version tag. Once this is done, CI automatically starts building and pushing to package managers. When this is done, you'll find a new release on the [releases page](https://github.com/TheThingsNetwork/lorawan-stack/releases).
```bash
$ git push origin ${version}
```

### Post Release

- [ ] Edit the release notes on the Github releases page, which is typically copied from `CHANGELOG.md`.

#### Check 2 (for reviewers)

- [ ] The new release contains only the intended commits. This can be checked using `https://github.com/TheThingsNetwork/lorawan-stack/compare/v<previous-version>...v<current-version>`
