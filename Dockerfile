FROM golang:1.25 AS builder

# 设置工作目录
WORKDIR /src

# 配置 Git 和 Go 环境（合并 RUN 命令减少层数）
RUN go env -w GO111MODULE=on && \
    go env -w GONOSUMDB=\* && \
    go env -w CGO_ENABLED=0 && \
    go env -w GOOS=linux && \
    go env -w GOARCH=amd64 && \
    go env -w GOPROXY=https://goproxy.cn,direct

# 安装 Task 工具
RUN go install github.com/go-task/task/v3/cmd/task@latest

# 复制 go.mod 和 go.sum 文件（利用 Docker 缓存）
COPY go.mod go.sum ./
COPY shared/go.mod shared/go.sum ./shared/

# 使用 BuildKit 缓存下载依赖
RUN --mount=type=cache,target=~/go/pkg/mod \
    --mount=type=cache,target=~/.cache/go-build \
    go mod download

# 复制源代码和 Taskfile
COPY . .

# 切换到服务目录并使用 BuildKit 缓存构建
WORKDIR /src
RUN --mount=type=cache,target=~/go/pkg/mod \
    --mount=type=cache,target=~/.cache/go-build \
    task build

# 运行阶段
FROM debian:13 AS runtime

# 一次复制所有文件并设置权限
COPY --from=builder --chmod=755 /src/bin/server ./server

# 设置工作目录
WORKDIR /app

CMD ["/app/server", "--config-dir", "/app/configs"]
