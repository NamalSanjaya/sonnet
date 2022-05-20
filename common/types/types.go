package types

type DataUnit struct {
	Id         string
	InsertedAt string
	Data       []byte
	Size       int // size in byte
}

type DbConfig struct {
	Schema, Username, Password, Server, Database string
	Port                                         int
}
