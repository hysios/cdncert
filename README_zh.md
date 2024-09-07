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
