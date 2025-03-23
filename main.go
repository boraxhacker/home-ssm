package main

import (
	"flag"
	"home-ssm/awslib"
	"home-ssm/ssm"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dgraph-io/badger/v4"
	"gopkg.in/yaml.v3"
)

type SsmCredentials struct {
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
	Username  string `yaml:"username"`
}

type HomeSsmConfig struct {
	Region      string           `yaml:"region"`
	Credentials []SsmCredentials `yaml:"credentials"`
}

const (
	ZeroAccountId string = "000000000000"
)

func main() {

	configFilePtr := flag.String("config", ".home-ssm-config.yaml", "Path to the home-ssm config file.")
	dbPathPtr := flag.String("db-path", ".home-ssm-db", "Path to badger database folder.")
	flag.Parse()

	ssmConfig := readAuthCredsOrDie(*configFilePtr)

	credentialsProvider := awslib.CredentialsProvider{
		Service:     awslib.ServiceSsm,
		Region:      ssmConfig.Region,
		Credentials: []aws.Credentials{},
	}

	credentialsProvider.Region = ssmConfig.Region
	log.Println("Region:", ssmConfig.Region)

	log.Println("Credentials:")
	for i, cred := range ssmConfig.Credentials {
		credentialsProvider.Credentials = append(credentialsProvider.Credentials, aws.Credentials{
			AccessKeyID:     cred.AccessKey,
			SecretAccessKey: cred.SecretKey,
			Source:          cred.Username,
			AccountID:       ZeroAccountId,
		})
		log.Printf("\tAccessKey %02d: %s\n", i+1, cred.AccessKey)
	}

	service := initialServiceOrDie(ssmConfig.Region, ZeroAccountId, *dbPathPtr)
	defer service.Close()

	api := ssm.NewParameterApi(service, &credentialsProvider)

	http.HandleFunc("/ssm", credentialsProvider.WithSigV4(api.Handle))

	addr := ":9080"
	log.Printf("Listening on %s", addr)
	http.ListenAndServe(addr, nil)
}

func readAuthCredsOrDie(configFileName string) *HomeSsmConfig {

	configFile, err := os.ReadFile(configFileName) // Replace with your yaml file name/path
	if err != nil {
		log.Panicln("Error reading config file:", err)
	}

	var config HomeSsmConfig
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Panicln("Error unmarshalling config:", err)
	}

	return &config
}

func initialServiceOrDie(region string, accountId string, databasePath string) *ssm.ParameterService {

	opts := badger.DefaultOptions(databasePath).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		log.Panicln("Error opening badger db:", err)
	}

	return ssm.NewParameterService(region, accountId, db)
}
