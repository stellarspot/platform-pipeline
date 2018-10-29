package main

import (
	"fmt"

	"github.com/DATA-DOG/godog/gherkin"
)

const grpcBasicTemplateDir = "Services/gRPC/Basic_Template"

func dnnmodelservicesIsRunning() (err error) {

	dir := dnnModelServicesDir + "/" + grpcBasicTemplateDir

	command := ExecCommand{
		Command:   "chmod",
		Directory: dir,
		Args:      []string{"u+x", "buildproto.sh"},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	command = ExecCommand{
		Command:   dir + "/buildproto.sh",
		Directory: dir,
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	outputFile = logPath + "/dnn-model-services.log"
	outputContainsStrings = []string{"created", "description"}
	outputSkipErrors = []string{"No Daemon config file!"}

	command = ExecCommand{
		Command:    "python3",
		Directory:  dir,
		Args:       []string{"run_basic_service.py"},
		OutputFile: outputFile,
	}

	err = runCommandAsync(command)

	if err != nil {
		return err
	}

	_, err = checkWithTimeout(5000, 500, checkFileContainsStrings)

	return
}

func snetdaemonIsStartedWithDnnmodelservices(table *gherkin.DataTable) (err error) {

	dir := dnnModelServicesDir + "/" + grpcBasicTemplateDir

	daemonPort := getTableValue(table, "daemon port")
	ethereumEndpointPort := getTableValue(table, "ethereum endpoint port")
	passthroughEndpointPort := getTableValue(table, "passthrough endpoint port")

	snetdConfigTemplate := `
	{
		"AGENT_CONTRACT_ADDRESS": "%s",
		"MULTI_PARTY_ESCROW_CONTRACT_ADDRESS": "0x5c7a4290f6f8ff64c69eeffdfafc8644a4ec3a4e",
		"PRIVATE_KEY": "%s",
		"DAEMON_LISTENING_PORT": %s,
		"ETHEREUM_JSON_RPC_ENDPOINT": "http://localhost:%s",
		"PASSTHROUGH_ENABLED": true,
		"PASSTHROUGH_ENDPOINT": "http://localhost:%s",
		"price_per_call": 10,
		"log": {
			"level": "debug",
			"output": {
			"type": "stdout"
			}
		}
	}
	`

	snetdConfig := fmt.Sprintf(snetdConfigTemplate,
		agentAddress, accountPrivateKey, daemonPort, ethereumEndpointPort, passthroughEndpointPort)

	file := dir + "/snetd.config.json"
	err = writeToFile(file, snetdConfig)

	if err != nil {
		return
	}

	// TBD: check error
	linkFile(envSingnetRepos+"/snet-daemon/build/snetd-linux-amd64", dir+"/snetd-linux-amd64")

	outputFile = logPath + "/dnn-model-services-snetd.log"
	outputContainsStrings = []string{"multi_party_escrow_contract_address: 0x5c7a4290f6f8ff64c69eeffdfafc8644a4ec3a4e"}
	outputSkipErrors = []string{}

	command := ExecCommand{
		Command:    "./snetd-linux-amd64",
		Directory:  dir,
		OutputFile: outputFile,
	}

	err = runCommandAsync(command)

	if err != nil {
		return err
	}

	_, err = checkWithTimeout(5000, 500, checkFileContainsStrings)

	return
}
