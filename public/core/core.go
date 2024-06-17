package core

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

var (
	gitlabToken string
	gitlabUrl   string

	gitlabUsers    = make([]*gitlab.User, 0)
	gitlabProjects = make([]*gitlab.Project, 0)
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

func init() {
	pwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()

	viper.SetConfigName("gitlab-auditor.conf")
	viper.SetConfigType("properties")
	viper.AddConfigPath(fmt.Sprintf("%s/etc", pwd))
	viper.AddConfigPath(fmt.Sprintf("%s/../etc", pwd))
	viper.AddConfigPath(home)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	cfg := viper.ConfigFileUsed()
	logger.Infof(strings.ReplaceAll(cfg, "\\", "/"))

	// Important: Viper configuration keys are case insensitive.
	configMap := map[string]string{}
	keys := viper.AllKeys()
	for _, k := range keys {
		configMap[k] = viper.GetString(k)
	}

	gitlabUrl = viper.GetString("gitlab.url")
	gitlabToken = viper.GetString("gitlab.token")
}

var mu sync.Mutex

func ParseNginxLogs(nginxLogPath string) {
	fs, err := os.Stat(nginxLogPath)
	if err != nil {
		log.Fatal("Can not find the log file of nginx_access: " + nginxLogPath)
	}

	var files = make([]string, 0)
	if fs.IsDir() {
		entrys, err := os.ReadDir(nginxLogPath)
		if err != nil {
			log.Fatal("Can not read the directory of nginx_access: " + nginxLogPath)
		}

		for _, entry := range entrys {
			files = append(files, fmt.Sprintf("%s/%s", nginxLogPath, entry.Name()))
		}
	} else {
		files = append(files, nginxLogPath)
	}

	// control the total goroutines for uploading
	cpuNum := runtime.NumCPU()
	w := sync.WaitGroup{}
	ch := make(chan bool, cpuNum)

	for _, nginxLogfile := range files {
		ch <- true
		w.Add(1)

		go func(logFile string) {
			defer func() {
				w.Done()
				<-ch
			}()

			file, err := os.Open(logFile)
			if err != nil {
				log.Fatal("Can not open the log file of nginx_access: " + logFile)
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

					mu.Lock()
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
					mu.Unlock()
				}
			}
		}(nginxLogfile)
	}
	w.Wait()
}

func GetGitlabMetadata() {
	git, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabUrl))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// control the total goroutines for uploading
	cpuNum := runtime.NumCPU()
	w := sync.WaitGroup{}
	ch := make(chan bool, cpuNum)

	for user := range userLogMap {
		ch <- true
		w.Add(1)

		go func(name string) {
			defer func() {
				w.Done()
				<-ch
			}()

			options := gitlab.ListUsersOptions{}
			options.Page = 1
			options.PerPage = 100
			options.Search = gitlab.Ptr(name)
			dd, _, err := git.Users.ListUsers(&options)
			if err != nil {
				logger.Errorf("Failed to get users: %v", err)
				return
			}

			if len(dd) == 0 {
				return
			}
			mu.Lock()
			gitlabUsers = append(gitlabUsers, dd...)
			mu.Unlock()
		}(user)
	}
	w.Wait()

	for repo := range repoLogMap {
		ch <- true
		w.Add(1)

		go func(name string) {
			defer func() {
				w.Done()
				<-ch
			}()
			opt := &gitlab.ListProjectsOptions{}
			opt.Page = 1
			opt.PerPage = 100

			tokens := strings.Split(name, "/")
			opt.Search = gitlab.Ptr(tokens[len(tokens)-1])
			projects, _, err := git.Projects.ListProjects(opt)
			if err != nil {
				logger.Errorf("Failed to get projects: %v", err)
				return
			}

			if len(projects) == 0 {
				return
			}
			gitlabProjects = append(gitlabProjects, projects...)
		}(repo)
	}
	w.Wait()
}

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
