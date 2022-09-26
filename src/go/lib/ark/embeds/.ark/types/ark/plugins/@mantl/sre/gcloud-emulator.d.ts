import * as shared from './shared'

export type PluginOptions = Pick<shared.ContainerOptions, 'name'> & {
    emulator: 'bigtable' | 'datastore' | 'firestore' | 'pubsub' | 'spanner',
    project: string,
};

export type GoogleCloudEmulator = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const gcloudEmulator = GoogleCloudEmulator({
 *      name: "firestore-emulator",
 *      emulator: "firestore",
 *      project: "01F8BBGYZR3TTGZ026XWTTNPWS"
 *  })
 */
declare const plugin: GoogleCloudEmulator

export default plugin
