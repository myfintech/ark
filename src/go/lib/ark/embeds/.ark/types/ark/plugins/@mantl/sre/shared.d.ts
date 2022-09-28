/**
 * Base plugins options that configure the deployment
 */

export type VaultOptions = {
  vaultTeam: string;
  vaultApp: string;
  vaultEnv: string;
  clusterEnv: string;
  vaultRole: string;
  vaultDefaultConfig: string;
  vaultAddress: string;
  enableVault: boolean;
};

export type HostAndPort = {
  /**
  * Name or number of the port to access on the container.
  * Number must be in the range 1 to 65535.
  * Name must be an IANA_SVC_NAME.
  */
  port: number;
  /**
  * Host name to connect to, defaults to the pod IP. You probably want to set
  * "Host" in httpHeaders instead.
  */
  host?: string;
};

/** ExecAction describes a "run in container" action. */
export type ExecAction = {
  /** 
  * Command is the command line to execute inside the container, the working directory for the
  * command  is root ('/') in the container's filesystem. The command is simply exec'd, it is
  * not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use
  * a shell, you need to explicitly call out to that shell.
  * Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
  */
  command?: string[];
};

/** URIScheme identifies the scheme used for connection to a host for Get actions */
export type URIScheme =
  /** URISchemeHTTP means that the scheme used will be http:// */
  "HTTP"
  /** URISchemeHTTPS means that the scheme used will be https:// */
  | "HTTPS";

/** HTTPHeader describes a custom header to be used in HTTP probes */
export type HTTPHeader = {
  /** The header field name */
  name: string;
  /** The header field value */
  value: string;
};

/** HTTPGetAction describes an action based on HTTP Get requests. */
export type HTTPGetAction = HostAndPort & {
  /** Path to access on the HTTP server. */
  path?: string;
  /**
    * Scheme to use for connecting to the host.
    * Defaults to HTTP.
  */
  scheme?: URIScheme;
  /** Custom headers to set in the request. HTTP allows repeated headers. */
  httpHeaders?: HTTPHeader[];
};


/** TCPSocketAction describes an action based on opening a socket */
export type TCPSocketAction = HostAndPort & {};

/**
* Handler defines a specific action that should be taken
* TODO: pass structured data to these actions, and document that data here.
*/
export type Handler = {
  /**
  * One and only one of the following should be specified.
  * Exec specifies the action to take.
*/
  exec?: ExecAction;
  /**
  * HTTPGet specifies the http request to perform.
  */
  httpGet?: HTTPGetAction;
  /**
  * TCPSocket specifies an action involving a TCP port.
  * TCP hooks not yet supported
  *  TODO: implement a realistic TCP lifecycle hook
  */
  tcpSocket?: TCPSocketAction;
};

/**
* ProbeOptions describes a health check to be performed against a container to determine whether it is
* alive or ready to receive traffic.
*/
export type Probe =
  /** The action taken to determine the health of a container */
  Handler & {
    /**
    * Number of seconds after the container has started before liveness probes are initiated.
    * More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
    */
    initialDelaySeconds?: number;
    /**
     * Number of seconds after which the probe times out.
     * Defaults to 1 second. Minimum value is 1.
     * More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
    */
    timeoutSeconds?: number;
    /**
     * How often (in seconds) to perform the probe.
     * Default to 10 seconds. Minimum value is 1.
    */
    periodSeconds?: number;
    /**
    * Minimum consecutive successes for the probe to be considered successful after having failed.
    * Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
    */
    successThreshold?: number;
    /**
    * Minimum consecutive failures for the probe to be considered failed after having succeeded.
    * Defaults to 3. Minimum value is 1.
    */
    failureThreshold?: number;
  };

export type ContainerOptions = {
  name: string;
  image: string;
  command: string[];
  env: { [key: string]: string };
  replicas: number;
  readinessProbe?: Probe;
  livenessProbe?: Probe;
};

export type BasicContainerOptions = Exclude<
  ContainerOptions,
  "command" | "env"
>;

export type NetworkingOptions = {
  ingress: IngressOptions;
  hostNetwork: boolean;
  includeKubeletHostIp: boolean;
  port?: number;
  servicePort?: number;
};

export type ServiceAccountOptions = {
  enableGoogleCloudServiceAccount: boolean;
  serviceAccountName: string;
};

export type MicroserviceAppOptions = VaultOptions &
  NetworkingOptions &
  ServiceAccountOptions &
  ContainerOptions;

export type StatefulAppOptions = VaultOptions &
  NetworkingOptions &
  Exclude<ServiceAccountOptions, "serviceAccountName"> &
  ContainerOptions & { dataDir: string };

export type PluginOptions =
  | {}
  | BasicContainerOptions
  | StatefulAppOptions
  | MicroserviceAppOptions;

/**
 * K8s manifest string representation
 */
export type Manifest = string;

/**
 * Factory type for plugins
 * @template {PluginOptions} T
 * @param {T} opts - The plugin options
 * @returns {Manifest} - K8s manifest string representation
 */
export type Plugin<T extends PluginOptions> = (opts: T) => Manifest;

export type IngressOptions = {
  name: string;
  rules: IngressRule[];
  tls?: IntgresTLS[];
  annotations?: { [key: string]: string };
  googleGlobalIPName?: string;
};

/**
 * IngressRule represents the rules mapping the paths under a specified host to
 * the related backend services. Incoming requests are first evaluated for a host
 * match, then routed to the backend associated with the matching IngressRuleValue.
 */
export type IngressRule = {
  /**
   * Host is the fully qualified domain name of a network host, as defined by RFC 3986.
   * Note the following deviations from the "host" part of the
   * URI as defined in RFC 3986:
   * 1. IPs are not allowed. Currently an IngressRuleValue can only apply to
   *    the IP in the Spec of the parent Ingress.
   * 2. The `:` delimiter is not respected because ports are not allowed.
   *	  Currently the port of an Ingress is implicitly :80 for http and
   *	  :443 for https.
   * Both these may change in the future.
   * Incoming requests are matched against the host before the
   * IngressRuleValue. If the host is unspecified, the Ingress routes all
   * traffic based on the specified IngressRuleValue.
   *
   * Host can be "precise" which is a domain name without the terminating dot of
   * a network host (e.g. "foo.bar.com") or "wildcard", which is a domain name
   * prefixed with a single wildcard label (e.g. "*.foo.com").
   * The wildcard character '*' must appear by itself as the first DNS label and
   * matches only a single label. You cannot have a wildcard label by itself (e.g. Host == "*").
   * Requests will be matched against the Host field in the following way:
   * 1. If Host is precise, the request matches this rule if the http host header is equal to Host.
   * 2. If Host is a wildcard, then the request matches this rule if the http host header
   * is to equal to the suffix (removing the first label) of the wildcard rule.
   */
  host?: string;
}
/**
 * IngressRuleValue represents a rule to route requests for this IngressRule.
 * If unspecified, the rule defaults to a http catch-all. Whether that sends
 * just traffic matching the host to the default backend or all traffic to the
 * default backend, is left to the controller fulfilling the Ingress. Http is
 * currently the only supported IngressRuleValue.
 */ & IngressRuleValue;

/**
 * IngressRuleValue represents a rule to apply against incoming requests. If the
 * rule is satisfied, the request is routed to the specified backend. Currently
 * mixing different types of rules in a single Ingress is disallowed, so exactly
 * one of the following must be set.
 */
export type IngressRuleValue = {
  http?: HTTPIngressRuleValue;
};

/**
 * HTTPIngressRuleValue is a list of http selectors pointing to backends.
 * In the example: http://<host>/<path>?<searchpart> -> backend where
 * where parts of the url correspond to RFC 3986, this resource will be used
 * to match against everything after the last '/' and before the first '?'
 * or '#'.
 */
export type HTTPIngressRuleValue = {
  /**
   * A collection of paths that map requests to backends.
   * +listType=atomic
   */
  paths: HTTPIngressPath[];
};

// PathType represents the type of path referred to by a HTTPIngressPath.
export type PathType =
  // PathTypeExact matches the URL path exactly and with case sensitivity.
  | "Exact"
  /**
   * PathTypePrefix matches based on a URL path prefix split by '/'. Matching
   * is case sensitive and done on a path element by element basis. A path
   * element refers to the list of labels in the path split by the '/'
   * separator. A request is a match for path p if every p is an element-wise
   * prefix of p of the request path. Note that if the last element of the
   * path is a substring of the last element in request path, it is not a
   * match (e.g. /foo/bar matches /foo/bar/baz, but does not match
   * /foo/barbaz). If multiple matching paths exist in an Ingress spec, the
   * longest matching path is given priority.
   * Examples:
   * - /foo/bar does not match requests to /foo/barbaz
   * - /foo/bar matches request to /foo/bar and /foo/bar/baz
   * - /foo and /foo/ both match requests to /foo and /foo/. If both paths are
   *   present in an Ingress spec, the longest matching path (/foo/) is given
   *   priority.
   */
  | "Prefix"
  /**
   * PathTypeImplementationSpecific matching is up to the IngressClass.
   * Implementations can treat this as a separate PathType or treat it
   * identically to Prefix or Exact path types.
   */
  | "ImplementationSpecific";

/**
 * HTTPIngressPath associates a path with a backend. Incoming urls matching the
 * path are forwarded to the backend.
 */
export type HTTPIngressPath = {
  /**
   * Path is matched against the path of an incoming request. Currently it can
   * contain characters disallowed from the conventional "path" part of a URL
   * as defined by RFC 3986. Paths must begin with a '/'. When unspecified,
   * all paths from incoming requests are matched.
   */
  path?: string;
  /**
   * PathType determines the interpretation of the Path matching. PathType can
   * be one of the following values:
   * * Exact: Matches the URL path exactly.
   * * Prefix: Matches based on a URL path prefix split by '/'. Matching is
   *   done on a path element by element basis. A path element refers is the
   *   list of labels in the path split by the '/' separator. A request is a
   *   match for path p if every p is an element-wise prefix of p of the
   *   request path. Note that if the last element of the path is a substring
   *   of the last element in request path, it is not a match (e.g. /foo/bar
   *   matches /foo/bar/baz, but does not match /foo/barbaz).
   * * ImplementationSpecific: Interpretation of the Path matching is up to
   *   the IngressClass. Implementations can treat this as a separate PathType
   *   or treat it identically to Prefix or Exact path types.
   * Implementations are required to support all path types.
   */
  pathType?: PathType;
  /**
   * Backend defines the referenced service endpoint to which the traffic
   * will be forwarded to.
   */
  backend: IngressBackend;
};

// IngressBackend describes all endpoints for a given service and port.
export type IngressBackend = {
  /**
   * Service references a Service as a Backend.
   * This is a mutually exclusive setting with "Resource".
   */
  service?: IngressServiceBackend;
  /**
   * Resource is an ObjectRef to another Kubernetes resource in the namespace
   * of the Ingress object. If resource is specified, a service.Name and
   * service.Port must not be specified.
   * This is a mutually exclusive setting with "Service".
   */
  resource?: TypeLocalObjectReference;
};

// IngressServiceBackend references a Kubernetes Service as a Backend.
export type IngressServiceBackend = {
  /**
   * Name is the referenced service. The service must exist in
   * the same namespace as the Ingress object.
   */
  name: string;
  /**
   * Port of the referenced service. A port name or port number
   * is required for a IngressServiceBackend.
   */
  port?: ServiceBackendPort;
};

// ServiceBackendPort is the service port being referenced.
export type ServiceBackendPort = {
  /**
   * Name is the name of the port on the Service.
   * This is a mutually exclusive setting with "Number".
   */
  name?: string;
  /**
    // Number is the numerical port number (e.g. 80) on the Service.
    // This is a mutually exclusive setting with "Name".
    */
  number?: number;
};

/**
 * TypedLocalObjectReference contains enough information to let you locate the
 * typed referenced object inside the same namespace.
 */
export type TypeLocalObjectReference = {
  // APIGroup is the group for the resource being referenced.
  /**
   * If APIGroup is not specified, the specified Kind must be in the core API group.
   * For any other third-party types, APIGroup is required.
   */
  apiGroup: string;
  // Kind is the type of resource being referenced
  kind: string;
  // Name is the name of resource being referenced
  name: string;
};

// IngressTLS describes the transport layer security associated with an Ingress.
export type IntgresTLS = {
  /**
   * Hosts are a list of hosts included in the TLS certificate. The values in
   * this list must match the name/s used in the tlsSecret. Defaults to the
   * wildcard host setting for the loadbalancer controller fulfilling this
   * Ingress, if left unspecified.
   * +listType=atomic
   */
  hosts?: string[];
  /**
   * SecretName is the name of the secret used to terminate TLS traffic on
   * port 443. Field is left optional to allow TLS routing based on SNI
   * hostname alone. If the SNI host in a listener conflicts with the "Host"
   * header field used by an IngressRule, the SNI host is used for termination
   * and value of the Host header is used for routing.
   */
  secretName?: string;
};
