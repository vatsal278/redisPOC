package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	wrapper "github.com/vatsal278/Redis-go-cache"
	"io/ioutil"
	"log"
	"net/http"
)

type Post struct {
	ID    string `json:"id"`
	Title string `json:"title"`
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
	db_string = fmt.Sprintf("CREATE Table IF NOT EXISTS %s(id text NOT NULL, title text, PRIMARY KEY (id));", "posts")
	_, err = db.Exec(db_string)
	if err != nil {
		log.Fatal(err.Error())
	}
	sdk := wrapper.RedisSdkI(wrapper.Config{
		Addr:     "localhost:9096",
		Password: "",
		DB:       0,
	})
	r := gin.Default()
	r.GET("/pong/:id", func(c *gin.Context) {
		var post Post
		x := c.Param("id")

		data, err := sdk.RedisGet(x)
		if err != nil && err != redis.Nil {
			log.Print(err)
			return
		}
		if err == nil {
			err = json.Unmarshal(data, &post)
			if err != nil {
				log.Print(err)
			}
			log.Print("cached data")
			json.NewEncoder(c.Writer).Encode(post)
			return
		}
		result, err := db.Query("SELECT id, title FROM posts WHERE id = ?", x)
		if err != nil {
			log.Print(err)
		}
		defer result.Close()
		for result.Next() {
			err := result.Scan(&post.ID, &post.Title)
			if err != nil {
				log.Print(err.Error())
			}
		}
		y, err := json.Marshal(post)
		err = sdk.RedisSet(post.ID, y)
		if err != nil {
			log.Print(err)
		}
		json.NewEncoder(c.Writer).Encode(post)
		return
		log.Print()

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
