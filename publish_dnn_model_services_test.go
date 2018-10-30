package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DATA-DOG/godog/gherkin"
)

const (
	grpcBasicTemplateDir = "Services/gRPC/Basic_Template"
	serviceName          = "basic_service_one"
)

func snetdaemonConfigFileIsCreated(table *gherkin.DataTable) (err error) {

	dir := dnnModelServicesDir + "/" + grpcBasicTemplateDir

	agentAddress = "0x3b07411493C72c5aEC01b6Cf3cd0981cF0586fA7"
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

	file := fmt.Sprintf("%s/snetd_%s_config.json", dir, serviceName)
	log.Printf("create snetd config: %s\n---\n:%s\n---\n", file, snetdConfig)

	err = writeToFile(file, snetdConfig)

	return
}

func dnnmodelservicesIsRunning() (err error) {

	dir := dnnModelServicesDir + "/" + grpcBasicTemplateDir

	err = os.Chmod(dir+"/buildproto.sh", 0544)

	if err != nil {
		return
	}

	command := ExecCommand{
		Command:   dir + "/buildproto.sh",
		Directory: dir,
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	outputFile = logPath + "/dnn-model-services-" + serviceName + ".log"
	outputContainsStrings = []string{"created", "description"}
	outputSkipErrors = []string{"No Daemon config file!"}

	command = ExecCommand{
		Command:    "python3",
		Directory:  dir,
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
