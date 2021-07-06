package generate

import (
	"embed"
	"fmt"
	"github.com/goombaio/namegenerator"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"
)

var (
	//go:embed assets/*
	assets embed.FS
)

func NewAppGenerator(config *Config) *AppGenerator {
	return &AppGenerator{
		config:        config,
		nameGenerator: namegenerator.NewNameGenerator(config.Seed),
	}
}

func (a *AppGenerator) Generate() error {

	// if unset, make random
	if a.config.Seed == 0 {
		a.config.Seed = time.Now().UnixNano()
	}

	rand.Seed(a.config.Seed)
	// generate namespaces
	// namespace names will be very simple ns1, ns2 etc
	var microservices []*Microservice
	for i := 1; i <= a.config.NumberOfNamespaces; i++ {
		namespaceName := fmt.Sprintf("ns-%d", i)

		// TODO right now its just easier to have 1 microservice entry point per namespace

		// namespaces[namespaceName] =
		// generate tiers in reverse order so we can connect them to parents
		tiers := map[int][]*Microservice{}
		for tier := a.config.NumberOfTiers; tier > 0; tier-- {

			if tier == 1 {
				// top tier
				tiers[tier] = append([]*Microservice{}, a.genMicroService(namespaceName, true, tier))
			} else {
				tiers[tier] = a.genTier(namespaceName, tier)
			}

			if tier != a.config.NumberOfTiers {
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
		microservices = append(microservices, flattenTiers(tiers)...)
	}

	// shuffle microservices to setup cross namespace calls
	rand.Shuffle(len(microservices), func(i, j int) { microservices[i], microservices[j] = microservices[j], microservices[i] })

	for _, ms := range microservices {
		i := rand.Intn(100) + 1
		if i <= a.config.ChanceOfCrossNamespaceChatter {
			// microservice should call another one from a different namespace
			giveUpAttempts := len(microservices)
			foundBackend := false
			for !foundBackend {
				// randomly grab a microservice and see if its compatible. we will give up after so many tries
				msIndex := rand.Intn(len(microservices))
				foundMs := microservices[msIndex]
				if foundMs.Namespace != ms.Namespace && foundMs.Tier > ms.Tier {
					foundBackend = true
					ms.Backends = append(ms.Backends, &Backend{
						Name:      foundMs.Name,
						Namespace: foundMs.Namespace,
					})
				}
				giveUpAttempts--
				if giveUpAttempts < 1 {
					// there are lots of reasons why a backend might not be found, the ms is in the bottom tier
					// there are no apps in other namespaces etc
					// fmt.Println("giving up finding backend ")
					foundBackend = true
				}
			}
		}
	}
	// sort the microservices so the yaml is generated the same if non random seed used
	sort.SliceStable(microservices, func(i, j int) bool {
		return microservices[i].Name < microservices[j].Name
	})

	return a.render(groupMicroservicesByNamespace(microservices))
}

func groupMicroservicesByNamespace(microservices []*Microservice) map[string][]*Microservice {
	namespaces := map[string][]*Microservice{}
	for _, ms := range microservices {
		namespaces[ms.Namespace] = append(namespaces[ms.Namespace], ms)
	}
	return namespaces
}

func (a *AppGenerator) render(microservices map[string][]*Microservice) error {

	testTemplate, err := template.New("template").Funcs(template.FuncMap{
		"genUpstream": func(backends []*Backend) string {
			var hostnames []string
			for _, backend := range backends {
				hostnames = append(hostnames, fmt.Sprintf("http://%s.%s:8080", backend.Name, backend.Namespace))
			}

			return strings.Join(hostnames, ",")
		},
		"genExternalServices": func(es []string) string {
			return strings.Join(es, ",")
		},
	}).ParseGlob("pkg/generate/assets/*")
	if err != nil {
		log.Println(" template: ", err)
		return err
	}
	if _, err := os.Stat(a.config.OutputDir); os.IsNotExist(err) {
		err := os.Mkdir(a.config.OutputDir, os.ModePerm)
		if err != nil {
			log.Printf("create folder %s/: %v\n", a.config.OutputDir, err)
			return err
		}
	}

	templateConfig := &TemplateConfig{
		Microservices:    microservices,
		ExternalServices: externalServices,
		Host:             a.config.Hostname,
		Seed:             a.config.Seed,
	}

	gatewayFile, err := os.Create(a.config.OutputDir + "/gateway.yaml")
	if err != nil {
		log.Println("create file: ", err)
		return err
	}

	if err := testTemplate.ExecuteTemplate(gatewayFile, "gateway.yaml.tmpl", templateConfig); err != nil {
		return err
	}
	log.Println("Generated Gateway config at " + gatewayFile.Name())

	virtualserviceFile, err := os.Create(a.config.OutputDir + "/virtualservice.yaml")
	if err != nil {
		log.Println("create file: ", err)
		return err
	}
	if err := testTemplate.ExecuteTemplate(virtualserviceFile, "virtualservice.yaml.tmpl", templateConfig); err != nil {
		return err
	}
	log.Println("Generated VirtualService config at " + virtualserviceFile.Name())

	appFile, err := os.Create(a.config.OutputDir + "/app.yaml")
	if err != nil {
		log.Println("create file: ", err)
		return err
	}
	if err := testTemplate.ExecuteTemplate(appFile, "app.yaml.tmpl", templateConfig); err != nil {
		return err
	}
	log.Println("Generated App config at " + appFile.Name())

	serviceEntryFile, err := os.Create(a.config.OutputDir + "/serviceentry.yaml")
	if err != nil {
		log.Println("create file: ", err)
		return err
	}

	if err := testTemplate.ExecuteTemplate(serviceEntryFile, "serviceentry.yaml.tmpl", templateConfig); err != nil {
		return err
	}
	log.Println("Generated ServiceEntry config at " + serviceEntryFile.Name())

	return nil
}

func flattenTiers(tiers map[int][]*Microservice) (microservices []*Microservice) {
	for _, ms := range tiers {
		microservices = append(microservices, ms...)
	}
	return microservices
}

func (a *AppGenerator) genTier(namespace string, tier int) []*Microservice {

	var microservices []*Microservice
	// the tier has to have 1 application because we cant have a nil tier
	appdx := rand.Intn(a.config.MaxAppsPerTier) + 1
	for i := 1; i <= appdx; i++ {
		ms := a.genMicroService(namespace, false, tier)
		microservices = append(microservices, ms)
	}
	return microservices
}

func (a *AppGenerator) genMicroService(namespace string, topTier bool, tier int) *Microservice {

	ms := &Microservice{
		Name:      a.nameGenerator.Generate(),
		Namespace: namespace,
		TopTier:   topTier,
		Tier:      tier,
	}

	// should it have errors
	i := rand.Intn(100) + 1
	if i <= a.config.ChanceOfErrors {
		ms.PercentErrors = a.config.ErrorPercent
	}
	// should we have versions
	i = rand.Intn(100) + 1
	if i <= a.config.ChanceOfVersions {
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
	if i <= a.config.ChanceToCallExternalService {
		// add one external service
		e := rand.Intn(len(externalServices))
		ms.ExternalServices = append(ms.ExternalServices, fmt.Sprintf("https://%s:443", externalServices[e]))

	}
	return ms
}
