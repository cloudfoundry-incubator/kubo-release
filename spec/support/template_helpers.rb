# frozen_string_literal: true

require 'bosh/template/renderer'
require 'yaml'

module TemplateHelpers
  include Bosh::Template::PropertyHelper

  def compiled_template(job_name, template_name, manifest_properties = {}, links = {}, network_properties = [], az = '')
    manifest = emulate_bosh_director_merge(job_name, manifest_properties, links, network_properties, az)
    renderer = Bosh::Template::Renderer.new(context: manifest)
    renderer.render("jobs/#{job_name}/templates/#{template_name}.erb")
  end

  # Trying to emulate bosh director Bosh::Director::DeploymentPlan::Job#extract_template_properties
  def emulate_bosh_director_merge(job_name, manifest_properties, links, network_properties, az)
    job_spec = YAML.load_file("jobs/#{job_name}/spec")
    spec_properties = job_spec['properties']

    default_property_values = {}
    spec_properties.each_pair do |name, definition|
      default_value = definition['default']
      copy_property(default_property_values, {}, name, default_value)
    end

    effective_properties = recursive_merge(default_property_values, manifest_properties)

    links.each do |link_name, contents|
      next unless link_name == 'cloud-provider'

      link_spec = YAML.load_file("jobs/#{link_name}/spec")
      link_spec_properties = link_spec['properties']

      default_link_property_values = {}
      link_spec_properties.each_pair do |name, definition|
        default_value = definition['default']
        copy_property(default_link_property_values, {}, name, default_value)
      end

      link_properties = recursive_merge(default_link_property_values, contents['properties'])
      contents['properties'] = link_properties
    end

    {
      'properties' => effective_properties,
      'networks' => network_properties,
      'links' => links,
      'az' => az
    }.to_json
  end

  def recursive_merge(first, other)
    first.merge(other) do |_, old_value, new_value|
      if old_value.class == Hash && new_value.class == Hash
        recursive_merge(old_value, new_value)
      else
        new_value
      end
    end
  end
end
