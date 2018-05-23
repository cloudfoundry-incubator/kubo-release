# frozen_string_literal: true

require 'rspec'
require 'spec_helper'

describe 'kubelet-pre-start-script' do
  let(:link_spec) do
    {
      'cloud-provider' => {
        'instances' => [
          {
            'az' => 'z1'
          }
        ],
        'properties' => properties
      }
    }
  end

  let(:rendered_template) { compiled_template('kubelet', 'bin/pre-start', vsphere_config, link_spec) }

  context 'vsphere' do
    let(:properties) { { 'cloud-provider' => { 'type' => 'vsphere' }}}
    let(:vsphere_config) { { 'http_proxy' => '10.74.42.132:8888', 'https_proxy' => '10.74.42.132:4443', 'no_proxy' => '10.74.42.132:2222' }}

    context 'configured for an http proxy' do
      it 'renders the correct template for an http proxy on vsphere' do
        expect(rendered_template).to include('http_proxy=10.74.42.132:8888')
        expect(rendered_template).to include('HTTP_PROXY=10.74.42.132:8888')
      end
    end

    context 'configured for an https proxy' do
      it 'renders the correct template for an https proxy on vsphere' do
        expect(rendered_template).to include('https_proxy=10.74.42.132:4443')
        expect(rendered_template).to include('HTTPS_PROXY=10.74.42.132:4443')
      end
    end

    context 'configured for no proxy' do
      it 'renders the correct template for no proxy on vsphere' do
        expect(rendered_template).to include('no_proxy=10.74.42.132:2222')
        expect(rendered_template).to include('NO_PROXY=10.74.42.132:2222')
      end
    end
  end
end
