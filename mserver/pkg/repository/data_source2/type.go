package datasource2

// DS2 history table metadata , don't change the order
type HistTbMetadata struct {
	UserId string
	Lastmsg, LastRead, LastDeleted, MemSize, State int
}

// unit of redis in memory DB
type MemoryRow struct {
	Timestamp int  `json:"timestamp"`
	Data string    `json:"data"`
	Size int       `json:"size"`
}

type MemoryRows []*MemoryRow
