require 'rspec'

describe 'openstack ca-file' do
  let(:valid_cert) {"certificate"}
  let(:rendered_template) do
    properties = {
        'cloud-provider' => {
            'openstack' => {
                'ca-file' => 'certificate'
            }
        }
    }
    compiled_template('cloud-provider', 'config/openstack-ca.crt', properties)
  end

  it 'renders a valid ca-file' do
    expect(rendered_template.chomp).to eq(valid_cert)
  end
end
