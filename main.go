package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	dockerCfg = `{"%s":{"username":"%s","password":"%s","email":"none"}}`
)

var (
	interval int
)

type Registry struct {
	accountID string
	region    string
}

type ECRcredentials struct {
	dockerServer   string
	dockerUsername string
	dockerPassword string
	dockerEmail    string
}

func refreshCredential(registry Registry, sess *session.Session) {
	instance := ecr.New(sess, &aws.Config{
		Region: aws.String(registry.region),
	})

	input := &ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{aws.String(registry.accountID)},
	}

	authToken, err := instance.GetAuthorizationToken(input)
	if err != nil {
		panic(err)
	}

	for _, data := range authToken.AuthorizationData {
		output, err := base64.StdEncoding.DecodeString(*data.AuthorizationToken)

		if err != nil {
			panic(err)
		}

		auth := strings.Split(string(output), ":")
		user, token := auth[0], auth[1]

		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		/*
			var kubeconfig *string
			home := homedir.HomeDir()
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
			config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
			if err != nil {
				panic(err)
			}
		*/

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		endpoint := "https://" + registry.accountID + ".dkr.ecr." + registry.region + ".amazonaws.com"
		ecrConfig := []byte(fmt.Sprintf(dockerCfg, endpoint, user, token))

		c := clientset.CoreV1().Secrets(apiv1.NamespaceDefault)
		_, err = c.Get("ecr-auth-refresh", metav1.GetOptions{})

		if err != nil {
			log.Println("Secret not found creating")

			_, err := c.Create(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ecr-auth-refresh",
				},
				Data: map[string][]byte{
					".dockerconfigjson": ecrConfig,
				},
				Type: "kubernetes.io/dockerconfigjson",
			})
			if err != nil {
				panic(err.Error())
			}
		} else {
			log.Println("Updating secret")

			_, err := c.Update(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ecr-auth-refresh",
				},
				Data: map[string][]byte{
					".dockerconfigjson": ecrConfig,
				},
				Type: "kubernetes.io/dockerconfigjson",
			})
			if err != nil {
				panic(err.Error())
			}
		}

	}
}

func main() {

	_, ok := os.LookupEnv("AWS_ACCESS_KEY_ID")
	if !ok {
		log.Fatal("Could not find AWS_ACCESS_KEY_ID environment vairable, exiting.")
	}

	_, ok = os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !ok {
		log.Fatal("Could not find AWS_SECRET_ACCESS_KEY environment vairable, exiting.")
	}

	accountID, ok := os.LookupEnv("ACCOUNT_ID")
	if !ok {
		log.Fatal("Could not find ACCOUNT_ID environment vairable, exiting.")
	}

	region, ok := os.LookupEnv("AWS_DEFAULT_REGION")
	if !ok {
		log.Fatal("Could not find AWS_DEFAULT_REGION environment vairable, exiting.")
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	registry := Registry{
		accountID: accountID,
		region:    region,
	}

	for {
		log.Printf("Getting ECR credentials for account %s in region %s", accountID, region)
		refreshCredential(registry, sess)
		time.Sleep(3 * time.Hour)
	}
}
