Exec { path => [ "/bin/", "/sbin/" , "/usr/bin/", "/usr/sbin/" ] }

include apt
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
