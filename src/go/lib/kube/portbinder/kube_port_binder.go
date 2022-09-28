package portbinder

import (
	"strings"
)

// Binding is a representation of a local and remote port pair
type Binding struct {
	HostPort   string
	RemotePort string
}

// PortMap a map of keys and their port bindings
//
//	{
//		grpc: {
//			HostPort: 9000
//			RemotePort: 9000
//		}
//	}
type PortMap map[string]Binding

// ToPairs returns a slice of port binding pairs
func (pm *PortMap) ToPairs() (pairs []string) {
	for _, binding := range *pm {
		pairs = append(pairs, strings.Join([]string{
			binding.HostPort,
			binding.RemotePort,
		}, ":"))
	}
	return
}

// Selector is a key value pair used to query kubernetes for matching resources
type Selector struct {
	Namespace  string
	Type       string
	LabelKey   string
	LabelValue string
}

// BindPortCommand the command to send via the portbinder channel to instruct it to attach to a port
type BindPortCommand struct {
	PortMap  PortMap
	Selector Selector
}

// CommandChannel a channel which recieved bind port commands
type CommandChannel chan BindPortCommand

// PortMapFromSlice parses a slice of port bindings into a port map where the key is the index in the slice
