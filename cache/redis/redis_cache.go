package redis

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"time"

	"github.com/go-redis/redis"
	"go.elastic.co/apm/module/apmgoredis"
)

const (
	//DefaultRedisAddress -
	DefaultRedisAddress = "localhost:6379"
	//DefaultMaxRetries -
	DefaultMaxRetries = 3
	//DefaultPoolSize -
	DefaultPoolSize = 100
	//DefaultRetryAfter
	DefaultRetryAfter = 5 * time.Second
)

//Connection - Config connection to Redis
type Connection struct {
	//MasterName - Master name of SentinelCluster
	MasterName string `json:"master_name"`
	//Address - Address of redis single
	Address string `json:"address"`
	//Addresses - Addresses of redis cluster or sentinel
	Addresses []string `json:"addresses"`
	//Password - Optional
	Password string `json:"password"`
	//DB - Default is 0
	DB int `json:"db"`
	//Max retries - Default is 3
	MaxRetries int `json:"max_retries"`
	//PoolSize - Default is 100
	PoolSize int `json:"pool_size"`
	//RetryAfter - Default is 5 seconds
	RetryAfter time.Duration `json:"retry_after"`
}

//ConfigWithDefault - Get Connection config with default
func (c *Connection) Default() {
	if c.Address == "" && len(c.Addresses) == 0 {
		c.Address = DefaultRedisAddress
		c.Addresses = []string{DefaultRedisAddress}
	}

	if c.MaxRetries <= 0 {
		c.MaxRetries = DefaultMaxRetries
	}

	if c.PoolSize <= 0 {
		c.PoolSize = DefaultPoolSize
	}

	if c.DB <= 0 {
		c.DB = 0
	}

	if c.RetryAfter <= 0 {
		c.RetryAfter = DefaultRetryAfter
	}
}

func ConnectionWithViper() *Connection {
	c := &Connection{
		MasterName: viper.GetString("redis.master_name"),
		Address:    viper.GetString("redis.address"),
		Addresses:  viper.GetStringSlice("redis.addresses"),
		Password:   viper.GetString("redis.password"),
		DB:         viper.GetInt("redis.db"),
		MaxRetries: viper.GetInt("redis.max_retries"),
		PoolSize:   viper.GetInt("redis.pool_size"),
		RetryAfter: viper.GetDuration("redis.retry_after"),
	}

	c.Default()

	return c
}

//NewUniversalRedisClient - New a redis client base on configuration
//If you want to use Sentinel, set MasterName
//If you want to use Cluster, Set Addresses more than one string
//If you want to use Single, Set Addresses or Address with a string
func NewUniversalRedisClient(c *Connection) (redis.UniversalClient, error) {
	if c.Address == "" && len(c.Addresses) == 0 {
		return nil, fmt.Errorf("address was not set")
	}

	options := &redis.UniversalOptions{
		Addrs:      c.Addresses,
		Password:   c.Password,
		DB:         c.DB,
		MasterName: c.MasterName,
		MaxRetries: c.MaxRetries,
		PoolSize:   c.PoolSize,
	}

	if len(options.Addrs) == 0 {
		options.Addrs = []string{c.Address}
	}

	client := redis.NewUniversalClient(options)

	retries := 0

	for {
		_, err := client.Ping().Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	return client, nil
}

//NewClusterRedisClient - New a redis client with cluster mode
func NewClusterRedisClient(c *Connection) (*redis.ClusterClient, error) {
	if len(c.Addresses) == 0 {
		return nil, fmt.Errorf("address was not set")
	}

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:      c.Addresses,
		Password:   c.Password,
		MaxRetries: c.MaxRetries,
		PoolSize:   c.PoolSize,
	})

	retries := 0

	for {
		_, err := client.Ping().Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	return client, nil
}

//NewSentinelRedisClient - New a redis client with sentinel mode
func NewSentinelRedisClient(c *Connection) (*redis.Client, error) {
	if len(c.Addresses) == 0 {
		return nil, fmt.Errorf("address was not set")
	}

	if c.MasterName == "" {
		return nil, fmt.Errorf("master name of sentinel cluster was not set")
	}

	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    c.MasterName,
		SentinelAddrs: c.Addresses,
		Password:      c.Password,
		DB:            c.DB,
		MaxRetries:    c.MaxRetries,
		PoolSize:      c.PoolSize,
	})

	retries := 0

	for {
		_, err := client.Ping().Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	return client, nil
}

//NewSingleRedisClient - New a redis client with single mode
func NewSingleRedisClient(c *Connection) (*redis.Client, error) {
	if c.Address == "" {
		return nil, fmt.Errorf("address was not set")
	}

	client := redis.NewClient(&redis.Options{
		Addr:       c.Address,
		Password:   c.Password,
		DB:         c.DB,
		MaxRetries: c.MaxRetries,
		PoolSize:   c.PoolSize,
	})

	retries := 0

	for {
		_, err := client.Ping().Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	return client, nil
}

//NewDefaultRedisClient - New a redis client with default address and single mode
func NewDefaultRedisClient() (*redis.Client, error) {
	return NewSingleRedisClient(&Connection{
		Address:    DefaultRedisAddress,
		Password:   "",
		DB:         0,
		MaxRetries: DefaultMaxRetries,
		PoolSize:   DefaultPoolSize,
	})
}

//NewDefaultRedisUniversalClient - New a redis universal client with default address and single mode
func NewDefaultRedisUniversalClient() (redis.UniversalClient, error) {
	c := &Connection{}
	c.Default()
	return NewUniversalRedisClient(c)
}

//Instrument - Wrap Redis UniversalClient with APM
func Instrument(ctx context.Context, client redis.UniversalClient) redis.UniversalClient {
	return apmgoredis.Wrap(client).WithContext(ctx)
}
