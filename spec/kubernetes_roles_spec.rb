# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'yaml'

describe 'kubernetes-roles' do
  let(:links) do
    {
      'kube-apiserver' => {
        'instances' => [],
        'properties' => {
          'tls' => {
            'kubernetes' => {
              'ca' => 'All scabbards desire scurvy, misty krakens.'
            }
          },
          'admin-username' => 'meatloaf',
          'admin-password' => 'madagascar-TEST',
          'port' => '2034'
        }
      }
    }
  end

  let(:rendered_ca) do
    compiled_template('kubernetes-roles', 'config/ca.pem', {}, links)
  end

  let(:rendered_kubeconfig) do
    YAML.safe_load(compiled_template('kubernetes-roles', 'config/kubeconfig', {}, links))
  end

  let(:kubeconfig_user) { rendered_kubeconfig['users'][0] }

  it 'uses the CA from the kube-apiserver link' do
    expect(rendered_ca).to eq('All scabbards desire scurvy, misty krakens.')
  end

  it 'uses the token from the kube-apiserver link' do
    expect(kubeconfig_user['user']['token']).to eq('madagascar-TEST')
  end

  it 'constructs the URL using the kube-apiserver link' do
    expect(rendered_kubeconfig['clusters'][0]['cluster']['server']).to eq('https://master.cfcr.internal:8443')
  end
end
