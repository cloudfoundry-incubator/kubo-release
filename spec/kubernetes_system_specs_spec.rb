# frozen_string_literal: true

require 'rspec'
require 'spec_helper'

describe 'kubernetes-system-specs' do
  let(:link_spec) { {} }
  let(:rendered_template) do
    properties = {
      'admin-password' => '1234'
    }
    links = link_spec

    compiled_template('kubernetes-system-specs', 'bin/deploy-specs', properties, links)
  end

  it 'does not apply the standard storage class by default' do
    expect(rendered_template).to_not include('apply_spec "storage-class-gce.yml"')
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
      expect(rendered_template).to include('apply_spec "storage-class-gce.yml"')
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
      expect(rendered_template).to_not include('apply_spec "storage-class-gce.yml"')
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
      expect(rendered_template).to_not include('apply_spec "storage-class-gce.yml"')
    end
  end

  context 'kubernetes-dashboard yaml' do
    context 'if authorization mode is set to abac' do
      let(:rendered_template) do
        properties = {
          'authorization-mode' => 'abac'
        }

        compiled_template('kubernetes-system-specs', 'config/kubernetes-dashboard.yml', properties)
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

        compiled_template('kubernetes-system-specs', 'config/kubernetes-dashboard.yml', properties)
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
end
