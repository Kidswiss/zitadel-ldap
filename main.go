package main

import (
	"flag"
	"fmt"

	ghandler "git.earthnet.ch/simon.beck/zitadel-ldap/handler"
	"git.earthnet.ch/simon.beck/zitadel-ldap/zitadel"
	"github.com/glauth/glauth/v2/pkg/handler"
	"github.com/rs/zerolog"
)

var (
	zitadelURL = flag.String("zitadelURL", "", "issuer of your ZITADEL instance (in the form: https://<instance>.zitadel.cloud or https://<yourdomain>)")
	pat        = flag.String("pat", "", "Pat for service account that can manage Zitadel users")
)

func main() {
	flag.Parse()

	// ctx := context.Background()

	c := zitadel.NewClient(*zitadelURL, *pat, &zerolog.Logger{})

	successful, err := c.Login("test", "Testtest1234.")
	if err != nil {
		panic(err)
	}

	if successful {
		fmt.Println("logged in!")
	} else {
		fmt.Println("loggin failed")
	}

}

func NewZitadelHandler(opts ...handler.Option) handler.Handler {
	return ghandler.NewZitadelHandler(opts...)
}
