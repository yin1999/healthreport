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
	6. 通过环境变量设置http代理(HTTP_PROXY/HTTPS_PROXY均需设置)

## 安装教程

适用类Unix，想直接使用的，请下载[release](https://github.com/yin1999/healthreport/releases/latest)版本后直接转到[使用说明](#使用说明) 

源码安装依赖[Golang](https://golang.google.cn/)-基于golang开发、[git](https://git-scm.com/)-版本管理工具、[make](https://www.gnu.org/software/make/)-快速构建，以及[tesseract-ocr](https://github.com/tesseract-ocr/tessdoc)——验证码识别，国内使用推荐开启golang的Go module并使用国内的Go proxy服务  
推荐使用[Goproxy.cn](https://goproxy.cn/)或[阿里云 Goproxy](https://developer.aliyun.com/mirror/goproxy)

## 使用说明

### Docker

应用支持Docker部署，具体使用方法请参考[yin199909/healthreport](https://hub.docker.com/r/yin199909/healthreport)

### Linux

请参考 [wiki](https://github.com/yin1999/healthreport/wiki)。
