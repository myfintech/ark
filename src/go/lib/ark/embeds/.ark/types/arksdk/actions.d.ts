import { RawArtifact, RawTarget} from "arksdk/interfaces";
import {
    DeployTarget,
    DockerImageTarget,
    GroupTarget,
    KubeExecTarget,
    LocalFileTarget,
    NixTarget,
    ProbeTarget,
    SyncKVTarget,
    TestTarget,
} from './targets'
import {
    DeployArtifact,
    DockerImageArtifact,
    GroupArtifact,
    KubeExecArtifact,
    LocalFileArtifact,
    NixArtifact,
    ProbeArtifact,
    SyncKVArtifact,
    TestArtifact,
} from './artifacts'

/**
 * A generic factory function type that construct Actions  given a target and an artifact
 * @param {T} target - The target options
 * @return {A} - Returns an artifact
 * @template T
 * @template A
 */
export type Action<T extends RawTarget, A extends RawArtifact> = (target: T) => A

export type BuildDockerImageAction = Action<DockerImageTarget, DockerImageArtifact>

export type DeployAction = Action<DeployTarget, DeployArtifact>

//not using generic because we exclude 'attributes'
export type GroupAction = (target: GroupTarget) => GroupTarget

export type KubeExecAction = Action<KubeExecTarget, KubeExecArtifact>

export type LocalFileAction = Action<LocalFileTarget, LocalFileArtifact>

export type NixAction = Action<NixTarget, NixArtifact>

export type ProbeAction = Action<ProbeTarget, ProbeArtifact>

export type SyncKVAction = Action<SyncKVTarget, SyncKVArtifact>

export type TestAction = Action<TestTarget, TestArtifact>


/**
 * buildDockerImage is an action that accepts a dockerImage target and produces dockerImage artifact as a result
 * this is typically passed to other images to reuse layers or used to deploy an application
 * @param  {DockerImageTarget>} target -
 * @returns {DockerImageArtifact>} -
 */
export const buildDockerImage: BuildDockerImageAction


/**
 * deploy is an action that accepts a deploy target and produces a reference to itself
 * this allows orchestrating the order of deployments if applications depend on one another
 * @param {DeployTarget} target -
 * @returns {DeployArtifact} -
 */
export const deploy: DeployAction

/**
 * group is an action that execute a group of targets
 * @param {GroupTarget} target -
 * @returns {GroupArtifact} -
 */
export const group: GroupAction

/**
 * kubeExec is an action that execute a command in already running pod in kubernetes
 * @param {KubeExecTarget} target -
 * @returns {KubeExecArtifact} -
 */
export const kubeExec: KubeExecAction

/**
 * localFile is an action that execute
 * @param {LocalFileTarget} target -
 * @returns {LocalFileArtifact} -
 */
export const localFile: LocalFileAction

/**
 * nix is an action that execute
 * @param {NixTarget} target -
 * @returns {NixArtifact} -
 */
export const nix: NixAction

/**
 * probe is an action that execute probe against pods in kubernetes
 * @param {ProbeTarget} target -
 * @returns {ProbeArtifact} -
 */
export const probe: ProbeAction

/**
 * syncKV is an action that triggers the synchronization of configuration store in source control
 * under the .ark/kv folder.
 * @param {SyncKVTarget} target -
 * @returns {SyncKVArtifact} -
 */
export const syncKV: SyncKVAction

/**
 * test is an action that execute a test for deployed resources in the k8s cluster
 * @param target {TestTarget} -
 * @returns {TestArtifact} -
 */
export const test: TestAction