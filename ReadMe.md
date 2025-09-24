# UrlServer

一个Windows平台的URL监控服务，用于根据配置规则匹配URL并自动切换输入法状态。
配合谷歌插件即可实现:打开规则的网址自动切换英文输入法

## 演示
<img src="https://raw.githubusercontent.com/icutool/img/main/switch.gif">

## 功能特性

- 接收URL请求并根据配置文件进行规则匹配
- 自动检测当前输入法状态（中文/英文）
- 当匹配规则且输入法为中文时，自动切换到英文输入法
- 发送桌面通知提醒用户

## 项目结构

```
UrlServer/
├── url-sender-extension            # 谷歌浏览器插件
├── UrlSwitchInput
    ├── main.go                     # 主程序入口
    ├── go.mod                      # Go模块文件
    ├── config.json                 # 配置文件
    └── internal/                   # 内部包
        ├── config/                 # 配置管理
        │   └── config.go
        ├── handler/                # HTTP处理器
        │   └── handler.go
        ├── ime/                    # 输入法控制
        │   └── ime.go
        ├── matcher/                # URL匹配器
        │   └── matcher.go
        └── notification/           # 通知服务
            └── notification.go
```

## 安装和运行

### 开箱即用
1. 下载release的包文件进行解压
2. 按照自己的需求,修改config.json文件
3. 双击运行UrlServer.exe程序
4. 谷歌浏览器扩展程序->管理扩展程序->加载未打包的扩展程序->选择url-sender-extension文件夹进行安装
5. enjoy

### 编译
1. 确保系统为Windows（代码使用了Windows API）
2. 安装Go 1.25或更高版本
3. 下载依赖：
   ```bash
   go mod download 或者 go mod tidy
   ```
4. 编译程序：
   ```bash
   go build -o UrlServer.exe main.go
   ```

服务将在 `http://localhost:7887` 启动。

## API接口

### POST /url

接收URL并进行处理。

请求体：
```json
{
  "url": "https://github.com/example/repo"
}
```

响应体：
```json
{
  "success": true,
  "message": "处理完成",
  "matched": true,
  "rule_name": "GitHub",
  "ime_status": "中文",
  "ime_switched": true
}
```

### GET /health

健康检查接口，返回服务状态。

## 配置文件

编辑 `config.json` 文件来配置URL匹配规则：

```json
{
  "rules": [
    {
      "name": "规则名称",
      "url_pattern": "匹配模式",
      "match_type": "匹配类型",
      "description": "规则描述",
      "enabled": true
    }
  ]
}
```

### 匹配类型详解

支持四种匹配类型：

#### 1. 精准匹配 (`exact`)
URL必须完全相同才能匹配
```json
{
  "name": "GitHub首页",
  "url_pattern": "https://github.com",
  "match_type": "exact"
}
```

#### 2. 关键词匹配 (`keyword`)
URL包含指定关键词即可匹配，支持多个关键词用逗号分隔
```json
{
  "name": "编程网站",
  "url_pattern": "github,stackoverflow,codepen",
  "match_type": "keyword"
}
```

#### 3. 通配符匹配 (`wildcard`)
支持 `*` 通配符，`*` 可以匹配任意字符
```json
{
  "name": "GitHub相关",
  "url_pattern": "*github.com*",
  "match_type": "wildcard"
}
```

#### 4. 正则表达式匹配 (`regex`)
支持完整的正则表达式语法
```json
{
  "name": "开发文档",
  "url_pattern": "^https?://[^/]*\\.(com|org)/(docs?|api)(/.*)?$",
  "match_type": "regex"
}
```

### 兼容性说明

- 如果不指定 `match_type`，系统会自动判断匹配类型
- 旧的配置文件无需修改，会自动兼容

## 工作流程

1. 接收URL请求
2. 根据配置文件中的规则进行匹配
3. 如果匹配成功：
    - 检测当前输入法状态
    - 如果是中文输入法，切换到英文输入法
    - 发送桌面通知
4. 返回处理结果

## 注意事项

- 仅支持Windows系统
- 需要管理员权限来控制输入法
- 确保系统已安装中文输入法
- 通知功能需要Windows 10或更高版本

## 依赖项

- `github.com/go-toast/toast` - Windows桌面通知
- `golang.org/x/sys/windows` - Windows API访问
