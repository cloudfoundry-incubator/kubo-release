require 'rspec'
require 'spec_helper'

describe 'kubernetes-system-specs' do
  let(:link_spec) { {} }
  let(:rendered_template) do
    properties = {
      'admin-password' => '1234'
    }
    links = link_spec

    compiled_template('kubernetes-system-specs', 'bin/deploy-specs', properties, links)
  end

  it 'does not apply the standard storage class by default' do
    expect(rendered_template).to_not include('apply_spec "storage-class-gce.yml"')
  end

  context 'on GCE' do
    let(:link_spec) do
      {
        'cloud-provider' => {
          'instances' => [],
          'properties' => {
            'cloud-provider' => {
              'type' => 'gce'
            }
          }
        }
      }
    end

    it 'applies the standard storage class' do
      expect(rendered_template).to include('apply_spec "storage-class-gce.yml"')
    end
  end

  context 'on non-GCE' do
    let(:link_spec) do
      {
        'cloud-provider' => {
          'instances' => [],
          'properties' => {
            'cloud-provider' => {
              'type' => 'anything'
            }
          }
        }
      }
    end

    it 'does not apply the standard storage class' do
      expect(rendered_template).to_not include('apply_spec "storage-class-gce.yml"')
    end
  end

  context 'on unspecified cloud-provider' do
    let(:link_spec) do
      {
        'cloud-provider' => {
          'instances' => [],
          'properties' => {}
        }
      }
    end

    it 'does not apply the standard storage class' do
      expect(rendered_template).to_not include('apply_spec "storage-class-gce.yml"')
    end
  end
end
