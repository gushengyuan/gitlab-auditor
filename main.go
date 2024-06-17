package main

import (
	"flag"

	"github.com/go-martini/martini"
	"github.com/kfei/gitlab-auditor/core"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
	logger "github.com/sirupsen/logrus"
)

var (
	h            bool
	shellLogPath string
	nginxLogPath string
)

func init() {
	flag.BoolVar(&h,
		"help",
		false,
		"display this help message")

	flag.StringVar(
		&shellLogPath,
		"shellLog",
		"/var/log/gitlab/gitlab-shell/gitlab-shell.log",
		"path to gitlab-shell.log")

	flag.StringVar(
		&nginxLogPath,
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
		core.DataInitialize(nginxLogPath)

		m := martini.Classic()

		m.Use(render.Renderer())

		m.Use(cors.Allow(&cors.Options{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "DELETE", "HEAD", "OPTIONS", "PUT", "PATCH"},
			AllowHeaders: []string{"Origin"},
		}))

		m.Get("/users", core.GetUsers)
		m.Get("/projects", core.GetRepos)

		m.Get("/user/:id", core.GetUserLogs)
		m.Get("/project/:name", core.GetRepoLogs)

		logger.Info("Starting Gitlab Auditor")
		m.Run()
	}
}
