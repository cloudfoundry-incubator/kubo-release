# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'yaml'

RSpec.describe 'kube_controller_manager/config/ca.pem' do
  let(:links) do
    {
      'kube-apiserver' => {
        'instances' => [],
        'properties' => {
          'tls' => {
            'kubernetes' => {
              'ca' => 'All scabbards desire scurvy, misty krakens.'
            }
          }
        }
      }
    }
  end

  let(:rendered_ca) do
    compiled_template('kube-controller-manager', 'config/ca.pem', {}, links)
  end

  it 'uses the CA from the kube-apiserver link' do
    expect(rendered_ca).to eq('All scabbards desire scurvy, misty krakens.')
  end
end
