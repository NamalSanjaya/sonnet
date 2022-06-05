package mssql

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/denisenkom/go-mssqldb"

)

type Config struct {
	Schema, Hostname, Database, Username, Password string
	Port int
}

type Client struct { 
	Pool   *sqlx.DB
	Read   map[string]*sqlx.NamedStmt
	Update map[string]*sqlx.NamedStmt
}

var _ Interface = (*Client)(nil)

// max query size 64MB

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
	dbct := &Client{Pool: connPool, Read: map[string]*sqlx.NamedStmt{}, Update: map[string]*sqlx.NamedStmt{}}
	return dbct
}

