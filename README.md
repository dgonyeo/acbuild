# acbuild

acbuild is a command line utility to build and modify App Container Images 
(ACIs), the container image format defined in the 
[App Container (appc) spec](https://github.com/appc/spec).

<iframe width="560" height="315" src="https://www.youtube.com/embed/WcnIDm80y68" frameborder="0" allowfullscreen></iframe>

## Rationale

We needed a powerful tool for constructing and manipulating ACIs that made it
easy to iteratively build containers, both from scratch and atop existing
images. We wanted that tool to integrate well with Unix mechanisms like the
shell and `Makefile`s so it would fit seamlessly into well-known administrator
and developer workflows.

## Installation

### Dependencies

acbuild can only be run on a Linux system, and has only been tested on the
amd64 architecture.

For trying out acbuild on Mac OS X, it's recommended to use Vagrant.
Instructions on how to do this are a little further down in this document.

acbuild requires a handful of commands be available on the system on 
which it's run:

- `systemd-nspawn`
- `cp`
- `modprobe`
- `gpg`

### Prebuilt Binaries

The easiest way to get `acbuild` is to download one of the
[releases](https://github.com/appc/acbuild/releases) from GitHub.

### Build from source

The other way to get `acbuild` is to build it from source.

Dependency versioning is done with
[glide](https://github.com/Masterminds/glide), and due to this a prerequisite
to building acbuild is that the glide binary is on your path. If you have a
working go setup, the easiest way to do this is to run `go get
github.com/Masterminds/glide` and then make sure that `$GOPATH/bin` is on your
path (if it's not, `export PATH=$PATH:$GOPATH/bin`).

Once glide is available, just perform the following steps.

1. Grab the source code for `acbuild` by `git clone`ing the source repository:
   ```
   cd ~
   git clone https://github.com/appc/acbuild
   ```

2. Run the `build` script from the root source repository directory:
   ```
   cd acbuild
   ./build
   ```

   Or, if you want to build in docker (assuming `$PWD` exists and contains
   `acbuild/` on your Docker host):

   ```
   cd acbuild
   ./build-docker
   ```

3. A `bin/` directory will be created that contains the `acbuild` tool. To make
   sure your shell can find this executable, append this directory to your
   environment's `$PATH` variable. You can do this in your `.bashrc` or similar
   file, for example:
   ```
   vi ~/.bashrc
   ```

and put the following lines at the end of the file:
   ```
   export ACBUILD_BIN_DIR=~/acbuild/bin
   export PATH=$PATH:$ACBUILD_BIN_DIR
   ```

### Trying out acbuild using Vagrant

For users with Vagrant 1.5.x or greater, there's a provided `Vagrantfile` that
can quickly get you set up with a Linux VM that has both acbuild and rkt. The
following steps will grab acbuild, set up the machine, and ssh into it.

```
git clone https://github.com/appc/acbuild
cd acbuild
vagrant up
vagrant ssh
```

## Documentation

Documentation about acbuild and many of its commands is available in the
[`Documentation`
directory](https://github.com/appc/acbuild/tree/master/Documentation) in this
repository.

## Examples

Check out the [`examples`
directory](https://github.com/appc/acbuild/tree/master/examples) for some common
applications being packaged into ACIs with `acbuild`.

## Related work

- https://github.com/sgotti/baci
- https://github.com/appc/spec/tree/master/actool - particularly the `build` and
  `patch-manifest` subcommands. `acbuild` may subsume such functionality,
  leaving `actool` as a validator only.
- https://github.com/blablacar/dgr
