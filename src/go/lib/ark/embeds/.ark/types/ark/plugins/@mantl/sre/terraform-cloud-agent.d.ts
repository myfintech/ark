import * as shared from './shared'

export type PluginOptions =
    Pick<shared.ContainerOptions, 'name' | 'replicas'> & {
    agentToken: string,
}

export type TerraformCloudAgent = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const terraformCloudAgent = TerraformCloudAgent({
 *      replicas: "1",
 *  })
 */
declare const plugin: TerraformCloudAgent

export default plugin