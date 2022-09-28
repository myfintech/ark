import * as shared from './shared'

export type PluginOptions = shared.BasicContainerOptions

export type Consul = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const consul = Consul({
 *      image: "",
 *      name: "",
 *      replicas: 1
 *  })
 */
declare const plugin: Consul

export default plugin