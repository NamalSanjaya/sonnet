package datasource2

// DS2 history table metadata , don't change the order
type HistTbMetadata struct {
	UserId string
	Lastmsg, LastRead, LastDeleted, MemSize, State int
}
