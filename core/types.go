package core

import (
	"time"

	"github.com/xanzy/go-gitlab"
)

type Log struct {
	IP        string `json:"userip"`
	UserName  string `json:"username"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	Info      string `json:"info"`
}

type GitLog struct {
	Logs []Log `json:"logs"`
}

func (log GitLog) Len() int {
	return len(log.Logs)
}

func (log GitLog) Less(i, j int) bool {
	ti, _ := time.Parse("02/Jan/2006:15:04:05", log.Logs[i].Timestamp)
	tj, _ := time.Parse("02/Jan/2006:15:04:05", log.Logs[j].Timestamp)
	return ti.Unix() > tj.Unix()
}

func (log GitLog) Swap(i, j int) {
	log.Logs[i], log.Logs[j] = log.Logs[j], log.Logs[i]
}

type GitlabUsers struct {
	Users []*gitlab.User
}

func (users GitlabUsers) Len() int {
	return len(users.Users)
}
func (users GitlabUsers) Less(i, j int) bool {
	return users.Users[i].Username < users.Users[j].Username
}

func (users GitlabUsers) Swap(i, j int) {
	users.Users[i], users.Users[j] = users.Users[j], users.Users[i]
}

type GitlabProjects struct {
	Projects []*gitlab.Project
}

func (projects GitlabProjects) Len() int {
	return len(projects.Projects)
}

func (projects GitlabProjects) Less(i, j int) bool {
	return projects.Projects[i].Name < projects.Projects[j].Name
}

func (projects GitlabProjects) Swap(i, j int) {
	projects.Projects[i], projects.Projects[j] = projects.Projects[j], projects.Projects[i]
}
