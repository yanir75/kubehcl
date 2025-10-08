package settings

import (
	"time"

	"github.com/spf13/pflag"
)

type RepoAddOptions struct {
	Name                 string
	Url                  string
	Username             string
	Password             string
	Timeout              time.Duration

	CertFile              string
	KeyFile               string
	CaFile                string
	InsecureSkipTLSverify bool
	PlainHttp bool
	RepoFile  string
	RepoCache string
}

func NewRepoSettings() *RepoAddOptions{
	return &RepoAddOptions{}
}


func AddRepoSettings(o *RepoAddOptions, fs *pflag.FlagSet) {

	fs.StringVar(&o.Username, "username", "", "chart repository username")
	fs.StringVar(&o.Password, "password", "", "chart repository password")
	fs.StringVar(&o.CertFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
	fs.StringVar(&o.KeyFile, "key-file", "", "identify HTTPS client using this SSL key file")
	fs.StringVar(&o.CaFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")
	fs.BoolVar(&o.InsecureSkipTLSverify, "insecure-skip-tls-verify", false, "skip tls certificate checks for the repository")
	fs.DurationVar(&o.Timeout, "timeout", 120*time.Second, "time to wait for the index file download to complete")
	fs.BoolVar(&o.PlainHttp, "plain-http",false, "use plain http when downloading from this repo")

}