Exec { path => [ "/bin/", "/sbin/" , "/usr/bin/", "/usr/sbin/" ] }

include apt
include golang
include sqlite
class { 'postgresql::server': }
class { '::mysql::server':
  root_password    => 'disgo',
}

mysql::db { 'disgo':
  user     => 'disgo',
  password => 'disgo',
  host     => 'localhost'
}


postgresql::server::db { 'disgo':
  user     => 'disgo',
  password => postgresql_password('disgo', 'disgo'),
}

package { "git":
    ensure => "installed",
    before => Exec["install_deps"]
}

package { "mercurial":
    ensure => "installed",
}

exec { "setup-path":
  command => "/bin/echo 'export PATH=/home/vagrant/go/bin:\$PATH' >> /home/vagrant/.profile",
  unless  => "/bin/grep -q /home/vagrant/go/bin /home/vagrant/.profile ; /usr/bin/test $? -eq 0"
}

exec { "setup-workspace":
  command => "/bin/echo 'export GOPATH=/home/vagrant/go' >> /home/vagrant/.profile",
  unless  => "/bin/grep -q GOPATH /home/vagrant/.profile ; /usr/bin/test $? -eq 0",
  before => Exec["install_gin"]
}

exec { "install_gin":
  environment => ["GOPATH=/home/vagrant/go"],
  command => 'go get github.com/codegangsta/gin',
  require => [Package['git'], Class["golang"]]
}

exec { "install_deps":
  cwd => '/home/vagrant/go/src/github.com/pascalj/disgo',
  environment => ["GOPATH=/home/vagrant/go"],
  command => 'go get',
  require => [Package['git'], Package['mercurial'], Class["golang"]]
}

