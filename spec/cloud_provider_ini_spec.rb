# frozen_string_literal: true

require 'rspec'
require 'spec_helper'

describe 'cloud-provider-ini' do
  let(:link_spec) do
    {
      'cloud-provider' => {
        'instances' => [
          {
            'az' => 'z1'
          }
        ],
        'properties' => properties
      }
    }
  end
  let(:rendered_template) { compiled_template('kube-apiserver', 'config/cloud-provider.ini', {}, link_spec, instance_name: 'master') }

  context 'passes through the cloud-config options as expected' do
    let(:properties)   { { 'cloud-provider' => { 'type' => 'garbage' }, 'cloud-config' => cloud_config } }
    let(:cloud_config) { { 'Global' => { 'foo' => 'bar', 'thing' => 'another' }, 'Random' => { 'junk' => 'food' } } }

    it 'renders the cloud-config options into the ini format' do
      expect(rendered_template).to include('[Global]')
      expect(rendered_template).to include('foo="bar"')
      expect(rendered_template).to include('thing="another"')
      expect(rendered_template).to include('[Random]')
      expect(rendered_template).to include('junk="food"')
    end
  end

  context 'strings are quoted and escaped where necessary in INI' do
    let(:properties) { { 'cloud-provider' => { 'type' => 'garbage' }, 'cloud-config' => cloud_config } }
    let(:cloud_config) { { 'Global' => { 'foo' => 'bar#123\\"' } } }

    it 'quotes the value and escapes quotation marks and backslashes in the string' do
      expect(rendered_template).to include('foo="bar#123\\\\\""')
    end
  end

  context 'errors on non-string arguments for sub-parameters' do
    let(:properties)   { { 'cloud-provider' => { 'type' => 'garbage' }, 'cloud-config' => cloud_config } }
    let(:cloud_config) { { 'Global' => { 'foo' => ['bar', 'baz'] } } }

    it 'should raise a TypeError' do
      expect { rendered_template }.to raise_error(TypeError)
    end
  end

  context 'Azure renders to YAML' do
    let(:properties)   { { 'cloud-provider' => { 'type' => 'azure' }, 'cloud-config' => cloud_config } }
    let(:cloud_config) { { 'cloud' => 'AzurePublicCloud', 'tenantId' => 'tenant' } }

    it 'should render to yaml' do
      expect { YAML::parse(rendered_template) }.not_to raise_error()
    end
  end
end
