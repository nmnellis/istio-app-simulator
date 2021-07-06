package generate

import "github.com/goombaio/namegenerator"

type AppGenerator struct {
	config        *Config
	nameGenerator namegenerator.Generator
}

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

type TemplateConfig struct {
	Microservices    map[string][]*Microservice
	ExternalServices []string
	Config           *Config
}

type Config struct {
	Seed                          int64
	NumberOfNamespaces            int
	NumberOfTiers                 int
	MaxAppsPerTier                int
	ChanceOfVersions              int
	ChanceOfCrossNamespaceChatter int
	ChanceOfErrors                int
	ErrorPercent                  float32
	ChanceToCallExternalService   int
	OutputDir                     string
	Hostname                      string
	MemoryLimit                   string
	CPULimit                      string
	MemoryRequest                 string
	CPURequest                    string
}
