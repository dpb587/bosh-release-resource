# bosh-release-resource

[![license](https://img.shields.io/badge/license-mit-blue.svg?longCache=true)](LICENSE) [![dockerhub](https://img.shields.io/badge/dockerhub-latest-green.svg?longCache=true)](https://hub.docker.com/r/dpb587/bosh-release-resource/)

A [Concourse](https://concourse-ci.org/) resource for working with versions of a [BOSH](https://bosh.io/) release. Specifically focused on...

 * supporting non-[bosh.io releases](https://bosh.io/releases), private release repositories, and firewalled environments;
 * consolidating `finalize-release` tasks across release repositories; and
 * supporting version constraints.


## Source Configuration

 * **`repository`** - location of the BOSH release git repository
 * `branch` - the branch to use (optional unless using `out`; uses default remote branch)
 * `name` - a specific release name to use (by default, `name` from `config/final.yml` will be used)
 * `private_config` - a hash of settings which will be serialized to `config/private.yml` for `in`/`out`
 * `private_key` - a SSH private key when using private git repositories
 * `version` - a [supported](https://github.com/Masterminds/semver#basic-comparisons) version constraint (e.g. `2.x`, `>= 2.3.4`, `>2.3.2, <3`)


## Operations

### `check`

Get the latest versions of the release.

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
 * `committer_name` - full name to use when committing (default `CI Bot`)
 * `committer_email` - email address to use when committing (default `ci@localhost`)
 * `rebase` - enable automatic rebasing if there are conflicts on push (default `false`)
 * `skip_tag` - disable creating an annotated tag pointing to the commit the release tarball was created with (default `false`)

Metadata:

 * `bosh` - version of `bosh` CLI used to finalize the release
 * `commit` - commit reference where the new version was finalized


## Usage

To use this resource type, you should configure it in the [`resource_types`](https://concourse-ci.org/resource-types.html) section of your pipeline.

    - name: bosh-release
      type: docker-image
      source:
        repository: dpb587/bosh-release-resource

The default `latest` tag will refer to the current, stable version of this Docker image. For using the latest development version, you can refer to the `master` tag. If you need to refer to an older version of this image, you can refer to the appropriate `v{version}` tag.


## Examples

A few example scenarios which may be helpful...


### Switching from `bosh-io-release`

If you originally used the `bosh-io-release`...

    - name: concourse
      type: bosh-io-release
      source:
        repository: concourse/concourse

After switching to this `bosh-release`, the resource would look like...

    - name: concourse
      type: bosh-release
      source:
        repository: https://github.com/concourse/concourse.git

And jobs using this `bosh-release` resource should continue to work as before unless they relied on the `url` or `sha1` files (which are no longer applicable since there is not a public URL to download the tarball).


### Simple Triggering Pipeline

The [bosh.yml`](examples/bosh.yml) example watches for new releases and sends a [Slack notification](https://github.com/cloudfoundry-community/slack-notification-resource).


## Caveats

Subtle details you might care about...

 * This tags the commit from which the release tarball was created (`commit_hash`), not the commit which finalizes the release in the `releases` directory. This is primarily to ensure git tags match `commit_hash` and refer to the underlying source where changes between versions occur (as opposed to when it was finalized which may have a different set of files).
 * This currently requires that versions match semver conventions. If you use a non-semver versioning strategy, this may not work for all releases. This is primarily for simplicity of implementation; if too many releases need to support other conventions, it is probably worth changing.
 * This currently requires an externally-managed version file rather than supporting `bosh`'s automatic major version-bumping strategy. This is primarily to encourage explicit, semver-based version management; if this becomes too burdensome, it is probably worth changing.


## Development

To run tests...

  $ ginkgo -r


## License

[MIT License](LICENSE)
