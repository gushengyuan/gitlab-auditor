package main

import (
	"flag"

	"github.com/go-martini/martini"
	"github.com/kfei/gitlab-auditor/public/core"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
)

var (
	h            bool
	shellLogFile string
	nginxLogFile string
)

func init() {
	flag.BoolVar(&h,
		"help",
		false,
		"display this help message")

	flag.StringVar(
		&shellLogFile,
		"shellLog",
		"/var/log/gitlab/gitlab-shell/gitlab-shell.log",
		"path to gitlab-shell.log")

	flag.StringVar(
		&nginxLogFile,
		"nginxLog",
		"/var/log/gitlab/nginx/gitlab-access.log",
		"path to gitlab-access.log")

	flag.Parse()
}

func main() {
	switch {
	case h:
		flag.PrintDefaults()
	default:
		core.ParseNginxLogs(nginxLogFile)

		m := martini.Classic()

		m.Use(render.Renderer())

		m.Use(cors.Allow(&cors.Options{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "DELETE", "HEAD", "OPTIONS", "PUT", "PATCH"},
			AllowHeaders: []string{"Origin"},
		}))

		m.Get("/users/:token/:url", core.GetUsers)
		m.Get("/projects/:token/:url", core.GetRepos)

		m.Get("/user/:id", core.GetUserLogs)
		m.Get("/project/:name", core.GetRepoLogs)

		m.Run()
	}
}
