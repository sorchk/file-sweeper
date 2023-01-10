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

var action string
var configPath string

func init() {
	flag.StringVar(&action, "a", "help", "交互命令：\n"+
		"-a help 查看帮助\n"+
		"-a clear 立即执行清理\n"+
		"-a install 安装服务\n"+
		"-a uninstall 卸载服务\n"+
		"-a start 启动服务\n"+
		"-a stop 停止服务\n"+
		"-a status 服务状态n\n"+
		"-a restart 重启服务")
	flag.StringVar(&configPath, "c", "", "配置文件路径")
}

func isServiceCmd(target string) bool {
	strArray := []string{"install", "uninstall", "start", "stop", "restart", "status"}
	for _, element := range strArray {
		if target == element {
			return true
		}
	}
	return false
}

func main() {
	flag.Parse()
	log.Printf("交互命令:%s\n", action)
	if configPath == "" {
		configPath = "config.yml"
	}
	log.Printf("配置文件:%s\n", configPath)
	if action == "clear" {
		//立即清理
		fileClear.ClearAll(configPath)
		return
	} else if isServiceCmd(action) {
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
			log.Fatal(":", err)
		}
		//errs := make(chan error, 5)
		//logger, err = s.Logger(errs)
		//if err != nil {
		//	log.Fatal(err)
		//	return
		//}
		if action == "install" {
			err := s.Install()
			if err != nil {
				log.Fatalf("安装服务成功:%v\n", err)
			} else {
				log.Println("安装服务成功")
			}
			return
		} else if action == "uninstall" {
			log.Println(action)
			err := s.Uninstall()
			if err != nil {
				log.Fatalf("卸载服务失败:%v\n", err)
			} else {
				log.Println("卸载服务成功")
			}
			return
		} else if action == "stop" {
			err := s.Stop()
			if err != nil {
				log.Fatalf("停止服务失败:%v\n", err)
			} else {
				log.Println("停止服务成功")
			}
			return
		} else if action == "restart" {
			err := s.Restart()
			if err != nil {
				log.Fatalf("重启服务失败:%v\n", err)
			} else {
				log.Println("重启服务成功")
			}
			return
		} else if action == "status" {
			status, err := s.Status()
			log.Printf("服务状态:%v  %s\n", status, " （0错误，1运行，2停止）")
			if err != nil {
				log.Fatalf("读取服务状态失败:%v\n", err)
			}
			return
		} else if action == "start" {
			err = s.Run() // 运行服务
			if err != nil {
				log.Fatalf("服务启动失败:%v\n", err)
			} else {
				log.Println("服务启动成功")
			}
		} else {
			log.Fatalf("不支持的服务命令:%v\n", action)
		}
	}
}

type Program struct {
	exit chan struct{}
}

func (p *Program) Start(s service.Service) error {
	log.Printf("启动服务 %v. action: %s\n", service.Platform(), action)
	p.exit = make(chan struct{})
	go p.run()
	return nil
}
func (p *Program) Stop(s service.Service) error {
	log.Println("停止服务!")
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
