import * as shared from './shared'

export type PluginOptions = shared.PluginOptions & {}

export type Kafka = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const kafka = Kafka({})
 */
declare const plugin: Kafka

export default plugin