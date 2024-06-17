package core

import (
	"strings"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

func GetUsers(params martini.Params, r render.Render) {
	r.JSON(200, gitlabUsers)
}

func GetRepos(params martini.Params, r render.Render) {
	r.JSON(200, gitlabProjects)
}

func GetUserLogs(params martini.Params, r render.Render) {
	userId := params["id"]

	userLog := userLogMap[userId]

	r.JSON(200, userLog)
}

func GetRepoLogs(params martini.Params, r render.Render) {
	var repo = strings.Replace(params["name"], "+", "/", -1)

	repoLog := repoLogMap[repo]

	r.JSON(200, repoLog)
}
