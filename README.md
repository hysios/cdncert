# CDNCert

CDNCert is a command-line tool for obtaining SSL certificates from Let's Encrypt and uploading them to Aliyun CDN.


[中文版本](README_zh.md)

## Features

- Obtain SSL certificates from Let's Encrypt using DNS challenge
- Upload certificates to Aliyun CDN
- Automatic mode for obtaining and uploading certificates in one step

## Prerequisites

- Go 1.16 or higher
- Aliyun account with CDN and DNS services enabled
- Aliyun Access Key and Secret Key with necessary permissions for both DNS and CDN services

## Installation

To install CDNCert, follow these steps:

1. Ensure you have Go 1.16 or higher installed on your system.

2. Clone the repository:
   ```
   git clone https://github.com/hysios/cdncert.git
   ```

3. Change to the project directory:
   ```
   cd cdncert
   ```

4. Build the executable:
   ```
   go build -o cdncert
   ```

5. (Optional) Move the executable to a directory in your PATH for easy access:
   ```
   sudo mv cdncert /usr/local/bin/
   ```

Now you can use the `cdncert` command from anywhere in your terminal.

## Usage

CDNCert supports the following command-line flags:

```
Usage: cdncert <command> [arguments]
  -cdn-access-key string
        Aliyun CDN Access Key
  -cdn-secret-key string
        Aliyun CDN Secret Key
  -dns-access-key string
        Aliyun DNS Access Key
  -dns-secret-key string
        Aliyun DNS Secret Key
  -domain string
        The domain for which to obtain/upload the SSL certificate
  -email string
        Contact email address for ACME registration
  -obtain
        Only obtain certificate, do not upload to Aliyun CDN
  -prod
        Set to true to use Let's Encrypt's production environment
  -region string
        Aliyun CDN region (default "cn-hangzhou")
```

Example usage:

```
cdncert -domain example.com -email your@email.com -dns-access-key YOUR_DNS_ACCESS_KEY -dns-secret-key YOUR_DNS_SECRET_KEY -cdn-access-key YOUR_CDN_ACCESS_KEY -cdn-secret-key YOUR_CDN_SECRET_KEY -prod
```

This command will obtain a certificate for example.com and upload it to Aliyun CDN.

To only obtain the certificate without uploading, add the `-obtain` flag:

```
cdncert -domain example.com -email your@email.com -dns-access-key YOUR_DNS_ACCESS_KEY -dns-secret-key YOUR_DNS_SECRET_KEY -cdn-access-key YOUR_CDN_ACCESS_KEY -cdn-secret-key YOUR_CDN_SECRET_KEY -obtain
```

## Automatic Renewal with Cron

To automatically renew your certificate every two months using cron on Linux, follow these steps:

1. Open your crontab file for editing:
   ```
   crontab -e
   ```

2. Add the following line to run the renewal process every 2 months:
   ```
   0 0 1 */2 * /path/to/cdncert -domain example.com -email your@email.com -dns-access-key YOUR_DNS_ACCESS_KEY -dns-secret-key YOUR_DNS_SECRET_KEY -cdn-access-key YOUR_CDN_ACCESS_KEY -cdn-secret-key YOUR_CDN_SECRET_KEY -prod >> /path/to/cdncert_renewal.log 2>&1
   ```

   Replace the placeholders with your actual values.

3. Save and exit the crontab editor.

Make sure the cdncert executable has the necessary permissions to run, and that the log file path is writable.

Note: Running the renewal process every two months ensures that your certificate is renewed well before its 90-day expiration, accounting for potential issues or delays.

## Release

This project uses GoReleaser to simplify the release process. To create a new release:

1. Ensure you have GoReleaser installed. If not, you can install it with:
   ```
   go install github.com/goreleaser/goreleaser@latest
   ```

2. Create and push a new tag:
   ```
   git tag -a v0.1.0 -m "First release"
   git push origin v0.1.0
   ```

3. Run GoReleaser:
   ```
   goreleaser release --clean
   ```

This will create a new GitHub release with binaries for different platforms.

For more information on using GoReleaser, please refer to the [GoReleaser documentation](https://goreleaser.com/).


### Automatic Releases

This project is configured with GitHub Actions to automatically create releases when a new tag is pushed. To trigger an automatic release:

1. Create a new tag locally:
   ```
   git tag -a v1.0.0 -m "Release version 1.0.0"
   ```

2. Push the tag to GitHub:
   ```
   git push origin v1.0.0
   ```

GitHub Actions will automatically run GoReleaser to build and publish the release.
