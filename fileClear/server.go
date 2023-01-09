package fileClear

import (
	"errors"
	"github.com/jinzhu/configor"
	"github.com/robfig/cron"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ConfigData struct {
	//  要清理的日志或备份文件所在目录
	Workdir string `default:""`
	//  定时执行清理任务
	Corn string `default:"0 0 0 * * ? *"`

	Filter struct {
		//清理文件还是目录
		File bool `default:"true"`
		//清理服务正则表达式的文件或目录
		Regex string `default:""`
	}
	Clear struct {
		//最少保留最近几个文件
		Keep int `default:"100"`
		//最少保留最近几天的文件
		Offset string `default:"190d"`
	}
}

func StartServer(configPath string) {
	c := cron.New()
	spec := "*/5 * * * * ?"
	c.AddFunc(spec, func() {
		Clear(configPath)
	})
	//启动计划任务
	c.Start()
	//关闭着计划任务, 但是不能关闭已经在执行中的任务.
	defer c.Stop()
	select {}
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
func Clear(configPath string) {
	if configPath == "" {
		configPath = "config.yml"
	}
	var config = ConfigData{}
	err := configor.Load(&config, configPath)
	if err != nil {
		log.Println("读取配置文件发生错误：", err)
	}
	if config.Workdir == "" {
		config.Workdir = GetCurrentDirectory()
	}
	log.Println("config：", config)
	//列文件或目录
	files, err := ListDir(config)
	if err != nil {
		log.Fatal("读取文件目录发生错误：", nil)
		return
	}
	//按修改时间倒叙排列
	sort.Sort(byModTime(files))
	log.Println("文件列表：")
	for _, fi := range files {
		log.Println(fi.Name())
	}
	//跳过最小保留文件数
	keep := len(files)
	log.Println("文件数量：" + strconv.Itoa(keep))
	if config.Clear.Keep < keep {
		keep = config.Clear.Keep
	}
	log.Println("保留文件数量：" + strconv.Itoa(keep))
	files = files[keep:]
	//计算偏移时间
	offsetTime := getDurationTime(config.Clear.Offset)
	log.Println("保留最近文件时间：" + offsetTime.String())
	for _, fi := range files {
		if fi.ModTime().UnixMilli() > (time.Now().UnixMilli() - offsetTime.Milliseconds()) {
			//跳过保留内的文件
			continue
		}
		delPath := config.Workdir + string(os.PathSeparator) + fi.Name()
		log.Println("删除文件：" + delPath)
		//删除文件
		err = os.RemoveAll(delPath)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		log.Fatal("清理时发生错误：", nil)
		return
	}
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

func ListDir(config ConfigData) (files []fs.FileInfo, err error) {
	files = make([]fs.FileInfo, 0)
	dir, err := ioutil.ReadDir(config.Workdir)
	if err != nil {
		return nil, err
	}
	reg1, err := regexp.Compile(config.Filter.Regex)

	if reg1 == nil { //失败
		return nil, errors.New("正则表达式错误")
	}

	exeFile, _ := exec.LookPath(os.Args[0])
	log.Println("exeFile:" + exeFile)
	for _, fi := range dir {
		if config.Filter.File {
			if fi.IsDir() {
				// 忽略目录
				continue
			}
		} else if !fi.IsDir() {
			// 忽略文件
			continue
		}
		if fi.Name() == "config.yml" || fi.Name() == path.Base(exeFile) {
			//忽略配置文件
			continue
		}
		if reg1.MatchString(fi.Name()) {
			files = append(files, fi)
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
