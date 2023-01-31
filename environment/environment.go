package environment

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

/*
 * loadEnv loads variables from .env files into ENVIRONMENT if it exists.
 * Then the variables are loaded from ENVIRONMENT
 * Returning the url, user, password needed for saltstack API
 */
func LoadEnv() (string, string, string) {
	err := godotenv.Load()
	if err != nil {
		log.WithFields(log.Fields{
			"environment": ".env",
		}).Println("Error loading environment file, assume env variables are set globaly.")
	}

	// saltUrl is the url for the saltstack API
	saltUrl := os.Getenv("SALTSTACK_API_URL")
	// saltUser is the user to login with
	saltUser := os.Getenv("SALTSTACK_API_USER")
	// saltPassword is the password associated to the saltUser
	saltPassword := os.Getenv("SALTSTACK_API_PASSWORD")

	return saltUrl, saltUser, saltPassword
}
