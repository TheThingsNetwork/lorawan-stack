---
name: Release
about: Checklist for releases

---

<!--
Please check items along as you follow the release process.
-->

#### Overview

This is a checklist for releases. This is filled in by both the releaser and the reviewer where necessary.

#### Preparation

- [ ] Create a [documentation release issue](https://github.com/TheThingsIndustries/lorawan-stack-docs/issues/new?title=Release+v3.x.x&labels=release&template=release.md).
- [ ] Run `tools/bin/prepare_release minor` or `tools/bin/prepare_release patch`, depending on the kind of release.
- [ ] Follow the instructions.

#### Check 1 (for reviewers)

- [ ] The Changelog is complete i.e., contains only the changes that are in the release (not more/less).
- [ ] The version files are correctly updated.

#### Release

- [ ] Once this PR is approved and merged, checkout the latest `v3.${minor}` branch locally.
- [ ] Create a version tag
  ```bash
  $ tools/bin/mage version:bumpXXX version:tag
  # For RCs, make sure to use the same bumping combination (ex: `version:bumpXXX version:bumpYYY`) as used in the bump step above.
  ```
- [ ] Push the version tag. Once this is done, CI automatically starts building and pushing to package managers. When this is done, you'll find a new release on the [releases page](https://github.com/TheThingsNetwork/lorawan-stack/releases).
  ```bash
  $ git push origin ${version}
  ```

#### Post Release

- [ ] For non RC releases, push the Docker latest tag.
    ```bash
    $ versionDockerTag=${$(tools/bin/mage version:current)#"v"}
    $ docker pull thethingsnetwork/lorawan-stack:${versionDockerTag}
    $ docker tag thethingsnetwork/lorawan-stack:${versionDockerTag} thethingsnetwork/lorawan-stack:latest
    $ docker push thethingsnetwork/lorawan-stack:latest
    ```

#### Check 2 (for reviewers)

- [ ] The new release contains only the intended commits. This can be checked using `https://github.com/TheThingsNetwork/lorawan-stack/compare/v<previous-version>...v<current-version>`
- [ ] The Docker latest tag is up to date.
- [ ] The [documentation site](https://www.thethingsindustries.com/docs/) has been updated.
