package main

import (
	"encoding/json"
	"fmt"
	"github.com/goombaio/namegenerator"
	"math/rand"
	"time"
)

const (
	NumberOfNamespaces            = 1
	NumberOfTiers                 = 5
	MaxAppsPerTier                = 5
	ChanceOfVersions              = 5 // 0-100
	maxAmountOfVersions           = 3
	ChanceOfCrossNamespaceChatter = 5   // 0-100
	ChanceOfErrors                = 5   // 0-100
	ErrorPercent                  = .05 // 0 - 1
	ChanceToCallExternalService   = 10   // 0-100
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
				tiers[tier] = append([]*Microservice{},genMicroService(namespaceName,true))
			}else {
				tiers[tier] = genTier(namespaceName)
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
	bytes, _ := json.MarshalIndent(namespaces, "", "  ")
	fmt.Println(string(bytes))

}

func flattenTiers(tiers map[int][]*Microservice) (microservices []*Microservice) {
	for _, ms := range tiers {
		microservices = append(microservices, ms...)
	}
	return microservices
}

func genTier(namespace string) []*Microservice {

	var microservices []*Microservice
	// the tier has to have 1 application because we cant have a nil tier
	a := rand.Intn(MaxAppsPerTier) + 1
	for i := 1; i <= a; i++ {
		ms := genMicroService(namespace,false)
		microservices = append(microservices, ms)
	}
	return microservices
}

func genMicroService(namespace string, topTier bool) *Microservice {
	ms := &Microservice{
		Name:      nameGenerator.Generate(),
		Namespace: namespace,
		TopTier: topTier,
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
	}
	// should we call external Services
	i = rand.Intn(100) + 1
	if i <= ChanceToCallExternalService {
		// add one external service
		e := rand.Intn(len(externalServices))
		ms.ExternalServices = append(ms.ExternalServices, externalServices[e])
	}
	return ms
}
