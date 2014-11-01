Puppet module for installing the Go from the JuJu packages on
[Launchpad](https://launchpad.net/~juju/+archive/golang).

This module is also available on the [Puppet
Forge](https://forge.puppetlabs.com/garethr/golang)

[![Build
Status](https://secure.travis-ci.org/garethr/garethr-golang.png)](http://travis-ci.org/garethr/garethr-golang)

## Usage

The module includes a single class:

    include 'golang'

By default this sets up the PPA and installs the golang package.
