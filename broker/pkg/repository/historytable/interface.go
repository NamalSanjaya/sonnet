package historytable

import (
	"context"
)

type Interface interface {
	InsetMsgs(ctx context.Context, histTb, dataStr string) error
}