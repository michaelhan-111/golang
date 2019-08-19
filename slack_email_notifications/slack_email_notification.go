package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

/* Declare variables here */

type YamlConfig struct {
	Recipients map[string]map[string][]string
}

var configFile = "notification_config.yaml"

func main() {
	/* read from yaml  */
	yamlFile, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Fatal(err)
	}

	var config YamlConfig

	err = yaml.Unmarshal(yamlFile, &config.Recipients)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("Value: %#v\n", config.Description)
	//fmt.Println("Value:", config.Recipients["recipients"]["sre"])
	//fmt.Println("Length is:", len(config.Recipients["recipients"]["sre"]))
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("What message did you want to post? ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	groupPtr := flag.String("group", "", "Need to define a group; i.e. -group=sre. Read the README for options and rules around usage.")
	flag.Parse()
	for i := 0; i < len(config.Recipients["recipients"][*groupPtr]); i++ {
		//fmt.Println("This text:", text, "will be sent to the following recipients:", config.Recipients["recipients"][*groupPtr][i])
		// fmt.Println(*groupPtr, ":", config.Recipients["recipients"][*groupPtr][i])
		send(text, config.Recipients["recipients"][*groupPtr][i], config.Recipients["recipients"]["fromEmail"][0], config.Recipients["recipients"]["emailPW"][0])
	}
}

func send(body string, emailRecipient string, fromEmailAddress string, emailPW string) {
	from := fromEmailAddress
	pass := emailPW
	to := emailRecipient

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + "Testing Notification tool" + "\n" +
		body + "\n"

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Print(emailRecipient, " sent!")
}

/* func (yamlInput *YamlConf) readFromYaml() *YamlConf {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("Error reading yaml:", err)
	}
	err = yaml.Unmarshal(yamlFile, yamlInput)
	if err != nil {
		log.Fatal("Error unmarshalling:", err)
	}
	return yamlInput
} */

/*
func yamlReader() {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	var cfy ConfigYaml
	err = yaml.Unmarshal([]byte(yamlFile), &cfy.Services)
	if err != nil {
		log.Fatalf("Failed to map the YAML file: %v", err)
	}

	// fmt.Printf("\nYAML Mapping:\n%v\n\n", cfy.Services)

	var es = cfy.Services["services"]

	//fmt.Println("Services:", es[0])
	//	fmt.Printf("%# v", pretty.Formatter(es))

	for _, index := range es {
		//	fmt.Println("Services:", index)
		var ServiceName = index
		fmt.Println(ServiceName)
	}
}
*/
