package middleware

type PairHistTb struct {
	Tx2Rx_HistTb string `json:"tx2rx"`
	Rx2Tx_HistTb string `json:"rx2tx"`
}

type DS1MetadataJson struct {
	Username      string                `json:"username"`
	Email         string                `json:"email"`
	BlockUserList []string              `json:"blockuserlist"`
	HistTbs       map[string]PairHistTb `json:"histtbs"`
}
