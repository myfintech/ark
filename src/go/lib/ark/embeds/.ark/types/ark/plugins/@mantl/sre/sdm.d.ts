import * as shared from './shared'

export type PluginOptions = shared.PluginOptions & {}

export type StrongDM = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const sdm = StrongDM({})
 */
declare const plugin: StrongDM

export default plugin
