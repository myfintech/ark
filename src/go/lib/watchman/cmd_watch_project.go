package watchman

// WatchProject Requests that the project containing the requested dir is watched for changes. Watchman will track all files and dirs rooted at the project path, and respond with the relative path difference between the project path and the requested dir.
// https://facebook.github.io/watchman/docs/cmd/watch-project.html
func (client *Client) WatchProject(opts WatchProjectOptions) (resp WatchProjectResponse, err error) {
	err = client.Exec(RawPDUCommand{"watch-project", opts.Directory}, &resp)
	return
}

// WatchProjectResponse
// https://facebook.github.io/watchman/docs/cmd/watch-project.html
type WatchProjectResponse struct {
	Version string `json:"version" mapstructure:"version"`
	Watch   string `json:"watch" mapstructure:"watch"`
	Watcher string `json:"watcher" mapstructure:"watcher"`
	RelPath string `json:"relative_path" mapstructure:"relative_path"`
}

// WatchProjectOptions
// https://facebook.github.io/watchman/docs/cmd/watch-project.html
type WatchProjectOptions struct {
	Directory string
}
