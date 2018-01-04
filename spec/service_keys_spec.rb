# frozen_string_literal: true

require 'rspec'

describe 'service key' do
  let(:valid_json) { { 'valid-json' => true }.to_json }
  let(:rendered_template) do
    properties = {
      'cloud-provider' => {
        'gce' => {
          'service_key' => valid_json
        }
      }
    }
    compiled_template('cloud-provider', 'config/service_key.json', properties)
  end

  it 'renders a valid json' do
    expect(rendered_template.chomp).to eq(valid_json)
  end
end
