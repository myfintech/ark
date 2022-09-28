package watchman

// CheckCapabilities checks the capabilities of the watchman socket service using the version command
// returns an error if the required capabilities are not present
// https://facebook.github.io/watchman/docs/cmd/version.html
func (client *Client) CheckCapabilities(opts VersionOptions) (resp VersionResponse, err error) {
	err = client.Exec(RawPDUCommand{"version", RawPDUCommandOptions{
		"optional": opts.Optional,
		"required": opts.Required,
	}}, &resp)
	return
}

// VersionResponse
// https://facebook.github.io/watchman/docs/cmd/version.html
type VersionResponse struct {
	Version      string          `json:"version" mapstructure:"version"`
	Capabilities map[string]bool `json:"capabilities" mapstructure:"capabilities"`
}

// VersionOptions
// https://facebook.github.io/watchman/docs/cmd/version.html
type VersionOptions struct {
	Optional []string
	Required []string
}
