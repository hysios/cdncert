# CDNCert

CDNCert 是一个命令行工具，用于从 Let's Encrypt 获取 SSL 证书并将其上传到阿里云 CDN。

## 功能特点

- 使用 DNS 验证从 Let's Encrypt 获取 SSL 证书
- 将证书上传到阿里云 CDN
- 自动模式，一步完成证书获取和上传

## 前提条件

- Go 1.16 或更高版本
- 已启用 CDN 和 DNS 服务的阿里云账户
- 具有必要权限的阿里云 Access Key 和 Secret Key（用于 DNS 和 CDN 服务）

## 安装

按照以下步骤安装 CDNCert：

1. 确保您的系统上安装了 Go 1.16 或更高版本。

2. 克隆仓库：
   ```
   git clone https://github.com/hysios/cdncert.git
   ```

3. 进入项目目录：
   ```
   cd cdncert
   ```

4. 构建可执行文件：
   ```
   go build -o cdncert
   ```

5. （可选）将可执行文件移动到 PATH 中的目录，以便于访问：
   ```
   sudo mv cdncert /usr/local/bin/
   ```

现在您可以在终端的任何位置使用 `cdncert` 命令。

## 使用方法

CDNCert 支持以下命令行参数：

```
使用方法: cdncert <命令> [参数]
  -cdn-access-key string
        阿里云 CDN Access Key
  -cdn-secret-key string
        阿里云 CDN Secret Key
  -dns-access-key string
        阿里云 DNS Access Key
  -dns-secret-key string
        阿里云 DNS Secret Key
  -domain string
        需要获取/上传 SSL 证书的域名
  -email string
        ACME 注册的联系邮箱
  -obtain
        仅获取证书，不上传到阿里云 CDN
  -prod
        设置为 true 以使用 Let's Encrypt 的生产环境
  -region string
        阿里云 CDN 区域（默认为 "cn-hangzhou"）
```

使用示例：

```
cdncert -domain example.com -email your@email.com -dns-access-key YOUR_DNS_ACCESS_KEY -dns-secret-key YOUR_DNS_SECRET_KEY -cdn-access-key YOUR_CDN_ACCESS_KEY -cdn-secret-key YOUR_CDN_SECRET_KEY -prod
```

此命令将为 example.com 获取证书并上传到阿里云 CDN。

要仅获取证书而不上传，请添加 `-obtain` 参数：

```
cdncert -domain example.com -email your@email.com -dns-access-key YOUR_DNS_ACCESS_KEY -dns-secret-key YOUR_DNS_SECRET_KEY -cdn-access-key YOUR_CDN_ACCESS_KEY -cdn-secret-key YOUR_CDN_SECRET_KEY -obtain
```

## 使用 Cron 自动续期

要使用 Linux 上的 cron 每两个月自动续期您的证书，请按照以下步骤操作：

1. 打开您的 crontab 文件进行编辑：
   ```
   crontab -e
   ```

2. 添加以下行以每 2 个月运行一次续期过程：
   ```
   0 0 1 */2 * /path/to/cdncert -domain example.com -email your@email.com -dns-access-key YOUR_DNS_ACCESS_KEY -dns-secret-key YOUR_DNS_SECRET_KEY -cdn-access-key YOUR_CDN_ACCESS_KEY -cdn-secret-key YOUR_CDN_SECRET_KEY -prod >> /path/to/cdncert_renewal.log 2>&1
   ```

   请将占位符替换为您的实际值。

3. 保存并退出 crontab 编辑器。

确保 cdncert 可执行文件具有必要的运行权限，并且日志文件路径是可写的。

注意：每两个月运行一次续期过程可以确保您的证书在 90 天的有效期到期之前得到更新，以应对潜在的问题或延迟。

## 发布

本项目使用 GoReleaser 来简化发布过程。要创建新的发布版本：

1. 确保已安装 GoReleaser。如果没有，可以使用以下命令安装：
   ```
   go install github.com/goreleaser/goreleaser@latest
   ```

2. 创建并推送新的标签：
   ```
   git tag -a v0.1.0 -m "首次发布"
   git push origin v0.1.0
   ```

3. 运行 GoReleaser：
   ```
   goreleaser release --clean
   ```

这将创建一个新的 GitHub 发布版本，其中包含适用于不同平台的二进制文件。

有关使用 GoReleaser 的更多信息，请参阅 [GoReleaser 文档](https://goreleaser.com/)。

### 自动发布

本项目配置了 GitHub Actions，当推送新标签时会自动创建发布版本。要触发自动发布：

1. 在本地创建新标签：
   ```
   git tag -a v1.0.0 -m "发布版本 1.0.0"
   ```

2. 将标签推送到 GitHub：
   ```
   git push origin v1.0.0
   ```

GitHub Actions 将自动运行 GoReleaser 来构建和发布新版本。
