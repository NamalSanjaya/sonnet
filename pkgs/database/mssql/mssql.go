package database

// import (
// 	"context"
// 	"fmt"
// 	"net/url"

// 	"github.com/jmoiron/sqlx"

// )

// type DbClient struct { 
// 	pool   *sqlx.DB
// 	read   map[string]*sqlx.NamedStmt
// 	update map[string]*sqlx.NamedStmt
// }

// func NewDBClient(ctx context.Context, dbConfig *cmtype.DbConfig) (*DbClient, error) {
// 	params := url.Values{}
// 	params.Add("database", dbConfig.Database)

// 	dataSrcUrl := &url.URL{
// 		Scheme:   dbConfig.Schema,
// 		User:     url.UserPassword(dbConfig.Username, dbConfig.Password),
// 		Host:     fmt.Sprintf("%s:%d", dbConfig.Server, dbConfig.Port),
// 		RawQuery: params.Encode(),
// 	}
// 	dbPool, err := sqlx.ConnectContext(ctx, "sqlserver", dataSrcUrl.String())
// 	if err != nil {
// 		panic(err)
// 	}
// 	dbct := &DbClient{pool: dbPool, read: map[string]*sqlx.NamedStmt{}, update: map[string]*sqlx.NamedStmt{}}

// 	if err = dbct.prepareAllStmts(ctx); err != nil {
// 		return nil, err
// 	}
// 	return dbct, nil
// }

// func (db *DbClient) prepareAllStmts(ctx context.Context) (err error) {
// 	return nil
// }
