// author: wsfuyibing <websearch@163.com>
// date: 2021-02-14

// Package redis lock.
package lock

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/fuyibing/log"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

const (
	DefaultExpiration = 32
	DefaultRenewal    = 15
)

// Redis lock interface.
type RedisLock interface {
	Get(ctx interface{}) (str string, found bool, err error)
	Set(ctx interface{}) (succeed bool, err error)
	Unset(ctx interface{}) (succeed bool, err error)
	Renewal(ctx interface{}) (succeed bool, err error)
}

// Redis lock struct.
type redisLock struct {
	ch        chan int
	listening bool
	mu        *sync.RWMutex
	key       string
	value     string
	succeed   bool
}

// Create redis lock instance.
func New(key string) RedisLock {
	o := &redisLock{key: key}
	o.ch = make(chan int)
	o.mu = new(sync.RWMutex)
	o.value = o.uuid()
	return o
}

// Get locked resource.
func (o *redisLock) Get(ctx interface{}) (string, bool, error) {
	// Get and release redis connection.
	conn := Config.pool.Get()
	defer func() {
		if e1 := conn.Close(); e1 != nil {
			log.Errorfc(ctx, "[lock][redis][key=%s] release redis connection error: %v.", o.key, e1)
		}
	}()
	// Send command.
	rep, e2 := conn.Do("GET", o.key)
	if e2 != nil {
		log.Errorfc(ctx, "[lock][redis][key=%s] %v.", o.key, e2)
		return "", false, e2
	}
	// Key not found.
	str, e3 := redis.String(rep, nil)
	if e3 != nil {
		return "", false, nil
	}
	// Key found.
	return str, true, nil
}

// Lock resource.
func (o *redisLock) Set(ctx interface{}) (bool, error) {
	// Get and release redis connection.
	conn := Config.pool.Get()
	defer func() {
		if e1 := conn.Close(); e1 != nil {
			log.Errorfc(ctx, "[lock][redis][key=%s] release redis connection error: %v.", o.key, e1)
		}
	}()
	// Send command.
	rep, e2 := conn.Do("SET", o.key, o.value, "NX", "EX", DefaultExpiration)
	if e2 != nil {
		log.Errorfc(ctx, "[lock][redis][key=%s] %v.", o.key, e2)
		return false, e2
	}
	// Send status.
	res, e3 := redis.String(rep, nil)
	if e3 != nil {
		log.Warnfc(ctx, "[lock][redis][key=%s] lock fail: %s.", o.key, e3)
		return false, nil
	}
	// Succeed.
	log.Infofc(ctx, "[lock][redis][key=%s] lock succeed response: %s.", o.key, res)
	o.succeed = true
	o.listen(ctx)
	return true, nil
}

// Release locked resource.
func (o *redisLock) Unset(ctx interface{}) (bool, error) {
	// not set.
	if !o.succeed {
		return true, nil
	}
	o.ch <- 1
	// Get and release redis connection.
	conn := Config.pool.Get()
	defer func() {
		if e1 := conn.Close(); e1 != nil {
			log.Errorfc(ctx, "[lock][redis][key=%s] release redis connection error: %v.", o.key, e1)
		}
	}()
	// Get before DELETE.
	get, e2 := conn.Do("GET", o.key)
	if e2 != nil {
		log.Errorfc(ctx, "[lock][redis][key=%s] %v.", o.key, e2)
		return false, e2
	}
	// Parse get string.
	str, e3 := redis.String(get, nil)
	if e3 != nil {
		log.Infofc(ctx, "[lock][redis][key=%s] resource expired.", o.key)
		return true, nil
	}
	// Return false if value not equal.
	if str != o.value {
		log.Warnfc(ctx, "[lock][redis][key=%s] different goroutine resource.", o.key)
		return false, nil
	}
	// Send delete command.
	rep, e4 := conn.Do("DEL", o.key)
	if e4 != nil {
		log.Errorfc(ctx, "[lock][redis][key=%s] %v.", o.key, e2)
		return false, e2
	}
	// Ended.
	num, _ := redis.Int(rep, nil)
	log.Debugfc(ctx, "[lock][redis][key=%s] delete locked resource: %d.", o.key, num)
	return num > 0, nil
}

// Listen channel.
func (o *redisLock) listen(ctx interface{}) {
	// read/write lock.
	o.mu.Lock()
	defer o.mu.Unlock()
	// status.
	o.listening = true
	go func() {
		redo := true
		defer func() {
			o.listening = false
			if redo {
				o.listen(ctx)
			} else {
				log.Debugfc(ctx, "[lock][redis][key=%s] close channel listening.", o.key)
			}
		}()
		t := time.NewTicker(time.Duration(DefaultRenewal) * time.Second)
		for {
			select {
			case <-t.C:
				go func() {
					_, _ = o.Renewal(ctx)
				}()
			case <-o.ch:
				redo = false
				t.Stop()
				return
			}
		}
	}()
}

// Update expiration.
func (o *redisLock) Renewal(ctx interface{}) (bool, error) {
	// Get and release redis connection.
	conn := Config.pool.Get()
	defer func() {
		if e1 := conn.Close(); e1 != nil {
			log.Errorfc(ctx, "[lock][redis][key=%s] release redis connection error: %v.", o.key, e1)
		}
	}()
	// Send command.
	rep, e1 := conn.Do("EXPIRE", o.key, DefaultExpiration)
	if e1 != nil {
		log.Errorfc(ctx, "[lock][redis][key=%s] %v.", o.key, e1)
		return false, e1
	}
	// Renewal check.
	res, e2 := redis.Int64(rep, nil)
	if e2 != nil {
		log.Errorfc(ctx, "[lock][redis][key=%s] renewal error: %v.", o.key, e2)
		return false, e2
	}
	// Renewal succeed.
	log.Debugfc(ctx, "[lock][redis][key=%s] renewal with %d seconds: %v.", o.key, DefaultExpiration, res > 0)
	return res != 0, nil
}

// Build uuid string.
func (o *redisLock) uuid() string {
	if u, e := uuid.NewUUID(); e == nil {
		return strings.ReplaceAll(u.String(), "-", "")
	}
	t := time.Now()
	return fmt.Sprintf("a%d%d%d", t.Unix(), t.UnixNano(), rand.Int63n(999999999999))
}
