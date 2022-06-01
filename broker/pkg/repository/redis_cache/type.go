package redis_cache

// DS2 history table metadata
type HistTbMetadata struct {
	UserId string
	Lastmsg, LastRead, LastDeleted, MemSize, State int
}