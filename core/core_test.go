package core

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	logger "github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

/*
nginx log

push

10.2.47.75 - - [13/Jun/2024:10:55:45 +0800] "GET /internal/gitlab-auditor.git/info/refs?service=git-receive-pack HTTP/1.1" 401 279 "" "git/2.39.2.windows.1" -
10.2.47.75 - somebody [13/Jun/2024:10:55:46 +0800] "GET /internal/gitlab-auditor.git/info/refs?service=git-receive-pack HTTP/1.1" 200 197 "" "git/2.39.2.windows.1" -
10.2.47.75 - somebody [13/Jun/2024:10:55:47 +0800] "POST /internal/gitlab-auditor.git/git-receive-pack HTTP/1.1" 200 52 "" "git/2.39.2.windows.1" -

pull

10.2.47.75 - - [13/Jun/2024:11:24:33 +0800] "GET /internal/quality.internal.net.git/info/refs?service=git-upload-pack HTTP/1.1" 401 279 "" "git/2.39.2.windows.1" -
10.2.47.75 - somebody [13/Jun/2024:11:24:34 +0800] "GET /internal/quality.internal.net.git/info/refs?service=git-upload-pack HTTP/1.1" 200 164 "" "git/2.39.2.windows.1" -
10.2.47.75 - somebody [13/Jun/2024:11:24:35 +0800] "POST /internal/quality.internal.net.git/git-upload-pack HTTP/1.1" 200 261 "" "git/2.39.2.windows.1" -
10.2.47.75 - somebody [13/Jun/2024:11:24:36 +0800] "POST /internal/quality.internal.net.git/git-upload-pack HTTP/1.1" 200 15487 "" "git/2.39.2.windows.1" -
*/

/*
shell log

pull

{"command":"*uploadpack.Command","correlation_id":"01J0AS0S1YBK2RNHQYY2JV1CZ8","env":{"GitProtocolVersion":"","IsSSHConnection":true,"OriginalCommand":"git-upload-pack 'internal/demo-test.git'","RemoteAddr":"10.3.233.7"},"level":"info","msg":"gitlab-shell: main: executing command","time":"2024-06-14T07:05:02Z"}
{"content_length_bytes":659,"correlation_id":"01J0AS0S1YBK2RNHQYY2JV1CZ8","duration_ms":113,"level":"info","method":"POST","msg":"Finished HTTP request","status":200,"time":"2024-06-14T07:05:02Z","url":"http://unix/api/v4/internal/allowed"}
{"command":"git-upload-pack","correlation_id":"01J0AS0S1YBK2RNHQYY2JV1CZ8","git_protocol":"","gl_key_id":714,"gl_key_type":"key","gl_project_path":"internal/demo-test","gl_repository":"project-5608","level":"info","msg":"executing git command","remote_ip":"10.3.233.7","time":"2024-06-14T07:05:02Z","user_id":"user-90","username":"somebody"}
{"correlation_id":"01J0AS0S1YBK2RNHQYY2JV1CZ8","level":"info","msg":"gitlab-shell: main: command executed successfully","time":"2024-06-14T07:05:02Z"}

push

{"command":"*receivepack.Command","correlation_id":"01J0AS262NVTYTTVXWRJGQWM3Y","env":{"GitProtocolVersion":"","IsSSHConnection":true,"OriginalCommand":"git-receive-pack 'internal/demo-test.git'","RemoteAddr":"10.3.233.7"},"level":"info","msg":"gitlab-shell: main: executing command","time":"2024-06-14T07:05:48Z"}
{"content_length_bytes":659,"correlation_id":"01J0AS262NVTYTTVXWRJGQWM3Y","duration_ms":133,"level":"info","method":"POST","msg":"Finished HTTP request","status":200,"time":"2024-06-14T07:05:49Z","url":"http://unix/api/v4/internal/allowed"}
{"command":"git-receive-pack","correlation_id":"01J0AS262NVTYTTVXWRJGQWM3Y","git_protocol":"","gl_key_id":714,"gl_key_type":"key","gl_project_path":"internal/demo-test","gl_repository":"project-5608","level":"info","msg":"executing git command","remote_ip":"10.3.233.7","time":"2024-06-14T07:05:49Z","user_id":"user-90","username":"somebody"}
{"correlation_id":"01J0AS262NVTYTTVXWRJGQWM3Y","level":"info","msg":"gitlab-shell: main: command executed successfully","time":"2024-06-14T07:05:49Z"}
*/

func TestRegexNginx(t *testing.T) {
	//text := `10.14.188.100 - Username for 'http [24/May/2024:00:11:01 +0800] "GET /internal/demo-test.git/info/refs?service=git-upload-pack HTTP/1.1" 401 279 "" "git/1.8.3.1" -`
	text := `10.77.251.228 - - [24/May/2024:00:11:01 +0800] "POST /api/v4/jobs/request HTTP/1.1" 204 0 "" "gitlab-runner 16.4.1 (16-4-stable; go1.20.5; linux/amd64)" -`
	//text := `10.14.201.3 - somebody [24/May/2024:00:11:01 +0800] "GET /internal/demo-test.git/info/refs?service=git-upload-pack HTTP/1.1" 200 287 "" "git/1.8.3.1" -`
	tokens := regexNginx.FindStringSubmatch(text)
	for _, token := range tokens {
		logger.Info(token)
	}
}

func TestNginxPost(t *testing.T) {
	text := `/internal/demo-test.git/git-upload-pack`
	tokens := regexNginxPost.FindStringSubmatch(text)
	for _, token := range tokens {
		logger.Info(token)
	}
}

func TestNginxGet(t *testing.T) {
	text := `/internal/demo-test.git/info/refs?service=git-upload-pack`
	tokens := regexNginxGet.FindStringSubmatch(text)
	for _, token := range tokens {
		logger.Info(token)
	}
}

func TestGitlabTimeformat(t *testing.T) {
	t1, err := time.Parse("2006-01-02 15:04:05", "2024-12-19 10:55:46")
	if err != nil {
		logger.Error(err)
	}
	logger.Info(t1.Unix())
}

func TestDateFormat1(t *testing.T) {
	t1, err := time.Parse("02/Jan/2006:15:04:05", "12/Jun/2024:10:55:46")
	if err != nil {
		logger.Error(err)
	}
	logger.Info(t1)
}

func TestDateFormat2(t *testing.T) {
	t1, err := time.Parse("2006-01-02T15:04:05Z", "2024-06-14T07:23:50Z")
	if err != nil {
		logger.Error(err)
	}
	logger.Info(t1.Format("02/Jan/2006:15:04:05"))
}

func TestExportUserLogs(t *testing.T) {
	DataInitialize("../nginx")

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	for _, user := range gitlabUsers {
		userName := user.Name
		// Create a new sheet.
		_, err := f.NewSheet(user.Name)
		if err != nil {
			fmt.Println(err)
			return
		}

		f.SetCellValue(userName, "A1", "Timestamp")
		f.SetCellValue(userName, "B1", "IP Address")
		f.SetCellValue(userName, "C1", "Action")
		f.SetCellValue(userName, "D1", "Repository")
		for i := 0; i < userLogMap[user.Username].Len(); i++ {
			log := userLogMap[user.Username].Logs[i]
			f.SetCellValue(userName, "A"+strconv.Itoa(i+2), log.Timestamp)
			f.SetCellValue(userName, "B"+strconv.Itoa(i+2), log.IP)
			f.SetCellValue(userName, "C"+strconv.Itoa(i+2), log.Action)
			f.SetCellValue(userName, "D"+strconv.Itoa(i+2), log.Info)
		}
	}

	// Delete the default sheet.
	err := f.DeleteSheet("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Save spreadsheet by the given path.
	if err := f.SaveAs("userLogs.xlsx"); err != nil {
		fmt.Println(err)
	}
}

func TestExportRepoLogs(t *testing.T) {
	DataInitialize("../nginx")

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	for _, repo := range gitlabProjects {
		repoName := repo.Name
		if len(repoName) > 31 {
			repoName = repoName[:31]
		}
		// Create a new sheet.
		_, err := f.NewSheet(repoName)
		if err != nil {
			fmt.Println(err)
			return
		}

		f.SetCellValue(repoName, "A1", "Timestamp")
		f.SetCellValue(repoName, "B1", "IP Address")
		f.SetCellValue(repoName, "C1", "Action")
		f.SetCellValue(repoName, "D1", "UserName")
		for i := 0; i < repoLogMap[repo.PathWithNamespace].Len(); i++ {
			log := repoLogMap[repo.PathWithNamespace].Logs[i]
			f.SetCellValue(repoName, "A"+strconv.Itoa(i+2), log.Timestamp)
			f.SetCellValue(repoName, "B"+strconv.Itoa(i+2), log.IP)
			f.SetCellValue(repoName, "C"+strconv.Itoa(i+2), log.Action)
			f.SetCellValue(repoName, "D"+strconv.Itoa(i+2), log.UserName)
		}
	}

	// Delete the default sheet.
	err := f.DeleteSheet("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Save spreadsheet by the given path.
	if err := f.SaveAs("repoLogs.xlsx"); err != nil {
		fmt.Println(err)
	}
}

func init() {
	log := Log{IP: "10.15.0.226", UserName: "_username", Action: "Push", Timestamp: "24/Apr/2024:11:20:06", Info: "/demo/walle-deploy"}
	t, _ := time.Parse("02/Jan/2006:15:04:05", "24/Apr/2024:11:20:06")
	tick := t.Unix()
	for i := 0; i < 4000; i++ {
		log.Timestamp = time.Unix(tick+int64(i), 0).Format("02/Jan/2006:15:04:05")
	}
}
