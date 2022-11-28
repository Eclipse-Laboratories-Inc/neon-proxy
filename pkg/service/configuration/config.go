package configuration

type Config struct {
	Name         string
	Storage      *ConfigStorageList
	IsConsoleApp bool
}

type ConfigStorageList struct {
	Postgres []string
}
