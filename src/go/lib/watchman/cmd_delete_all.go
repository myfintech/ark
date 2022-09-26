package watchman

// DeleteAll Requests that the project containing the requested dir is watched for changes. Watchman will track all files and dirs rooted at the project path, and respond with the relative path difference between the project path and the requested dir.
// https://facebook.github.io/watchman/docs/cmd/watch-del-all.html
func (client *Client) DeleteAll() (resp DeleteAllResponse, err error) {
	err = client.Exec(RawPDUCommand{"watch-del-all"}, &resp)
	return
}

// DeleteAllResponse
// https://facebook.github.io/watchman/docs/cmd/watch-del-all.html
type DeleteAllResponse struct {
	Version string   `json:"version" mapstructure:"version"`
	Root    []string `json:"root" mapstructure:"root"`
}
