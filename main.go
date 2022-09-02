package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Post struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
type opt struct {
	Addr string
	// Use the specified Username to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	Username string
	// Optional password. Must match the password specified in the
	// requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string
	// Database to be selected after connecting to the server.
	DB int
	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int
	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration
	// Maximum number of socket connections.
	// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	PoolSize int
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
}

//client args as

func main() {
	var db *sql.DB
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "pass"
	dbName := "goex"
	connection_string := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True", dbUser, dbPass, "localhost", "9095")

	db, err := sql.Open(dbDriver, connection_string)
	if err != nil {
		log.Print(err.Error())
	}
	db_string := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", dbName)
	_, err = db.Exec(db_string)
	if err != nil {
		panic(err)
	}
	db.Close()
	connection_string = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True", dbUser, dbPass, "localhost", "9095", dbName)

	db, err = sql.Open(dbDriver, connection_string)
	if err != nil {
		log.Print(err.Error())
	}
	db_string = fmt.Sprintf("CREATE Table IF NOT EXISTS %s(id text NOT NULL, title text, PRIMARY KEY (id));", "posts")
	_, err = db.Exec(db_string)
	if err != nil {
		log.Fatal(err.Error())
	}
	r := gin.Default()
	r.GET("/pong/:id", func(c *gin.Context) {
		var post Post
		x := c.Param("id")

		data, err := redisGet(x)
		if err != nil {
			if err != redis.Nil {
				log.Print(err)
				return
			}
			result, err := db.Query("SELECT id, title FROM posts WHERE id = ?", x)
			if err != nil {
				log.Print(err.Error())
			}
			defer result.Close()
			for result.Next() {
				err := result.Scan(&post.ID, &post.Title)
				if err != nil {
					log.Print(err.Error())
				}
			}
			x, err := json.Marshal(post)
			err = redisSet(post.ID, x)
			if err != nil {
				log.Print(err)
			}
			json.NewEncoder(c.Writer).Encode(post)
			return
			log.Print()
		}
		err = json.Unmarshal(data, &post)
		if err != nil {
			log.Print(err)
		}
		log.Print("cached data")
		json.NewEncoder(c.Writer).Encode(post)

	})
	r.POST("/ping", func(c *gin.Context) {
		var posts Post
		stmt, err := db.Prepare("INSERT INTO posts(title) VALUES(?)")
		if err != nil {
			panic(err.Error())
		}
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			panic(err.Error())
		}

		json.Unmarshal(body, &posts)

		_, err = stmt.Exec(posts.Title)
		if err != nil {
			panic(err.Error())
		}
		c.JSON(http.StatusCreated, posts)
	})

	r.Run(":8086")
}
