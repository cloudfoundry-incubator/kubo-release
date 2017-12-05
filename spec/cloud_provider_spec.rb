require 'rspec'
require 'spec_helper'

describe 'cloud-provider' do
  let(:rendered_template) do
    properties = {
      'cloud-provider' => {
        'gce' => {
          'service_key' => 'foo'
        }
      }
    }

    compiled_template('cloud-provider', 'bin/cloud-provider_utils', properties)
  end

  it 'does not apply the standard storage class by default' do
    expect(rendered_template).to include('export GOOGLE_APPLICATION_CREDENTIALS')
  end
end
