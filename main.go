package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	configPath := flag.String(`config`, ``, `Path to ninurta.yml`)
	flag.Parse()

	configFile, err := os.Open(*configPath) // For read access.
	if err != nil {
		log.Fatal(err)
	}

	var configRaw []byte
	_, err = configFile.Read(configRaw)
	if err != io.EOF && err != nil {
		log.Fatal(err)
	}

	err=os.Chdir(filepath.Dir(*configPath))
	if err != nil {
		log.Fatal(err)
	}

	config := ninurtaConfig{}

	err = yaml.Unmarshal(configRaw, &config)
	if err != nil {
		log.Fatal(err)
	}

	// TODO if only http is set, serve http
	// TODO if only https is set, serve https
	// TODO if http and https are set, redirect to https
	// TODO look for certs and check to see which ones we have available
	// TODO generate any certs that are missing, see https://kalyanchakravarthy.net/blog/https-server-with-go-letsencrypt/
	// TODO load certs, see https://kalyanchakravarthy.net/blog/tls-server-in-go/

	httpMux := http.NewServeMux()

	for _, site := range config.Sites {
		info, err := os.Stat(site.Path)
		if os.IsNotExist(err) || !info.IsDir() {
			log.Fatalf(`%s is not a directory`, site.Path)
		}

		httpMux.Handle(fmt.Sprintf(`%s/`, site.Domain), http.FileServer(http.Dir(site.Path)))
		log.Printf(`Serving %s at %s`, site.Path, site.Domain)
		for _, redirect := range site.Redirects{
			httpMux.HandleFunc(fmt.Sprintf(`%s/`, redirect), makeRedirect(site.Domain))
			log.Printf(`Redirecting %s to %s`, redirect, site.Domain)
		}
	}
}

func makeRedirect(toDomain string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(`location`, fmt.Sprintf(`%s%s`, toDomain, r.URL))
		w.WriteHeader(http.StatusMovedPermanently)
	}
}

type ninurtaConfig struct {
	Ports	struct{
		Http	uint16	``
		Https	uint16
	}
	Sites	[]site
}

type site struct{
	Path	string
	Domain	string
	Redirects	[]string
}
