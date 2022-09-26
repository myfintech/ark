/**
 * customTemplateDelimiters defaults to
 * ${} for objects and {% ~%} to match the liquid templating syntax
 */
export type customTemplateDelimiters = {
  objectLeft: string
  objectRight: string
  tagLeft?: string
  tagRight?: string
}

/**
 * load reads the contents of a file as a string
 * @param path
 */
export function load(path: string): string

/**
 * loadAsTemplate reads the contents of a file, executes it against Go's built in template engine,
 * and returns a string.
 * @param path The path to the file to be loaded as a template
 * @param vars {T} The variables injected into the template at runtime
 * @param allowNullVars when true template parsing will inject "<no value>" where template variables are null or nil
 * @param customDelimiters when defined overrides the default template engine delimiters
 * @template T
 */
export function loadAsTemplate<T>(path: string, vars: T, allowNullVars?: boolean, customDelimiters?: customTemplateDelimiters): string

/**
 Glob returns the names of all files matching pattern or null if there is no matching file.
 Glob ignores file system errors such as I/O errors reading directories. The only possible thrown error is ErrBadPattern, when pattern is malformed.
 The pattern may describe hierarchical names such as /usr/*\/bin/ed (assuming the Separator is '/').
 */
export function glob(...patterns: string[]): string[]

/**
 Join joins any number of path elements into a single path,
 separating them with an OS specific Separator. Empty elements
 are ignored. The result is Cleaned. However, if the argument
 list is empty or all its elements are empty, Join returns
 an empty string.
 */
export function join(...segment: string[]): string

/**
 * fromRoot returns a path to a file relative to the workspace root
 * @param path {String}
 */
export function fromRoot(path: string): string


/**
 * Return the folder name of the current file
 */
export function getFolderNameFromCurrentLocation(): string