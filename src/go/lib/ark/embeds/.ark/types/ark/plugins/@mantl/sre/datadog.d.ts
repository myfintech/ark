import * as shared from './shared'

export type PluginOptions = Pick<shared.ContainerOptions, 'image'> & {
    imagePullPolicy: string,
    arkVersion: string,
    enableLogging: boolean,
    loggingInclude: string,
    loggingExclude: string
};

export type DataDog = shared.Plugin<Partial<PluginOptions>>

/**
 * Factory function that k8s manifest used to deploy a container into the cluster
 * @param {PluginOptions} opts - The plugin options
 * @returns {shared.Manifest} - A string representing a K8s YAML file
 * @example
 *  const dataDog = DataDog({
 *      image: "",
 *      loggingInclude: false,
 *      loggingInclude: false
 *  })
 */
declare const plugin: DataDog

export default plugin;