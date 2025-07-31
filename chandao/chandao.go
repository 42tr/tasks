package chandao

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"tasks/config"

	"github.com/robfig/cron/v3"
)

var resolvedCounts map[string]int
var unresolvedCounts map[string]int

func setBugs() {
	var err error
	resolvedCounts, unresolvedCounts, err = getBugs()
	if err != nil {
		log.Println("获取禅道信息失败")
	}
}

func init() {
	setBugs()
	c := cron.New()
	c.AddFunc("* * * * *", func() {
		log.Println("Get bugs from chandao")
		setBugs()
	})
	c.Start()
}

func GetBugs() (map[string]int, map[string]int) {
	return resolvedCounts, unresolvedCounts
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

func getBugs() (map[string]int, map[string]int, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.CHANDAO_HOST+"/api.php/v1/products/1/bugs?limit=100000&status=all", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	token, err := getToken()
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Token", token)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	type BugResp struct {
		Bugs []struct {
			ResolvedBy struct {
				RealName string `json:"realname"`
			} `json:"resolvedBy"`
			AssignedTo struct {
				RealName string `json:"realname"`
			} `json:"assignedTo"`
		} `json:"bugs"`
	}
	var bugResp BugResp
	err = json.Unmarshal(body, &bugResp)
	if err != nil {
		return nil, nil, err
	}

	resolvedCounts := make(map[string]int)
	unresolvedCounts := make(map[string]int)
	for _, bug := range bugResp.Bugs {
		resolvedCounts[bug.ResolvedBy.RealName]++
		unresolvedCounts[bug.AssignedTo.RealName]++
	}

	return resolvedCounts, unresolvedCounts, nil
}
