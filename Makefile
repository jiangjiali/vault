OUTPUT?=bin

mac:
	GOARCH=amd64 CGO_ENABLED=1 GOOS=darwin go build -ldflags "-w -s" -a -installsuffix cgo -o ${OUTPUT}/vault-mac ./main.go
	upx -9 ${OUTPUT}/vault-mac

linux:
	GOARCH=amd64 CGO_ENABLED=1 GOOS=linux go build -ldflags "-w -s" -a -installsuffix cgo -o ${OUTPUT}/vault-linux ./main.go
	upx -9 ${OUTPUT}/vault-linux

win:
	GOARCH=amd64 CGO_ENABLED=1 GOOS=windows go build -ldflags "-w -s" -a -installsuffix cgo -o ${OUTPUT}/vault-windows.exe ./main.go
	upx -9 ${OUTPUT}/vault-windows.exe

