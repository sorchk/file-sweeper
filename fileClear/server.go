package fileClear

import (
	"errors"
	"fileClear/utils"

	//"github.com/robfig/cron"
	_ "encoding/json"
	"github.com/jakecoffman/cron"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"time"
)

var mainCron = cron.New()
var dateLayout = "2006-01-02 15:04:05"

func StartServer(config utils.AppConfig) {
	for _, task := range config.Tasks {
		spec := task.Corn
		log.Debugf("添加计划任务：%s", task.Name)
		mainCron.AddFunc(spec, func() {
			Clear(task)
			PrintNextJob(task.Name)
		}, task.Name)
	}
	//启动计划任务
	mainCron.Start()
	log.Debugf("启动计划任务")
	PrintNextJobs()
	//关闭着计划任务, 但是不能关闭已经在执行中的任务.
	defer mainCron.Stop()
	select {} //阻塞主线程不退出
}

func PrintJobInfo(i int, e *cron.Entry) {
	nextTime := time.Unix(e.Next.Unix(), 0).Format(dateLayout)
	log.Infof("[%s]-下次执行时间为：%v", e.Name, nextTime)
}
func PrintNextJobs() {
	entries := mainCron.Entries()
	for i, e := range entries {
		PrintJobInfo(i, e)
	}
}
func PrintNextJob(name string) {
	entries := mainCron.Entries()
	for i, e := range entries {
		if e.Name == name {
			PrintJobInfo(i, e)
		}
	}
}

func ClearAll(configPath string) {
	var config, err = utils.LoadAppConfig(configPath)
	if err != nil {
		log.Fatalf("读取配置文件发生错误：%v", err)
		return
	}
	for _, task := range config.Tasks {
		Clear(task)
	}
}

func Clear(task utils.TaskConfig) {
	log.Infof("[%s]-开始执行清理任务,%v", task.Name)
	if task.Test {
		log.Warnf("[%s]-测试模式", task.Name)
	}
	//列文件或目录
	files, err := ListDir(task)
	if err != nil {
		log.Warnf("[%s]-读取文件目录发生错误：%v", task.Name, err)
		return
	}
	//按修改时间倒叙排列
	sort.Sort(utils.ByModTime(files))
	log.Debugf("[%s]-文件列表：", task.Name)
	for _, fi := range files {
		log.Debugf("[%s]-%s", task.Name, fi.Name())
	}
	//跳过最小保留文件数
	keep := len(files)
	log.Debugf("[%s]-文件数量：%d", task.Name, keep)
	if task.Keep < keep {
		keep = task.Keep
	}
	log.Debugf("[%s]-保留文件数量：%d", task.Name, keep)
	files = files[keep:]
	//计算偏移时间
	offsetTime := utils.GetDurationTime(task.Offset)
	log.Debugf("[%s]-保留最近文件时间：%s", task.Name, offsetTime.String())
	for _, fi := range files {
		if fi.ModTime().UnixMilli() > (time.Now().UnixMilli() - offsetTime.Milliseconds()) {
			//跳过保留内的文件
			continue
		}
		delPath := task.Workdir + string(os.PathSeparator) + fi.Name()
		log.Infof("[%s]-删除文件：%s", task.Name, delPath)
		if task.Test {
			continue
		}
		//删除文件
		err := os.RemoveAll(delPath)
		if err != nil {
			log.Warnf("删除文件失败：%v", err)
		}
	}
	log.Infof("[%s]-成功执行清理任务", task.Name)
}
func ListDir(task utils.TaskConfig) (files []fs.FileInfo, err error) {
	files = make([]fs.FileInfo, 0)
	dir, err := ioutil.ReadDir(task.Workdir)
	if err != nil {
		return nil, err
	}
	Regx, _ := regexp.Compile(task.Regex)
	if Regx == nil { //失败
		return nil, errors.New("正则表达式错误")
	}
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
		isExclude := false
		for _, Exclude := range task.Excludes {
			ExcludeRegx, _ := regexp.Compile(Exclude)
			if ExcludeRegx == nil { //失败
				return nil, errors.New("排除文件正则表达式错误：" + Exclude)
			}
			if task.Test {
				log.Debugf("排除文件 正则：%s 文件名：%s 匹配结果：%v", Exclude, fi.Name(), ExcludeRegx.MatchString(fi.Name()))
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
