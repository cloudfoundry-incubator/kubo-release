require 'rspec'
require 'spec_helper'

describe 'cloud-provider-ini' do
  let(:rendered_template) { compiled_template('cloud-provider', 'config/cloud-provider.ini', properties) }

  context 'if cloud provider is gce' do
    let(:properties) { {'cloud-provider' => { 'type' => 'gce', 'gce' => gce_config }} }
    let(:gce_config) { {'project-id' => 'fake-project-id', 'network-name' => 'fake-network-name', 'service_key' => 'foo', 'worker-node-tag' => 'fake-worker-node-tag'} }

    it 'renders the correct template for gce' do
      expect(rendered_template).to include('project-id=fake-project-id')
      expect(rendered_template).to include('network-name=fake-network-name')
      expect(rendered_template).to include('node-tags=fake-worker-node-tag')
    end

    it 'sets token-url to nil if service_key is set' do
      expect(rendered_template).to include('token-url=nil')
    end

    context 'if gce service key not defined' do
      let(:gce_config) { {'project-id' => 'fake-project-id', 'network-name' => 'fake-network-name', 'worker-node-tag' => 'fake-worker-node-tag'} }

      it 'does not set token-url to nil' do
        expect(rendered_template).not_to include('token-url=nil')
      end
    end
  end

  context 'if cloud provider is vsphere' do
    let(:properties) { {'cloud-provider' => { 'type' => 'vsphere', 'vsphere' => vsphere_config }} }
    let(:vsphere_config) do
      {
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
    end

    it 'renders the correct template for vsphere' do
      vsphere_config.each { |k,v| expect(rendered_template).to include("#{k}=#{v}") }
    end
  end

  context 'if cloud provider is openstack' do
    let(:properties) { {'cloud-provider' => { 'type' => 'openstack', 'openstack' => openstack_config }} }
    let(:required_openstack_config) { { 'auth-url' => 'fake-url', 'username' => 'fake-username', 'password' => 'fake-password', 'tenant-id' => 'fake-tenant-id' } }
    let(:optional_openstack_config) do
      {
         'tenant-name' => 'fake-tenant-name',
         'trust-id' => 'fake-trust-id',
         'domain-id' => 'fake-domain-id',
         'domain-name' => 'fake-domain-name',
         'region' => 'fake-region',
         'ca-file' => 'fake-ca-file'
      }
    end

    context 'required properties' do
      let(:openstack_config) { required_openstack_config }
      it 'renders the correct template for openstack' do
        openstack_config.each { |k,v| expect(rendered_template).to include("#{k}=#{v}") }
        optional_openstack_config.each { |k,v| expect(rendered_template).to_not include("#{k}=#{v}") }
      end

      context 'error handling' do
        let(:openstack_config) { required_openstack_config.delete 'auth-url' }
        it 'errors if a required property is not specified' do
          expect{rendered_template}.to raise_error(Bosh::Template::UnknownProperty, /Can't find property/ )
        end
      end
    end

    context 'optional properties' do
      let(:openstack_config) { required_openstack_config.merge optional_openstack_config }
      it 'renders the correct template for openstack' do
        openstack_config.each { |k,v| expect(rendered_template).to include("#{k}=#{v}") }
      end

      context 'error handling' do
        it 'does not error if an optional property is not specified' do
          expect{rendered_template}.to_not raise_error
        end
      end
    end
  end

  context 'if cloud provider is aws' do
    let(:properties) { {'cloud-provider' => { 'type' => 'aws', 'aws' => aws_config }} }

    context 'optional properties provided' do
      let(:aws_config) do
        {
          'vpc' => 'fake-vpc',
          'subnet-id' => 'fake-subnet-id',
          'route-table-id' => 'fake-route-table-id',
          'kubernetes-cluster-tag' => 'fake-kubernetes-cluster-tag',
          'kubernetes-cluster-id' => 'fake-kubernetes-cluster-id',
          'disable-security-group-ingress' => true,
          'elb-security-group' => 'fake-elb-security-group'
        }
      end

      let(:key_map) do
        {
          'vpc' => 'VPC',
          'subnet-id' => 'SubnetID',
          'route-table-id' => 'RouteTableID',
          'kubernetes-cluster-tag' => 'KubernetesClusterTag',
          'kubernetes-cluster-id' => 'KubernetesClusterID',
          'disable-security-group-ingress' => 'DisableSecurityGroupIngress',
          'elb-security-group' => 'ElbSecurityGroup'
        }
      end

      it 'renders the correct template for aws' do
        aws_config.each { |k,v| expect(rendered_template).to include("#{key_map[k]}=#{v}") }
      end
    end

    context 'optional properties not provided' do
      let(:aws_config) {{}}

      it 'renders the correct template for aws' do
        expect(rendered_template).to eq("\n[Global]\nDisableSecurityGroupIngress=false\n\n\n")
      end
    end
  end
end
