package gredis

import (
	"time"

	"github.com/EasonZhao/tables/setting"
	"github.com/gomodule/redigo/redis"
)

// RedisConn 连接池
var RedisConn *redis.Pool

// Setup 初始化连接池
func Setup(host string, password string) error {

	RedisConn = &redis.Pool{
		MaxIdle:     30,
		MaxActive:   30,
		IdleTimeout: 200,
		// 提供创建和配置应用程序连接的一个函数
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", setting.RedisSetting.Password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		// 可选的应用程序检查健康功能
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return nil
}

// Set 设置
func Set(key string, data string, time int) error {
	// 在连接池中获取一个活跃连接
	conn := RedisConn.Get()
	defer conn.Close()

	// value, err := json.Marshal(data)
	// if err != nil {
	// 	return err
	// }
	// 向 Redis 服务器发送命令并返回收到的答复
	_, err := conn.Do("SET", key, data)
	if err != nil {
		return err
	}
	// 到期时间
	//_, err = conn.Do("EXPIRE", key, time)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// Exists 判断存在
func Exists(key string) bool {
	conn := RedisConn.Get()
	defer conn.Close()

	// 将命令返回转为布尔值
	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false
	}

	return exists
}

// Get 读取字段
func Get(key string) (string, error) {
	conn := RedisConn.Get()
	defer conn.Close()

	// 将命令返回转为 Bytes
	reply, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", err
	}

	return reply, nil
}

// Delete 删除
func Delete(key string) (bool, error) {
	conn := RedisConn.Get()
	defer conn.Close()

	return redis.Bool(conn.Do("DEL", key))
}

// LikeDeletes comment
func LikeDeletes(key string) error {
	conn := RedisConn.Get()
	defer conn.Close()

	// 将命令返回转为 []string
	keys, err := redis.Strings(conn.Do("KEYS", "*"+key+"*"))
	if err != nil {
		return err
	}

	for _, key := range keys {
		_, err = Delete(key)
		if err != nil {
			return err
		}
	}

	return nil
}
