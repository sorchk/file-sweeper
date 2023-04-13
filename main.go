package main

import (
	"fileClean/sweeper"
	"fileClean/utils"
	"fmt"
	"github.com/kardianos/service"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
)

var appConfig utils.AppConfig

type Program struct {
	exit chan struct{}
}

func (p *Program) Start(s service.Service) error {
	log.Debugln("启动服务...", service.Platform())
	p.exit = make(chan struct{})
	go p.run()
	log.Debugln("服务启动完成.")
	return nil
}
func (p *Program) Stop(s service.Service) error {
	log.Println("停止服务.")
	close(p.exit)
	return nil
}
func (p *Program) run() {
	log.Debugln("服务运行中...")
	// 启动服务
	sweeper.StartServer(appConfig)
}

var (
	configName  = "config.yml"
	serviceName = "fileSweeper"
	action      string
	workDir     string
	exeName     string
	configPath  string
)

func initLog() {
	appConfig, _ = utils.LoadAppConfig(configPath)
	logPath := workDir + string(os.PathSeparator) + exeName
	writer, _ := rotatelogs.New(
		logPath+".%Y-%m-%d.log",
		rotatelogs.WithLinkName(logPath+".log"),
		rotatelogs.WithRotationTime(utils.GetDurationTime(appConfig.Log.Time)),
		rotatelogs.WithRotationCount(appConfig.Log.Count),
		rotatelogs.WithRotationSize(100*1024*1024),
	)
	writers := []io.Writer{
		writer,
		os.Stdout}
	//同时写文件和屏幕
	fileAndStdoutWriter := io.MultiWriter(writers...)
	log.SetOutput(fileAndStdoutWriter)
	//log.SetFormatter(&log.JSONFormatter{})
	var level log.Level
	level.UnmarshalText([]byte(appConfig.Log.Level))
	log.Printf("日志级别:%v", level)
	log.SetLevel(level)

	//log.SetFlags(log.Ldate | log.Lmicroseconds)
	//logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if err != nil {
	//	log.Fatalf("打开日志文件错误:%v", err)
	//}
	//mw := io.MultiWriter(os.Stdout, logFile)
	//log.SetOutput(mw)
}

func isServiceAction(target string) bool {
	strArray := []string{"install", "uninstall", "start", "stop", "restart", "status", "run", ""}
	for _, element := range strArray {
		if target == element {
			return true
		}
	}
	return false
}
func isAction(target string) bool {
	strArray := []string{"install", "uninstall", "start", "stop", "restart", "status", "run", "", "clear", "clean"}
	for _, element := range strArray {
		if target == element {
			return true
		}
	}
	return false
}
func initService() (service.Service, error) {
	//服务的配置信息
	options := make(service.KeyValue)
	options["LogOutput"] = true
	options["HasOutputFileSupport"] = true
	options["WorkingDirectory"] = workDir
	if runtime.GOOS == "windows" {
	} else {
		options["Restart"] = "on-failure"
		options["SuccessExitStatus"] = "1 2 8 SIGKILL"
	}
	svcConfig := &service.Config{
		Name:        serviceName,
		DisplayName: "文件清道夫",
		Description: "定时计划执行日志文件清理，备份文件清理等任务",
		Option:      options,
		//Arguments :
	}
	if runtime.GOOS == "windows" {
	} else {
		svcConfig.Dependencies = []string{
			"Requires=network.target",
			"After=network-online.target syslog.target"}
		svcConfig.UserName = "root"
	}
	pro := &Program{}
	s, err := service.New(pro, svcConfig)
	return s, err
}
func main() {
	configPath = os.Getenv("fc_conf_path")
	//初始化常用变量
	workDir = utils.GetExeFileDirectory()
	exeFile, _ := exec.LookPath(os.Args[0])
	exeName = path.Base(exeFile)
	if configPath == "" {
		configPath = workDir + string(os.PathSeparator) + configName
	}

	if len(os.Args) > 1 {
		action = os.Args[1]
	}
	if !isAction(action) {
		help()
		return
	}

	//初始化日志
	initLog()

	log.Infof("工作目录:%s", workDir)
	log.Debugf("配置文件:%s", configPath)

	if action == "clean" || action == "clear" {
		sweeper.CleanAll(configPath)
	} else if isServiceAction(action) {
		s, err := initService()
		if err != nil {
			log.Fatalf("初始化服务错误：%v", err)
		}
		errs := make(chan error, 5)
		_, err = s.SystemLogger(errs)
		if err != nil {
			log.Fatalf("初始化服务日志错误：%v", err)
		}
		go func() {
			for {
				err := <-errs
				if err != nil {
					log.Errorf("服务错误：%v", err)
				}
			}
		}()
		log.Infof("初始化服务：%v", s)
		if action == "" {
			err = s.Run()
			if err != nil {
				log.Fatalf("发生错误：%v", err)
			}
		} else if action == "status" {
			status, err := s.Status()
			log.Infof("服务状态:%v  %s", status, " （0错误，1运行中，2停止）")
			if err != nil {
				log.Fatalf("读取服务状态失败:%v", err)
			}
		} else if action == "install" {
			err := s.Install()
			if err != nil {
				log.Fatalf("安装服务失败:%v", err)
			} else {
				log.Println("安装服务成功,服务名：" + serviceName)
			}
		} else if action == "uninstall" {
			err := s.Uninstall()
			if err != nil {
				log.Fatalf("卸载服务失败:%v", err)
			} else {
				log.Println("卸载服务成功")
			}
		} else if action == "start" {
			err := s.Start()
			if err != nil {
				log.Fatalf("启动服务失败:%v", err)
			} else {
				log.Infof("启动服务成功")
			}
			return
		} else if action == "stop" {
			err := s.Stop()
			if err != nil {
				log.Fatalf("停止服务失败:%v", err)
			} else {
				log.Println("停止服务成功")
			}
			return
		} else if action == "restart" {
			err := s.Restart()
			if err != nil {
				log.Fatalf("重启服务失败:%v", err)
			} else {
				log.Println("重启服务成功")
			}
			return
		} else if action == "run" {
			err := s.Run()
			if err != nil {
				log.Fatal(err)
			}
			return
		}
	} else {
		help()
	}
}
func help() {
	fmt.Println("")
	fmt.Println("---------------命令使用说明--------------------------------")
	fmt.Println(exeName + " clean 无需安装服务，立即运行一次清理任务")
	fmt.Println(exeName + " install 安装服务")
	fmt.Println(exeName + " uninstall 卸载服务")
	fmt.Println(exeName + " start 启动服务")
	fmt.Println(exeName + " stop 停止服务")
	fmt.Println(exeName + " restart 重启服务")
	fmt.Println(exeName + " status 查看服务状态")
	fmt.Println(exeName + " run 控制台运行定时任务服务")
	fmt.Println("--------------------------------------------------------")
	fmt.Println("")
}
