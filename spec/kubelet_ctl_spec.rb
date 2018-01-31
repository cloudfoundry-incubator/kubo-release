# frozen_string_literal: true

require 'rspec'
require 'spec_helper'

describe 'kubelet_ctl' do
  let(:rendered_template) { compiled_template('kubelet', 'bin/kubelet_ctl', {}, {}, {}, 'z1') }

  it 'labels the kubelet with its own az' do
    expect(rendered_template).to include(',failure-domain.beta.kubernetes.io/zone=z1')
  end
end
