Vault 1.1.8 rc1
--------------------

安装必须组件 (GO Version 1.16+ is *required*)

## 编译
```shell
# 编译前，必须安装 upx 工具
# upx 工具用于压缩程序

# 编译MAC版
make mac

# 编译Linux版
make linux

# 编译windows版
make win
```

## 添加命令
```markdown
vault file 文件加密、解密
vault finance 金融模块
```

## 使用说明

### 配置文件
config.json
```json
{
	"backend": {
		"file": {
			"path": "vault/file"
		}
	},
	"ui": true,
	"default_lease_ttl": "168h",
	"max_lease_ttl": "720h",
	"listener": {
		"tcp": {
			"address": "0.0.0.0:8200",
			"tls_disable": "1",
			"max_request_size": 8388608,
			"max_request_duration": "30s"
		}
	},
	"disable_mlock": true,
	"api_addr": "http://127.0.0.1:8200",
	"cluster_name": "jiangjiali",
	"plugin_directory": "vault/plugins"
}

```

### 运行
```shell
# 运行测试模式
vault server -dev

# 通过配置文件运行
vault server -config=config.json
```

## API接口
http://127.0.0.1:8200/v1/sys/internal/specs/openapi
