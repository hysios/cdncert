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
	flag.StringVar(&domain, "domain", "", "The domain for which to obtain/upload the SSL certificate")
	flag.StringVar(&email, "email", "", "Contact email address for ACME registration")
	flag.StringVar(&accessKey, "access-key", "", "Aliyun Access Key")
	flag.StringVar(&secretKey, "secret-key", "", "Aliyun Secret Key")
	flag.BoolVar(&production, "prod", false, "Set to true to use Let's Encrypt's production environment")
	flag.StringVar(&region, "region", "cn-hangzhou", "Aliyun CDN region")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	flag.Parse()

	if domain == "" || email == "" || accessKey == "" || secretKey == "" {
		log.Fatal("All parameters (domain, email, access-key, and secret-key) are required.")
	}

	// 设置 Aliyun DNS 所需的环境变量
	os.Setenv("ALICLOUD_ACCESS_KEY", accessKey)
	os.Setenv("ALICLOUD_SECRET_KEY", secretKey)

	switch os.Args[1] {
	case "obtain":
		obtainCertificate()
	case "upload":
		uploadCertificate()
	case "auto":
		err := autoObtainAndUpload()
		if err != nil {
			log.Fatalf("Error in auto mode: %v", err)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: cdncert <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  obtain  - Obtain a new SSL certificate")
	fmt.Println("  upload  - Upload an existing certificate to Aliyun CDN")
	fmt.Println("  auto    - Automatically obtain and upload certificate")
	fmt.Println("\nRun 'cdncert <command> -h' for more information on a command.")
}

func obtainCertificate() (*certificate.Resource, error) {
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
	if production {
		config.CADirURL = lego.LEDirectoryProduction
	} else {
		config.CADirURL = lego.LEDirectoryStaging
	}

	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatalf("Error creating ACME client: %v", err)
	}

	// Set up the DNS provider
	provider, err := alidns.NewDNSProvider()
	if err != nil {
		log.Fatalf("Error creating DNS provider: %v", err)
	}

	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		log.Fatalf("Error setting DNS provider: %v", err)
	}

	// Obtain the certificate
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("error obtaining certificate: %v", err)
	}

	err = saveCertificateAndKey(certificates)
	if err != nil {
		return nil, fmt.Errorf("error saving certificate and key: %v", err)
	}

	fmt.Println("Certificate obtained successfully!")
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

func uploadCertificate() {
	certPath := filepath.Join("certificates", domain+".crt")
	keyPath := filepath.Join("certificates", domain+".key")

	// Check if certificate files exist
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		log.Fatalf("Certificate file not found: %s", certPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Fatalf("Private key file not found: %s", keyPath)
	}

	err := uploadCertificateToAliyunCDN(certPath, keyPath)
	if err != nil {
		log.Fatalf("Error uploading certificate: %v", err)
	}

	fmt.Println("Certificate uploaded successfully!")
}

func autoObtainAndUpload() error {
	cert, err := obtainCertificate()
	if err != nil {
		return err
	}

	// Save the certificate and key
	err = saveCertificateAndKey(cert)
	if err != nil {
		return fmt.Errorf("error saving certificate and key: %v", err)
	}

	// Upload the certificate
	uploadCertificate()

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
