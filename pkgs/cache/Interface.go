package cache

type Interface interface{
	HSet(key, field, value string) error
	HGet(key, field string) (string, error)
	HDel(key string, fields ...string) error
}
