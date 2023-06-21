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
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/sweeper_linux_amd64 main.go
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/sweeper_linux_arm64 main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/sweeper_mac_amd64 main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o build/sweeper_windows_amd64.exe main.go

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/sweeper main.go

```

## 使用说明

#### 参数说明
配置文件默认为程序所在目录下的config.yml,使用环境变量FC_CONF_PATH可以修改配置文件路径
```yaml
#日志配置
log:
  #日志级别 FATAL ERROR WARN INFO DEBUG 默认为INFO
  level: "DEBUG"
  #日志文件分割 每24小时重新生成一个文件 默认24小时
  time: "24h"
  #最大保留日志文件个数 默认190个也就是保证6个月，小于10则默认为190
  count: 190
#任务配置
tasks:
    #任务名不可重复 必须配置
  - name: "样例任务"
    #要清理的日志或备份文件所在目录 必须配置
    workdir: "/datadisk/testdata2"
    #定时执行清理任务 默认 每天0点 0 0 0 * * ?
    corn: "0 * * * * ?"
    #清理符合类型的 1 文件 2目录 3递归目录文件 4递归目录文件清理空目录
    filter-type: 1
    #清理符合正则表达式的文件或目录 默认.log结尾
    filter-regex: ".+\\.txt"
    #排除文件 排除优先与 上面的规则 默认无
    excludes-regex:
      - ".+副本 +4.+\\.txt"
      - ".+副本 +5.+\\.txt"
    #最少保留最近几个文件 默认190个文件 小于0则默认190
    clean-keep: 190
    #最少保留最近几天的文件 默认190天
    time-offset: "190d"
    #最大处理文件数，超出后将在下次任务处理 每次任务最多处理文件数 默认1000
    max-batch: 1000
    #测试模式不会删除文件 默认false
    test: true
  - name: "任务1"
    workdir: "/datadisk/adp/logs/info"
    corn: "0 * * * * ?"
    #清理符合类型的 1 文件 2目录
    filter-type: 1
    #清理符合正则表达式的文件或目录
    filter-regex: ".+\\.log"
    #排除文件 排除优先与 上面的规则
    excludes-regex:
    #最少保留最近几个文件
    clean-keep: 10
    #最少保留最近几天的文件
    time-offset: "240h"
    #最大处理文件数，超出后将在下次任务处理
    max-batch: 200
```

#### 命令实例
```shell

# 给执行权限
chmod +x sweeper 
# 查看命令帮助
./sweeper help
# 无需安装服务立即运行一次清理任务
./sweeper clean
# 服务端安装
./sweeper install
# 启动服务
systemctl start sweeper
# 停止服务
systemctl stop sweeper
# 查看服务状态
systemctl status sweeper
  
```