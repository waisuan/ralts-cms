package provider

import (
	"fmt"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/utils"
	l "log"
	"net"
	"net/http"
	"os"
	"ralts-cms/internal/pact"
	"testing"
)

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
var port, _ = utils.GetFreePort()

// Provider States data sets
var sallyExists = &UserRepository{
	Users: map[string]*pact.User{
		"sally": {
			FirstName: "Jean-Marie",
			LastName:  "de La Beaujardière😀😍",
			Username:  "sally",
			Type:      "admin",
			ID:        10,
		},
	},
}

var sallyDoesNotExist = &UserRepository{}

var sallyUnauthorized = &UserRepository{
	Users: map[string]*pact.User{
		"sally": {
			FirstName: "Jean-Marie",
			LastName:  "de La Beaujardière😀😍",
			Username:  "sally",
			Type:      "blocked",
			ID:        10,
		},
	},
}

func TestPactProvider(t *testing.T) {
	log.SetLogLevel("INFO")

	go startInstrumentedProvider()

	verifier := provider.NewVerifier()

	// Verify the Provider - Branch-based Published Pacts for any known consumers
	err := verifier.VerifyProvider(t, provider.VerifyRequest{
		Provider:           "GoUserService",
		ProviderBaseURL:    fmt.Sprintf("http://127.0.0.1:%d", port),
		ProviderBranch:     os.Getenv("VERSION_BRANCH"),
		FailIfNoPactsFound: false,
		// Use this if you want to test without the Pact Broker
		// PactFiles:                   []string{filepath.FromSlash(fmt.Sprintf("%s/GoAdminService-GoUserService.json", os.Getenv("PACT_DIR")))},
		BrokerURL:                  fmt.Sprintf("%s://%s", os.Getenv("PACT_BROKER_PROTO"), os.Getenv("PACT_BROKER_URL")),
		BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
		PublishVerificationResults: true,
		ProviderVersion:            os.Getenv("VERSION_COMMIT"),
		StateHandlers:              stateHandlers,
		RequestFilter:              fixBearerToken,
		BeforeEach: func() error {
			userRepository = sallyExists
			return nil
		},
	})

	if err != nil {
		t.Log(err)
	}
}

// Simulates the need to set a time-bound authorization token,
// such as an OAuth bearer token
func fixBearerToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only set the correct bearer token, if one was provided in the first place
		if r.Header.Get("Authorization") != "" {
			r.Header.Set("Authorization", getAuthToken())
		}
		next.ServeHTTP(w, r)
	})
}

var stateHandlers = models.StateHandlers{
	"User sally exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
		userRepository = sallyExists
		return models.ProviderStateResponse{}, nil
	},
	"User sally does not exist": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
		userRepository = sallyDoesNotExist
		return models.ProviderStateResponse{}, nil
	},
}

// Starts the provider API with hooks for provider states.
// This essentially mirrors the main.go file, with extra routes added.
func startInstrumentedProvider() {
	mux := GetHTTPHandler()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		l.Fatal(err)
	}
	defer ln.Close()

	l.Printf("API starting: port %d (%s)", port, ln.Addr())
	l.Printf("API terminating: %v", http.Serve(ln, mux))

}
