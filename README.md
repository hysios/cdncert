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
- Aliyun Access Key and Secret Key with necessary permissions

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

CDNCert supports three main commands:

1. Obtain a certificate:
   ```
   cdncert obtain -domain example.com -email your@email.com -access-key YOUR_ACCESS_KEY -secret-key YOUR_SECRET_KEY
   ```

2. Upload a certificate:
   ```
   cdncert upload -domain example.com -access-key YOUR_ACCESS_KEY -secret-key YOUR_SECRET_KEY
   ```

3. Automatically obtain and upload a certificate:
   ```
   cdncert auto -domain example.com -email your@email.com -access-key YOUR_ACCESS_KEY -secret-key YOUR_SECRET_KEY
   ```

Additional flags:
- `-prod`: Set to true to use Let's Encrypt's production environment (default is staging)
- `-region`: Specify the Aliyun CDN region (default is cn-hangzhou)

For more information on each command, use the `-h` flag:

## Automatic Renewal with Cron

To automatically renew your certificate every three months using cron on Linux, follow these steps:

1. Open your crontab file for editing:
   ```
   crontab -e
   ```

2. Add the following line to run the renewal process every 3 months:
   ```
   0 0 1 */3 * /path/to/cdncert auto -domain example.com -email your@email.com -access-key YOUR_ACCESS_KEY -secret-key YOUR_SECRET_KEY -prod true >> /path/to/cdncert_renewal.log 2>&1
   ```

   Replace the following:
   - `/path/to/cdncert` with the actual path to your cdncert executable
   - `example.com` with your domain
   - `your@email.com` with your email address
   - `YOUR_ACCESS_KEY` and `YOUR_SECRET_KEY` with your Aliyun credentials
   - `/path/to/cdncert_renewal.log` with the path where you want to store the log file

   This cron job will run at midnight on the first day of every third month.

3. Save and exit the crontab editor.

Make sure the cdncert executable has the necessary permissions to run, and that the log file path is writable.

Note: It's recommended to run the renewal process more frequently than the actual expiration date of your certificate to account for potential issues. Let's Encrypt certificates are valid for 90 days, so running the renewal every 60 days (2 months) might be a safer option:

