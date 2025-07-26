package chandao

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"tasks/config"

	"github.com/robfig/cron"
)

var bugs map[string]int

func setBugs() {
	var err error
	bugs, err = getBugs()
	if err != nil {
		log.Fatalln("获取禅道信息失败")
	}
}

func init() {
	setBugs()
	c := cron.New()
	c.AddFunc("0 0 * * * *", func() {
		setBugs()
	})
	c.Start()
}

func GetBugs() map[string]int {
	return bugs
}

func getToken() (string, error) {
	data := fmt.Sprintf(`{"account": "%s","password": "%s"}`, config.CHANDAO_ACCOUNT, config.CHANDAO_PASSWORD)
	resp, err := http.Post(config.CHANDAO_HOST+"/api.php/v1/tokens", "application/json", strings.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	type TokenResp struct {
		Token string `json:"token"`
	}
	var tokenResp TokenResp
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.Token, nil
}

func getBugs() (map[string]int, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.CHANDAO_HOST+"/api.php/v1/products/1/bugs?limit=100000&status=all", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	token, err := getToken()
	if err != nil {
		return nil, err
	}
	req.Header.Add("Token", token)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type BugResp struct {
		Bugs []struct {
			ResolvedBy struct {
				RealName string `json:"realname"`
			} `json:"resolvedBy"`
		} `json:"bugs"`
	}
	var bugResp BugResp
	err = json.Unmarshal(body, &bugResp)
	if err != nil {
		return nil, err
	}

	statusCounts := make(map[string]int)
	for _, bug := range bugResp.Bugs {
		statusCounts[bug.ResolvedBy.RealName]++
	}

	return statusCounts, nil
}
