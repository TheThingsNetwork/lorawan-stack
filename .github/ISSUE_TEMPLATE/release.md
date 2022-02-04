---
name: Release
about: Checklist for releases

---

<!--
Please check items along as you follow the release process.
-->

#### Overview

This is a checklist for releases.

#### Release candidate

- [ ] Create a [documentation release issue](https://github.com/TheThingsIndustries/lorawan-stack-docs/issues/new?title=Release+v3.x.x&labels=release&template=release.md).
- [ ] Run the `create-release-branch.yml` (Create release branch) Github Action.
- [ ] Check that the versioning and `CHANGELOG.md` on the new branch are correct.
- [ ] Run the `release-rc.yml` (Release-candidate release) Github Action.
- [ ] You can repeatedly run the `release-rc.yml` (Release-candidate release) Github Action for multiple release candidates.

#### Release

- [ ] Checkout the latest `release/v3.${minor}.${patch}` branch locally.
- [ ] Create the version tag.
  ```bash
  $ git tag v3.${minor}.${patch}
  ```
- [ ] Push the version tag. Once this is done, CI automatically starts building and pushing to package managers. When this is done, you'll find a new release on the [releases page](https://github.com/TheThingsNetwork/lorawan-stack/releases).
  ```bash
  $ git push origin v3.${minor}.${patch}
  ```

#### Post release

- [ ] Merge `release/v3.${minor}.${patch}` back to `v3.${minor}`.
