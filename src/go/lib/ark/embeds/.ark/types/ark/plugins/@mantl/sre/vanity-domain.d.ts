import * as shared from './shared'

export type PluginOptions = Pick<shared.ContainerOptions, 'name' | 'image' | 'replicas'> &
    shared.VaultOptions & {
        address: string,
        port: number,
};

export type VDS = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that produces a K8s manifest used to deploy a container into a cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const vds = VDS({
 *      name: "vds",
 *      image: "example-image:latest",
 *      replicas: 3,
 *      vaultEnv: "integration",
 *      clusterEnv: "int"
 *      vaultRole: "vanity-domain-int"
 *  })
 */
declare const plugin: VDS

export default plugin
