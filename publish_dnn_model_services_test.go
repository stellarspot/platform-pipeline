package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DATA-DOG/godog/gherkin"
)

const (
	serviceName = "basic_service_one"
)

func dnnmodelServiceIsRegistered(table *gherkin.DataTable) (err error) {
	err = serviceIsRegistered(table)
	return
}

func dnnmodelServiceIsPublishedToNetwork() (err error) {
	err = serviceIsPublishedToNetwork("./service.json")
	return
}

func dnnmodelMpeServiceIsRegistered(table *gherkin.DataTable) (err error) {

	name := getTableValue(table, "name")
	group := getTableValue(table, "group")
	endpoint := getTableValue(table, "endpoint")

	log.Println("dnnModelServicesDir: ", dnnModelServicesDir)

	output := "output.txt"

	// snet mpe-service publish_proto
	command := ExecCommand{
		Command:    "snet",
		Directory:  dnnModelServicesDir,
		Args:       []string{"mpe-service", "publish_proto", "service/service_spec/"},
		OutputFile: output,
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	modelIpfsHash, err := readFile(output)
	log.Printf("modelIpfsHash: '%s'\n", modelIpfsHash)
	log.Printf("registryAddress: '%s'\n", registryAddress)
	log.Printf("multiPartyEscrow: '%s'\n", multiPartyEscrow)
	log.Printf("organizationAddress: '%s'\n", organizationAddress)
	log.Printf("agentAddress: '%s'\n", agentAddress)

	if err != nil {
		return
	}

	//snet mpe-service  metadata_init
	command = ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args:      []string{"mpe-service", "metadata_init", modelIpfsHash, multiPartyEscrow},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	// snet mpe-service metadata_add_group
	command = ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args:      []string{"mpe-service", "metadata_add_group", group, organizationAddress},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	// snet mpe-service metadata_add_endpoints
	command = ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args:      []string{"mpe-service", "metadata_add_endpoints", group, endpoint},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	// snet mpe-service  publish_service
	command = ExecCommand{
		Command:   "snet",
		Directory: dnnModelServicesDir,
		Args:      []string{"mpe-service", "publish_service", registryAddress, name, "Basic_Template", "-y"},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	return
}

func dnnmodelServiceSnetdaemonConfigFileIsCreated(table *gherkin.DataTable) (err error) {

	daemonPort := getTableValue(table, "daemon port")
	price := getTableValue(table, "price")
	ethereumEndpointPort := getTableValue(table, "ethereum endpoint port")
	passthroughEndpointPort := getTableValue(table, "passthrough endpoint port")

	snetdConfigTemplate := `
	{
		"AGENT_CONTRACT_ADDRESS": "%s",
		"MULTI_PARTY_ESCROW_CONTRACT_ADDRESS": "%s",
		"PRIVATE_KEY": "%s",
		"DAEMON_LISTENING_PORT": %s,
		"ETHEREUM_JSON_RPC_ENDPOINT": "http://localhost:%s",
		"PASSTHROUGH_ENABLED": true,
		"PASSTHROUGH_ENDPOINT": "http://localhost:%s",
		"price_per_call": %s,
		"log": {
			"level": "debug",
			"output": {
			"type": "stdout"
			}
		}
	}
	`

	snetdConfig := fmt.Sprintf(
		snetdConfigTemplate,
		agentAddress,
		multiPartyEscrow,
		accountPrivateKey,
		daemonPort,
		ethereumEndpointPort,
		passthroughEndpointPort,
		price,
	)

	file := fmt.Sprintf("%s/snetd_%s_config.json", dnnModelServicesDir, serviceName)
	log.Printf("create snetd config: %s\n---\n:%s\n---\n", file, snetdConfig)

	err = writeToFile(file, snetdConfig)

	return
}

func dnnmodelServiceIsRunning() (err error) {

	err = os.Chmod(dnnModelServicesDir+"/buildproto.sh", 0544)

	if err != nil {
		return
	}

	command := ExecCommand{
		Command:   dnnModelServicesDir + "/buildproto.sh",
		Directory: dnnModelServicesDir,
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	outputFile = logPath + "/dnn-model-services-" + serviceName + ".log"
	outputContainsStrings = []string{"multi_party_escrow_contract_address"}
	outputSkipErrors = []string{}

	command = ExecCommand{
		Command:    "python3",
		Directory:  dnnModelServicesDir,
		Args:       []string{"run_basic_service.py", "--daemon-config-path", "."},
		OutputFile: outputFile,
	}

	err = runCommandAsync(command)

	if err != nil {
		return err
	}

	_, err = checkWithTimeout(5000, 500, checkFileContainsStrings)

	return
}
