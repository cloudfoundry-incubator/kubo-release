# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'yaml'

describe 'flag_generation_tests' do

  k8s_args = {
    'k8s-args' => {
      'hash': {
        'key1' => 'value1',
        'key2' => 'value2'
      },
      'array' => [ 'value1', 'value2' ],
      'true' => true,
      'false' => false,
      'string' => "value",
      'flagNil' => nil,
      'colonSuffix' => "value:"
    }
  }

  def test_bpm(template)
    yaml = YAML.safe_load(template)
    expect(yaml['processes'][0]['args']).to include('--hash=key1=value1,key2=value2')
    expect(yaml['processes'][0]['args']).to include('--array=value1,value2')
    expect(yaml['processes'][0]['args']).to include('--true=true')
    expect(yaml['processes'][0]['args']).to include('--false=false')
    expect(yaml['processes'][0]['args']).to include('--string=value')
    expect(yaml['processes'][0]['args']).to include('--flagNil')
    expect(yaml['processes'][0]['args']).to include('--colonSuffix=value:')
  end

  def test_ctl(template)
    expect(template).to include('--hash=key1=value1,key2=value2')
    expect(template).to include('--array=value1,value2')
    expect(template).to include('--true=true')
    expect(template).to include('--false=false')
    expect(template).to include('--string=value')
    expect(template).to include('--flagNil')
    expect(template).to include('--colonSuffix=value:')
  end

  context 'kube-controller-manager' do
    it 'passes through args correctly' do
      kube_controller_manager = compiled_template(
        'kube-controller-manager',
        'config/bpm.yml',
        k8s_args)

      test_bpm(kube_controller_manager)
    end
  end

  context 'kube-apiserver' do
    let(:link_spec) do
      {
        'kube-apiserver' => {
          'address' => 'fake.kube-api-address',
          'instances' => [],
          'properties' => {}
        },
        'etcd' => {
          'address' => 'fake-etcd-address',
          'properties' => { 'etcd' => { 'advertise_urls_dns_suffix' => 'dns-suffix' } },
          'instances' => [
            {
              'name' => 'etcd',
              'index' => 0,
              'address' => 'fake-etcd-address-0'
            },
            {
              'name' => 'etcd',
              'index' => 1,
              'address' => 'fake-etcd-address-1'
            }
          ]
        }
      }
    end

    it 'passes through args correctly' do
      kube_apiserver = compiled_template(
        'kube-apiserver',
        'config/bpm.yml',
        k8s_args,
        link_spec)

      test_bpm(kube_apiserver)
    end
  end

  context 'kubelet' do
    it 'passes through args correctly' do
      kubelet = compiled_template(
        'kubelet',
        'bin/kubelet_ctl',
        k8s_args)

      test_ctl(kubelet)
    end
  end
end
