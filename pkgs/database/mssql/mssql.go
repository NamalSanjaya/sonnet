package mssql

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jmoiron/sqlx"

)

type Config struct {
	Schema, Hostname, Database, Username, Password string
	Port int
}

type Client struct { 
	pool   *sqlx.DB
	read   map[string]*sqlx.NamedStmt
	update map[string]*sqlx.NamedStmt
}

func NewClient(ctx context.Context, config *Config) *Client {
	params := url.Values{}
	params.Add("database", config.Database)

	dataSrcUrl := &url.URL{
		Scheme:   config.Schema,
		User:     url.UserPassword(config.Username,config.Password),
		Host:     fmt.Sprintf("%s:%d", config.Hostname,config.Port),
		RawQuery: params.Encode(),
	}
	connPool, err := sqlx.ConnectContext(ctx, "sqlserver", dataSrcUrl.String())
	if err != nil {
		panic(err)
	}
	dbct := &Client{pool: connPool, read: map[string]*sqlx.NamedStmt{}, update: map[string]*sqlx.NamedStmt{}}
	return dbct
}

