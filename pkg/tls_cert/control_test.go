package tlscert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"mkBlog/config"
	"testing"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/registration"
)

func i() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("GenerateKey failed: ", err)
		return
	}
	leuser = MyUser{
		Email: "mkitsdts@outlook.com",
		key:   privateKey,
	}
	newconfig := lego.NewConfig(&leuser)
	newconfig.Certificate.KeyType = certcrypto.RSA2048
	client, err = lego.NewClient(newconfig)
	if err != nil {
		fmt.Println("Failed to create lego client", " err :", err)
	}
	cfg := alidns.NewDefaultConfig()
	cfg.APIKey = config.Cfg.CertControl.Key
	cfg.SecretKey = config.Cfg.CertControl.Secret
	if p, err = alidns.NewDNSProviderConfig(cfg); err != nil {
		fmt.Println("Failed to create dns provider config", " err : ", err)
		return
	}

	if err := client.Challenge.SetDNS01Provider(p); err != nil {
		fmt.Println("Failed to set dns to provider", " err : ", err)
		return
	}
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		fmt.Println("Failed to register ", "err : ", err)
	}
	leuser.Registration = reg
}

func TestCheckExpireDate(t *testing.T) {
	i()
	if checkExpireDate("./static/server.pem") {
		fmt.Println("Need update")
	} else {
		fmt.Println("Don't need to update")
	}
}

func TestUpdateCert(t *testing.T) {
	i()
	if err := applyTLSCert("./static/server.key", "./static/server.crt"); err != nil {
		fmt.Println("Apply cert failed")
	}
	fmt.Println("Apply cert success")
}
