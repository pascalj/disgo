require 'spec_helper_system'

describe 'golang' do
  it 'class should install without errors' do
    pp = "class { 'golang': }"

    puppet_apply(pp) do |r|
      r.exit_code.should_not eq(1)
      r.refresh
      r.exit_code.should == 0
    end

    shell('go version') do |r|
      r.exit_code.should be_zero
      r.stderr.should be_empty
      r.stdout.should == "go version go1.1.2 linux/amd64"
    end
  end
end
