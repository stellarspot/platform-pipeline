package main

const grpcBasicTemplateDir = "dnn-model-services/Services/gRPC/Basic_Template"

func dnnmodelservicesIsRunning() (err error) {

	dir := dnnModelServicesDir + "/" + grpcBasicTemplateDir

	command := ExecCommand{
		Command:   ".",
		Directory: dir,
		Args:      []string{"buildproto.sh"},
	}

	err = runCommand(command)

	if err != nil {
		return
	}

	// python3 run_basic_service.py >$GOPATH/log/dnn-model-services.log 2>&1 &

	outputFile = logPath + "/dnn-model-services.log"
	outputContainsStrings = []string{"created", "description"}

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

	if err != nil {
		return err
	}

	return
}
