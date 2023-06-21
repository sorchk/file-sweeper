package utils

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type LogConfig struct {
	Level string `default:"INFO"`
	Time  string `default:"24h"`
	Count uint   `default:"190"`
}
type TaskConfig struct {
	Name string `required:"true"`
	//  要清理的日志或备份文件所在目录
	Workdir string `required:"true"`
	//  定时执行清理任务
	Corn string `default:"0 0 0 * * ?"`
	//清理文件还是目录  1文件2目录
	Type int `yaml:"filter-type",default:"1"`
	//清理服务正则表达式的文件或目录
	Regex    string   `yaml:"filter-regex",default:""`
	Excludes []string `yaml:"excludes-regex"`
	//最少保留最近几个文件 兼容以前的配置
	KeepOld int `yaml:"clear-keep",default:"0"`
	//最少保留最近几个文件
	Keep int `yaml:"clean-keep",default:"0"`
	//最少保留最近几天(多久)的文件
	Offset string `yaml:"time-offset",default:"190d"`
	//批量处理文件数
	Batch int `yaml:"max-batch",default:"1000""`
	//测试模式不会删除文件
	Test bool `yaml:"test"`
}
type AppConfig struct {
	Log   LogConfig    `yaml:"log"`
	Tasks []TaskConfig `yaml:"tasks"`
}

func LoadAppConfig(path string) (AppConfig, error) {
	var appConfig = AppConfig{}
	f, err := os.Open(path)
	if err != nil {
		return appConfig, err
	}
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&appConfig)
	if err == nil {
		if appConfig.Log.Count < 10 {
			appConfig.Log.Count = 190
		}
		if appConfig.Log.Time == "" {
			appConfig.Log.Time = "24h"
		}
		if appConfig.Log.Level == "" {
			appConfig.Log.Level = "INFO"
		}
		for i, _ := range appConfig.Tasks {
			var task = &appConfig.Tasks[i]
			if task.Corn == "" {
				task.Corn = "0 0 0 * * ?"
			}
			if task.Regex == "" {
				task.Regex = ".+/.log"
			}
			if task.Type < 1 {
				task.Type = 1
			}
			//兼容以前的配置
			if task.KeepOld != 0 && task.Keep == 0 {
				task.Keep = task.KeepOld
			}
			if task.Keep < 1 {
				task.Keep = 190
			}
			if task.Batch < 1 {
				task.Batch = 1000
			}
			if task.Batch < task.Keep {
				task.Batch = task.Keep
			}
			if task.Offset == "" {
				task.Offset = "190d"
			}
		}
	}
	jsonStr, jsonErr := json.MarshalIndent(appConfig, "", "\t")
	if jsonErr != nil {
		fmt.Printf("将配置格式化为字符串错误:%v", jsonErr)
	} else {
		fmt.Printf("配置信息：%s", string(jsonStr))
	}

	return appConfig, err
}
func GetDurationTime(timeStr string) time.Duration {
	start := 0
	i := strings.Index(timeStr, "d")
	dayTime := time.Duration(0)
	if i != -1 {
		day, err := strconv.Atoi(timeStr[start:i])
		start = i + 1
		dayTime = time.Duration(day) * time.Hour * 24
		if err == nil {
			dayTime = time.Duration(day) * time.Hour * 24
		}
	}
	i = strings.Index(timeStr, "h")
	hourTime := time.Duration(0)
	if i != -1 {
		hour, err := strconv.Atoi(timeStr[start:i])
		start = i + 1
		if err == nil {
			hourTime = time.Duration(hour) * time.Hour
		}
	}
	i = strings.Index(timeStr, "m")
	minuteTime := time.Duration(0)
	if i != -1 {
		minute, err := strconv.Atoi(timeStr[start:i])
		start = i + 1
		if err == nil {
			minuteTime = time.Duration(minute) * time.Minute
		}
	}
	i = strings.Index(timeStr, "s")
	secondTime := time.Duration(0)
	if i != -1 {
		second, err := strconv.Atoi(timeStr[start:i])
		if err == nil {
			secondTime = time.Duration(second) * time.Second
		}
	}
	return dayTime + hourTime + minuteTime + secondTime
}

// 按文件名排序，可扩展至文件时间
type ByModTime []fs.FileInfo

func (f ByModTime) Less(i, j int) bool {
	return f[i].ModTime().UnixMilli() > f[j].ModTime().UnixMilli()
}                                 // 文件名倒序
func (f ByModTime) Len() int      { return len(f) }
func (f ByModTime) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		log.Fatal(err)
	}
	return dir
}
func GetExeFileDirectory() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}
