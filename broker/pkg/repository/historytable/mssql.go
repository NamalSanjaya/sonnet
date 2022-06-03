package historytable

import (
	"context"

	"github.com/NamalSanjaya/sonnet/pkgs/database/mssql"
)

type mssqlRepo struct {
	client *mssql.Client
}

func NewRepo(ctx context.Context, config *mssql.Config) *mssqlRepo {
	return &mssqlRepo{
		client: mssql.NewClient(ctx, config),
	}
}

// we have think about which method is use when data type is NVARCHAR, binary data
// file system base import
// 10010000 1100101 01101100  01101100 01101111