import * as shared from './shared'

// NOTE: This one uses stateful app src/go/lib/kube/statefulapp/statefulapp.go
export type PluginOptions = shared.StatefulAppOptions

// const opts: Options = {
//     replicas: 1,
//     port: 6379,
//     servicePort: 6379,
//     image: 'gcr.io/managed-infrastructure/mantl/redis-ark:latest',
//     serviceType: 'ClusterIP',
//     dataDir: 'data',
// }

export type Redis = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const redis = Redis({
 *      replicas: "1",
 *  })
 */
declare const plugin: Redis

export default plugin