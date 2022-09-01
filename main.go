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
)

type Post struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

func redisGet() ([]byte, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	data, err := rdb.Get(context.Background(), "id").Bytes()
	if err != nil {
		panic(err)
	}
	return data, err
}
func redisSet(i interface{}) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	err := rdb.Set(context.Background(), "id", i, 0).Err()
	if err != nil {
		panic(err)
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
		panic(err.Error())
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
		panic(err.Error())
	}
	defer db.Close()
	db_string = fmt.Sprintf("CREATE Table IF NOT EXISTS %s(id int NOT NULL AUTO_INCREMENT, title text, PRIMARY KEY (id));", "posts")
	_, err = db.Exec(db_string)
	if err != nil {
		log.Fatal(err.Error())
	}
	r := gin.Default()
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
	r.GET("/pong", func(c *gin.Context) {
		var post Post
		var request int64
		//x:= c.Request.URL.Query()
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			panic(err.Error())
		}

		json.Unmarshal(body, &request)
		data, err := redisGet()
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(data, &post)
		if err != nil {
			panic(err)
		}
		if post.ID != request {

			result, err := db.Query("SELECT id, title FROM posts WHERE id = ?", request)
			if err != nil {
				panic(err.Error())
			}
			defer result.Close()
			for result.Next() {
				err := result.Scan(&post.ID, &post.Title)
				if err != nil {
					panic(err.Error())
				}
			}
			err = redisSet(post)
			if err != nil {
				panic(err)
			}
			json.NewEncoder(c.Writer).Encode(post)
			return
		}
		json.NewEncoder(c.Writer).Encode(post)

	})
	r.Run(":8086")
}
