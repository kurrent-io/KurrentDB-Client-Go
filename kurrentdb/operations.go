package kurrentdb

// ServerVersion Represents the version of a KurrentDB node.
type ServerVersion struct {
	Major int
	Minor int
	Patch int
}

// GetServerVersion Returns the version of the KurrentDB node to which the client is currently connected.
func (client *Client) GetServerVersion() (*ServerVersion, error) {
	handle, err := client.grpcClient.getConnectionHandle()
	if err != nil {
		return nil, err
	}

	return handle.GetServerVersion()
}

func (v *ServerVersion) IsAtLeast(major, minor, patch int) bool {
	if v.Major > major {
		return true
	}
	if v.Major == major {
		if v.Minor > minor {
			return true
		}
		if v.Minor == minor {
			return v.Patch >= patch
		}
	}
	return false
}

func (v *ServerVersion) IsBelow(major, minor, patch int) bool {
	if v.Major < major {
		return true
	}
	if v.Major == major {
		if v.Minor < minor {
			return true
		}
		if v.Minor == minor {
			return v.Patch < patch
		}
	}
	return false
}
