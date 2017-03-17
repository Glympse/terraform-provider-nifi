package nifi

// Config is the structure that stores the configuration to talk to a
// NiFi API compatible host.
type Config struct {
	Host               string
	ApiPath            string
}
