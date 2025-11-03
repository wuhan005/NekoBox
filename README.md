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

## Deployment

### Docker Deployment

1. Create a Configuration File

Create a configuration file `app.ini` based on the template `conf/app.sample.ini`. Adjust the settings as needed by
referring to the comments in the file.

2. Start the Container

```bash
# Pull the latest image
docker pull ghcr.io/wuhan005/nekobox:master

# Start the container (listen on port 80 and mount the configuration file)
docker run -dt --name NekoBox -p 80:80 -v $(pwd)/app.ini:/app/conf/app.ini ghcr.io/wuhan005/nekobox:master
```

### Build from Source

1. Requirements

* [Go](https://golang.org/dl/) (v1.19 or higher)
* [MySQL](https://www.mysql.com/downloads/) (v5.7 or higher)
* [Redis](https://redis.io/download/) (v6.0 or higher)

2. Compile the Source Code

```bash
# Clone the source code
git clone https://github.com/wuhan005/NekoBox.git

# Enter the project directory
cd NekoBox

# Build the binary for the current system and architecture
go build -v -ldflags "-w -s -extldflags '-static'" -o NekoBox ./cmd/

# Build the binary for Linux, AMD64 architecture
GOOS=linux GOARCH=amd64 go build -v -ldflags "-w -s -extldflags '-static'" -o NekoBox ./cmd/
```

3. Edit the Configuration File

Create a configuration file based on the template `conf/app.sample.ini`. Adjust the settings as needed by referring to
the comments in the file.

```bash
cp conf/app.sample.ini conf/app.ini
```

4. Run

```bash
./NekoBox web
```

## License

MIT License
