package core

import (
	"bufio"
	"encoding/base64"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	logger "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

var (
	/*
	   nginx_access.log format

	   log_format gitlab_access            '$remote_addr - $remote_user [$time_local] "$request_method $filtered_request_uri $server_protocol" $status $body_bytes_sent "$filtered_http_referer" "$http_user_agent" $gzip_ratio';
	   log_format gitlab_mattermost_access '$remote_addr - $remote_user [$time_local] "$request_method $filtered_request_uri $server_protocol" $status $body_bytes_sent "$filtered_http_referer" "$http_user_agent" $gzip_ratio';
	*/
	regexNginx     = regexp.MustCompile(`([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}) - (.*?) \[(.*) .*\] "(.*?) (.*?) (.*?)" (\d+) (\d+) "(.*?)" "(.*?)" (.*?)`)
	regexNginxPost = regexp.MustCompile(`/(.+)\.git/(.+)`)
	regexNginxGet  = regexp.MustCompile(`/(.+)\.git/info/refs\?service=(.+)`)
)

type Log struct {
	IP        string `json:"userip"`
	UserName  string `json:"username"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	Info      string `json:"info"`
}

type UserLog struct {
	Logs []Log `json:"logs"`
}

type RepoLog struct {
	Logs []Log `json:"logs"`
}

var userLogMap = make(map[string]UserLog)
var repoLogMap = make(map[string]RepoLog)

func ParseNginxLogs(nginxLogFile string) {
	file, err := os.Open(nginxLogFile)
	if err != nil {
		log.Fatal("Can not open the log file of nginx_access: " + nginxLogFile)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		tokens := regexNginx.FindStringSubmatch(text)
		if len(tokens) > 0 {
			_userIP := tokens[1]
			_username := tokens[2]
			_timestamp := tokens[3]
			_method := tokens[4]
			_url := tokens[5]
			//_protocol := tokens[6]
			_status := tokens[7]
			//_bodyBytesSent := tokens[8]
			//_httpReferer := tokens[9]
			_httpUserAgent := tokens[10]
			//_gzipRatio := tokens[11]
			_action := "Push"
			_repo := ""

			// skip the gitlab-runner log
			if strings.Contains(_httpUserAgent, "gitlab-runner") {
				continue
			}

			// skip the non-200 status
			if _status != "200" {
				continue
			}

			// get the repo and action
			var repoTokens []string
			if _method == "POST" {
				// push
				repoTokens = regexNginxPost.FindStringSubmatch(_url)
				if len(repoTokens) <= 0 {
					continue
				}
				_repo = repoTokens[1]
				_action = "Push"
			} else {
				// pull/fetch
				repoTokens = regexNginxGet.FindStringSubmatch(_url)
				if len(repoTokens) <= 0 {
					continue
				}
				_action = "Fetch"
				_repo = repoTokens[1]
			}

			if strings.Contains(_username, "@") {
				clips := strings.Split(_username, "@")
				_username = clips[0]
			}

			// set the user log map
			userLog := userLogMap[_username]
			if userLog.Logs == nil {
				userLog.Logs = make([]Log, 0)
			}
			ulog := Log{IP: _userIP, UserName: _username, Action: _action, Timestamp: _timestamp, Info: _repo}
			userLog.Logs = append(userLog.Logs, ulog)
			userLogMap[_username] = userLog

			// set the repo log map
			repoLog := repoLogMap[_repo]
			if repoLog.Logs == nil {
				repoLog.Logs = make([]Log, 0)
			}

			rlog := Log{IP: _userIP, UserName: _username, Action: _action, Timestamp: _timestamp, Info: _repo}
			repoLog.Logs = append(repoLog.Logs, rlog)
			repoLogMap[_repo] = repoLog
		}
	}
}

func GetUsers(params martini.Params, r render.Render) {
	users := make([]string, 0, len(userLogMap))
	for user := range userLogMap {
		users = append(users, user)
	}

	url, _ := base64.StdEncoding.DecodeString(params["url"])
	git, err := gitlab.NewClient(params["token"], gitlab.WithBaseURL((string)(url)))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	var gitlabUsers = make([]*gitlab.User, 0)

	var i = 1
	for {
		options := gitlab.ListUsersOptions{}
		options.Page = i
		options.PerPage = 100
		dd, _, err := git.Users.ListUsers(&options)
		if err != nil {
			logger.Errorf("Failed to get users: %v", err)
			break
		}

		if len(dd) == 0 {
			break
		}
		gitlabUsers = append(gitlabUsers, dd...)
		i = i + 1
	}

	var rc = make([]*gitlab.User, 0)
	for _, user := range gitlabUsers {
		for _, userName := range users {
			if user.Username == userName {
				rc = append(rc, user)
			}
		}
	}

	r.JSON(200, rc)
}

func GetRepos(params martini.Params, r render.Render) {
	repos := make([]string, 0, len(repoLogMap))
	for repo := range repoLogMap {
		repos = append(repos, repo)
	}

	url, _ := base64.StdEncoding.DecodeString(params["url"])
	git, err := gitlab.NewClient(params["token"], gitlab.WithBaseURL((string)(url)))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	var gitlabProjects = make([]*gitlab.Project, 0)

	for _, repo := range repos {
		opt := &gitlab.ListProjectsOptions{}
		opt.Page = 1
		opt.PerPage = 100

		tokens := strings.Split(repo, "/")
		opt.Search = gitlab.Ptr(tokens[len(tokens)-1])
		projects, _, err := git.Projects.ListProjects(opt)
		if err != nil {
			logger.Errorf("Failed to get projects: %v", err)
			break
		}

		if len(projects) == 0 {
			break
		}
		gitlabProjects = append(gitlabProjects, projects...)
	}

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
