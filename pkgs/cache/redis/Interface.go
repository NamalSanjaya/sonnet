package redis

type Interface interface{
	HSet(key, field, value string) error
	HGet(key, field string) (string, error)
	HDel(key string, fields ...string) error
	HMGet(key string, fields ...string)([]interface{}, error)
}
