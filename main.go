package main

import (
	"fileClear/fileClear"
	"flag"
	"github.com/kardianos/service"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var logger service.Logger

var action string
var configPath string

func init() {
	flag.StringVar(&action, "a", "help", "交互命令： clear 服务命令： install uninstall start stop restart status")
	flag.StringVar(&configPath, "c", "", "配置文件路径")
}

func main() {
	flag.Parse()
	log.Println("action:" + action)
	log.Println("configPath:" + configPath)

	if action == "clear" {
		//立即清理
		fileClear.Clear(configPath)
		return
	} else if action == "help" {
		log.Println("-a help 查看帮助")
		log.Println("-a clear 立即执行清理")
		log.Println("-a install 安装服务")
		log.Println("-a uninstall 卸载服务")
		log.Println("-a start 启动服务")
		log.Println("-a stop 停止服务")
		log.Println("-a restart 重启服务")
		log.Println("-a status 服务状态")
		//帮助信息
		return
	} else {
		options := make(service.KeyValue)
		options["LogOutput"] = true
		options["HasOutputFileSupport"] = true
		options["WorkingDirectory"] = GetCurrentDirectory()

		svcConfig := &service.Config{
			Name:        "fileClear",
			DisplayName: "文件清理服务",
			Description: "文件清理服务",
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
		if err != nil {
			log.Fatal("service.New() err:", err)
		}
		errs := make(chan error, 5)
		logger, err = s.Logger(errs)
		if err != nil {
			log.Fatal(err)
		}
		if action == "install" {
			err = s.Install()
			if err != nil {
				logger.Error("安装服务成功:", err)
			} else {
				logger.Info("安装服务成功")
			}
			return
		} else if action == "uninstall" {
			log.Println(action)
			err = s.Uninstall()
			if err != nil {
				logger.Error("卸载服务失败:", err)
			} else {
				logger.Info("卸载服务成功")
			}
			return
		} else if action == "stop" {
			log.Println(action)
			err = s.Stop()
			return
		} else if action == "restart" {
			log.Println(action)
			err = s.Restart()
			return
		} else if action == "status" {
			log.Println(action)
			status, err := s.Status()
			logger.Info("status:", status, " （1运行，2停止）")
			if err != nil {
				logger.Error("服务启动失败:", err)
			}
			return
		}
		err = s.Run() // 运行服务
		if err != nil {
			logger.Error("服务启动失败:", err)
		}
	}
}

type Program struct {
	exit chan struct{}
}

func (p *Program) Start(s service.Service) error {
	logger.Infof("启动服务 %v. action: %v", service.Platform(), action)
	p.exit = make(chan struct{})
	go p.run()
	return nil
}
func (p *Program) Stop(s service.Service) error {
	logger.Info("停止服务!")
	close(p.exit)
	return nil
}

func (p *Program) run() {
	// 启动服务
	fileClear.StartServer(configPath)
}
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}
