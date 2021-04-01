@echo off 
Title ����
Color 0A 
set GOARCH=amd64
:caozuo  
echo. 
echo �T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T�T
echo    ˵����ʹ��upx���߽���ѹ�������ļ�
echo    1.���� - Win64
echo    2.���� - Darwin
echo    3.���� - Linux64
echo.
set /p n=��������: 
if "%n%"=="" cls&goto :caozuo 
if "%n%"=="1" call :1 
if "%n%"=="2" call :2 
if "%n%"=="3" call :3 
if /i "%n%"=="n" exit 
pause 
goto :eof 

:1 
echo ���뿪ʼ

set GOOS=windows
set File=.\bin\vault-win64.exe
if exist %File% (
   del %File%
)

go build -ldflags "-w -s" -a -installsuffix cgo -o bin/vault-win64.exe ./main.go
upx -9 .\bin\vault-win64.exe

echo �������
goto :caozuo 

:2 
echo ���뿪ʼ

set GOOS=darwin
set File=.\bin\vault-darwin
if exist %File% (
   del %File%
)

go build -ldflags "-w -s" -a -installsuffix cgo -o bin/vault-darwin ./main.go
upx -9 .\bin\vault-darwin
echo �������
goto :caozuo 

:3 
echo ���뿪ʼ
set GOOS=linux
set File=.\bin\vault-linux64
if exist %File% (
   del %File%
)

go build -ldflags "-w -s" -a -installsuffix cgo -o bin/vault-linux64 ./main.go
upx -9 .\bin\vault-linux64
echo �������
goto :caozuo 