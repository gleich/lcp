package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gleich/lumber/v2"
)

func main() {
	lumber.Info("booted")

	r := gin.Default()
	r.GET("/", rootRedirect)

	err := r.Run()
	if err != nil {
		lumber.Fatal(err, "running gin failed")
	}
}

func rootRedirect(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "https://mattglei.ch")
}
