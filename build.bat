@echo off 
Title 编译
Color 0A 
set GOARCH=amd64
:caozuo  
echo. 
echo ══════════════════════════════════════
echo    说明：使用upx工具进行压缩编译文件
echo    1.编译 - Win64
echo    2.编译 - Darwin
echo    3.编译 - Linux64
echo.
set /p n=请输入编号: 
if "%n%"=="" cls&goto :caozuo 
if "%n%"=="1" call :1 
if "%n%"=="2" call :2 
if "%n%"=="3" call :3 
if /i "%n%"=="n" exit 
pause 
goto :eof 

:1 
echo 编译开始

set GOOS=windows
set File=.\bin\vault-win64.exe
if exist %File% (
   del %File%
)

go build -ldflags "-w -s" -a -installsuffix cgo -o bin/vault-win64.exe ./main.go
upx -9 .\bin\vault-win64.exe

echo 编译完成
goto :caozuo 

:2 
echo 编译开始

set GOOS=darwin
set File=.\bin\vault-darwin
if exist %File% (
   del %File%
)

go build -ldflags "-w -s" -a -installsuffix cgo -o bin/vault-darwin ./main.go
upx -9 .\bin\vault-darwin
echo 编译完成
goto :caozuo 

:3 
echo 编译开始
set GOOS=linux
set File=.\bin\vault-linux64
if exist %File% (
   del %File%
)

go build -ldflags "-w -s" -a -installsuffix cgo -o bin/vault-linux64 ./main.go
upx -9 .\bin\vault-linux64
echo 编译完成
goto :caozuo 