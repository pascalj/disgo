# == Class: golang
#
# Module to install an up-to-date version of Go from the
# JuJu PPA. The use of the PPA means this only works
# on Ubuntu.
#
# === Parameters
# [*version*]
#   The package version to install, passed to ensure.
#   Defaults to present.
#
class golang(
  $version = 'present'
) {
  include apt
  validate_string($version)
  validate_re($::osfamily, '^Debian$', 'This module uses PPA repos and only works with Debian based distros')

  apt::ppa { 'ppa:juju/golang':}

  package { 'golang':
    ensure  => $version,
    require => Apt::Ppa['ppa:juju/golang'],
  }
}
