export type Operation = "encode" | "decode";

/**
 * base64 encode and decode strings
 * @param {string} str The string to encode/decode
 * @param {Operation} opt The operation: encode/decode. The default value is encode so can be omitted.
 * @returns {string} Depending of the direction can return an encode value or a decoded value.
 * @example
 *  const encodedValue = base64('encode', 'myValue') // bXlWYWx1ZQ==
 */
export function base64(str: string, opt?: Operation): string;

/**
 * json2string encode an object to a json stringify
 * @param {{[key: string]: any}} obj - The json object to stringify
 */
export function json2string(obj: { [key: string]: any }): string;
