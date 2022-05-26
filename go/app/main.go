package main

import (
	// "encoding/json"
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
	ImgDir = "images"
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
	//sqlを開く
	db,err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err !=nil{
		c.Logger().Error("error occured while opening database:%s",err)
	}
	rows,err:=db.Query("select name,category from items")
	var result itemlist
	defer rows.Close()
	for rows.Next() {
		var category string
		var name string
		// var image string
		rows.Scan(&name, &category)
		result_json := item{Name: name, Category: category}
		result.Items = append(result.Items, result_json)
	}
	return c.JSON(http.StatusOK, result.Items)

}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	//sqlを開く
	db,err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err !=nil{
		c.Logger().Error("error occured while opening database:%s",err)
	}
	// var els itemlist

	//sqlに入れる
	rows, err := db.Prepare("insert into items(name, category) values(?,?)")
	_, err = rows.Exec(name, category)
	if err != nil {
		c.Logger().Error("error occured while Exec")
	}
	c.Logger().Infof("Receive item: %s, category: %s", name, category)


	message := fmt.Sprintf("item received: %s,category: %s", name, category)
	res:=Response{Message:message}

	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

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
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
