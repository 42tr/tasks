package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"tasks/chandao"
	"tasks/config"
	"tasks/util/jwt"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed all:static
var staticFiles embed.FS

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Owner     string `json:"owner"`
	Status    string `json:"status"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type TaskHistory struct {
	TaskID    int       `json:"taskId"`
	TaskTitle string    `json:"taskTitle"`
	TaskOwner string    `json:"taskOwner"`
	Timestamp time.Time `json:"timestamp"`
	Field     string    `json:"field"`
	OldValue  string    `json:"oldValue"`
	NewValue  string    `json:"newValue"`
}

var tasks []Task
var taskHistory []TaskHistory
var nextID = 1

func main() {
	loadTasks()
	loadTaskHistory()

	r := gin.Default()

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	r.StaticFS("/static", http.FS(staticFS))

	r.GET("/", func(c *gin.Context) {
		index, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			log.Fatal(err)
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})

	r.Use(AuthRequired())

	r.GET("/api/tasks", func(c *gin.Context) {
		c.JSON(http.StatusOK, tasks)
	})

	r.POST("/api/tasks", func(c *gin.Context) {
		var newTask Task
		if err := c.BindJSON(&newTask); err != nil {
			return
		}
		newTask.ID = nextID
		nextID++
		newTask.StartTime = time.Now().Format("2006-01-02")
		tasks = append(tasks, newTask)
		saveTasks()
		c.JSON(http.StatusOK, newTask)
	})

	r.PUT("/api/tasks/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var updates map[string]interface{}
		if err := c.BindJSON(&updates); err != nil {
			return
		}

		for i, t := range tasks {
			if t.ID == id {
				if status, ok := updates["status"].(string); ok && t.Status != status {
					recordHistory(t.ID, "status", t.Status, status)
					tasks[i].Status = status
				}
				if owner, ok := updates["owner"].(string); ok && t.Owner != owner {
					recordHistory(t.ID, "owner", t.Owner, owner)
					tasks[i].Owner = owner
				}
				if startTime, ok := updates["startTime"].(string); ok && t.StartTime != startTime {
					recordHistory(t.ID, "startTime", t.StartTime, startTime)
					tasks[i].StartTime = startTime
				}
				if endTime, ok := updates["endTime"].(string); ok && t.EndTime != endTime {
					recordHistory(t.ID, "endTime", t.EndTime, endTime)
					tasks[i].EndTime = endTime
				}
				saveTasks()
				c.JSON(http.StatusOK, tasks[i])
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
	})

	r.DELETE("/api/tasks/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		found := false
		for i, t := range tasks {
			if t.ID == id {
				tasks = append(tasks[:i], tasks[i+1:]...)
				saveTasks()
				found = true
				break
			}
		}

		if found {
			c.Status(http.StatusNoContent) // 204 No Content
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		}
	})

	r.GET("/api/tasks/:id/history", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var historyForTask []TaskHistory
		for _, h := range taskHistory {
			if h.TaskID == id {
				historyForTask = append(historyForTask, h)
			}
		}
		c.JSON(http.StatusOK, historyForTask)
	})

	r.GET("/api/history", func(c *gin.Context) {
		c.JSON(http.StatusOK, taskHistory)
	})

	r.GET("/api/bugs", func(c *gin.Context) {
		resolvedCounts, unresolvedCounts := chandao.GetBugs()
		c.JSON(http.StatusOK, map[string]any{
			"resolved":   resolvedCounts,
			"unresolved": unresolvedCounts,
		})
	})

	r.Run(":8084")
}

func loadTasks() {
	file, err := os.ReadFile("tasks.json")
	if err != nil {
		if os.IsNotExist(err) {
			tasks = []Task{}
			return
		}
		log.Fatal(err)
	}
	json.Unmarshal(file, &tasks)

	if len(tasks) > 0 {
		maxID := 0
		for _, t := range tasks {
			if t.ID > maxID {
				maxID = t.ID
			}
		}
		nextID = maxID + 1
	}
}

func saveTasks() {
	file, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("tasks.json", file, 0644)
}

func loadTaskHistory() {
	file, err := os.ReadFile("task_history.json")
	if err != nil {
		if os.IsNotExist(err) {
			taskHistory = []TaskHistory{}
			return
		}
		log.Fatal(err)
	}
	json.Unmarshal(file, &taskHistory)
}

func saveTaskHistory() {
	file, err := json.MarshalIndent(taskHistory, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("task_history.json", file, 0644)
}

func recordHistory(taskId int, field, oldValue, newValue string) {
	var taskTitle, taskOwner string
	for _, t := range tasks {
		if t.ID == taskId {
			taskTitle = t.Title
			taskOwner = t.Owner
			break
		}
	}

	history := TaskHistory{
		TaskID:    taskId,
		TaskTitle: taskTitle,
		TaskOwner: taskOwner,
		Timestamp: time.Now(),
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
	}
	taskHistory = append(taskHistory, history)
	saveTaskHistory()
}

// AuthRequired 需要登录
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		if strings.HasPrefix(host, "localhost") {
			c.Set("uid", 1)
			c.Next()
			return
		}
		uid, err := jwt.Get(c)
		if err == nil {
			c.Set("uid", uid)
			c.Next()
			return
		}
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		fullReturnURL := scheme + "://" + host
		loginURL := config.AUTH_URL + "?url=" + url.QueryEscape(fullReturnURL)
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
		c.Abort()
	}
}

// Response 基础序列化器
type Response struct {
	Code int    `json:"code"`
	Data any    `json:"data,omitempty"`
	Msg  string `json:"msg"`
}

// Err 通用错误处理
func Err(errCode int, err error) Response {
	res := Response{
		Code: errCode,
		Msg:  err.Error(),
	}
	return res
}
