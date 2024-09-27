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

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/registration"
)

var (
	domain       string
	email        string
	dnsAccessKey string
	dnsSecretKey string
	cdnAccessKey string
	cdnSecretKey string
	production   bool
	region       string
	onlyObtain   bool
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
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	if domain == "" || email == "" || dnsAccessKey == "" || dnsSecretKey == "" || cdnAccessKey == "" || cdnSecretKey == "" {
		log.Fatal("All parameters (domain, email, dns-access-key, dns-secret-key, cdn-access-key, and cdn-secret-key) are required.")
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

	// Set up the DNS provider
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
