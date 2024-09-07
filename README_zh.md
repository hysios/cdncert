# CDNCert

CDNCert 是一个命令行工具，用于从 Let's Encrypt 获取 SSL 证书并将其上传到阿里云 CDN。

## 功能特点

- 使用 DNS 验证从 Let's Encrypt 获取 SSL 证书
- 将证书上传到阿里云 CDN
- 自动模式，一步完成证书获取和上传

## 前提条件

- Go 1.16 或更高版本
- 已启用 CDN 和 DNS 服务的阿里云账户
- 具有必要权限的阿里云 Access Key 和 Secret Key

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

CDNCert 支持三个主要命令：

1. 获取证书：
   ```
   cdncert obtain -domain example.com -email your@email.com -access-key YOUR_ACCESS_KEY -secret-key YOUR_SECRET_KEY
   ```

2. 上传证书：
   ```
   cdncert upload -domain example.com -access-key YOUR_ACCESS_KEY -secret-key YOUR_SECRET_KEY
   ```

3. 自动获取并上传证书：
   ```
   cdncert auto -domain example.com -email your@email.com -access-key YOUR_ACCESS_KEY -secret-key YOUR_SECRET_KEY
   ```

附加标志：
- `-prod`：设置为 true 以使用 Let's Encrypt 的生产环境（默认为测试环境）
- `-region`：指定阿里云 CDN 区域（默认为 cn-hangzhou）

要获取每个命令的更多信息，请使用 `-h` 标志。

## 使用 Cron 自动续期

要使用 Linux 上的 cron 每三个月自动续期您的证书，请按照以下步骤操作：

1. 打开您的 crontab 文件进行编辑：
   ```
   crontab -e
   ```

2. 添加以下行以每 3 个月运行一次续期过程：
   ```
   0 0 1 */3 * /path/to/cdncert auto -domain example.com -email your@email.com -access-key YOUR_ACCESS_KEY -secret-key YOUR_SECRET_KEY -prod true >> /path/to/cdncert_renewal.log 2>&1
   ```

   替换以下内容：
   - `/path/to/cdncert` 为您的 cdncert 可执行文件的实际路径
   - `example.com` 为您的域名
   - `your@email.com` 为您的电子邮件地址
   - `YOUR_ACCESS_KEY` 和 `YOUR_SECRET_KEY` 为您的阿里云凭证
   - `/path/to/cdncert_renewal.log` 为您想要存储日志文件的路径

   这个 cron 任务将在每三个月的第一天午夜运行。

3. 保存并退出 crontab 编辑器。

确保 cdncert 可执行文件具有必要的运行权限，并且日志文件路径是可写的。

注意：建议比证书实际过期日期更频繁地运行续期过程，以应对潜在的问题。Let's Encrypt 证书的有效期为 90 天，所以每 60 天（2 个月）运行一次续期可能是一个更安全的选择：

