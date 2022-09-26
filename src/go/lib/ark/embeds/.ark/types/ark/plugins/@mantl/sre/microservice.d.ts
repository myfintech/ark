import * as shared from './shared'

// NOTE: this options are coming from src/go/lib/kube/microservice/microservice.go
export type PluginOptions = shared.MicroserviceAppOptions

export type Microservice = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const Microservice = Microservice({
 *      name: "",
 *      replicas: "1",
 *  })
 */
declare const plugin: Microservice

export default plugin;