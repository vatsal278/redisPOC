package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"strconv"
)

type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func redisGet(key string) ([]byte, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:9096",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	data, err := rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		log.Print(err)
	}
	return data, err
}
func redisSet(i interface{}) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:9096",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := rdb.Set(context.Background(), "id", i, 0).Err()
	if err != nil {
		log.Print(err)
	}
	return err
}

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
	db_string = fmt.Sprintf("CREATE Table IF NOT EXISTS %s(id int NOT NULL AUTO_INCREMENT, title text, PRIMARY KEY (id));", "posts")
	_, err = db.Exec(db_string)
	if err != nil {
		log.Fatal(err.Error())
	}
	r := gin.Default()
	r.GET("/pong/:id", func(c *gin.Context) {
		var post Post
		x := c.Param("id")
		log.Print(x)
		intVar, err := strconv.Atoi(x)
		if err != nil {
			log.Print(err)
		}
		data, err := redisGet("id")
		if err != nil {
			log.Print(err)
		}
		err = json.Unmarshal(data, &post)
		if err != nil {
			log.Print(err)
		}
		if post.ID != intVar {

			result, err := db.Query("SELECT id, title FROM posts WHERE id = ?", intVar)
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
			err = redisSet(x)
			if err != nil {
				log.Print(err)
			}
			json.NewEncoder(c.Writer).Encode(post)
			return
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
		json.NewEncoder(c.Writer).Encode(posts)
	})
	r.GET("/pon", func(c *gin.Context) {})

	r.Run(":8086")
}
