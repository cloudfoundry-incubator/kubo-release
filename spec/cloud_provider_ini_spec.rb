require 'rspec'
require 'spec_helper'

describe 'cloud-provider-ini' do


  context 'if cloud provider is gce' do
    let(:rendered_template) do
      properties = {
          'cloud-provider' => {
              'type' => 'gce',
              'gce' => {
                  'project-id' => 'fake-project-id',
                  'network-name' => 'fake-network-name',
                  'service_key' => 'foo',
                  'worker-node-tag' => 'fake-worker-node-tag'
              }
          }
      }
      compiled_template('cloud-provider', 'config/cloud-provider.ini', properties)
    end

    it 'renders the correct template for gce' do
      expect(rendered_template).to include('project-id="fake-project-id"')
      expect(rendered_template).to include('network-name="fake-network-name"')
      expect(rendered_template).to include('node-tags="fake-worker-node-tag"')
    end

    it 'sets token-url to nil if service_key is set' do
      expect(rendered_template).to include('token-url=nil')
    end

    context 'if gce service key is empty' do
      let(:rendered_template) do
        properties = {
            'cloud-provider' => {
                'type' => 'gce',
                'gce' => {
                    'project-id' => 'fake-project-id',
                    'network-name' => 'fake-network-name',
                    'service_key' => '',
                    'worker-node-tag' => 'fake-worker-node-tag'
                }
            }
        }
        compiled_template('cloud-provider', 'config/cloud-provider.ini', properties)
      end

      it 'does not set token-url to nil' do
        expect(rendered_template).not_to include('token-url=nil')
      end
    end
  end

  context 'if cloud provider is vsphere' do
    let(:rendered_template) do
      properties = {
          'cloud-provider' => {
              'type' => 'vsphere',
              'vsphere' => {
                  'user' => 'fake-user',
                  'password' => 'fake-password',
                  'server' => 'fake-server',
                  'port' => 'fake-port',
                  'insecure-flag' => 'fake-insecure-flag',
                  'datacenter' => 'fake-datacenter',
                  'datastore' => 'fake-datastore',
                  'working-dir' => 'fake-working-dir',
                  'vm-uuid' => 'fake-vm-uuid',
                  'scsicontrollertype' => 'fake-scsicontrollertype'
              }
          }
      }

      compiled_template('cloud-provider', 'config/cloud-provider.ini', properties)
    end

    it 'renders the correct template for vsphere' do
      expect(rendered_template).to include('user="fake-user"')
      expect(rendered_template).to include('password="fake-password"')
      expect(rendered_template).to include('server="fake-server"')
      expect(rendered_template).to include('port="fake-port"')
      expect(rendered_template).to include('insecure-flag="fake-insecure-flag"')
      expect(rendered_template).to include('datacenter="fake-datacenter"')
      expect(rendered_template).to include('datastore="fake-datastore"')
      expect(rendered_template).to include('working-dir="fake-working-dir"')
      expect(rendered_template).to include('vm-uuid="fake-vm-uuid"')
      expect(rendered_template).to include('scsicontrollertype="fake-scsicontrollertype"')
    end
  end
end
