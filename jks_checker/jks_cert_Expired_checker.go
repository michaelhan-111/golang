package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

type ConfigYaml struct {
	Jks map[string]map[string][]string
}

var ServiceName = ""
var NodeFeatureName = ""
var jksYamlPath = "/usr/bin/jks_checker/jks_scanpath.yaml"
var findFileExt = ".jks"
var Ihp_Hostname_Check = exec.Command("hostname")
var IhpFound = false
var AwsFound = false
var MatchServiceNameIhp = ""
var MatchServiceNameAws = ""

func main() {
	serviceName := serviceCheck()

	jksFileDir := mappingJKSFolderPath(serviceName)
	jksFilePaths := mappingJKSFiles(jksFileDir)
	fmt.Println("the loop file path", jksFilePaths)
	for _, pathname := range jksFilePaths {
		SplitName := strings.Split(pathname, "/")
		TrimName := SplitName[len(SplitName)-1]
		validateEachJKS(pathname, TrimName)
	}
}

func validateEachJKS(jksFilePath string, jksFileName string) {
	//validate which service
	jksRawDate := decodeJKS(jksFilePath, jksFileName)
	if jksRawDate != "" {
		jksline := strings.Split(jksRawDate, "\n")
		var expiredDate []time.Time
		var owner []string
		for _, j := range jksline {
			if strings.Contains(j, "until:") {
				//convert each date info to RFC1123 format
				Validatetime := strings.TrimSpace(strings.Split(j, "until:")[1])
				temptimefield := strings.Split(Validatetime, " ")
				yearjks := temptimefield[5]
				monthjks := temptimefield[1]
				dayjks := temptimefield[2]
				MSjks := temptimefield[0]
				timehmsjks := temptimefield[3]
				timezonejks := strings.TrimSpace(temptimefield[4])
				regroupjks := MSjks + ", " + dayjks + " " + monthjks + " " + yearjks + " " + timehmsjks + " " + timezonejks
				time, err := time.Parse(time.RFC1123, regroupjks)
				if err != nil {
					fmt.Println("error  ", err)
				}
				// Sat May 25 19:54:12 GMT 2019
				expiredDate = append(expiredDate, time)
			}
		}
		for _, j := range jksline {
			if strings.Contains(j, "Owner:") {
				jksOwner := strings.TrimSpace(strings.Split(j, "=")[1])
				temptimefield := strings.Split(jksOwner, ",")[0]
				owner = append(owner, temptimefield)
			}
		}

		closedExpiredDate, difference := findClosedExpiredDate(expiredDate) //because expireddate is match the order as CN name
		for ind, o := range owner {
			t := time.Now()
			fmt.Println(t.Format(time.ANSIC), ":", jksFileName, ":", o, "is expiring:", difference[ind], "days")
		}
		color.Set(color.FgRed)
		fmt.Println(jksFileName, " closed expiring Date is: ", closedExpiredDate, " days")
		color.Unset()
	}
}

func decodeJKS(jksFilePath string, jksFileName string) string {
	cmd1 := exec.Command("keytool", "-list", "-v", "-keystore", jksFilePath)
	outputaws1, err := cmd1.Output()
	if err != nil {
		color.Set(color.FgYellow)
		t := time.Now()
		fmt.Println(t.Format(time.ANSIC), ":", jksFileName, ": Invalid keystore format")
		color.Unset()
		return ""
	}
	return string(outputaws1)
}

func findClosedExpiredDate(expiredDate []time.Time) (int, []int) {
	var currentTime = time.Now().UTC()
	// var closedtime time.Time
	var closedtime = expiredDate[0]
	for i := 0; i < len(expiredDate)-1; i++ {
		testingtime := expiredDate[i+1]
		if closedtime.After(testingtime) {
			closedtime = testingtime
		}
	}
	closedExpiredDate := closedtime.Sub(currentTime).Hours() / 24

	//push each time to each certificate
	var eachCertDiffDate []int
	for i := 0; i < len(expiredDate); i++ {
		temp := expiredDate[i].Sub(currentTime).Hours() / 24
		eachCertDiffDate = append(eachCertDiffDate, int(temp))
	}
	return int(closedExpiredDate), eachCertDiffDate
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func mappingJKSFolderPath(ServiceName string) []string {
	yamlFile, err := ioutil.ReadFile(jksYamlPath)
	if err != nil {
		log.Fatal(err)
	}

	var cfy ConfigYaml
	err = yaml.Unmarshal([]byte(yamlFile), &cfy.Jks)
	if err != nil {
		log.Fatalf("Failed to map the YAML file: %v", err)
	}

	var jksFileDir []string
	var Paths = cfy.Jks["jks"][ServiceName]
	for _, n := range Paths {
		var trimPathDupSpace = standardizeSpaces(n)
		fmt.Println("files want to validate ", trimPathDupSpace)
		jksFileDir = append(jksFileDir, trimPathDupSpace)
	}

	if len(jksFileDir) == 0 {
		fmt.Println("failed to map path with yaml file, please validate serviceName")
	}
	return jksFileDir
}

func mappingJKSFiles(paths []string) []string {
	var files []os.FileInfo
	var filesAbsolutePaths []string
	//based on the dir, find all files
	for _, j := range paths {
		if _, err := os.Stat(j); os.IsNotExist(err) {
			fmt.Printf("Dir does not exist: %s\n", j)
		} else {
			err := filepath.Walk(j, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					files = append(files, info)
					filesAbsolutePaths = append(filesAbsolutePaths, path)
				}
				return nil
			})
			if err != nil {
				fmt.Printf("Failed to read files under path: %s \n", err)
			}
		}
	}

	//find all Match findFileExt *.jks
	var filterFilesAbsolutePaths []string
	for i, f := range files {
		if strings.Contains(f.Name(), findFileExt) {
			filterFilesAbsolutePaths = append(filterFilesAbsolutePaths, filesAbsolutePaths[i])
		}
	}
	return filterFilesAbsolutePaths
}

func serviceCheck() string {

	yamlFile, err := ioutil.ReadFile(jksYamlPath)
	if err != nil {
		log.Fatal(err)
	}

	var cfy ConfigYaml
	err = yaml.Unmarshal([]byte(yamlFile), &cfy.Jks)
	if err != nil {
		log.Fatalf("Failed to map the YAML file: %v", err)
	}

	var es = cfy.Jks["jks"]["DEFAULTSERVICES"]
	var nodebyIHPservice = cfy.Jks["jks"]["NODENAMEBYDEFAULTIHPSERVICES"]
	var nodebyAWSservice = cfy.Jks["jks"]["NODENAMEBYDEFAULTAWSSERVICES"]

	output, err := Ihp_Hostname_Check.Output()
	if err != nil {
		log.Fatal(err)
	}
	for _, index := range es {
		ServiceName = index
		if strings.Contains(strings.ToLower(string(output)), strings.ToLower(ServiceName)) {
			IhpFound = true
			MatchServiceNameIhp = ServiceName
		}
	}

	// run addition validate service by ihp nodename (sepcial case)
	if !IhpFound {
		for _, ns := range nodebyIHPservice {
			ServiceName = strings.Split(ns, ":")[0]
			NodeFeatureName = strings.Split(ns, ":")[1]
			checkHostname := string(output)
			if strings.Contains(strings.ToLower(checkHostname), strings.ToLower(NodeFeatureName)) {
				IhpFound = true
				if ServiceName == "PBI|HDS" {
					testName := strings.Split(checkHostname, ".")[0]
					chars := []rune(testName)
					if unicode.IsLetter(chars[len(chars)-1]) { // validate hostname distinguish
						MatchServiceNameIhp = "PBI"
					} else {
						MatchServiceNameIhp = "HDS"
					}
				} else {
					MatchServiceNameIhp = ServiceName
				}
			}
		}
	}

	if IhpFound {
		fmt.Println(MatchServiceNameIhp, "host is in IHP")
	} else {
		fmt.Println("This host is not in IHP")
	}
	// Voila!
	for _, index := range es {
		ServiceName = index
		outputaws := getAWSenv(ServiceName)
		// fmt.Println("host name ", string(outputaws))
		// fmt.Println("yaml host ", strings.ToLower(ServiceName))
		if strings.Contains(strings.ToLower(string(outputaws)), strings.ToLower(ServiceName)) {
			AwsFound = true
			MatchServiceNameAws = ServiceName
		}
	}

	if MatchServiceNameIhp == "" { ///make same error output
		MatchServiceNameIhp = "This"
	}

	// run addition validate service by aws nodename (sepcial case)
	if !AwsFound {
		for _, ns := range nodebyAWSservice {
			ServiceName = ns
			outputaws := getAWSenv(ServiceName)
			if strings.Contains(strings.ToLower(string(outputaws)), strings.ToLower(ServiceName)) {
				AwsFound = true
				MatchServiceNameAws = ServiceName
			}
		}
	}

	if AwsFound {
		fmt.Print(MatchServiceNameAws, " host is in AWS\n")
	} else {
		fmt.Print(MatchServiceNameIhp, " host is not in AWS\n")
	}

	//handle failed auto mapping services
	if !IhpFound && !AwsFound {
		err := "Host does not belong to one of the services in the jks_scanpath.yaml file"
		// fmt.Println(err)
		log.Fatal(err)
	}
	if IhpFound {
		return MatchServiceNameIhp
	} else if AwsFound {
		return MatchServiceNameAws
	}
	return "Not Found"
}

func getAWSenv(ServiceName string) []byte {
	cmd1 := exec.Command("env")
	cmd2 := exec.Command("grep", "PS1")
	cmd3 := exec.Command("grep", strings.ToLower(ServiceName))

	// Get the pipe of Stdout from cmd1 and assign it
	// to the Stdin of cmd2.
	pipe1, err := cmd1.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd2.Stdin = pipe1

	// Start() cmd1, so we don't block on it.
	err = cmd1.Start()
	if err != nil {
		log.Fatal(err)
	}

	pipe2, err := cmd2.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd3.Stdin = pipe2
	// Start() cmd1, so we don't block on it.
	err = cmd2.Start()
	if err != nil {
		log.Fatal(err)
	}

	// Run Output() on cmd2 to capture the output.
	outputaws, _ := cmd3.Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	return outputaws
}
