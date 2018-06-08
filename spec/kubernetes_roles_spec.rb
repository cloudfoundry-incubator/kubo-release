# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'yaml'

describe 'kubernetes-roles' do
  let(:link_disallow_privileged) do
    {
      'kube-apiserver' => {
        'address' => 'fake.kube-api-address',
        'instances' => [],
        'properties' => {
          'allow_privileged' => false
        }
      }
    }
  end

  let(:link_allow_privileged) do
    {
      'kube-apiserver' => {
        'address' => 'fake.kube-api-address',
        'instances' => [],
        'properties' => {
          'allow_privileged' => true
        }
      }
    }
  end

  it 'disallows privileged containers when allow_privileged is false' do
    rendered_k8s_role_psp = compiled_template(
      'kubernetes-roles',
      'config/policies/podsecuritypolicy.yml',
      {},
      link_disallow_privileged)

    psp = YAML.safe_load(rendered_k8s_role_psp)
    expect(psp['spec']['privileged']).to equal(false)
  end

  it 'allows privileged containers escalation when allow_privileged is true' do
    rendered_k8s_role_psp = compiled_template(
      'kubernetes-roles',
      'config/policies/podsecuritypolicy.yml',
      {},
      link_allow_privileged)

    psp = YAML.safe_load(rendered_k8s_role_psp)
    expect(psp['spec']['privileged']).to equal(true)
  end
end
