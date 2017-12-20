require 'rspec'
require 'spec_helper'

describe 'kubo-dns-aliases' do
  let(:link_spec) do
      {
        'kube-apiserver' => {
          'address' => 'fake.kube-api-address',
          'instances' => []
        },
        'etcd' => {
          'address' => 'fake-etcd-address',
          'properties' =>  { 'etcd' => { 'advertise_urls_dns_suffix' => 'dns-suffix'} },
          'instances' => [
            {
             'name' => 'etcd',
             'index' => 0,
             'address' => 'fake-etcd-address-0',
           },
           {
             'name' => 'etcd',
             'index' => 1,
             'address' => 'fake-etcd-address-1',
           }
          ]
        }
      }
  end
  let(:properties) { {} }
  let(:rendered_template) { compiled_template('kubo-dns-aliases', 'dns/aliases.json', properties, link_spec) }
  let(:aliases) { JSON.parse(rendered_template) }

  it 'aliases.json is rendered without error' do
    expect{ JSON.parse(rendered_template) }.to_not raise_error
  end

  context 'master node' do
    it 'generates an alias for the master node' do
      expect(aliases.has_key?('master.kubo')).to be(true)
    end

    it 'sets the master node alias to be the wildcard dns name' do
      expect(aliases.fetch('master.kubo')).to eq(['*.kube-api-address'])
    end
  end

  context 'etcd' do
    let(:expected_dns_suffix) { link_spec['etcd']['properties']['etcd']['advertise_urls_dns_suffix'] }
    let(:etcd_instance) { link_spec['etcd']['instances'].first }
    let(:expected_node_prefix) { "#{etcd_instance['name']}-#{etcd_instance['index']}" }

    it 'generates aliases for the etcd nodes' do
      expect(aliases.has_key?(expected_dns_suffix)).to be(true)
      expect(aliases.has_key?("#{expected_node_prefix}.#{expected_dns_suffix}")).to be(true)
    end

    it 'sets the global etcd alias to the array of etcd instances' do
      expect(aliases.fetch(expected_dns_suffix)).to eq(['fake-etcd-address-0', 'fake-etcd-address-1'])
    end

    it 'sets the node etcd alias to an array of etcd instance' do
      expect(aliases.fetch("#{expected_node_prefix}.#{expected_dns_suffix}")).to eq(['fake-etcd-address-0'])
    end
  end
end
