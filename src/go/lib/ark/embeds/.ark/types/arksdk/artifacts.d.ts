import {Attributes, RawArtifact} from './interfaces'

/**
 * dockerImage artifacts are a content addressable container image that is remotely cacheable in a container registry
 */
type DockerImageArtifactAttributes = Attributes & {
    url: string
}

/**
 * NixArtifactAttributes artifacts are a content addressable container
 */
export type NixArtifactAttributes = Attributes & {
    packages: string[]
}

/**
 * LocalFileArtifactAttributes contain the contents of the rendered local file
 */
export type LocalFileArtifactAttributes = Attributes & {
    renderedFilePath: string
}


export type DeployArtifact = RawArtifact

export type DockerImageArtifact = RawArtifact<DockerImageArtifactAttributes>

export type GroupArtifact = RawArtifact

export type KubeExecArtifact = RawArtifact

export type LocalFileArtifact = RawArtifact<LocalFileArtifactAttributes>

export type NixArtifact = RawArtifact<NixArtifactAttributes>

export type ProbeArtifact = RawArtifact

export type SyncKVArtifact = RawArtifact

export type TestArtifact = RawArtifact


