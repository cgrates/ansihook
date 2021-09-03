/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package main

import (
	"flag"
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
	// stdout := new(bytes.Buffer)
	// stderr := new(bytes.Buffer)
	// cmd.Stdout = stdout
	// cmd.Stderr = stderr
	if err = cmd.Run(); err != nil {
		// fmt.Println(ansiblePath, scriptPath)
		// fmt.Print(stdout, stderr)
		log.Printf("Failed to run ansible script because: %s", err)
	}
	return
}
