package main

import (
	"errors"
	"log"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
)

var outputFile string
var outputContainsStrings []string
var outputSkipErrors []string

var platformContractsDir string
var exampleServiceDir string
var dnnModelServicesDir string
var snetConfigFile string

var accountPrivateKey string
var identiyPrivateKey string
var agentFactoryAddress string
var registryAddress string
var organizationAddress string
var agentAddress string

func init() {
	platformContractsDir = envSingnetRepos + "/platform-contracts"
	exampleServiceDir = envSingnetRepos + "/example-service"
	dnnModelServicesDir = envSingnetRepos + "/dnn-model-services"
	snetConfigFile = envHome + "/.snet/config"
}

func ethereumNetworkIsRunningOnPort(port int) error {

	outputFile = logPath + "/ganache.log"
	outputContainsStrings = []string{"Listening on 127.0.0.1:" + toString(port)}

	args := []string{"--mnemonic", "gauge enact biology destroy normal tunnel slight slide wide sauce ladder produce"}
	command := ExecCommand{
		Command:    "./node_modules/.bin/ganache-cli",
		Directory:  platformContractsDir,
		OutputFile: outputFile,
		Args:       args,
	}

	err := runCommandAsync(command)

	if err != nil {
		return err
	}

	exists, err := checkWithTimeout(5000, 500, checkFileContainsStrings)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("Etherium networks is not started")
	}

	organizationAddress, err = getPropertyFromFile(outputFile, "(1)")
	if err != nil {
		return err
	}

	accountPrivateKey, err = getPropertyWithIndexFromFile(outputFile, "(2)", 1)
	if err != nil {
		return err
	}

	if len(accountPrivateKey) < 3 {
		return errors.New("Len of account privite key is to small: " + accountPrivateKey)
	}

	accountPrivateKey = accountPrivateKey[2:len(accountPrivateKey)]

	identiyPrivateKey, err = getPropertyWithIndexFromFile(outputFile, "(0)", 1)
	if err != nil {
		return err
	}

	return nil
}

func contractsAreDeployedUsingTruffle() error {

	command := ExecCommand{
		Command:   "./node_modules/.bin/truffle",
		Directory: platformContractsDir,
		Args:      []string{"compile"},
	}

	err := runCommand(command)

	if err != nil {
		return err
	}

	output := "migrate.out"
	command.Args = []string{"migrate", "--network", "local"}
	command.OutputFile = output
	err = runCommand(command)

	registryAddress, err = getPropertyFromFile(output, "Registry:")
	if err != nil {
		return err
	}

	agentFactoryAddress, err = getPropertyFromFile(output, "AgentFactory:")
	if err != nil {
		return err
	}

	return err
}

func ipfsIsRunning(portAPI int, portGateway int) error {

	env := []string{"IPFS_PATH=" + envGoPath + "/ipfs"}

	command := ExecCommand{
		Command: "ipfs",
		Env:     env,
		Args:    []string{"init"},
	}

	err := runCommand(command)

	if err != nil {
		return err
	}

	command.Args = []string{"bootstrap", "rm", "--all"}
	err = runCommand(command)

	if err != nil {
		return err
	}

	addressAPI := "/ip4/127.0.0.1/tcp/" + toString(portAPI)
	command.Args = []string{"config", "Addresses.API", addressAPI}
	err = runCommand(command)

	if err != nil {
		return err
	}

	addressGateway := "/ip4/0.0.0.0/tcp/" + toString(portGateway)
	command.Args = []string{"config", "Addresses.Gateway", addressGateway}
	err = runCommand(command)

	if err != nil {
		return err
	}

	outputFile = logPath + "/ipfs.log"
	command.OutputFile = outputFile
	command.Args = []string{"daemon"}
	err = runCommandAsync(command)

	if err != nil {
		return err
	}

	outputContainsStrings = []string{
		"Daemon is ready",
		"server listening on " + addressAPI,
		"server listening on " + addressGateway,
	}
	exists, err := checkWithTimeout(5000, 500, checkFileContainsStrings)

	if err != nil {
		return err
	}

	if !exists {
		return errors.New("Etherium networks is not started")
	}

	return nil
}

func identityIsCreatedWithUserAndPrivateKey(user string, privateKey string) error {

	command := ExecCommand{
		Command: "snet",
		Args:    []string{"identity", "create", user, "key", "--private-key", identiyPrivateKey},
	}
	err := runCommand(command)

	if err != nil {
		return err
	}

	command.Args = []string{"identity", "snet-user"}
	return runCommand(command)
}

func snetIsConfiguredWithEthereumRPCEndpoint(endpointEthereumRPC int) error {

	config := `
[network.local]
default_eth_rpc_endpoint = http://localhost:` + toString(endpointEthereumRPC)

	err := appendToFile(snetConfigFile, config)

	if err != nil {
		return err
	}

	command := ExecCommand{
		Command: "snet",
		Args:    []string{"network", "local"},
	}
	err = runCommand(command)

	if err != nil {
		return err
	}

	outputFile = snetConfigFile
	outputContainsStrings = []string{"session"}
	exists, e := checkWithTimeout(5000, 500, checkFileContainsStrings)

	if !exists {
		return errors.New("snet config file is not created: " + snetConfigFile)
	}

	return e
}

func snetIsConfiguredWithIPFSEndpoint(endpointIPFS int) error {

	command := ExecCommand{
		Command: "sed",
		Args:    []string{"-ie", "/ipfs/,+2d", snetConfigFile},
	}

	err := runCommand(command)

	if err != nil {
		return err
	}

	config := `
[ipfs]
default_ipfs_endpoint = http://localhost:` + toString(endpointIPFS)

	return appendToFile(snetConfigFile, config)
}

func organizationIsAdded(table *gherkin.DataTable) error {

	organization := getTableValue(table, "organization")

	args := []string{
		"contract", "Registry",
		"--at", registryAddress,
		"createOrganization", organization,
		"[\"" + organizationAddress + "\"]",
		"--transact",
		"--yes",
	}

	command := ExecCommand{
		Command: "snet",
		Args:    args,
	}

	return runCommand(command)
}

func FeatureContext(s *godog.Suite) {

	// background
	s.Step(`^Ethereum network is running on port (\d+)$`, ethereumNetworkIsRunningOnPort)
	s.Step(`^Contracts are deployed using Truffle$`, contractsAreDeployedUsingTruffle)
	s.Step(`^IPFS is running with API port (\d+) and Gateway port (\d+)$`, ipfsIsRunning)
	s.Step(`^Identity is created with user "([^"]*)" and private key "([^"]*)"$`,
		identityIsCreatedWithUserAndPrivateKey)
	s.Step(`^snet is configured with Ethereum RPC endpoint (\d+)$`, snetIsConfiguredWithEthereumRPCEndpoint)
	s.Step(`^snet is configured with IPFS endpoint (\d+)$`, snetIsConfiguredWithIPFSEndpoint)
	s.Step(`^Organization is added:$`, organizationIsAdded)

	// example-service sample
	s.Step(`^example-service is registered$`, exampleserviceIsRegistered)
	s.Step(`^example-service is published to network$`, exampleserviceIsPublishedToNetwork)
	s.Step(`^example-service is run with snet-daemon$`, exampleserviceIsRunWithSnetdaemon)
	s.Step(`^SingularityNET job is created$`, singularityNETJobIsCreated)

	// dnn-model-services sample
	s.Step(`^dnn-model-services is running$`, dnnmodelservicesIsRunning)
	s.Step(`^snet-daemon is started with dnn-model-services$`, snetdaemonIsStartedWithDnnmodelservices)
}

func checkFileContainsStrings() (bool, error) {

	log.Printf("check output file: '%s'\n", outputFile)
	log.Printf("check output file contains string: '%s'\n", strings.Join(outputContainsStrings, ","))
	if len(outputSkipErrors) > 0 {
		log.Printf("check output with skipped errors: '%s'\n", strings.Join(outputSkipErrors, ","))
	}

	out, err := readFile(outputFile)
	if err != nil {
		return false, err
	}

	if out != "" {
		log.Printf("Output: %s\n", out)
	}

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(strings.ToLower(line), "error") {
			skip := false
			for _, skipErr := range outputSkipErrors {
				if strings.Contains(out, skipErr) {
					log.Printf("skipp error: '%s'\n", skipErr)
					skip = true
					break
				}
			}
			if !skip {
				return false, errors.New("Output contains error: '" + line + "'")
			}
		}
	}

	for _, str := range outputContainsStrings {
		if !strings.Contains(out, str) {
			return false, nil
		}
	}

	return true, nil
}

func getTableValue(table *gherkin.DataTable, column string) string {

	names := table.Rows[0].Cells
	for i, cell := range names {
		if cell.Value == column {
			return table.Rows[1].Cells[i].Value
		}
	}

	log.Printf("column: %s has not been found in table", column)
	return ""
}
