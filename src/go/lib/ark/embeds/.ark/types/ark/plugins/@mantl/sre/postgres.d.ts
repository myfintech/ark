import * as shared from './shared'

// NOTE: This one uses stateful app src/go/lib/kube/statefulapp/statefulapp.go
export type PluginOptions = shared.StatefulAppOptions


// const opts: Options = {
//     replicas: 1,
//     port: 5432,
//     servicePort: 5432,
//     image: "postgres:9",
//     serviceType: "ClusterIP",
//     env: {
//         "POSTGRES_PASSWORD": "domain",
//     },
// };

export type Postgres = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const postgres = Postgres({
 *      replicas: "1",
 *  })
 */
declare const plugin: Postgres

export default plugin
