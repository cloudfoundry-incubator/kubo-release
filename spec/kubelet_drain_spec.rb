# frozen_string_literal: true

require 'rspec'
require 'spec_helper'

def get_k8s_disks_dir(rendered_drain)
  lsblk_line = rendered_drain.split("\n").select { |line| line[/MOUNTPOINT/i] }
  expect(lsblk_line.length).to be(1)
  lsblk_line[0].match(/awk '\/(.+)\/ {print \$1}'/).captures[0]
end

describe 'kubelet drain' do

  it 'substitutes in the k8s-args.root-dir property for get_k8s_disks()' do
    manifest_properties = {
      'k8s-args' => {
        'root-dir' => '/var/vcap/data/kubelet'
      }
    }

    rendered_drain = compiled_template('kubelet', 'bin/drain', manifest_properties, {}, {}, nil, nil, nil)
    k8s_disks_dir = get_k8s_disks_dir(rendered_drain)
    expect(k8s_disks_dir).to eq(manifest_properties['k8s-args']['root-dir'].gsub("/", "\\/"))
  end

  it 'uses /var/lib/kubelet for get_k8s_disks() by default' do
    rendered_drain = compiled_template('kubelet', 'bin/drain', {}, {}, {}, nil, nil, nil)
    k8s_disks_dir = get_k8s_disks_dir(rendered_drain)
    expect(k8s_disks_dir).to eq("/var/lib/kubelet".gsub("/", "\\/"))
  end
end
