package main

import (
	"embed"
	"fmt"
	"github.com/goombaio/namegenerator"
	"log"
	"math/rand"
	"os"
	"strings"
	"text/template"
	"time"
)

const (
	NumberOfNamespaces  = 1
	NumberOfTiers       = 5
	MaxAppsPerTier      = 5
	ChanceOfVersions    = 5 // 0-100
	maxAmountOfVersions = 3
	// TODO not implemented yet
	ChanceOfCrossNamespaceChatter = 5   // 0-100
	ChanceOfErrors                = 5   // 0-100
	ErrorPercent                  = .05 // 0 - 1
	ChanceToCallExternalService   = 0   // 0-100
)

var (
	//go:embed assets/*
	assets embed.FS
)

type Microservice struct {
	Name      string
	Namespace string
	Versions  []string
	Backends  []*Backend
	// list of external urls it will call
	ExternalServices []string
	// percent of requests that will throw errors
	PercentErrors float32
	// denotes a top tier application in the namespace
	TopTier bool
	Tier    int
}

type Backend struct {
	Name      string
	Namespace string
}

var (
	nameGenerator = namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())
)

func main() {

	namespaces := map[string][]*Microservice{}
	// generate namespaces
	// namespace names will be very simple ns1, ns2 etc
	for i := 1; i <= NumberOfNamespaces; i++ {
		namespaceName := fmt.Sprintf("ns-%d", i)

		// TODO right now its just easier to have 1 microservice entry point per namespace

		// namespaces[namespaceName] =
		// generate tiers in reverse order so we can connect them to parents
		tiers := map[int][]*Microservice{}
		for tier := NumberOfTiers; tier > 0; tier-- {

			if tier == 1 {
				// top tier
				tiers[tier] = append([]*Microservice{}, genMicroService(namespaceName, true, tier))
			} else {
				tiers[tier] = genTier(namespaceName, tier)
			}

			if tier != NumberOfTiers {
				// there is a tier below current tier
				// all of the microservices below this tier need a parent or they will be orphans
				previousTier := tiers[tier+1]
				for _, ms := range previousTier {
					// grab a random app from this tier
					app := rand.Intn(len(tiers[tier]))
					tiers[tier][app].Backends = append(tiers[tier][app].Backends, &Backend{
						Name:      ms.Name,
						Namespace: ms.Namespace,
					})
				}
			}
		}
		namespaces[namespaceName] = flattenTiers(tiers)
	}
	// bytes, _ := json.MarshalIndent(namespaces, "", "  ")
	// fmt.Println(string(bytes))
	render(namespaces)
}

func render(microservices map[string][]*Microservice) {

	testTemplate, err := template.New("template").Funcs(template.FuncMap{
		"genUpstream": func(backends []*Backend, externalServices []string) string {
			hostnames := externalServices
			for _, backend := range backends {
				hostnames = append(hostnames, fmt.Sprintf("http://%s.%s:8080", backend.Name, backend.Namespace))
			}

			return strings.Join(hostnames, ",")
		},
	}).ParseGlob("assets/*")

	if _, err := os.Stat("out"); os.IsNotExist(err) {
		err := os.Mkdir("out", os.ModePerm)
		if err != nil {
			log.Println("create folder out/: ", err)
			return
		}
	}

	gatewayFile, err := os.Create("out/gateway.yaml")
	if err != nil {
		log.Println("create file: ", err)
		return
	}

	if err != nil {
		panic(err)
	}
	err = testTemplate.ExecuteTemplate(gatewayFile, "gateway.yaml.tmpl", microservices)
	if err != nil {
		panic(err)
	}

	virtualserviceFile, err := os.Create("out/virtualservice.yaml")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = testTemplate.ExecuteTemplate(virtualserviceFile, "virtualservice.yaml.tmpl", microservices)
	if err != nil {
		panic(err)
	}

	appFile, err := os.Create("out/app.yaml")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = testTemplate.ExecuteTemplate(appFile, "app.yaml.tmpl", microservices)
	if err != nil {
		panic(err)
	}

	serviceEntryFile, err := os.Create("out/serviceentry.yaml")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = testTemplate.ExecuteTemplate(serviceEntryFile, "serviceentry.yaml.tmpl", externalServices)
	if err != nil {
		panic(err)
	}
	// fmt.Println(string(bytes))
}

func flattenTiers(tiers map[int][]*Microservice) (microservices []*Microservice) {
	for _, ms := range tiers {
		microservices = append(microservices, ms...)
	}
	return microservices
}

func genTier(namespace string, tier int) []*Microservice {

	var microservices []*Microservice
	// the tier has to have 1 application because we cant have a nil tier
	a := rand.Intn(MaxAppsPerTier) + 1
	for i := 1; i <= a; i++ {
		ms := genMicroService(namespace, false, tier)
		microservices = append(microservices, ms)
	}
	return microservices
}

func genMicroService(namespace string, topTier bool, tier int) *Microservice {
	ms := &Microservice{
		Name:      nameGenerator.Generate(),
		Namespace: namespace,
		TopTier:   topTier,
		Tier:      tier,
	}

	// should it have errors
	i := rand.Intn(100) + 1
	if i <= ChanceOfErrors {
		ms.PercentErrors = ErrorPercent
	}
	// should we have versions
	i = rand.Intn(100) + 1
	if i <= ChanceOfVersions {
		i = rand.Intn(maxAmountOfVersions) + 1
		for v := 1; v <= i; v++ {
			ms.Versions = append(ms.Versions, fmt.Sprintf("v%d", v))
		}
	} else {
		// default to 1 version
		ms.Versions = append(ms.Versions, "v1")
	}
	// should we call external Services
	i = rand.Intn(100) + 1
	if i <= ChanceToCallExternalService {
		// add one external service
		e := rand.Intn(len(externalServices))
		ms.ExternalServices = append(ms.ExternalServices, fmt.Sprintf("https://%s:443", externalServices[e]))

	}
	return ms
}
