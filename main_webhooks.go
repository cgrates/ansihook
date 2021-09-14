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
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
	"github.com/go-playground/webhooks/v6/bitbucket"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/go-playground/webhooks/v6/gitlab"
)

const (
	GITHUB    = "*github"
	GITLAB    = "*gitlab"
	BITBUCKET = "*bitbucket"
)

func NewWebhookParser(wType, secret string) func(*http.Request) (interface{}, error) {
	switch wType {
	case GITHUB:
		hook, _ := github.New(github.Options.Secret(secret))
		return func(r *http.Request) (interface{}, error) {
			return hook.Parse(r, github.PushEvent)
		}
	case GITLAB:
		hook, _ := gitlab.New(gitlab.Options.Secret(secret))
		return func(r *http.Request) (interface{}, error) {
			return hook.Parse(r, gitlab.PushEvents)
		}
	case BITBUCKET:
		hook, _ := bitbucket.New(bitbucket.Options.UUID(secret))
		return func(r *http.Request) (interface{}, error) {
			return hook.Parse(r, bitbucket.RepoPushEvent)
		}
	default:
		return func(*http.Request) (interface{}, error) {
			return nil, errors.New("Webhook type not found")
		}
	}
}

var (
	secret            = flag.String("secret", "", "The secret for webhook")
	pattern           = flag.String("http_path", "/webhooks", "The webhook path")
	address           = flag.String("address", ":8080", "The addres the server is created")
	ansibleScriptPath = flag.String("path", "./main.yaml", "The path to the ansible script")
	ansibleInventory  = flag.String("inventory", "./hosts", "The path to the ansible inventory")
	service           = flag.String("service", "github", "The service ansihook will use i.e: Github, Gitlab, Bitbucket")
)

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	hook, _ := github.New(github.Options.Secret(*secret))

	http.HandleFunc(*pattern, func(rw http.ResponseWriter, r *http.Request) {
		event, err := hook.Parse(r, github.PushEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				log.Println(err)
			}
		}

		switch event.(type) {
		case github.PushPayload:
			log.Println("Received a push event")
			go executeAnsible(*ansibleScriptPath)
		default:
			log.Printf("unknown event type %T\n", event)
			return
		}
	})
}

func main() {
	flag.Parse()
	var err error
	log.Println("server started at: ", *address+*pattern)
	http.HandleFunc(*pattern, handleWebhook)
	if err = http.ListenAndServe(*address, nil); err != nil {
		log.Fatalf("Unable to start server: %s", err)
	}
}

func executeAnsible(scriptPath string) (err error) {
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
