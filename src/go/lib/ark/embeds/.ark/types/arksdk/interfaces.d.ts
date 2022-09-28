/**
 * Represent the structure of an ancestor
 */
export type Ancestor = {
  key: string;
  hash: string;
};

export type ExcludeFromHash = {
  sourceFiles: string[];
};

/**
 * Represent attributes for RawTarget and RawArtifact that are injected as a parametrized type
 * @example
 *  type TestTargetAttributes = Attributes & {
 *   command: string[],
 *   environment: { [key: string]: string }
 *   image: string,
 *   timeoutSeconds: number
 *   workingDirectory: string,
 * }
 */
export type Attributes = Object;

/**
 * Represent the absence of attributes for a RawTarget or RawArtifact
 * @example
 *  type GroupTarget = RawTarget<EmptyAttributes>
 */
export type EmptyAttributes = Attributes;

/**
 * A generic factory function type that construct a RawTarget which attributes are parametrized
 * @template T
 * @example
 *   type TestTargetAttributes = {
 *       command: string[],
 *       environment: { [key: string]: string }
 *       image: string,
 *       timeoutSeconds: number
 *       workingDirectory: string,
 *   }
 *  type DeployTarget = RawTarget<DeployTargetAttributes>
 */
export type RawTarget<T extends Attributes = Attributes> = {
  name: string;
  attributes: T;
  sourceFiles?: string[];
  dependsOn?: Ancestor[];
  excludeFromHash?: ExcludeFromHash;
  ignoreFileNotExistsError?: boolean;
};

/**
 * A generic factory function type that construct a RawArtifact which attributes are parametrized
 * @example
 *   type DeployArtifactAttributes = {
 *       workingDirectory: string,
 *   }
 *  type DeployArtifact = RawTarget<DeployArtifactAttributes>
 */
export type RawArtifact<T extends Attributes = Attributes> = Ancestor & {
  attributes: T;
  dependsOn?: Ancestor[];
};

/**
 * Represents the structure of a graph edge
 */
export type GraphEdge = {
  src: string;
  dst: string;
};
