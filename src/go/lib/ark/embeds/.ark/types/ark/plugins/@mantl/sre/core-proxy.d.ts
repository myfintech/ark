import * as shared from './shared'

type PluginOptions = shared.BasicContainerOptions
    & Pick<shared.ServiceAccountOptions, 'serviceAccountName'>
    & Pick<shared.NetworkingOptions, 'port'>
    & Pick<shared.VaultOptions, 'vaultAddress'>
    & {
    envVars: { [key: string]: string }
    environment: string,
    loadBalancerIP: string,
    disableHostNetworking?: boolean,
    disableOpenTelemetry?: boolean,
    disableGoogleServiceAccount?: boolean,
    disableVaultSidecar?: boolean
}


export type CoreProxy = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const coreProxy = CoreProxy({
 *      image: "",
 *      name: "",
 *      replicas: 1
 *  })
 */
declare let plugin: CoreProxy

export default plugin;
