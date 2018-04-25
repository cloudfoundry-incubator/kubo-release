# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'yaml'

RSpec.describe 'ca.pem' do
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

  %w( kube-controller-manager apply-specs kubernetes-roles kube-proxy kubelet route-sync).each do |component|
    context "for #{component}" do
      let(:rendered_ca) do
        compiled_template(component, 'config/ca.pem', {}, links)
      end

      it 'uses the CA from the kube-apiserver link' do
        expect(rendered_ca).to eq('All scabbards desire scurvy, misty krakens.')
      end
    end
  end
end
