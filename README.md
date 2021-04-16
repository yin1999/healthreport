# 健康打卡

[![build](https://github.com/yin1999/healthreport/actions/workflows/Build.yml/badge.svg)](https://github.com/yin1999/healthreport/actions/workflows/Build.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/yin1999/healthreport)](https://goreportcard.com/report/github.com/yin1999/healthreport) [![Go Reference](https://pkg.go.dev/badge/github.com/yin1999/healthreport.svg)](https://pkg.go.dev/github.com/yin1999/healthreport)

项目使用http请求模拟整个打卡过程，速度很快！  
一键打卡，用到就是爽到  
云函数版本请访问[健康打卡_河海大学版_FC](https://gitee.com/allo123/healthreport_fc)

## 状态

目前，[最新版本](https://github.com/yin1999/healthreport/releases/latest)具有以下特性

    1. 每日自动打卡
    2. 一次打卡失败，自动重新尝试，可设置最大打卡尝试次数以及重新打卡的等待时间
    3. 日志同步输出到Stdout以及log文件
    4. 版本查询
    5. 打卡失败邮件通知推送功能(目前支持STARTTLS/TLS端口+PlainAuth登录到SMTP服务器)

## 安装教程

适用类Unix/windows，想直接使用的，请下载[release](https://gitee.com/allo123/healthreport/releases)版本后直接转到[使用说明](#使用说明) 

源码安装依赖[Golang](https://golang.google.cn/)-基于golang开发、[git](https://git-scm.com/)-版本管理工具以及[make](https://www.gnu.org/software/make/)-快速构建，国内使用推荐开启golang的Go module并使用国内的Go proxy服务  
推荐使用[Goproxy 中国](https://goproxy.cn/)或[阿里云 Goproxy](https://developer.aliyun.com/mirror/goproxy)。

### 安装步骤

1. 环境配置，以CentOS 7为例

    安装软件：Golang[>= 1.16]、screen、git、make

       yum install -y golang screen git make

2. 通过源码下载、编译

       go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct  #配置Goproxy

       #下载编译
       git clone https://github.com/yin1999/healthreport.git
       cd healthreport
       make

3. 运行

       screen ./healthreport
       # 输入必要信息跳出验证账号密码成功后
       # 键盘CTRL+A+D离开screen进程，后台会每天自动运行

## 使用说明

### linux

1. 安装screen

       sudo yum install screen  #CentOS
       sudo apt-get install screen  #Ubuntu

2. 运行

       chmod +x healthreport
       screen ./healthreport

**请使用ctrl+a+d退出screen进程，ctrl+c是用来终止程序的**

### Windows

命令行中执行

    .\healthreport

### 邮件通知

1. 复制**email-template.json**命名为**email.json**

       cp email-template.json email.json  #Linux命令

2. 修改**email.json**的权限为仅使用者可读、可修改(0600权限)，保证数据安全，Windows用户请通过文件属性删除**Users**的读取权限

       chmod 600 email.json  #Linux命令

3. 修改**email.json**中的的配置，具体说明如下:

       to:    收件邮箱(string list)  
       SMTP:  SMTP 配置  
         username:   SMTP用户名(string)  
         password:   SMTP用户密码(string)  
         TLS:        是否为TLS端口(bool)  
         host:       SMTP服务地址(string)  
         port:       SMTP服务端口(需支持STARTTLS/TLS)(int)

4. 重启打卡服务，若提示**Email deliver enabled**，则邮件通知服务已启用

### 其它说明

1. 查看版本信息

       ./healthreport -v

2. 帮助信息

       ./healthreport -h

3. 验证SMTP服务

       ./healthreport -c

4. 版本更新

       git pull
       make clean && make
