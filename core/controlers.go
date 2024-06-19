package core

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gitlabUsers)
}

func GetRepos(c *gin.Context) {
	c.JSON(http.StatusOK, gitlabProjects)
}

func GetUserLogs(c *gin.Context) {
	userId := c.Param("id")

	userLog := userLogMap[userId]

	c.JSON(http.StatusOK, userLog)
}

func GetRepoLogs(c *gin.Context) {
	name := c.Param("name")
	repo := strings.Replace(name, "+", "/", -1)

	repoLog := repoLogMap[repo]

	c.JSON(http.StatusOK, repoLog)
}
