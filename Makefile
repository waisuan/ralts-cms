include config.mk

generate:
	go generate ./...

fmt:
	go fmt ./...

pact-run-provider:
	go run internal/pact/provider/cmd/main.go

pact-consumer: export PACT_TEST := true
pact-consumer:
	@echo "--- 🔨Running Consumer Pact tests "
	go test -tags=integration -count=1 ralts-cms/internal/pact/consumer -run 'TestClientPact' -v

pact-provider: export PACT_TEST := true
pact-provider:
	@echo "--- 🔨Running Provider Pact tests "
	go test -count=1 -tags=integration ralts-cms/internal/pact/provider -run "TestPactProvider" -v

pact-install-cli:
	@if [ ! -d pact/bin ]; then\
		echo "--- Installing Pact CLI dependencies";\
		curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash;\
    fi

pact-publish:
	@echo "--- 📝 Publishing Pacts"
	pact/bin/pact-broker publish ${PWD}/pacts --consumer-app-version ${VERSION_COMMIT} --branch ${VERSION_BRANCH} \
		-b $(PACT_BROKER_PROTO)://$(PACT_BROKER_URL) -u ${PACT_BROKER_USERNAME} -p ${PACT_BROKER_PASSWORD}
	@echo
	@echo "Pact contract publishing complete!"
	@echo
	@echo "Head over to $(PACT_BROKER_PROTO)://$(PACT_BROKER_URL) and login with $(PACT_BROKER_USERNAME)/$(PACT_BROKER_PASSWORD)"
	@echo "to see your published contracts.	"

pact-deploy-consumer:
	@echo "--- ✅ Checking if we can deploy consumer"
	pact/bin/pact-broker can-i-deploy \
		--pacticipant $(CONSUMER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--to-environment production

pact-deploy-provider:
	@echo "--- ✅ Checking if we can deploy provider"
	pact/bin/pact-broker can-i-deploy \
		--pacticipant $(PROVIDER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--to-environment production

pact-record-deploy-consumer:
	@echo "--- ✅ Recording deployment of consumer"
	pact/bin/pact-broker record-deployment \
		--pacticipant $(CONSUMER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--environment production

pact-record-deploy-provider:
	@echo "--- ✅ Recording deployment of provider"
	pact/bin/pact-broker record-deployment \
		--pacticipant $(PROVIDER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--environment production