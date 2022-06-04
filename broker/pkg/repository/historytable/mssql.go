package historytable

import (
	"context"
	"fmt"

	"github.com/NamalSanjaya/sonnet/pkgs/database/mssql"
)

const insetMsgsQuery string = "INSERT INTO %s VALUES %s"

type mssqlRepo struct {
	client *mssql.Client
}

func NewRepo(ctx context.Context, config *mssql.Config) *mssqlRepo {
	return &mssqlRepo{
		client: mssql.NewClient(ctx, config),
	}
}

func (msr *mssqlRepo) InsetMsgs(ctx context.Context, histTb, dataStr string) error {
	query := fmt.Sprintf(insetMsgsQuery, histTb, dataStr)
	_, err := msr.client.Pool.ExecContext(ctx, query)
	return err
}