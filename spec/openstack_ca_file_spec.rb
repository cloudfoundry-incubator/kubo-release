# frozen_string_literal: true

require 'rspec'
require 'spec_helper'

describe 'openstack ca-file' do
  let(:valid_cert) { 'certificate' }
  let(:link_spec) do
    {
      'cloud-provider' => {
        'instances' => [],
        'properties' => {
          'cloud-provider' => {
            'openstack' => {
              'ca-file' => 'certificate'
            }
          }
        }
      }
    }
  end

  let(:rendered_template) do
    compiled_template('kube-apiserver', 'config/openstack-ca.crt', {}, link_spec)
  end

  it 'renders a valid ca-file' do
    expect(rendered_template.chomp).to eq(valid_cert)
  end
end
