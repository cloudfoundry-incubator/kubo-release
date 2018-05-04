# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'fileutils'
require 'open3'

def run_get_hostname_override(rendered_contents, executable_path)
  File.open(executable_path, 'w', 0777) do |f|
    f.write(rendered_contents)
  end

  # exercise bash function by first changing path for any necessary executables to our mocks
  cmd = 'PATH=%s:%s /bin/bash -c "source %s && get_hostname_override"' % [
      File.dirname(executable_path),
      ENV['PATH'],
      executable_path
  ]
  # capturing stderr (ignored) prevents expected warnings from showing up in test console
  result, _, _ = Open3.capture3(cmd)
  result
end

describe 'kube-proxy can set hostnameOverride' do
  let(:rendered_template) { compiled_template('kube-proxy', 'config/config.yml', {}, {}) }

  # Check that the config file has HOSTNAMEOVERRIDE so that start script can find
  # and replace it at runtime
  it 'kube-proxy config has HOSTNAMEOVERRIDE key word' do
    expect(rendered_template).to include('HOSTNAMEOVERRIDE')
  end

end

describe 'kube_proxy_ctl setting of hostnameOverride property' do
  let(:test_context) {
    mock_dir = '/tmp/kube_proxy_mock'
    FileUtils.remove_dir(mock_dir, true)
    FileUtils.mkdir(mock_dir)
    kube_proxy_ctl_file = mock_dir + '/kube_proxy_ctl'

    test_context = {'mock_dir' => mock_dir, 'kube_proxy_ctl_file' => kube_proxy_ctl_file}
  }
  after(:each) do
    FileUtils.remove_dir(test_context['mock_dir'], true)
  end

  describe 'when cloud-provider is NOT gce' do

    it 'sets hostname_override to IP address of container IP' do
      expected_spec_ip = '1111'
      rendered_kube_proxy_ctl = compiled_template('kube-proxy', 'bin/kube_proxy_ctl', {}, {}, {}, 'az1', expected_spec_ip)
      result = run_get_hostname_override(rendered_kube_proxy_ctl, test_context['kube_proxy_ctl_file'])

      expect(result).to include(expected_spec_ip)
    end
  end

  describe 'when cloud-provider is gce' do

    it 'sets hostname_override to google container hostname' do
      expected_google_hostname = 'foohost'

      echo_mock_file = test_context['mock_dir'] + '/curl'
      File.open(echo_mock_file, 'w', 0777) {|f|
        f.write("#!/bin/bash\n")
        f.write("echo #{expected_google_hostname}")
      }

      test_link = {'cloud-provider' => {
          'instances' => [],
          'properties' => {
              'cloud-provider' => {
                  'type' => 'gce',
                  'gce' => {
                      'project-id' => 'f',
                      'network-name' => 'ff',
                      'worker-node-tag' => 'fff',
                      'service_key' => 'ffff'
                  }}}}}
      rendered_kube_proxy_ctl = compiled_template('kube-proxy', 'bin/kube_proxy_ctl', {}, test_link)
      result = run_get_hostname_override(rendered_kube_proxy_ctl, test_context['kube_proxy_ctl_file'])

      expect(result).to include(expected_google_hostname)
    end
  end
end
