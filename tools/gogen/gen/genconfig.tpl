package config

type Server struct {
	Addr string `toml:"addr" default:":8080"`
}

type Config struct {
	Server Server `toml:"server"`
}
