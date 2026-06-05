package config

type Config struct {
	Routes map[string]Route
}

type Route struct {
	Path       string
	Url        string
	TrimPrefix bool
	Protected  bool
}
