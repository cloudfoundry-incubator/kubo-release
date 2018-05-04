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

  context 'if cloud provider is vsphere' do
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
    let(:properties) { { 'cloud-provider' => { 'type' => 'vsphere', 'vsphere' => vsphere_config } } }
    let(:vsphere_config) do
      {
        'user' => 'fake-user',
        'password' => 'fake-password',
        'server' => 'fake-server',
        'port' => 'fake-port',
        'insecure-flag' => 'fake-insecure-flag',
        'datacenter' => 'fake-datacenter'
      }
    end

    it 'renders the correct template for vsphere' do
      expect(rendered_template).to include(<<-EOF
  export GOVC_URL="fake-server:fake-port"
  export GOVC_USERNAME="fake-user"
  export GOVC_PASSWORD=$'fake-password'
  export GOVC_INSECURE="fake-insecure-flag"
  export GOVC_DATACENTER="fake-datacenter"
      EOF
                                          )
    end

    context 'and password has special characters' do
      it '- single quote [\']' do
        vsphere_config['password'] = "foo'bar"
        expect(rendered_template).to include(%q(export GOVC_PASSWORD=$'foo\'bar'))
      end
    end

    context 'and password has multiple special characters' do
      it 'has a special character in the rendered template' do
        vsphere_config['password'] = %q(#&;^\ !%'Az1)
        expect(rendered_template).to include(%q(export GOVC_PASSWORD=$'#&;^\ !%\'Az1'))
      end
    end
  end
end
