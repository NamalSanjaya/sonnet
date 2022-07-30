package redis_cache

import (
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
)

type DS1Metadata struct {
	Info mdw.DS1MetadataJson
	State int
}
