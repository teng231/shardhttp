package shardhttp

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRemote(t *testing.T) {
	paths := map[string]bool{
		"webapp-0":              false,
		"http://webapi-0":       true,
		"http://api:4000":       true,
		"http://api:4000/aob/1": true,
		"http://api":            true,
	}
	for path := range paths {
		remote, err := url.Parse(path)
		log.Print(remote.Scheme, remote.Host, err)
	}
}

func TestShardKeyGet(t *testing.T) {
	for i := 0; i < 100; i++ {
		extract, index := GetShardAddressFromShardKey("top"+strconv.Itoa(i), []string{"http://localhost:3000", "http://localhost:3001"})
		if index == 1 {
			log.Print(index, extract, ",", i)
		}
	}
}

func serve(index int) *gin.Engine {
	r := gin.New()

	// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// By default gin.DefaultWriter = os.Stdout
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		// your custom format
		return fmt.Sprintf("[%s]\"%s %s %s %d %s \"%s\" %s\"\n",
			param.Request.Host,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	// r.Use(gin.Recovery())
	r.Use(GinShardHook([]string{"http://localhost:3000", "http://localhost:3001"}, index))
	r.GET("/apple", func(c *gin.Context) {
		log.Print("XXXXXXX")
		c.JSON(http.StatusOK, gin.H{
			"message": "get apple",
		})
	})
	r.POST("/orange", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "update orange",
		})
	})
	return r
}
func TestProxy(t *testing.T) {
	serve0 := serve(0)
	serve1 := serve(1)
	go serve0.Run(":3000")
	go serve1.Run(":3001")
	time.Sleep(1000 * time.Microsecond)
	lock := make(chan bool)

	shardsKeys := map[string]string{
		"top1": "extract host",
		"top4": "wrong host",
	}
	for shardKey, result := range shardsKeys {
		log.Print(result)
		resp1, err := http.Get("http://localhost:3000/apple?shard_key=" + shardKey)
		if err != nil {
			log.Print(err)
		}
		defer resp1.Body.Close()
		body, err := io.ReadAll(resp1.Body)
		log.Print(string(body), err)

		resp2, err := http.Post("http://localhost:3000/orange?shard_key="+shardKey, "application/json", nil)
		if err != nil {
			log.Print(err)
		}
		defer resp2.Body.Close()
		body, err = io.ReadAll(resp2.Body)
		log.Print(string(body), err)

	}
	// time.Sleep(10 * time.Millisecond)
	<-lock
}
