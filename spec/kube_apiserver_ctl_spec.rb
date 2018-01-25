# frozen_string_literal: true

require 'rspec'
require 'spec_helper'
require 'yaml'

describe 'kube-apiserver' do
  let(:link_spec) do
    {
      'kube-apiserver' => {
        'address' => 'fake.kube-api-address',
        'instances' => []
      },
      'etcd' => {
        'address' => 'fake-etcd-address',
        'properties' => { 'etcd' => { 'advertise_urls_dns_suffix' => 'dns-suffix' } },
        'instances' => [
          {
            'name' => 'etcd',
            'index' => 0,
            'address' => 'fake-etcd-address-0'
          },
          {
            'name' => 'etcd',
            'index' => 1,
            'address' => 'fake-etcd-address-1'
          }
        ]
      }
    }
  end

  it 'configures audit logging by default' do
    rendered_kube_apiserver_ctl = compiled_template(
      'kube-apiserver',
      'bin/kube_apiserver_ctl',
      {},
      link_spec
    )

    expect(rendered_kube_apiserver_ctl).to include('--audit-log-path="$AUDIT_LOG_FILE"')
    expect(rendered_kube_apiserver_ctl).to include('--audit-log-maxage=0')
    expect(rendered_kube_apiserver_ctl).to include('--audit-log-maxsize=0')
    expect(rendered_kube_apiserver_ctl).to include('--audit-log-maxbackup=0')
    expect(rendered_kube_apiserver_ctl).to include('--audit-policy-file="$CONFIG_DIR/audit_policy.yml"')
  end

  it 'does not configure audit logging when enable audit logs is false' do
    rendered_kube_apiserver_ctl = compiled_template(
      'kube-apiserver',
      'bin/kube_apiserver_ctl',
      { 'enable_audit_logs' => false },
      link_spec
    )

    expect(rendered_kube_apiserver_ctl).to_not include('--audit-log-path="$AUDIT_LOG_FILE"')
    expect(rendered_kube_apiserver_ctl).to_not include('--audit-log-maxage=0')
    expect(rendered_kube_apiserver_ctl).to_not include('--audit-log-maxsize=0')
    expect(rendered_kube_apiserver_ctl).to_not include('--audit-log-maxbackup=0')
    expect(rendered_kube_apiserver_ctl).to_not include('--audit-policy-file="$CONFIG_DIR/audit_policy.yml"')
  end

  it 'configures audit logging when enable audit logs is true' do
    rendered_kube_apiserver_ctl = compiled_template(
      'kube-apiserver',
      'bin/kube_apiserver_ctl',
      { 'enable_audit_logs' => true },
      link_spec
    )

    expect(rendered_kube_apiserver_ctl).to include('--audit-log-path="$AUDIT_LOG_FILE"')
    expect(rendered_kube_apiserver_ctl).to include('--audit-log-maxage=0')
    expect(rendered_kube_apiserver_ctl).to include('--audit-log-maxsize=0')
    expect(rendered_kube_apiserver_ctl).to include('--audit-log-maxbackup=0')
    expect(rendered_kube_apiserver_ctl).to include('--audit-policy-file="$CONFIG_DIR/audit_policy.yml"')
  end
end
