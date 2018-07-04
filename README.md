# bosh-release-resource

[![license](https://img.shields.io/badge/license-mit-blue.svg?longCache=true)](LICENSE) [![build](https://travis-ci.org/dpb587/bosh-release-resource.svg)](https://travis-ci.org/dpb587/bosh-release-resource) [![dockerhub](https://img.shields.io/badge/dockerhub-latest-green.svg?longCache=true)](https://hub.docker.com/r/dpb587/bosh-release-resource/)

A [Concourse](https://concourse-ci.org/) resource for working with versions of a [BOSH](https://bosh.io/) release. Specifically focused on support for non-[bosh.io releases](https://bosh.io/releases), private release repositories, version constraints, dev releases, and `finalize-release` tasks.


## Source Configuration

 * **`repository`** - location of the BOSH release git repository
 * `branch` - the branch to use (optional unless using `out`; uses default remote branch)
 * `dev_releases` - set to `true` to create dev releases from every commit
 * `name` - a specific release name to use (default is `name` from `config/final.yml`)
 * `private_config` - a hash of settings which will be serialized to `config/private.yml` for `in`/`out`
 * `private_key` - a SSH private key when using private git repositories
 * `version` - a [supported](https://github.com/Masterminds/semver#basic-comparisons) version constraint (e.g. `2.x`, `>= 2.3.4`, `>2.3.2, <3`)


## Operations

### `check`

Get the latest versions of the release.

When `dev_releases` is enabled, the version will be in the format of `((version))-dev.((commit-date-utc))+commit.((short-commit-hash))`. The version number is an incremented patch from the latest final version (as of the referenced commit), followed by the commit-based, pre-release data. For example, if the last final release was `5.0.0` and the last commit was made on `2018-06-13` in `dd7c33e1d`... the version would be `5.0.1-dev.20180613T040837Z.commit.dd7c33e1d`).

Version:

 * `version` - release version


### `in`

Get a specific version of the release.

Parameters:

 * `tarball` - create a release tarball (default `true`)

Resource:

 * `name` - release name
 * `release.tgz` - source release tarball
 * `version` - release version

Metadata:

 * `bosh` - version of `bosh` CLI used to create the tarball
 * `time` - timestamp when the tarball was created


### `out`

Create a new version of the release from an existing tarball or repository checkout.

Parameters:

 * **`repository`** - path to a repository checkout from which to create a release (one of `repository` or `tarball` must be configured)
 * **`tarball`** - path to an existing release tarball to finalize (one of `repository` or `tarball` must be configured)
 * **`version`** - path to the file with contents of a specific version to use
 * `commit_file` - path to the file with contents of a commit message (default message `Version {version}`)
 * `author_name` - full name to use as commit author (default `CI Bot`)
 * `author_email` - email address to use as commit author (default `ci@localhost`)
 * `rebase` - enable automatic rebasing if there are conflicts on push (default `false`)
 * `skip_tag` - disable creating an annotated tag pointing to the commit the release tarball was created with (default `false`)

Metadata:

 * `bosh` - version of `bosh` CLI used to finalize the release
 * `commit` - commit reference where the new version was finalized


### `create-dev-release`

Another common task used outside of resource lifecycles is to generically create dev releases. The `create-dev-release` executable can be used to create a release tarball from the current working directory. See [`create-dev-release.yml`](tasks/create-dev-release.yml) for an example [task config](https://concourse-ci.org/tasks.html).

Arguments:

 * Output director for creating the release tarball


## Usage

To use this resource type, you should configure it in the [`resource_types`](https://concourse-ci.org/resource-types.html) section of your pipeline.

    - name: bosh-release
      type: docker-image
      source:
        repository: dpb587/bosh-release-resource

The default `latest` tag will refer to the current, stable version of this Docker image. For using the latest development version, you can refer to the `master` tag. If you need to refer to an older version of this image, you can refer to the appropriate `v{version}` tag.


### Alternative to `bosh-io-release`

This resource is generally equivalent to the `check`/`get` behaviors of [`bosh-io-release`](https://github.com/concourse/bosh-io-release-resource). The notable difference is that `url` and `sha1` are not provided since local tarballs are always created.

If you originally used the `bosh-io-release`...

    - name: concourse
      type: bosh-io-release
      source:
        repository: concourse/concourse

The equivalent `bosh-release` resource would be...

    - name: concourse
      type: bosh-release
      source:
        repository: https://github.com/concourse/concourse.git


## Examples

A few examples which may be helpful...

 * [BOSH Release Notifications](examples/bosh-release-notifications.yml) - a pipeline to send a [Slack](https://slack.com/) notification when there is a new release of [cloudfoundry/bosh](https://github.com/cloudfoundry/bosh)


## Caveats

Subtle details you might care about...

 * This tags the commit from which the release tarball was created (`commit_hash`), not the commit which finalizes the release in the `releases` directory. This is primarily to ensure git tags match `commit_hash` and refer to the underlying source where changes between versions occur (as opposed to when it was finalized which may have a different set of files).
 * This currently requires that versions match semver conventions. If you use a non-semver versioning strategy, this may not work for all releases. This is primarily for simplicity of implementation; if too many releases need to support other conventions, it is probably worth changing.
 * This currently requires an externally-managed version file rather than supporting `bosh`'s automatic major version-bumping strategy. This is primarily to encourage explicit, semver-based version management; if this becomes too burdensome, it is probably worth changing.


## Development

Before committing, tests can be run locally with [`bin/test`](bin/test). After pushing, [Travis CI](https://travis-ci.org/) should automatically run tests for commits and pull requests.


## License

[MIT License](LICENSE)
