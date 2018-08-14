# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'yaml'

describe 'kube_controller_manager' do
  context 'horizontal pod autoscaling' do
    it 'sets the properties' do
      rendered_kube_controller_manager_bpm_yml = compiled_template(
        'kube-controller-manager',
        'config/bpm.yml',
        'horizontal-pod-autoscaler' => {
          'downscale-delay' => '2m0s',
          'upscale-delay' => '2m0s',
          'sync-period' => '40s',
          'tolerance' => '0.2',
          'use-rest-clients' => false
        }
      )

      bpm_yml = YAML.safe_load(rendered_kube_controller_manager_bpm_yml)
      expect(bpm_yml['processes'][0]['args']).to include('--horizontal-pod-autoscaler-downscale-delay=2m0s')
      expect(bpm_yml['processes'][0]['args']).to include('--horizontal-pod-autoscaler-upscale-delay=2m0s')
      expect(bpm_yml['processes'][0]['args']).to include('--horizontal-pod-autoscaler-sync-period=40s')
      expect(bpm_yml['processes'][0]['args']).to include('--horizontal-pod-autoscaler-tolerance=0.2')
      expect(bpm_yml['processes'][0]['args']).to include('--horizontal-pod-autoscaler-use-rest-clients=false')
    end
  end
  it 'has no http proxy when no proxy is defined' do
    rendered_kube_controller_manager_bpm_yml = compiled_template(
      'kube-controller-manager',
      'config/bpm.yml',
      {}
    )

    bpm_yml = YAML.safe_load(rendered_kube_controller_manager_bpm_yml)
    expect(bpm_yml['processes'][0]['env']).to be_nil
  end

  it 'sets http_proxy when an http proxy is defined' do
    rendered_kube_controller_manager_bpm_yml = compiled_template(
      'kube-controller-manager',
      'config/bpm.yml',
      'http_proxy' => 'proxy.example.com:8090'
    )

    bpm_yml = YAML.safe_load(rendered_kube_controller_manager_bpm_yml)
    expect(bpm_yml['processes'][0]['env']['http_proxy']).to eq('proxy.example.com:8090')
    expect(bpm_yml['processes'][0]['env']['HTTP_PROXY']).to eq('proxy.example.com:8090')
  end

  it 'sets https_proxy when an https proxy is defined' do
    rendered_kube_controller_manager_bpm_yml = compiled_template(
      'kube-controller-manager',
      'config/bpm.yml',
      'https_proxy' => 'proxy.example.com:8100'
    )

    bpm_yml = YAML.safe_load(rendered_kube_controller_manager_bpm_yml)
    expect(bpm_yml['processes'][0]['env']['https_proxy']).to eq('proxy.example.com:8100')
    expect(bpm_yml['processes'][0]['env']['HTTPS_PROXY']).to eq('proxy.example.com:8100')
  end

  it 'sets no_proxy when no proxy property is set' do
    rendered_kube_controller_manager_bpm_yml = compiled_template(
      'kube-controller-manager',
      'config/bpm.yml',
      'no_proxy' => 'noproxy.example.com,noproxy.example.net'
    )

    bpm_yml = YAML.safe_load(rendered_kube_controller_manager_bpm_yml)
    expect(bpm_yml['processes'][0]['env']['no_proxy']).to eq('noproxy.example.com,noproxy.example.net')
    expect(bpm_yml['processes'][0]['env']['NO_PROXY']).to eq('noproxy.example.com,noproxy.example.net')
  end
end
