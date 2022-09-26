/**
 * get returns a unencrypted secrets from the kv engine
 * @param {string} path The secret path to be reveal
 * @returns {{ [key: string]: any }} an json object containing the secret value
 * @example
 *  var dbPwd = get('path/to/secret') // returns 'my-super-secret'
 */
export function get(path: string): { [key: string]: any };
