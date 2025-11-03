<h1 align="center">
<img src="https://box-user-assets.n3ko.cc/public/Neko.png" width=100px/>

NekoBox
</h1>

<p align="center">
匿名提问箱 / Anonymous Question Box
</p>
<p align="center">
<a href="https://goreportcard.com/badge/github.com/wuhan005/NekoBox">
    <img src="https://github.com/wuhan005/NekoBox/workflows/Go/badge.svg" alt="Go Report Card">
</a>
<a href="https://sourcegraph.com/github.com/wuhan005/NekoBox">
    <img src="https://img.shields.io/badge/view%20on-Sourcegraph-brightgreen.svg?logo=sourcegraph" alt="Sourcegraph">
</a>
<a href="https://deepsource.io/gh/wuhan005/NekoBox/?ref=repository-badge">
    <img src="https://deepsource.io/gh/wuhan005/NekoBox.svg/?label=active+issues&token=7nuU5C-4QG3CP_5g9qFf3Bl9" alt="DeepSource">
</a>
<a href="https://goreportcard.com/report/github.com/wuhan005/NekoBox">
    <img src="https://goreportcard.com/badge/github.com/wuhan005/NekoBox" alt="Go Report Card">
<a>
</p>

<p align="center">
<a href="/README.zh-CN.md">简体中文</a> | <a href="/README.md">English</a>
</p>

![Screenshot](./dev/screenshot.svg)

## 部署

### Docker 部署

1. 创建配置文件

基于配置文件模板 `conf/app.sample.ini` 创建配置文件 `app.ini`，相关配置可参考注释进行调整。

2. 启动容器

```bash
# 拉取最新镜像
docker pull ghcr.io/wuhan005/nekobox:master

# 启动容器（监听 80 端口并挂载配置文件）
docker run -dt --name NekoBox -p 80:80 -v $(pwd)/app.ini:/app/conf/app.ini ghcr.io/wuhan005/nekobox:master
```

### 从源码构建

1. 环境需求

* [Go](https://golang.org/dl/) (v1.19 或更高版本)
* [MySQL](https://www.mysql.com/downloads/) (v5.7 或更高版本)
* [Redis](https://redis.io/download/) (v6.0 或更高版本)

2. 编译源码

```bash
# 克隆源码
git clone https://github.com/wuhan005/NekoBox.git

# 进入项目目录
cd NekoBox

# 构建当前机器系统与架构的二进制文件
go build -v -ldflags "-w -s -extldflags '-static'" -o NekoBox ./cmd/

# 构建 Linux、AMD64 架构的二进制文件
GOOS=linux GOARCH=amd64 go build -v -ldflags "-w -s -extldflags '-static'" -o NekoBox ./cmd/
```

3. 编辑配置文件

基于配置文件模板 `conf/app.sample.ini` 创建配置文件，相关配置可参考注释进行调整。

```bash
cp conf/app.sample.ini conf/app.ini
```

4. 运行

```bash
./NekoBox web
```

## 开源协议

MIT License
