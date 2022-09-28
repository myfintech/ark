package watchman

// Watch Requests that the specified dir is watched for changes. Watchman will track all files and dirs rooted at the specified path.
// https://facebook.github.io/watchman/docs/cmd/watch.html
//
// Deprecated:starting in version 3.1. We recommend that clients adopt the watch-project command.
// https://facebook.github.io/watchman/docs/cmd/watch-project.html
func (client *Client) Watch(opts WatchOptions) (resp WatchResponse, err error) {
	err = client.Exec(RawPDUCommand{"watch", opts.Directory}, &resp)
	return
}

// WatchResponse
// https://facebook.github.io/watchman/docs/cmd/watch.html
// Deprecated: starting in version 3.1. We recommend that clients adopt the watch-project command.
// https://facebook.github.io/watchman/docs/cmd/watch-project.html
type WatchResponse struct {
	Version string `json:"version" mapstructure:"version"`
}

/*
WatchOptions
https://facebook.github.io/watchman/docs/cmd/watch.html
Deprecated: starting in version 3.1. We recommend that clients adopt the watch-project command.
https://facebook.github.io/watchman/docs/cmd/watch-project.html
*/
type WatchOptions struct {
	Directory string
}
