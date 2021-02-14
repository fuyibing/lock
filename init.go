// author: wsfuyibing <websearch@163.com>
// date: 2021-02-14

package lock

import (
	"sync"
)

var (
	Config *configuration
)

func init() {
	new(sync.Once).Do(func() {
		Config = new(configuration)
		Config.initialize()
	})
}
