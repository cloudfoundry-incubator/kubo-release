# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'fileutils'
require 'open3'

def run_get_hostname_override(rendered_contents, executable_path)
  File.open(executable_path, 'w', 0o777) do |f|
    f.write(rendered_contents)
  end

  # exercise bash function by first changing path for any necessary executables to our mocks
  cmd = format('PATH=%<dirname>s:%<env_path>s /bin/bash -c "source %<exe>s && get_hostname_override"',
               dirname: File.dirname(executable_path), env_path: ENV['PATH'], exe: executable_path)
  # capturing stderr (ignored) prevents expected warnings from showing up in test console
  result, = Open3.capture3(cmd)
  result
end

describe 'kube-proxy can set hostnameOverride' do
  let(:config_properties) do
  {
    'kube-proxy-configuration' => { 'mode': 'iptables' }
  }
  end
  let(:rendered_template) { compiled_template('kube-proxy', 'config/config.yml', config_properties, {}) }

  # Check that the config file has HOSTNAMEOVERRIDE so that start script can find
  # and replace it at runtime
  it 'kube-proxy config has HOSTNAMEOVERRIDE key word' do
    expect(rendered_template).to include('HOSTNAMEOVERRIDE')
  end
end

describe 'kube_proxy_ctl setting of hostnameOverride property' do
  let(:test_context) do
    mock_dir = '/tmp/kube_proxy_mock'
    FileUtils.remove_dir(mock_dir, true)
    FileUtils.mkdir(mock_dir)
    kube_proxy_ctl_file = mock_dir + '/kube_proxy_ctl'

    { 'mock_dir' => mock_dir, 'kube_proxy_ctl_file' => kube_proxy_ctl_file }
  end
  after(:each) do
    FileUtils.remove_dir(test_context['mock_dir'], true)
  end

  describe 'when cloud-provider is AWS' do
    it 'sets hostname_override to aws container hostname' do
      expected_aws_hostname = 'foohost.aws.internal'

      echo_mock_file = test_context['mock_dir'] + '/curl'
      File.open(echo_mock_file, 'w', 0o777) do |f|
        f.write("#!/bin/bash\n")
        f.write("echo #{expected_aws_hostname}")
      end

      test_link = { 'cloud-provider' => {
        'instances' => [],
        'properties' => {
          'cloud-provider' => {
            'type' => 'aws',
          }
        }
      } }
      rendered_kube_proxy_ctl = compiled_template('kube-proxy', 'bin/kube_proxy_ctl', { 'cloud-provider' => 'aws' }, test_link)
      result = run_get_hostname_override(rendered_kube_proxy_ctl, test_context['kube_proxy_ctl_file'])

      expect(result).to include(expected_aws_hostname)
    end
  end
  describe 'when cloud-provider is GCP' do
    it 'sets hostname_override to gcp cloud id' do
      expected_gcp_hostname = '123456'

      echo_mock_file = test_context['mock_dir'] + '/curl'
      File.open(echo_mock_file, 'w', 0o777) do |f|
        f.write("#!/bin/bash\n")
        f.write("echo #{expected_gcp_hostname}")
      end

      test_link = { 'cloud-provider' => {
        'instances' => [],
        'properties' => {
          'cloud-provider' => {
            'type' => 'gce',
          }
        }
      } }
      rendered_kube_proxy_ctl = compiled_template('kube-proxy', 'bin/kube_proxy_ctl', { 'cloud-provider' => 'gce' }, test_link)
      result = run_get_hostname_override(rendered_kube_proxy_ctl, test_context['kube_proxy_ctl_file'])

      expect(result).to include(expected_gcp_hostname)
    end
  end
end
