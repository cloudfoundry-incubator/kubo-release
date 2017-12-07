require 'rspec'
require 'spec_helper'
require 'yaml'

describe 'apply-specs' do
  let(:links) do
    {
        'kubernetes-api' => {
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
    compiled_template('apply-specs', 'config/ca.pem', {}, links)
  end

  let(:rendered_kubeconfig) do
    YAML.load(compiled_template('apply-specs', 'config/kubeconfig', {}, links))
  end

  let(:kubeconfig_user) { rendered_kubeconfig['users'][0] }


  it 'uses the CA from the kubernetes-api link' do
    expect(rendered_ca).to eq('All scabbards desire scurvy, misty krakens.')
  end

  it 'uses the admin name from the kubernetes-api link' do
    expect(rendered_kubeconfig['contexts'][0]['context']['user']).to eq('meatloaf')
    expect(kubeconfig_user['name']).to eq('meatloaf')
  end

  it 'uses the token from the kubernetes-api link' do
    expect(kubeconfig_user['user']['token']).to eq('madagascar-TEST')
  end

  it 'constructs the URL using the kubernetes-api link' do
    expect(rendered_kubeconfig['clusters'][0]['cluster']['server']).to eq('https://master.kubo:8443')
  end
end
