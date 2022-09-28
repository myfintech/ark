import * as shared from './shared';

/**
 * GitHub Action Runner plugins options that configure the deployment
 */
export type PluginOptions =
    shared.BasicContainerOptions & Pick<shared.ServiceAccountOptions, 'serviceAccountName'>

export type GitHubActionRunner = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const ghaRunner = GitHubActionRunner({
 *      image: "",
 *      name: "",
 *      serviceAccountName: "",
 *      replicas: 1
 *  })
 */
declare const plugin: GitHubActionRunner

export default plugin