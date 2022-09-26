/**
 * args is an string that contains all remaining parameters passed to the ark CLI after termination characters `--`
 * example:
 *  ark run target name -- extra args
 * result:
 *  "extra args"
 */
export const args: string;
/**
 * argsList is an Array<string> that contains all remaining parameters passed to the ark CLI after termination characters `--`
 * example:
 *  ark run target name -- extra args
 * result:
 *  ["extra", "args"]
 */
export const argsList: string[];
/**
 * flags are options forwarded from the CLI.
 * Useful for control flow and conditional configuration loading.
 */
export const flags: {
  namespace: string;
  environment: string;
  ci: boolean;
  // NOT YET IMPLEMENTED
  // context: string
};
