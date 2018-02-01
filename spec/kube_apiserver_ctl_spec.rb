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

  it 'has no security context deny when privileged containers are enabled' do
    rendered_kube_apiserver_ctl = compiled_template(
      'kube-apiserver',
      'bin/kube_apiserver_ctl',
      { 'allow_privileged' => true },
      link_spec
    )

    expect(rendered_kube_apiserver_ctl).to include(
      '--admission-control=DenyEscalatingExec,LimitRanger,' \
      'NamespaceExists,NamespaceLifecycle,ResourceQuota,' \
      'ServiceAccount,DefaultStorageClass'
    )
  end

  it 'denies security context when privileged containers are not enabled' do
    rendered_kube_apiserver_ctl = compiled_template(
      'kube-apiserver',
      'bin/kube_apiserver_ctl',
      {},
      link_spec
    )

    expect(rendered_kube_apiserver_ctl).to match(/--admission-control=.*,SecurityContextDeny/)
  end

  it 'has no http proxy when no proxy is defined' do
    rendered_kube_apiserver_ctl = compiled_template(
      'kube-apiserver',
      'bin/kube_apiserver_ctl',
      {},
      link_spec
    )

    expect(rendered_kube_apiserver_ctl).not_to include('export http_proxy')
    expect(rendered_kube_apiserver_ctl).not_to include('export https_proxy')
    expect(rendered_kube_apiserver_ctl).not_to include('export no_proxy')
  end

  it 'sets http_proxy when an http proxy is defined' do
    rendered_kube_apiserver_ctl = compiled_template(
      'kube-apiserver',
      'bin/kube_apiserver_ctl',
      {
        'http_proxy' => 'proxy.example.com:8090'
      },
      link_spec
    )

    expect(rendered_kube_apiserver_ctl).to include('export http_proxy=proxy.example.com:8090')
  end

  it 'sets https_proxy when an https proxy is defined' do
    rendered_kube_apiserver_ctl = compiled_template(
      'kube-apiserver',
      'bin/kube_apiserver_ctl',
      {
        'https_proxy' => 'proxy.example.com:8100'
      },
      link_spec
    )

    expect(rendered_kube_apiserver_ctl).to include('export https_proxy=proxy.example.com:8100')
  end

  it 'sets no_proxy when no proxy property is set' do
    rendered_kube_apiserver_ctl = compiled_template(
      'kube-apiserver',
      'bin/kube_apiserver_ctl',
      {
        'no_proxy' => 'noproxy.example.com,noproxy.example.net'
      },
      link_spec
    )

    expect(rendered_kube_apiserver_ctl).to include('export no_proxy=noproxy.example.com,noproxy.example.net')
    expect(rendered_kube_apiserver_ctl).to include('export NO_PROXY=noproxy.example.com,noproxy.example.net')
  end
end
