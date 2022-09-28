import * as shared from './shared'

// NOTE: This one uses stateful app src/go/lib/kube/statefulapp/statefulapp.go
export type PluginOptions = shared.StatefulAppOptions

// const opts: PluginOptions = {
//     replicas: 1,
//     port: 8200,
//     servicePort: 8200,
//     image: 'gcr.io/[insert-google-project]/ark/dev/vault:438a183',
//     serviceType: 'ClusterIP',
// }

export type Vault = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const vault = Vault({
 *      replicas: "1",
 *  })
 */
declare const plugin: Vault

export default plugin
