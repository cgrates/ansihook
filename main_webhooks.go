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
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
	"github.com/go-playground/webhooks/v6/github"
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
	hook, _ := github.New(github.Options.Secret(*secret))

	http.HandleFunc(*pattern, func(rw http.ResponseWriter, r *http.Request) {
		event, err := hook.Parse(r, github.PushEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				fmt.Println(err)
			}
		}

		switch event.(type) {
		case github.PushPayload:
			log.Println("Received a push event")
			go executeAnsible(ansiblePath, *ansibleScriptPath)
		default:
			log.Printf("unknown event type %T\n", event)
			return
		}
	})
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
	ansiblePlaybookConnectionOptions := &options.AnsibleConnectionOptions{
		Connection: "local",
		User:       "nick",
	}

	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Inventory: *ansibleInventory,
	}

	playbook := &playbook.AnsiblePlaybookCmd{
		Playbooks:         []string{scriptPath},
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
	}
	buff := os.Stdout
	errorLogger := log.New(buff, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	err = playbook.Run(context.TODO())
	fmt.Println("[*] Running script")
	if err != nil {
		errorLogger.Println(err)
	}
	return
}
