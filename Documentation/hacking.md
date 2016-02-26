# Hacking on acbuild

## Dependencies

acbuild uses glide to control the versions of its dependencies. General
documentation for glide can be found in [the project's
README](https://github.com/Masterminds/glide/blob/master/README.md).

Here you'll find whatever weird quirks we've found that developers of acbuild
should be aware of.

- The version of `appc/spec` specified in the `glide.yaml` file needs to match
  whatever version is vendored by rkt, or the build will fail.
