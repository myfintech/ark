package watchman

// ListCapabilities lists the capabilities of the watchman socket service
// https://facebook.github.io/watchman/docs/cmd/list-capabilities.html
func (client *Client) ListCapabilities() (resp ListCapabilitiesResponse, err error) {
	err = client.Exec(RawPDUCommand{"list-capabilities"}, &resp)
	return
}

// ListCapabilitiesResponse
// https://facebook.github.io/watchman/docs/cmd/list-capabilities.html
type ListCapabilitiesResponse struct {
	Version      string   `json:"version" mapstructure:"version"`
	Capabilities []string `json:"capabilities" mapstructure:"capabilities"`
}

// HasCapability returns true if the capability is in the list
func (l ListCapabilitiesResponse) HasCapability(check string) bool {
	for _, capability := range l.Capabilities {
		if capability == check {
			return true
		}
	}
	return false
}
