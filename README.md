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



