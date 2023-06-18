#64bit
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/wechatbot-amd64.exe main.go

# 64-bit
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/wechatbot-amd64-linux main.go

# 64-bit 
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o bin/wechatbot-amd64-darwin main.go 