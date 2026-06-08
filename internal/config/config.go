package config

type Config struct {
	Routes map[string]Route
}

type Route struct {
	Backends     []string
	TrimPrefix   bool
	Protected    bool
	AllowedRoles []string
}
