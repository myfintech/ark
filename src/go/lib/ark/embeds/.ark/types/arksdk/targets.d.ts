import { Attributes, RawTarget } from "./interfaces"; import { Manifest } from "ark/plugins/@ark/sre/shared";

export type PortMap = {
  [key: string]: {
    remotePort: string
    hostPort: string
  }
}

export type Step = {
  command: string[]
  workDir: string
  patterns: string[]
}

/**
 * Represents the target state of a k8s deployment
 */
export type DeployTargetAttributes = Attributes & {
  env?: { [key: string]: string };
  liveSyncEnabled?: boolean;
  liveSyncRestartMode?: string;
  liveSyncOnStep?: Step[];
  manifest: Manifest;
  portForward?: PortMap;
};

/**
 * Represents the target state of building a docker image
 */
export type DockerImageTargetAttributes = Attributes & {
  cacheFrom?: string[];
  cacheInline?: boolean;
  disableEntrypointInjection?: boolean;
  dockerfile: string;
  output?: string;
  repo: string;
  secrets?: string[];
};

/**
 * Represents the target state of executing a command in a k8s container
 */
export type KubeExecTargetAttributes = Attributes & {
  command: string[];
  containerName: string;
  resourceName: string;
  resourceType: string;
  timeoutSeconds: number;
};

export type LocalFileTargetAttributes = Attributes & {
  filename: string;
  content: string;
};

/**
 * Represents the target state of
 */
export type NixTargetAttributes = Attributes & {
  packages: string[];
};

/**
 * Represents the target state of
 */
export type ProbeTargetAttributes = Attributes & {
  delay: string;
  dialAddress: string;
  expectedStatus: number;
  maxRetries: number;
  timeout: string;
};

/**
 * Represents the target state of
 */
export type SyncKVTargetAttributes = Attributes & {
  engine: string;
  engineURL: string;
  token: string;
  timeoutSeconds?: number;
  maxRetries?: number;
};

/**
 * Represents the target state of
 */
export type TestTargetAttributes = Attributes & {
  args: string[];
  command: string[];
  environment: { [key: string]: string };
  image: string;
  timeoutSeconds: number;
  workingDirectory: string;
  disableCleanup?: boolean;
};

export type DeployTarget = RawTarget<DeployTargetAttributes>;

export type DockerImageTarget = RawTarget<DockerImageTargetAttributes>;

export type GroupTarget = Omit<RawTarget, "attributes">;

export type KubeExecTarget = RawTarget<KubeExecTargetAttributes>;

export type LocalFileTarget = RawTarget<LocalFileTargetAttributes>;

export type NixTarget = RawTarget<NixTargetAttributes>;

export type ProbeTarget = RawTarget<ProbeTargetAttributes>;

export type SyncKVTarget = RawTarget<SyncKVTargetAttributes>;

export type TestTarget = RawTarget<TestTargetAttributes>;
