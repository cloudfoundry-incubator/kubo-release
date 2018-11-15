# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'yaml'

describe 'Kubernetes Post Start Custom Specs' do

  let(:manifest_properties) do
    {
      'post-start-custom-specs' => [
        {
          'name' => 'psp1',
          'value' => %q{
---
apiVersion: "no"
kind: nope
---
apiVersion: beta
kind: alwaysBe
}
        },
        {
          'name' => 'psp2',
          'value' => %q{
---
apiVersion: ye
kind: YAML is weird
}
        }
      ]
    }
  end

  it 'renders custom specs as yaml' do
    expected_psp = [
      {'apiVersion' => 'no', 'kind' => 'nope'},
      {'apiVersion' => 'beta', 'kind' => 'alwaysBe'},
      {'apiVersion' => 'ye', 'kind' => 'YAML is weird'}
    ]
    rendered_kubernetes_roles_yml = compiled_template(
      'kubernetes-roles',
      'config/policies/post-start-custom-specs.yml',
      manifest_properties,
      {}
    )
    specs_hash = YAML.load_stream(rendered_kubernetes_roles_yml).reject(&:nil?)
    expect(expected_psp).to eq(specs_hash)
  end

  it 'handles empty post-start-custom-specs property' do
    rendered_kubernetes_roles_yml = compiled_template(
      'kubernetes-roles',
      'config/policies/post-start-custom-specs.yml',
      {},
      {}
    )
    specs_hash = YAML.load_stream(rendered_kubernetes_roles_yml).reject(&:nil?)
    expect(specs_hash).to be_empty
  end
end
