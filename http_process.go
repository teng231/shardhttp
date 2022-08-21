package shardhttp

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	hostname = os.Getenv("HOSTNAME")
)

// MakeHostNameWithStatefullset
func GenServicesWithStatefullset(totalShard int) ([]string, int, error) {
	arr := strings.Split(hostname, "-")
	if len(arr) != 2 {
		return nil, -1, fmt.Errorf("hostname '%s' not valid form xxx-i ", hostname)
	}
	index, err := strconv.Atoi(arr[1])
	if err != nil {
		return nil, -1, fmt.Errorf("hostname not include index, err: %s", err.Error())
	}
	// create list pod of service.
	serviceAddrs := []string{}
	for i := 0; i < totalShard; i++ {
		serviceAddrs = append(serviceAddrs, "http://"+arr[0]+"-"+strconv.FormatInt(int64(i), 10))
	}
	return serviceAddrs, index, nil
}

// GinShardHook create gin shard middleware
// if true extract shard -> process -> return
// if not redirect using proxy to extract shard then return.
func GinShardHook(serviceAddrs []string, index int) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Print("calling ", serviceAddrs[index])
		if c.GetHeader("s-from") != "" {
			log.Print("run at shard")
		}
		shardKey := c.GetHeader("s-shard-key")
		if shardKey == "" {
			shardKey, _ = c.GetQuery("shard_key")
		}
		if shardKey == "" {
			// non shardkey is just simple process
			c.Next()
			return
		}
		// get extract address with shardkey
		extractAddr, idx := GetShardAddressFromShardKey(shardKey, serviceAddrs)
		log.Print(extractAddr, ",", idx)
		if index == idx {
			// process inside extract shardkey
			c.Next()
			return
		}
		// override header
		headers := http.Header{
			"s-from": []string{serviceAddrs[index]},
			"s-to":   []string{extractAddr},
		}
		proxy(extractAddr, c.Request, c.Writer, headers)
		c.Abort()
	}
}
