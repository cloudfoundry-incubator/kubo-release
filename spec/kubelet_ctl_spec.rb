# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'fileutils'
require 'open3'

describe 'kubelet_ctl' do
  let(:rendered_template) { compiled_template('kubelet', 'bin/kubelet_ctl', {}, {}, {}, 'z1', 'fake-bosh-ip', 'fake-bosh-id') }

  it 'labels the kubelet with its own az' do
    expect(rendered_template).to include(',failure-domain.beta.kubernetes.io/zone=z1')
  end

  it 'labels the kubelet with the spec ip' do
    expect(rendered_template).to include('spec.ip=fake-bosh-ip')
  end

  it 'labels the kubelet with the bosh id' do
    expect(rendered_template).to include(',bosh.id=fake-bosh-id')
  end

  it 'has no http proxy when no proxy is defined' do
    rendered_kubelet_ctl = compiled_template(
      'kubelet',
      'bin/kubelet_ctl',
      {}
    )

    expect(rendered_kubelet_ctl).not_to include('export http_proxy')
    expect(rendered_kubelet_ctl).not_to include('export https_proxy')
    expect(rendered_kubelet_ctl).not_to include('export no_proxy')
  end

  it 'sets http_proxy when an http proxy is defined' do
    rendered_kubelet_ctl = compiled_template(
      'kubelet',
      'bin/kubelet_ctl',
      'http_proxy' => 'proxy.example.com:8090'
    )

    expect(rendered_kubelet_ctl).to include('export http_proxy=proxy.example.com:8090')
  end

  it 'sets https_proxy when an https proxy is defined' do
    rendered_kubelet_ctl = compiled_template(
      'kubelet',
      'bin/kubelet_ctl',
      'https_proxy' => 'proxy.example.com:8100'
    )

    expect(rendered_kubelet_ctl).to include('export https_proxy=proxy.example.com:8100')
  end

  it 'sets no_proxy when no proxy property is set' do
    rendered_kubelet_ctl = compiled_template(
      'kubelet',
      'bin/kubelet_ctl',
      'no_proxy' => 'noproxy.example.com,noproxy.example.net'
    )

    expect(rendered_kubelet_ctl).to include('export no_proxy=noproxy.example.com,noproxy.example.net')
    expect(rendered_kubelet_ctl).to include('export NO_PROXY=noproxy.example.com,noproxy.example.net')
  end
end

def call_get_hostname_override(rendered_kubelet_ctl, executable_path)
  File.open(executable_path, 'w', 0o777) do |f|
    f.write(rendered_kubelet_ctl)
  end

  # exercise bash function by changing path for any necessary executables to our mocks in /tmp/mock/*
  cmd = format('PATH=%<dirname>s:%<env_path>s /bin/bash -c "source %<exe>s && get_hostname_override"',
               dirname: File.dirname(executable_path), env_path: ENV['PATH'], exe: executable_path)

  # capturing stderr (ignored) prevents expected warnings from showing up in test console
  result, = Open3.capture3(cmd)
  result
end

describe 'kubelet_ctl setting of --hostname-override property' do
  let(:test_context) do
    mock_dir = '/tmp/kubelet_mock'
    FileUtils.remove_dir(mock_dir, true)
    FileUtils.mkdir(mock_dir)
    kubelet_ctl_file = mock_dir + '/kubelet_ctl'

    { 'mock_dir' => mock_dir, 'kubelet_ctl_file' => kubelet_ctl_file }
  end
  after(:each) do
    FileUtils.remove_dir(test_context['mock_dir'], true)
  end

  describe 'when cloud-provider is NOT gce' do
    it 'sets hostname_override to IP address of container IP' do
      expected_spec_ip = '1111'
      rendered_kubelet_ctl = compiled_template('kubelet', 'bin/kubelet_ctl', { 'cloud-provider' => 'nonsense' }, {}, {}, 'az1', expected_spec_ip)
      result = call_get_hostname_override(rendered_kubelet_ctl, test_context['kubelet_ctl_file'])

      expect(result).to include(expected_spec_ip)
    end
  end

  describe 'when cloud-provider is gce' do
    it 'sets hostname_override to google container hostname' do
      expected_google_hostname = 'foohost'

      # mock out curl because this code path will try to use it.
      echo_mock_file = test_context['mock_dir'] + '/curl'
      File.open(echo_mock_file, 'w', 0o777) do |f|
        f.write("#!/bin/bash\n")
        f.write("echo #{expected_google_hostname}")
      end

      manifest_properties = {
        'cloud-provider' => 'gce'
      }

      test_link = {
        'cloud-provider' => {
          'instances' => [],
          'properties' => {
            'cloud-provider' => {
              'type' => 'gce',
              'gce' => {
                'project-id' => 'f',
                'network-name' => 'ff',
                'worker-node-tag' => 'fff',
                'service_key' => 'ffff'
              }
            }
          }
        }
      }
      rendered_kubelet_ctl = compiled_template('kubelet', 'bin/kubelet_ctl', manifest_properties, test_link)
      expect(rendered_kubelet_ctl).to include('cloud_provider="gce"')

      result = call_get_hostname_override(rendered_kubelet_ctl, test_context['kubelet_ctl_file'])
      expect(result).to include(expected_google_hostname)
    end
  end
end

context 'when cloud provider is vsphere' do
  it 'does not set cloud-config' do
    manifest_properties = {
      'cloud-provider' => 'vsphere'
    }

    test_link = {
      'cloud-provider' => {
        'instances' => [],
        'properties' => {
          'cloud-provider' => {
            'type' => 'vsphere',
            'vsphere' => {
              'user' => 'fake-user',
              'password' => 'fake-password',
              'server' => 'fake-server',
              'port' => 'fake-port',
              'insecure-flag' => 'fake-insecure-flag',
              'datacenter' => 'fake-datacenter',
              'datastore' => 'fake-datastore',
              'working-dir' => 'fake-working-dir',
              'vm-uuid' => 'fake-vm-uuid',
              'scsicontrollertype' => 'fake-scsicontrollertype'
            }
          }
        }
      }
    }
    rendered_kubelet_ctl = compiled_template('kubelet', 'bin/kubelet_ctl', manifest_properties, test_link)
    expect(rendered_kubelet_ctl).not_to include('--cloud-config')
    expect(rendered_kubelet_ctl).to include('cloud_provider="vsphere"')
  end
end

context 'when there is no cloud-provider link' do
  it 'does not set cloud options' do
    rendered_kubelet_ctl = compiled_template('kubelet', 'bin/kubelet_ctl', {}, {})
    expect(rendered_kubelet_ctl).not_to include('--cloud-config')
    expect(rendered_kubelet_ctl).not_to include('--cloud-provider')
  end
end
