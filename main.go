package main

import (
	"flag"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/kfei/gitlab-auditor/core"
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
		core.DataInitialize(nginxLogPath, shellLogPath)

		m := gin.Default()
		//gin.SetMode(gin.ReleaseMode)

		//m.Static("public", "./public")
		//https://stackoverflow.com/questions/36357791/gin-router-path-segment-conflicts-with-existing-wildcard
		m.Use(static.Serve("/", static.LocalFile("./public", false)))

		m.GET("/users", core.GetUsers)
		m.GET("/projects", core.GetRepos)

		m.GET("/user/:id", core.GetUserLogs)
		m.GET("/project/:name", core.GetRepoLogs)

		logger.Info("Starting Gitlab Auditor")
		m.Run(":3000")
	}
}
