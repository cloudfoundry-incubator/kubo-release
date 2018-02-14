# frozen_string_literal: true

require 'rspec'

describe 'service key' do
  let(:valid_json) { { 'valid-json' => true }.to_json }
  let(:link_spec) do
    {
      'cloud-provider' => {
        'instances' => [],
        'properties' => {
          'cloud-provider' => {
            'gce' => {
              'service_key' => valid_json
            }
          }
        }
      }
    }
  end

  let(:rendered_template) do
    compiled_template('kube-apiserver', 'config/service_key.json', {}, link_spec)
  end

  it 'renders a valid json' do
    expect(rendered_template.chomp).to eq(valid_json)
  end
end
