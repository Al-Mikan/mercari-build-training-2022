package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
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
	//jsonをGoに持ってきている
	jsonFromFile, err := os.ReadFile("./items.json")

	if err != nil {
		c.Logger().Error("Notfound ./items.json")
	}
	//jsonを構造体に変換
	var els itemlist
	err = json.Unmarshal(jsonFromFile, &els)

	return c.JSON(http.StatusOK, els)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	newItem := item{Name: name, Category: category}
	//jsonをGoに持ってきている
	jsonFile, err := os.ReadFile("./items.json")
	//error処理
	if err != nil {
		c.Logger().Error("Notfound ./items.json")
	}

	var els itemlist
	//jsonを構造体に変換
	err = json.Unmarshal(jsonFile, &els)
	if err != nil {
		c.Logger().Error("error occured while unmarshalling json")
	}
	//els.Itemsに新しく追加
	els.Items = append(els.Items, newItem)
	file, _ := json.MarshalIndent(els, "", " ")
	os.WriteFile("./items.json", file, 0644)

	message := fmt.Sprintf("item received: %s,category: %s", name, category)

	return c.JSON(http.StatusOK, Response{Message: message})
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
	e.GET("/image/:itemImg", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
