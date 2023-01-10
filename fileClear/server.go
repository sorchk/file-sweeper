package fileClear

import (
	"errors"
	"github.com/robfig/cron"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	Name string `required:"true"`
	//  要清理的日志或备份文件所在目录
	Workdir string `required:"true"`
	//  定时执行清理任务
	Corn string `default:"0 0 0 * * ? *"`
	//清理文件还是目录  1文件2目录
	Type int `yaml:"filter-type"`
	//清理服务正则表达式的文件或目录
	Regex    string   `yaml:"filter-regex"`
	Excludes []string `yaml:"excludes-regex"`
	//最少保留最近几个文件
	Keep int `yaml:"clear-keep"`
	//最少保留最近几天(多久)的文件
	Offset string `yaml:"time-offset"`
	//批量处理文件数
	Batch int `yaml:"max-batch"`
	//测试模式不会删除文件
	Test bool `yaml:"test"`
}
type TaskConfig struct {
	Tasks []Task `yaml:"tasks"`
}

func readYaml(path string) (TaskConfig, error) {
	var taskConfig = TaskConfig{}
	//err := configor.Load(&taskConfig, path)
	f, err := os.Open(path)
	if err != nil {
		return taskConfig, err
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&taskConfig)
	if err == nil {
		for i, _ := range taskConfig.Tasks {
			var task = &taskConfig.Tasks[i]
			if task.Corn == "" {
				task.Corn = "0 0 0 * * ? *"
			}
			if task.Regex == "" {
				task.Regex = ".+/.log"
			}
			if task.Type < 1 {
				task.Type = 1
			}
			if task.Keep < 1 {
				task.Keep = 100
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
	log.Printf("%v\n", taskConfig)

	return taskConfig, err
}
func StartServer(configPath string) {
	var config, err = readYaml(configPath)
	if err != nil {
		log.Fatalf("读取配置文件发生错误：%v", err)
		return
	}
	c := cron.New()
	for _, task := range config.Tasks {
		spec := task.Corn
		c.AddFunc(spec, func() {
			//Clear(task)
		})
	}

	//启动计划任务
	c.Start()
	//关闭着计划任务, 但是不能关闭已经在执行中的任务.
	defer c.Stop()
	select {}
}
func ClearAll(configPath string) {
	var config, err = readYaml(configPath)
	if err != nil {
		log.Fatalf("读取配置文件发生错误：%v\n", err)
		return
	}
	Clear(config)
}
func getDurationTime(timeStr string) time.Duration {
	i := strings.Index(timeStr, "d")
	dayTime := time.Duration(0)
	start := 0
	if i != -1 {
		day, err := strconv.Atoi(timeStr[start:i])
		start = i + 1
		dayTime = time.Duration(day) * time.Hour * 24
		if err != nil {
			dayTime = time.Duration(day) * time.Hour * 24
		}
	}
	i = strings.Index(timeStr, "h")
	hourTime := time.Duration(0)
	if i != -1 {
		hour, err := strconv.Atoi(timeStr[start:i])
		start = i + 1
		if err != nil {
			hourTime = time.Duration(hour) * time.Hour
		}
	}
	i = strings.Index(timeStr, "m")
	minuteTime := time.Duration(0)
	if i != -1 {
		minute, err := strconv.Atoi(timeStr[start:i])
		start = i + 1
		if err != nil {
			minuteTime = time.Duration(minute) * time.Minute
		}
	}
	i = strings.Index(timeStr, "s")
	secondTime := time.Duration(0)
	if i != -1 {
		second, err := strconv.Atoi(timeStr[start:i])
		if err != nil {
			secondTime = time.Duration(second) * time.Second
		}
	}
	return dayTime + hourTime + minuteTime + secondTime
}

func Clear(config TaskConfig) {
	for _, task := range config.Tasks {
		log.Printf("[%s]-开始执行清理任务,%v\n", task.Name, task)
		if task.Test {
			log.Println("[%s]-测试模式\n", task.Name)
		}
		//列文件或目录
		files, err := ListDir(task)
		if err != nil {
			log.Printf("[%s]-读取文件目录发生错误：%v\n", task.Name, err)
			return
		}
		//按修改时间倒叙排列
		sort.Sort(byModTime(files))
		log.Printf("[%s]-文件列表：\n", task.Name)
		for _, fi := range files {
			log.Printf("[%s]-%s\n", task.Name, fi.Name())
		}
		//跳过最小保留文件数
		keep := len(files)
		log.Printf("[%s]-文件数量：%d\n", task.Name, keep)
		if task.Keep < keep {
			keep = task.Keep
		}
		log.Printf("[%s]-保留文件数量：%d\n", task.Name, keep)
		files = files[keep:]
		//计算偏移时间
		offsetTime := getDurationTime(task.Offset)
		log.Printf("[%s]-保留最近文件时间：%s\n", task.Name, offsetTime.String())
		for _, fi := range files {
			if fi.ModTime().UnixMilli() > (time.Now().UnixMilli() - offsetTime.Milliseconds()) {
				//跳过保留内的文件
				continue
			}
			delPath := task.Workdir + string(os.PathSeparator) + fi.Name()
			log.Printf("[%s]-删除文件：%s\n", task.Name, delPath)
			if task.Test {
				continue
			}
			//删除文件
			err := os.RemoveAll(delPath)
			if err != nil {
				log.Printf("删除文件失败：%v\n", err)
			}
		}
		log.Printf("[%s]-成功执行清理任务\n", task.Name)
	}
}
func ListDir(task Task) (files []fs.FileInfo, err error) {
	files = make([]fs.FileInfo, 0)
	dir, err := ioutil.ReadDir(task.Workdir)
	if err != nil {
		return nil, err
	}
	Regx, _ := regexp.Compile(task.Regex)
	if Regx == nil { //失败
		return nil, errors.New("正则表达式错误")
	}
	exeFile, _ := exec.LookPath(os.Args[0])
	for _, fi := range dir {
		if task.Type == 1 && fi.IsDir() {
			// 忽略目录
			continue
		} else if task.Type == 2 && !fi.IsDir() {
			// 忽略文件
			continue
		}
		if task.Type > 3 || task.Type < 1 {
			continue
		}
		if fi.Name() == "config.yml" || fi.Name() == path.Base(exeFile) {
			//忽略配置文件
			continue
		}
		isExclude := false
		for _, Exclude := range task.Excludes {
			ExcludeRegx, _ := regexp.Compile(Exclude)
			if ExcludeRegx == nil { //失败
				return nil, errors.New("排除文件正则表达式错误：" + Exclude)
			}
			if task.Test {
				log.Printf("排除文件 正则：%s 文件名：%s 匹配结果：%v\n", Exclude, fi.Name(), ExcludeRegx.MatchString(fi.Name()))
			}
			if ExcludeRegx.MatchString(fi.Name()) {
				isExclude = true
				break
			}
		}
		if isExclude {
			continue
		}
		if Regx.MatchString(fi.Name()) {
			files = append(files, fi)
		}
		if len(files) >= task.Batch {
			break
		}
	}
	return files, nil
}

// 按文件名排序，可扩展至文件时间
type byModTime []os.FileInfo

func (f byModTime) Less(i, j int) bool {
	return f[i].ModTime().UnixMilli() > f[j].ModTime().UnixMilli()
}                                 // 文件名倒序
func (f byModTime) Len() int      { return len(f) }
func (f byModTime) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
