# frozen_string_literal: true

require 'rspec'
require 'spec_helper'

describe 'kubo-dns-aliases' do
  let(:link_spec) do
    {
      'kube-apiserver' => {
        'address' => 'fake.kube-api-address',
        'instances' => []
      }
    }
  end
  let(:properties) { {} }
  let(:rendered_template) { compiled_template('kubo-dns-aliases', 'dns/aliases.json', properties, link_spec) }
  let(:aliases) { JSON.parse(rendered_template) }

  it 'aliases.json is rendered without error' do
    expect { JSON.parse(rendered_template) }.to_not raise_error
  end

  context 'master node' do
    it 'generates an alias for the master node' do
      expect(aliases.key?('master.cfcr.internal')).to be(true)
    end

    it 'sets the master node alias to be the wildcard dns name' do
      expect(aliases.fetch('master.cfcr.internal')).to eq(['*.kube-api-address'])
    end
  end
end
