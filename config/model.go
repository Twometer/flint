package config

type Config struct {
	Ip        string            `toml:"ip"`
	Port      uint16            `toml:"port"`
	Messages  MessageCollection `toml:"messages"`
	Upstreams []Upstream        `toml:"upstream"`
}

type MessageCollection struct {
	ServerNotFound string `toml:"server_not_found"`
	ServerDown     string `toml:"server_down"`
	Maintenance    string `toml:"maintenance"`
}

type Upstream struct {
	Name    string `toml:"name"`
	Host    string `toml:"host"`
	Address string `toml:"address"`
}
