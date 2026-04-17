# 📋 OpenClipboard

基于 Go + WebSocket 的实时共享剪切板系统。打开网页即可获得一个房间，将房间号分享给他人，即可实现多设备间的文本实时同步。

同时提供命令行客户端，可在无浏览器的环境中使用，并支持自动同步本机剪切板。

## 功能特性

- **自动分配房间** — 打开页面 / 启动客户端自动创建 6 位房间号
- **加入已有房间** — 输入房间号加入他人的共享空间
- **实时文本同步** — 同房间内所有用户的内容实时同步
- **命令行客户端** — 支持 Windows / Linux，可自动同步本机剪切板
- **开箱即用下载** — 网页下载的客户端文件名已内嵌服务器地址，无需配置
- **断线自动重连** — WebSocket 断开后 2 秒自动重连
- **字数限制** — 最大 10000 字符

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go + [gorilla/websocket](https://github.com/gorilla/websocket) |
| 前端 | 单文件 HTML + [TailwindCSS](https://tailwindcss.com/) (CDN) |
| 通信 | WebSocket (JSON) |

## 快速开始

### 服务端

1. 从 [Releases](https://github.com/Clizo1209/OpenClipboard/releases) 下载对应平台的服务端可执行文件
2. 运行：

```bash
# Windows
openclipboard-windows-amd64.exe

# Linux / macOS
chmod +x openclipboard-linux-amd64
./openclipboard-linux-amd64
```

3. 打开浏览器访问 `http://localhost:8080`

默认监听 `:8080`，可通过 `-port` 参数自定义：

```bash
openclipboard-windows-amd64.exe -port 3000
```

### 命令行客户端

#### 方式一：从网页下载（推荐）

在 Web 界面底部点击对应平台的下载按钮。下载的文件名中已自动内嵌当前服务器地址，下载后直接运行，无需任何配置。

#### 方式二：手动下载

从 Releases 下载客户端后，首次启动时程序会提示输入服务器地址（`ws://` 或 `wss://` 开头）。

也可以将服务器地址写入文件名，程序启动时自动识别：

```
opencb-client(ws)[myserver.com]{8080}-windows-amd64.exe
opencb-client(wss)[myserver.com]{8443}-linux-amd64
opencb-client(wss)[myserver.com]-linux-arm64    # 省略端口
```

#### 客户端命令

```
join <房间号>   加入指定房间（自动转大写）
push            将本机剪切板内容推送到房间
pull            将房间内容写入本机剪切板
show            显示当前房间内容
status          显示房间状态（人数 / 字符数）
auto push       本机剪切板变化时自动推送
auto pull       房间更新时自动写入本机剪切板
auto both       双向自动同步
auto off        关闭自动同步
help            显示帮助
quit / exit     退出
```

> Linux 剪切板需要系统安装 `xclip` 或 `xsel`：`apt install xclip`

## 项目结构

```
OpenClipboard/
├── main.go           # 服务端：HTTP + WebSocket + 房间管理
├── go.mod / go.sum
├── public/
│   ├── index.html    # 前端页面
│   └── client/       # 编译后的客户端二进制（由 build.bat 生成）
├── client/
│   ├── main.go       # 命令行客户端源码
│   ├── go.mod / go.sum
│   └── build.bat     # 客户端多平台编译（输出到 public/client/）
├── build.bat         # 一键编译：先编客户端，再编服务端（服务端嵌入 public/）
└── README.md
```

## 从源码编译

需要 **Go 1.21+**。

```bash
git clone https://github.com/Clizo1209/OpenClipboard.git
cd OpenClipboard

# 一键编译所有平台（服务端 + 客户端）
build.bat

# 仅本机运行
go run main.go
```

产物输出到 `build/` 目录（服务端）和 `public/client/` 目录（客户端）。

服务端二进制通过 `go:embed` 将整个 `public/` 目录（含客户端二进制）打包进单一可执行文件，部署时只需一个文件。
