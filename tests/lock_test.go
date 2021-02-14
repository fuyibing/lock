// author: wsfuyibing <websearch@163.com>
// date: 2021-02-14

package tests

import (
	"testing"
	"time"

	"github.com/fuyibing/log"

	"github.com/fuyibing/lock"
)

func init() {
	log.Config.TimeFormat = "15:04:05.999999"
	log.Logger.SetAdapter(log.AdapterTerm)
	log.Logger.SetLevel(log.LevelDebug)
}

func TestLock(t *testing.T) {

	c := log.NewContext()

	l := lock.New("test")

	if o1, _ := l.Set(c); o1 {
		time.Sleep(time.Second * 16)
		l.Unset(c)
	}

	// Unset
	// b0, e0 := l.Renewal(c)
	// log.Infofc(c, "b0: %v.", b0)
	// log.Infofc(c, "e0: %v.", e0)
	// log.Infofc(c, "b3: %v.", b3)

	// Get
	// s1, b1, e1 := l.Get(c)
	// log.Infofc(c, "e1: %v", e1)
	// log.Infofc(c, "b1: %v", b1)
	// log.Infofc(c, "s1: %s", s1)
	// Set
	// b2, e2 := l.Set(c)
	// log.Infofc(c, "e2: %v.", e2)
	// log.Infofc(c, "b2: %v.", b2)
	// Unset
	// b3, e3 := l.Unset(c)
	// log.Infofc(c, "e3: %v.", e3)
	// log.Infofc(c, "b3: %v.", b3)
}
