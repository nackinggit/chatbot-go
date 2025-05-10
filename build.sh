go env -w GOPROXY=https://goproxy.cn

go mod tidy

EXEC_FILE=chatbot-go
MAIN_FILE=cmd/main.go

BRANCH=$(git rev-parse --abbrev-ref HEAD)
RELEASE=$(git describe --tags)
COMMMIT=$(git rev-parse --short HEAD)
DATE=$(date +%FT%T%z)
RELEASE=${RELEASE:-v0.0.1}
GEN_TARGET=${GEN_TARGET:-yes}

VER_PKG=git.imgo.tv/ft/go-ceres/pkg/version
LDFLAGS="-w -s -X ${VER_PKG}.version=${RELEASE} -X ${VER_PKG}.date=${DATE} -X ${VER_PKG}.commit=${COMMMIT} -X ${VER_PKG}.branch=${BRANCH}"

echo "start build"
CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o ${EXEC_FILE} ${MAIN_FILE}
