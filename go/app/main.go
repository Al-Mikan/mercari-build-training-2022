package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "image"
)

type item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}
type itemlist struct {
	Items []item `json:"items"`
}
type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}
func getItems(c echo.Context) error {
	// jsonをGoに持ってきている
	// データベースのコネクションを開く
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		panic(err)
	}

	// 複数レコード取得 データベースへクエリを送信
	rows, err := db.Query("select name, category from items")
	if err != nil {
		panic(err)
	}
	// 処理が終わったらカーソルを閉じる
	defer rows.Close()
	var els itemlist

	// 行セットに対して繰り返し処理
	for rows.Next() {
		var name string
		var category string
		err := rows.Scan(&name, &category)
		if err != nil {
			c.Logger().Error("error occured while scan rows:%s", err)
		}
		r_json := item{Name: name, Category: category}
		els.Items = append(els.Items, r_json)
	}
	return c.JSON(http.StatusOK, els.Items)

}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		c.Logger().Error("error occured open database:%s", err)
	}

	rows, err := db.Prepare("insert into items(name,category) values(?,?);")
	if err != nil {
		c.Logger().Error("error occured while prepare database:%s", err)
	}
	_, err = rows.Exec(name, category)
	if err != nil {
		c.Logger().Error("error occured while insert database:%s", err)
	}
	// 処理が終わったらカーソルを閉じる
	defer rows.Close()

	message := fmt.Sprintf("item received: %s,category: %s", name, category)
	res := Response{Message: message}
	fmt.Print(1)
	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("itemImg"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.GET("/items", getItems)
	e.POST("/items", addItem)
	// e.GET("/image/:itemImg", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
