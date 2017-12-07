require 'rspec'

describe 'service key' do
  let(:rendered_template) do
    properties = {
        'cloud-provider' => {
            'gce' => {
                'service_key' => { 'valid-json' => true }
            }
        }
    }

    compiled_template('cloud-provider', 'config/service_key.json', properties)
  end

  it 'renders a valid json' do
    expect(rendered_template.to_json).to eq("\"{\\\"valid-json\\\"=>true}\\n\"")
  end
end