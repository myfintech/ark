import * as shared from './shared'
//NOTE: this is using the same type as microservice

export type PluginOptions = shared.MicroserviceAppOptions

export type NsReaper = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const nsReaper = NsReaper({
 *      replicas: "1",
 *  })
 */
declare const plugin: NsReaper

export default plugin