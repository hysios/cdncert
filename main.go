package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"path"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/registration"
)

var (
	domain        string
	email         string
	dnsAccessKey  string
	dnsSecretKey  string
	cdnAccessKey  string
	cdnSecretKey  string
	production    bool
	region        string
	onlyObtain    bool
	challengeType string
	s3Bucket      string
	s3Endpoint    string
	s3Region      string
	s3Token       string
)

func init() {
	flag.StringVar(&domain, "domain", "", "The domain for which to obtain/upload the SSL certificate")
	flag.StringVar(&email, "email", "", "Contact email address for ACME registration")
	flag.StringVar(&dnsAccessKey, "dns-access-key", "", "Aliyun Access Key")
	flag.StringVar(&dnsSecretKey, "dns-secret-key", "", "Aliyun Secret Key")
	flag.StringVar(&cdnAccessKey, "cdn-access-key", "", "Aliyun CDN Access Key")
	flag.StringVar(&cdnSecretKey, "cdn-secret-key", "", "Aliyun CDN Secret Key")
	flag.BoolVar(&production, "prod", false, "Set to true to use Let's Encrypt's production environment")
	flag.StringVar(&region, "region", "cn-hangzhou", "Aliyun CDN region")
	flag.BoolVar(&onlyObtain, "obtain", false, "Only obtain certificate, do not upload to Aliyun CDN")
	flag.StringVar(&challengeType, "challenge", "s3", "Challenge type (dns or s3)")
	flag.StringVar(&s3Bucket, "s3-bucket", "", "Aliyun OSS Bucket")
	flag.StringVar(&s3Endpoint, "s3-endpoint", "", "Aliyun OSS endpoint")
	flag.StringVar(&s3Region, "s3-region", "", "Aliyun OSS region")
	flag.StringVar(&s3Token, "s3-token", "", "Aliyun OSS token")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	if domain == "" || email == "" || cdnAccessKey == "" || cdnSecretKey == "" {
		log.Fatal("All parameters (domain, email, dns-access-key, dns-secret-key, cdn-access-key, and cdn-secret-key) are required.")
	}

	if dnsAccessKey == "" {
		dnsAccessKey = cdnAccessKey
	}

	if dnsSecretKey == "" {
		dnsSecretKey = cdnSecretKey
	}

	certs, err := obtainCertificate(dnsAccessKey, dnsSecretKey)
	if err != nil {
		log.Fatalf("Error obtaining certificate: %v", err)
	}

	err = saveCertificateAndKey(certs)
	if err != nil {
		log.Fatalf("Error saving certificate and key: %v", err)
	}

	if !onlyObtain {
		uploadCertificate(cdnAccessKey, cdnSecretKey)
	}
}

func printUsage() {
	fmt.Println("Usage: cdncert <command> [arguments]")
	flag.PrintDefaults()
}

func obtainCertificate(aliAccessKey, aliSecretKey string) (*certificate.Resource, error) {
	// Create a new ACME user
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Error generating private key: %v", err)
	}

	user := &User{
		Email: email,
		key:   privateKey,
	}

	// Create a new ACME client
	config := lego.NewConfig(user)

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("无法创建 ACME 客户端: %v", err)
	}

	// 根据不同的 challenge 类型设置不同的验证方式
	switch challengeType {
	case "dns":
		aliconfig := alidns.NewDefaultConfig()
		aliconfig.APIKey = aliAccessKey
		aliconfig.SecretKey = aliSecretKey

		dnsProvider, err := alidns.NewDNSProviderConfig(aliconfig)
		if err != nil {
			return nil, fmt.Errorf("无法创建阿里云 DNS 提供商: %v", err)
		}

		err = client.Challenge.SetDNS01Provider(dnsProvider, dns01.AddRecursiveNameservers([]string{"223.5.5.5:53", "223.6.6.6:53"}))
		if err != nil {
			return nil, fmt.Errorf("无法设置 DNS 提供商: %v", err)
		}

	case "s3":
		if s3Endpoint == "" || s3Region == "" || s3Bucket == "" {
			return nil, fmt.Errorf("使用 s3 验证方式时必须提供 s3-endpoint 和 s3-region 参数")
		}

		err = client.Challenge.SetHTTP01Provider(&s3Provider{
			accessKey: aliAccessKey,
			secretKey: aliSecretKey,
			bucket:    s3Bucket,
			endpoint:  s3Endpoint,
			region:    s3Region,
			token:     s3Token,
		})
		if err != nil {
			return nil, fmt.Errorf("无法设置 S3 提供商: %v", err)
		}

	default:
		return nil, fmt.Errorf("不支持的验证方式: %s", challengeType)
	}

	// 注册
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("无法注册: %v", err)
	}
	user.Registration = reg

	// Obtain the certificate
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("error obtaining certificate: %v", err)
	}

	return certificates, nil
}

func saveCertificateAndKey(cert *certificate.Resource) error {
	// Create a directory to store the certificates
	certDir := "certificates"
	err := os.MkdirAll(certDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating certificate directory: %v", err)
	}

	// Save the certificate
	certPath := filepath.Join(certDir, domain+".crt")
	err = ioutil.WriteFile(certPath, cert.Certificate, 0644)
	if err != nil {
		return fmt.Errorf("error saving certificate: %v", err)
	}

	// Save the private key
	keyPath := filepath.Join(certDir, domain+".key")
	err = ioutil.WriteFile(keyPath, cert.PrivateKey, 0600)
	if err != nil {
		return fmt.Errorf("error saving private key: %v", err)
	}

	fmt.Printf("Certificate saved to: %s\n", certPath)
	fmt.Printf("Private key saved to: %s\n", keyPath)

	return nil
}

func uploadCertificate(cdnAccessKey, cdnSecretKey string) {
	certPath := filepath.Join("certificates", domain+".crt")
	keyPath := filepath.Join("certificates", domain+".key")

	// Check if certificate files exist
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		log.Fatalf("Certificate file not found: %s", certPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Fatalf("Private key file not found: %s", keyPath)
	}

	err := uploadCertificateToAliyunCDN(certPath, keyPath, cdnAccessKey, cdnSecretKey)
	if err != nil {
		log.Fatalf("Error uploading certificate: %v", err)
	}

	fmt.Println("Certificate uploaded successfully!")
}

func autoObtainAndUpload(cdnAccessKey, cdnSecretKey string, dnsAccessKey, dnsSecretKey string) error {
	cert, err := obtainCertificate(cdnAccessKey, cdnSecretKey)
	if err != nil {
		return err
	}

	// Save the certificate and key
	err = saveCertificateAndKey(cert)
	if err != nil {
		return fmt.Errorf("error saving certificate and key: %v", err)
	}

	// Upload the certificate
	uploadCertificate(cdnAccessKey, cdnSecretKey)

	return nil
}

// User implements the acme.User interface
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

func uploadCertificateToAliyunCDN(certPath, keyPath, cdnAccessKey, cdnSecretKey string) error {
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
	client, err := cdn.NewClientWithAccessKey(region, cdnAccessKey, cdnSecretKey)
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

// 在文件末尾添加 S3 provider 实现
type s3Provider struct {
	accessKey string
	secretKey string
	bucket    string
	endpoint  string
	region    string
	token     string
}

func (p *s3Provider) Present(domain, token, keyAuth string) error {
	creds := credentials.NewStaticCredentials(p.accessKey, p.secretKey, p.token)
	config := &aws.Config{
		Region:           aws.String(p.region),
		Endpoint:         &p.endpoint,
		S3ForcePathStyle: aws.Bool(false),
		Credentials:      creds,
	}
	sess, err := session.NewSession(config)
	if err != nil {
		return fmt.Errorf("创建 session 失败: %v", err)
	}
	service := s3.New(sess)
	// ACME 协议要求的验证文件路径
	wellKnownPath := path.Join(".well-known", "acme-challenge", token)
	// 使用字符串读取器上传内容
	reader := strings.NewReader(keyAuth)

	// 上传验证文件
	_, err = service.PutObject(&s3.PutObjectInput{
		Bucket: &p.bucket,
		Body:   reader,
		Key:    &wellKnownPath,
	})
	if err != nil {
		return fmt.Errorf("上传验证文件失败: %v", err)
	}

	return nil
}

func (p *s3Provider) CleanUp(domain, token, keyAuth string) error {
	creds := credentials.NewStaticCredentials(p.accessKey, p.secretKey, p.token)
	config := &aws.Config{
		Region:           aws.String(p.region),
		Endpoint:         &p.endpoint,
		S3ForcePathStyle: aws.Bool(false),
		Credentials:      creds,
	}
	sess, err := session.NewSession(config)
	if err != nil {
		return fmt.Errorf("创建 session 失败: %v", err)
	}
	service := s3.New(sess)

	// 删除验证文件
	wellKnownPath := path.Join(".well-known", "acme-challenge", token)
	_, err = service.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &p.bucket,
		Key:    &wellKnownPath,
	})
	if err != nil {
		return fmt.Errorf("删除验证文件失败: %v", err)
	}

	return nil
}
