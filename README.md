# 📋 OpenClipboard

基于 Go + WebSocket 的实时共享剪切板系统。打开网页即可获得一个房间，将房间号分享给他人，即可实现多设备间的文本实时同步。

## 功能特性

- **自动分配房间** — 打开页面自动创建 6 位房间号
- **加入已有房间** — 输入房间号加入他人的共享空间
- **实时文本同步** — 同房间内所有用户的文本框内容实时同步
- **一键复制 / 清空** — 快捷操作按钮
- **点击复制房间号** — 方便分享
- **自动解散房间** — 所有人离开后自动清除房间数据
- **断线自动重连** — WebSocket 断开后 2 秒自动重连
- **字数限制** — 最大 10000 字符

## 技术栈

| 组件 | 技术                                                       |
| ---- | ---------------------------------------------------------- |
| 后端 | Go +[gorilla/websocket](https://github.com/gorilla/websocket) |
| 前端 | 单文件 HTML +[TailwindCSS](https://tailwindcss.com/) (CDN)    |
| 通信 | WebSocket (JSON)                                           |

## 快速开始

1. 从 [Releases](https://github.com/Clizo1209/OpenClipboard/releases) 下载对应平台的可执行文件
2. 运行：

```bash
# Windows
openclipboard-windows-amd64.exe

# Linux / macOS
chmod +x openclipboard-linux-amd64
./openclipboard-linux-amd64
```

3. 打开浏览器访问 `http://localhost:8080` 即可使用

`index.html` 已通过 `go:embed` 嵌入二进制文件，只需 **一个可执行文件** 即可运行，无需额外文件。

可用平台：

| 文件名 | 平台 |
|--------|------|
| `openclipboard-windows-amd64.exe` | Windows x64 |
| `openclipboard-windows-arm64.exe` | Windows ARM64 |
| `openclipboard-linux-amd64` | Linux x64 |
| `openclipboard-linux-arm64` | Linux ARM64 |
| `openclipboard-linux-armv7` | Linux ARMv7 (树莓派等) |
| `openclipboard-darwin-amd64` | macOS Intel |
| `openclipboard-darwin-arm64` | macOS Apple Silicon |
| `openclipboard-freebsd-amd64` | FreeBSD x64 |

默认监听 `:8080` 端口。可通过 `-port` 参数自定义：

```bash
# 使用 3000 端口
openclipboard-windows-amd64.exe -port 3000

# Linux / macOS
./openclipboard-linux-amd64 -port 3000
```

## 从源码编译

需要 **Go 1.21+** 环境。

```bash
# 克隆项目
git clone https://github.com/Clizo1209/OpenClipboard.git
cd OpenClipboard

# 安装依赖
go mod tidy

# 直接运行
go run main.go
```

### 本机编译

```bash
go build -ldflags="-s -w" -trimpath -buildvcs=false -o openclipboard.exe main.go
```

### 一键多平台编译

运行 `build.bat` 自动编译所有常见平台：

```bash
build.bat
```

产物输出到 `build/` 目录。

## 项目结构

```
OpenClipboard/
├── go.mod          # Go 模块定义
├── go.sum          # 依赖校验
├── main.go         # 后端：HTTP 服务 + WebSocket + 房间管理
├── index.html      # 前端：单页面 UI + WebSocket 客户端
├── build.bat       # 多平台编译脚本
└── README.md
```
