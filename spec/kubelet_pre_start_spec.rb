# frozen_string_literal: true

require 'rspec'
require 'spec_helper'

describe 'kubelet-pre-start-script' do
  let(:rendered_template) { compiled_template('kubelet', 'bin/pre-start', {}, link_spec) }
  let(:properties) { {} }
  let(:link_spec) { {} }

  it 'renders' do
    expect(rendered_template).to_not include('GOVC')
  end
end
