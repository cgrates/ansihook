// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/google/go-github/v38/github"
)

var (
	secret            = flag.String("secret", "", "The secret for webhook")
	pattern           = flag.String("http_path", "/webhooks", "The webhook path")
	address           = flag.String("address", ":8080", "The addres the server is created")
	ansibleScriptPath = flag.String("path", "./main.yaml", "The path to the ansible script")
	ansibleInventory  = flag.String("inventory", "./hosts", "The path to the ansible inventory")

	ansiblePath string
)

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(*secret))
	if err != nil {
		log.Printf("error validating request body: err=%s\n", err)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	r.Body.Close()
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		return
	}

	switch event.(type) {
	case *github.PushEvent:
		log.Println("Received a push event")
		go executeAnsible(ansiblePath, *ansibleScriptPath)
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(r))
		return
	}
}

func main() {
	flag.Parse()
	var err error
	if ansiblePath, err = exec.LookPath("ansible-playbook"); err != nil {
		log.Fatalf("Unable to find ansible-playbook: %s", err)
	}
	log.Println("server started at: ", *address+*pattern)
	http.HandleFunc(*pattern, handleWebhook)
	if err = http.ListenAndServe(*address, nil); err != nil {
		log.Fatalf("Unable to start server: %s", err)
	}
}

func executeAnsible(ansiblePath, scriptPath string) (err error) {
	cmd := exec.Command(ansiblePath, scriptPath, "-i", *ansibleInventory)
	fmt.Println("Running:", cmd.String())
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err = cmd.Run(); err != nil {
		fmt.Println(ansiblePath, scriptPath)
		fmt.Print(stdout, stderr)
		log.Printf("Failed to run ansible script because: %s", err)
	}
	fmt.Println(ansiblePath, scriptPath)
	return
}
