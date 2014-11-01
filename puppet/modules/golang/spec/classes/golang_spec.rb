require 'spec_helper'

describe 'golang', :type => :class do
  let(:facts) { {
    :osfamily => 'Debian',
    :operatingsystem => 'Ubuntu',
    :lsbdistcodename => 'precise',
    :lsbdistid => 'ubuntu'
  } }

  context 'with no parameters' do
    it { should contain_class('apt') }
    it { should contain_package('golang').with_ensure('present') }
    it { should contain_apt__ppa('ppa:juju/golang') }
  end

  context 'with a custom version' do
    let(:params) { {'version' => 'absent' } }
    it { should contain_package('golang').with_ensure('absent') }
  end

  context 'with an invalid distro name' do
    let(:facts) { {
      :osfamily => 'RedHat',
      :lsbdistid => 'redhat'
    } }
    it do
      expect {
        should contain_package('new-golang')
      }.to raise_error(Puppet::Error, /This module only works on Debian or derivatives/)
    end
  end

end
