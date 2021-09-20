# 河海大学健康打卡

[![build](https://github.com/yin1999/healthreport/actions/workflows/Build.yml/badge.svg)](https://github.com/yin1999/healthreport/actions/workflows/Build.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/yin1999/healthreport)](https://goreportcard.com/report/github.com/yin1999/healthreport) [![Go Reference](https://pkg.go.dev/badge/github.com/yin1999/healthreport.svg)](https://pkg.go.dev/github.com/yin1999/healthreport)

项目使用http请求模拟整个打卡过程，速度很快！  
一键打卡，用到就是爽到  
云函数版本请访问[健康打卡_河海大学版_FC](https://github.com/yin1999/healthreport_fc)(无服务器版，配置方便，**零成本**)

## 状态

目前，[最新版本](https://github.com/yin1999/healthreport/releases/latest)具有以下特性:

    1. 每日自动打卡
    2. 一次打卡失败，自动重新尝试，可设置最大打卡尝试次数
    3. 日志同步输出到Stderr
    4. 版本查询
    5. 打卡失败邮件通知推送功能(目前支持STARTTLS/TLS端口+PlainAuth登录到SMTP服务器)

## 安装教程

适用类Unix/windows，想直接使用的，请下载[release](https://github.com/yin1999/healthreport/releases/latest)版本后直接转到[使用说明](#使用说明) 

源码安装依赖[Golang](https://golang.google.cn/)-基于golang开发、[git](https://git-scm.com/)-版本管理工具以及[make](https://www.gnu.org/software/make/)-快速构建，国内使用推荐开启golang的Go module并使用国内的Go proxy服务  
推荐使用[Goproxy.cn](https://goproxy.cn/)或[阿里云 Goproxy](https://developer.aliyun.com/mirror/goproxy)

### 安装步骤

1. 环境配置，以`Centos`/`Debian`为例

	- 安装Golang[>= 1.16]: [golang.google.cn/doc/install](https://golang.google.cn/doc/install)

	- 安装 git、make:

	   ```bash
	   # Centos
	   sudo yum install git make

	   # Debian/Ubuntu
	   sudo apt install git make
	   ```

2. 通过源码下载、编译

	```bash
	# 配置Goproxy
	go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct  

	# 下载编译
	git clone --depth 1 https://github.com/yin1999/healthreport.git
	cd healthreport
	make # 若没有安装make，可以使用命令: go run _script/make.go 代替
	```

## 使用说明

### Docker

应用支持Docker部署，具体使用方法请参考[yin199909/healthreport](https://hub.docker.com/repository/docker/yin199909/healthreport)

### Linux/Windows

1. 授予可执行权限(`源码编译`或`Windows`平台的可以跳过此步)

	```bash
	chmod +x healthreport
	```

2. 运行

	```bash
	# example(set punch time as 9:52)
	./healthreport -u username -p password -t 9:52 -save
	# -save: 保存账户信息至文件，第二次启动程序时可不设置用户名、密码两个参数（仅使用: ./healthreport -t 9:52）
	```

	**Linux**用户可使用`systemd`(recommend，配置模板：`_script/healthreport.service`)或者`screen`管理打卡进程

### 邮件通知

1. 生成**email.json**

	```bash
	./healthreport -g  # 可使用 '-email' 指定配置文件生成目录
	```

2. 修改**email.json**中的的配置，具体说明如下:

	```properties
	to:    收件邮箱(string list)
	SMTP:  SMTP 配置
	    username:   SMTP用户名(string)
	    password:   SMTP用户密码(string)
	    TLS:        是否为TLS端口(bool)
	    host:       SMTP服务地址(string)
	    port:       SMTP服务端口(需支持STARTTLS/TLS)(int)
	```

3. 重启打卡服务，若提示**Email deliver enabled**，则邮件通知服务已启用

### 其它说明

1. 查看版本信息

	```bash
	./healthreport -v
	```

2. 帮助信息（提供程序的所有命令行参数）

	```bash
	./healthreport -h
	```

4. 版本更新

	```bash
	git pull
	make
	```
