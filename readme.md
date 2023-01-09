# 定时文件清理

## 功能说明
定时清理某目录下的日志文件或备份文件，可自定义配置最小保留文件数，保留最近几天几小时的文件。配置正则表达式过滤文件

## 编译安装
### 配置仓库
```shell
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/
```
### 交叉编译
```shell
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/fileClear_linux_amd64 main.go
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/fileClear_linux_arm64 main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/fileClear_mac_amd64 main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o build/fileClear_windows_amd64.exe main.go

```

## 使用说明

#### 参数说明
- -a     #程序模式，clear立即清理 install安装服务，uninstall卸载服务 start,restart, stop,status 
- -c       #配置文件路径 不适用默认文当前目录下的config.yml

#### 命令实例
```shell
#curl -o dfw https://gitee.com/sorc/log-clear/attach_files/1103945/download/logClear_linux_amd64 -O -L

# 给执行权限
chmod +x logClear 
# 服务端安装
./logClear -a install -c /home/user/config.yml
# 启动服务
systemctl start logClear
# 停止服务
systemctl stop logClear
# 查看服务状态
systemctl status logClear
  
```
 