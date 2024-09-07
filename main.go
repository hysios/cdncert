package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/registration"
)

var (
	domain     string
	email      string
	accessKey  string
	secretKey  string
	production bool
	region     string
)

func init() {
	flag.StringVar(&domain, "domain", "", "The domain for which to obtain the SSL certificate")
	flag.StringVar(&email, "email", "", "Contact email address for ACME registration")
	flag.StringVar(&accessKey, "access-key", "", "Aliyun Access Key")
	flag.StringVar(&secretKey, "secret-key", "", "Aliyun Secret Key")
	flag.BoolVar(&production, "prod", false, "Set to true to use Let's Encrypt's production environment")
	flag.StringVar(&region, "region", "cn-hangzhou", "Aliyun CDN region")
	flag.Parse()

	if domain == "" || email == "" || accessKey == "" || secretKey == "" {
		log.Fatal("All parameters (domain, email, access-key, and secret-key) are required.")
	}

	// 设置 Aliyun DNS 所需的环境变量
	os.Setenv("ALICLOUD_ACCESS_KEY", accessKey)
	os.Setenv("ALICLOUD_SECRET_KEY", secretKey)
}

func main() {
	// 创建新的 ECDSA 私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Error generating private key: %v", err)
	}

	// 创建用户
	user := &User{
		Email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(user)

	// 这里我们使用 ACME v2 的 URL
	if production {
		config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	} else {
		config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	}

	// 创建 ACME 客户端
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	// 创建 DNS 提供商
	dnsProvider, err := alidns.NewDNSProvider()
	if err != nil {
		log.Fatalf("Error creating DNS provider: %v", err)
	}

	err = client.Challenge.SetDNS01Provider(dnsProvider, dns01.AddRecursiveNameservers([]string{"223.5.5.5:53", "223.6.6.6:53"}))

	if err != nil {
		log.Fatalf("Error setting DNS provider: %v", err)
	}

	// 新建账户
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatalf("Error registering account: %v", err)
	}
	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatalf("Error obtaining certificate: %v", err)
	}

	// 创建 certs 目录
	certsDir := "certs"
	err = os.MkdirAll(certsDir, 0755)
	if err != nil {
		log.Fatalf("Error creating certs directory: %v", err)
	}

	// 写入证书和私钥到 certs 目录
	certPath := filepath.Join(certsDir, domain+".crt")
	keyPath := filepath.Join(certsDir, domain+".key")

	err = ioutil.WriteFile(certPath, certificates.Certificate, 0644)
	if err != nil {
		log.Fatalf("Error writing certificate: %v", err)
	}

	err = ioutil.WriteFile(keyPath, certificates.PrivateKey, 0600)
	if err != nil {
		log.Fatalf("Error writing private key: %v", err)
	}

	log.Printf("Certificate obtained for domain: %s\n", domain)
	log.Printf("Certificate Path: %s\n", certPath)
	log.Printf("Private Key Path: %s\n", keyPath)

	// 上传证书到阿里云 CDN
	err = uploadCertificateToAliyunCDN(certPath, keyPath)
	if err != nil {
		log.Printf("Failed to upload certificate to Aliyun CDN: %v", err)
	} else {
		log.Println("Certificate uploaded to Aliyun CDN successfully")
	}
}

// User 实现了 acme.User 接口
type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func uploadCertificateToAliyunCDN(certPath, keyPath string) error {
	// 读取证书和私钥文件
	certContent, err := ioutil.ReadFile(certPath)
	if err != nil {
		return err
	}
	keyContent, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return err
	}

	// 创建 CDN 客户端，使用 flag 参数 region
	client, err := cdn.NewClientWithAccessKey(region, accessKey, secretKey)
	if err != nil {
		return err
	}

	// 创建 SetDomainServerCertificate 请求
	request := cdn.CreateSetDomainServerCertificateRequest()
	request.Scheme = "https"
	request.DomainName = domain
	request.CertType = "upload"
	request.ServerCertificateStatus = "on"
	request.ServerCertificate = string(certContent)
	request.PrivateKey = string(keyContent)

	// 发送请求
	_, err = client.SetDomainServerCertificate(request)
	if err != nil {
		return err
	}

	return nil
}
