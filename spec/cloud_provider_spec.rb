# frozen_string_literal: true

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

  it 'exports GOOGLE_APPLICATION_CREDENTIALS in the cloud-provider_utils' do
    expect(rendered_template).to include('export GOOGLE_APPLICATION_CREDENTIALS')
  end
end
