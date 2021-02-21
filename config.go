// author: wsfuyibing <websearch@163.com>
// date: 2021-02-14

package lock

import (
	"io/ioutil"
	"time"

	"github.com/gomodule/redigo/redis"
	"gopkg.in/yaml.v2"
)

// Configuration.
type configuration struct {
	Addr            string `yaml:"addr"`
	Index           int    `yaml:"index"`
	Network         string `yaml:"network"`
	Password        string `yaml:"password"`
	MaxActive       int    `yaml:"max-active"`
	MaxIdle         int    `yaml:"max-idle"`
	Wait            bool   `yaml:"wait"`
	IdleTimeout     int    `yaml:"idle-timeout"`
	MaxConnLifetime int    `yaml:"max-conn-lifetime"`
	pool            *redis.Pool
}

// Initialize configuration.
func (o *configuration) initialize() {
	// Parse fields from yaml.
	for _, file := range []string{"./tmp/lock.yaml", "./config/lock.yaml", "../config/lock.yaml"} {
		bs, e1 := ioutil.ReadFile(file)
		if e1 != nil {
			continue
		}
		if e2 := yaml.Unmarshal(bs, o); e2 != nil {
			continue
		}
	}
	// Create pool.
	o.pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				Config.Network,
				Config.Addr,
				redis.DialPassword(Config.Password),
				redis.DialDatabase(Config.Index),
			)
		},
		Wait:            Config.Wait,
		MaxActive:       Config.MaxActive,
		MaxIdle:         Config.MaxIdle,
		IdleTimeout:     time.Duration(Config.IdleTimeout) * time.Second,
		MaxConnLifetime: time.Duration(Config.MaxConnLifetime) * time.Second,
	}
}
