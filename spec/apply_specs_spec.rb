# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'yaml'

describe 'apply-specs' do
  let(:links) do
    {
      'kube-apiserver' => {
        'instances' => [],
        'properties' => {
          'tls' => {
            'kubernetes' => {
              'ca' => 'All scabbards desire scurvy, misty krakens.'
            }
          },
          'admin-username' => 'meatloaf',
          'admin-password' => 'madagascar-TEST',
          'port' => '2034'
        }
      }
    }
  end

  let(:rendered_ca) do
    compiled_template('apply-specs', 'config/ca.pem', {}, links)
  end

  let(:rendered_kubeconfig) do
    YAML.safe_load(compiled_template('apply-specs', 'config/kubeconfig', {}, links))
  end

  let(:kubeconfig_user) { rendered_kubeconfig['users'][0] }

  it 'uses the CA from the kube-apiserver link' do
    expect(rendered_ca).to eq('All scabbards desire scurvy, misty krakens.')
  end

  it 'uses the admin name from the kube-apiserver link' do
    expect(rendered_kubeconfig['contexts'][0]['context']['user']).to eq('meatloaf')
    expect(kubeconfig_user['name']).to eq('meatloaf')
  end

  it 'uses the token from the kube-apiserver link' do
    expect(kubeconfig_user['user']['token']).to eq('madagascar-TEST')
  end

  it 'constructs the URL using the kube-apiserver link' do
    expect(rendered_kubeconfig['clusters'][0]['cluster']['server']).to eq('https://master.cfcr.internal:8443')
  end

  context 'kubernetes-dashboard yaml' do
    context 'if authorization mode is set to abac' do
      let(:rendered_template) do
        properties = {
          'authorization-mode' => 'abac'
        }

        compiled_template('apply-specs', 'specs/kubernetes-dashboard.yml', properties)
      end

      it 'should include the kubconfig' do
        str = '- --kubeconfig=/var/vcap/jobs/kubelet/config/kubeconfig'

        expect(rendered_template).to include(str)
      end

      it 'should include the mountPath of kubconfig' do
        str = '- mountPath: /var/vcap/jobs/kubelet/config/'

        expect(rendered_template).to include(str)
      end
    end

    context 'if authorization mode is set to rbac' do
      let(:rendered_template) do
        properties = {
          'authorization-mode' => 'rbac'
        }

        compiled_template('apply-specs', 'specs/kubernetes-dashboard.yml', properties)
      end

      it 'should include a service account' do
        str = <<~FOO
          kind: ServiceAccount
          metadata:
            labels:
              k8s-app: kubernetes-dashboard
            name: kubernetes-dashboard
            namespace: kube-system
        FOO
        expect(rendered_template).to include(str)
      end
    end
  end

  let(:link_spec) { {} }
  let(:default_properties) do
    {
      'admin-password' => '1234'
    }
  end
  let(:rendered_deploy_specs) do
    compiled_template('apply-specs', 'bin/deploy-specs', default_properties, link_spec)
  end

  let(:rendered_errand_run) do
    compiled_template('apply-specs', 'bin/run', default_properties, link_spec)
  end

  it 'sets the post-deploy timeout to 1200 by default' do
    expect(rendered_errand_run).to include('TIMEOUT=1200')
  end

  context 'when errand run timeout is re-configured' do
    let(:default_properties) do
      {
        'admin-password' => '1234',
        'timeout-sec' => '1122'
      }
    end

    it 'overrides the default timeout' do
      expect(rendered_errand_run).to include('TIMEOUT=1122')
    end
  end

  it 'does not apply the standard storage class by default' do
    expect(rendered_deploy_specs).to_not include('apply_spec "storage-class-gce.yml"')
  end

  context 'on GCE' do
    let(:link_spec) do
      {
        'cloud-provider' => {
          'instances' => [],
          'properties' => {
            'cloud-provider' => {
              'type' => 'gce'
            }
          }
        }
      }
    end

    it 'applies the standard storage class' do
      expect(rendered_deploy_specs).to include('apply_spec "storage-class-gce.yml"')
    end
  end

  context 'on non-GCE' do
    let(:link_spec) do
      {
        'cloud-provider' => {
          'instances' => [],
          'properties' => {
            'cloud-provider' => {
              'type' => 'anything'
            }
          }
        }
      }
    end

    it 'does not apply the standard storage class' do
      expect(rendered_deploy_specs).to_not include('apply_spec "storage-class-gce.yml"')
    end
  end

  context 'on unspecified cloud-provider' do
    let(:link_spec) do
      {
        'cloud-provider' => {
          'instances' => [],
          'properties' => {}
        }
      }
    end

    it 'does not apply the standard storage class' do
      expect(rendered_deploy_specs).to_not include('apply_spec "storage-class-gce.yml"')
    end
  end
end
