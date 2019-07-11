package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time" // "gopkg.in/gomail.v2"

	"gopkg.in/yaml.v2"
)

// Variable declarations
const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type MountDirPath struct {
	mountpath string
	dirpath   []string
}

type FilesExtToDeletePattern struct {
	ext       string
	extingore []string
}

type ConfigYaml struct {
	Services map[string]map[string][]string
}

type DfData struct {
	fs        string
	size      int64
	used      int64
	available int64
	percents  int64
	mount     string
}

var MatchServiceNameIhp = ""
var MatchServiceNameAws = ""
var ServiceName = ""
var DiskThreshold = 0.80
var NumDaysToDelete = 15
var MinNumDaysToDelete = 3
var Ihp_Hostname_Check = exec.Command("hostname")
var IhpFound = false
var AwsFound = false
var serviceYamlPath = "/usr/bin/oncall_helper/services.yaml"

func main() {
	//validate which service
	ServiceName = serviceCheck()

	///maping test-config
	paths, FileExtToDelete := mappingPath(ServiceName)

	fsinfo, _, _ := checkSpace("nodetail")

	fmt.Printf("Current Status is: " + "\nService Name = " + ServiceName)

	//allocated todo Path
	var todoPaths []MountDirPath
	for _, dfinfo := range fsinfo {
		if float64(dfinfo.percents)/100.00 > DiskThreshold {
			for _, yamlPath := range paths {
				if dfinfo.mount == yamlPath.mountpath {
					todoPaths = append(todoPaths, yamlPath)
				}
			}
		}
	}
	// fmt.Printf("\nChecking disk usage for todoPath(s):\n%s\n\n", todoPaths)
	if len(todoPaths) == 0 {
		fmt.Printf("\nDisk space is good, no need to delete. Here's the disk space usage:\n")
		_, _, dfrawinfo := checkSpace("")
		fmt.Println("\n", string(dfrawinfo))
	} else {
		fmt.Printf("File Extension we're deleting = " + FileExtToDelete.ext + "\nDeleting files older than = " + strconv.Itoa(NumDaysToDelete) + " days\n")
		fmt.Printf("Deleting from these path(s):\n%v\n", paths)
	}

	//disk space check based on the ServiceName
	var diskRate float64
	var deleteAction bool
	for _, n := range todoPaths {
		deleteAction = true
		for deleteAction {

			_, diskRate, _ = checkSpace(n.mountpath)
			fmt.Println("diskRate. ", diskRate)
			if diskRate > DiskThreshold {
				//search files need to delete
				FilesToDelete, FilesToDeletePaths := checkFiles(n.dirpath, FileExtToDelete, NumDaysToDelete)
				fmt.Printf("Files to delete: %s", FileExtToDelete.ext+"  with number of days to delete: "+strconv.Itoa(NumDaysToDelete))
				fmt.Printf("\nDeleting files: %s \n", FilesToDelete)
				//end search files

				//delete files based on serach result
				removeFiles(FilesToDelete, FilesToDeletePaths)
				//end delete files
				// check MaxNum Days Limited
				if NumDaysToDelete > MinNumDaysToDelete+1 {
					NumDaysToDelete--
				} else {
					fmt.Printf("\nFiles 3 days and older have been deleted, but Disk Usage is still high. We need to reach out to the SME for Service Name:" + ServiceName + "\n\n")
					// SendEmail(Email)
					goto End
				}
			} else {
				_, _, dfrawinfo := checkSpace(n.mountpath)
				fmt.Printf("Current Disk usage is fine\n\n")
				fmt.Println("\n", string(dfrawinfo))
				deleteAction = false
			}
		}
	}
End:
	// start of block for service restart
	yamlFile, err := ioutil.ReadFile(serviceYamlPath)
	if err != nil {
		log.Fatal(err)
	}

	var cfy ConfigYaml
	err = yaml.Unmarshal([]byte(yamlFile), &cfy.Services)
	if err != nil {
		log.Fatalf("Failed to map the YAML file: %v", err)
	}

	var es = cfy.Services["services"]["SERVICESTORESTART"]
	for _, index := range es {
		//fmt.Println("Here are the services that will be started:", index)
		//cmd1 := exec.Command("sudo /sbin/service start", index)
		cmd1 := exec.Command("echo", index)
		pipe, err := cmd1.Output()
		if err != nil {
			log.Fatal(err)
		}
		var service_name = (string(pipe))
		//		var kv map[string]string
		var ss []string
		ss = strings.Split(service_name, ":")
		var user = ss[0]
		var service = ss[1]
		var user_trim = strings.TrimSpace(user)
		var service_trim = strings.TrimSpace(service)

		//fmt.Println(results_kill)
		var numProcesses = returnNumProcesses(service_trim)
		if numProcesses == 0 {
			//fmt.Println("No Processes found, need to start it up bro.. ")
			//var results_kill = returnProcesses(user_trim, service_trim)
			//killProcesses(results_kill)
			cmd := exec.Command("/sbin/service", service_trim, "start")
			output, err := cmd.Output()
			time.Sleep(2 * time.Second)
			if err != nil {
				fmt.Println(err.Error())
			}
			if strings.Contains(strings.ToLower(string(output)), "ok") {
				var service_count = returnNumProcesses(service_trim)
				if service_count == 0 {
					fmt.Println("Successfully started", service_trim, "using /sbin/service but it didn't stay up; need to look into it")
				} else {
					fmt.Println("Successfully started", service_trim, "inside the loop that checks for the number of process after attempting to start it up")
				}
			} else {
				fmt.Println("/sbin/service failed to start up", service_trim, "; need to check out why")
			}
		} else {
			//fmt.Println("1 or more processes were found, need to kill them and start it up bro.. ")
			var results_kill = returnProcesses(user_trim, service_trim)
			killProcesses(results_kill)
			time.Sleep(2 * time.Second)
			// fmt.Println("Successfully killed", service_trim)
			cmd := exec.Command("/sbin/service", service_trim, "start")
			// fmt.Println("Successfully started", service_trim)
			output, err := cmd.Output()
			if err != nil {
				fmt.Println(err.Error())
			}
			if strings.Contains(strings.ToLower(string(output)), "ok") {
				fmt.Println("Successfully started", service_trim)
			} else {
				fmt.Println("Failed to startup after killing existing processes", service_trim, "; need to check out why")
			}
		}
	}
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func removeFiles(filetodelete []string, filesPath []string) string {
	if len(filesPath) == 0 {
		return "No files to delete!"
	}

	for _, d := range filesPath {
		var err = os.Remove(d)
		if err != nil {
			fmt.Printf("Failed to delete file: filename: %s\n", d)
			fmt.Println("Failed reason: ", err.Error())
		} else {
			fmt.Printf("File is successfully deleted. Filename: %s\n", d)
		}
	}
	return "Success"
}

func mappingPath(ServiceName string) ([]MountDirPath, FilesExtToDeletePattern) {
	yamlFile, err := ioutil.ReadFile(serviceYamlPath)
	if err != nil {
		log.Fatal(err)
	}

	var cfy ConfigYaml
	err = yaml.Unmarshal([]byte(yamlFile), &cfy.Services)
	if err != nil {
		log.Fatalf("Failed to map the YAML file: %v", err)
	}

	// fmt.Printf("\nYAML Mapping:\n%v\n\n", cfy.Services)

	var FileExtDel FilesExtToDeletePattern
	var MDPaths []MountDirPath
	var MDPath MountDirPath
	var Paths = cfy.Services["services"][ServiceName]
	for _, n := range Paths {
		var trimPathDupSpace = standardizeSpaces(n)
		var MountDirPath = strings.Split(trimPathDupSpace, " ")
		// fmt.Println("MountDirPath.  ", MountDirPath)
		if MountDirPath[0] == "extensions" {
			FileExtDel.ext = strings.Join(MountDirPath[1:], " ")
		} else if MountDirPath[0] == "extensions_ignore" {
			FileExtDel.extingore = MountDirPath[1:]
		} else {
			MDPath.mountpath = MountDirPath[0]
			MDPath.dirpath = MountDirPath[1:]
			MDPaths = append(MDPaths, MDPath)
		}
	}

	// fmt.Printf("\n MountDirPaths :\n%s\n", MDPaths)
	// fmt.Printf("\n FileExtDel :\n%s\n", FileExtDel)

	return MDPaths, FileExtDel
}

func checkFiles(paths []string, FileExtToDelete FilesExtToDeletePattern, NumDaysToDelete int) ([]string, []string) {
	if FileExtToDelete.ext == "" {
		FileExtToDelete.ext = ".log .gz"
	}
	if NumDaysToDelete == -1 {
		NumDaysToDelete = 15
	}
	var files []os.FileInfo
	var filesAbsolutePaths []string
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
	var filterFiles []string
	var filterFilesAbsolutePaths []string

	for i, f := range files {
		var fileTime = f.ModTime().UTC()
		var checkTime = fileTime.AddDate(0, 0, NumDaysToDelete)
		var currentTime = time.Now().UTC()
		var ValidFileExtToDelete []string

		for _, ext := range strings.Split(FileExtToDelete.ext, " ") {
			if ext != "" {
				ValidFileExtToDelete = append(ValidFileExtToDelete, ext)
			}
		}
		for _, eachExt := range ValidFileExtToDelete {
			if strings.Contains(f.Name(), eachExt) && checkTime.Before(currentTime) {
				filterFiles = append(filterFiles, f.Name())
				filterFilesAbsolutePaths = append(filterFilesAbsolutePaths, filesAbsolutePaths[i])
			}
		}
	}
	// fmt.Println("filterFiles", len(filterFiles), filterFiles)

	//filter out ingore parts (because index in filterFiles and filterFilesAbsolutePaths is one by one match, will reuse index to filter both arrays)
	var Files []string
	var FilesAbsolutePaths []string
	var sameFlag = false
	for m, n := range filterFiles {
		for _, k := range FileExtToDelete.extingore {
			if k == n { //found same element
				sameFlag = true
			}
		}
		if sameFlag {
			sameFlag = false //reset sameFlag and do nothing
		} else {
			sameFlag = false //reset sameFlag and add element to the finai remove list
			Files = append(Files, n)
			FilesAbsolutePaths = append(FilesAbsolutePaths, filterFilesAbsolutePaths[m])
		}
	}

	// fmt.Println("Files", len(Files), Files)
	return Files, FilesAbsolutePaths
}

//space check
func checkSpace(path string) (fsinfo []DfData, diskRate float64, dfrawinfo []byte) {
	cmd := "df"
	arg := "-Pkh"
	output, err := exec.Command(cmd, arg).CombinedOutput()
	if err != nil {
		log.Fatal()
	}
	if path != "nodetail" {
		fmt.Println("\n", string(output))
	}
	dfLines := strings.Split(string(output), "\n")
	// var fsinfo []DfData
	var f DfData
	for index, line := range dfLines {
		if index > 0 && strings.Replace(line, " ", "", -1) != "" {
			// fmt.Println("index ", index, " line  ", line)
			pm := strings.Split(standardizeSpaces(line), " ")
			f.fs = pm[0]
			f.size, err = strconv.ParseInt(strings.TrimRight(pm[1], "G"), 10, 64)
			f.used, err = strconv.ParseInt(strings.TrimRight(pm[2], "G"), 10, 64)
			f.available, err = strconv.ParseInt(strings.TrimRight(pm[3], "G"), 10, 64)
			f.percents, err = strconv.ParseInt(strings.TrimRight(pm[4], "%%"), 10, 64)
			f.mount = pm[5]
			fsinfo = append(fsinfo, f)
			if err != nil {
				fmt.Println("Failed to conevert Df -Pkh", err.Error())
			}
		}
	}
	if path == "" {
		return fsinfo, 0.00, nil
	} else {
		diskRate, err := mappingDiskRate(fsinfo, path)
		if err != "" {
			fmt.Println("")
		}
		return fsinfo, diskRate, dfrawinfo
	}
}

func mappingDiskRate(fsinfo []DfData, path string) (float64, string) {
	for _, lineinfo := range fsinfo {
		if path == lineinfo.mount {
			return float64(lineinfo.percents), ""
		}
	}
	return 0.00, "Cannot Match the Mount on Path"
}

func serviceCheck() string {

	yamlFile, err := ioutil.ReadFile(serviceYamlPath)
	if err != nil {
		log.Fatal(err)
	}

	var cfy ConfigYaml
	err = yaml.Unmarshal([]byte(yamlFile), &cfy.Services)
	if err != nil {
		log.Fatalf("Failed to map the YAML file: %v", err)
	}

	var es = cfy.Services["services"]["DEFAULTSERVICES"]

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

	if IhpFound {
		fmt.Println(MatchServiceNameIhp, "host is in IHP")
	} else {
		fmt.Println("This host is not in IHP")
	}

	// Voila!
	for _, index := range es {
		ServiceName = index
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

		pipe2, err := cmd3.StdoutPipe()
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
		// if err != nil {
		// 	fmt.Println(err.Error())
		// }
		if strings.Contains(strings.ToLower(string(outputaws)), strings.ToLower(ServiceName)) {
			AwsFound = true
			MatchServiceNameAws = ServiceName
		}
	}

	if MatchServiceNameIhp == "" { ///make same error output
		MatchServiceNameIhp = "This"
	}

	if AwsFound {
		fmt.Print(MatchServiceNameAws, " host is in AWS\n")
	} else {
		fmt.Print(MatchServiceNameIhp, " host is not in AWS\n")
	}

	//handle failed auto mapping services
	if !IhpFound && !AwsFound {
		err := "Host does not belong to one of the services in the services.yaml file"
		log.Fatal(err)
	}
	if IhpFound {
		return MatchServiceNameIhp
	} else if AwsFound {
		return MatchServiceNameAws
	}
	return "Not Found"
}

/* new command to return PID; need to account for process running as different user (i.e. splunk running as root):
   ps -u mhan-admin -o user,pid,comm | grep sshd | awk '{print $2}' */
func returnProcesses(user string, service string) string {

	cmd1 := exec.Command("ps", "-u", user, "-o", "user,pid,comm")
	cmd2 := exec.Command("grep", service)
	cmd3 := exec.Command("grep", "-v", "grep")
	cmd4 := exec.Command("awk", "{print $2}")

	pipe1, err := cmd1.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
		//		fmt.Println("Got through 1st pipe")
	}

	cmd2.Stdin = pipe1

	// Start() cmd1, so we don't block on it.
	err = cmd1.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	pipe2, err := cmd2.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
		//		fmt.Println("Got through 2nd pipe")
	}

	cmd3.Stdin = pipe2

	err = cmd2.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	pipe3, err := cmd3.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
		//		fmt.Println("Got through 2nd pipe")
	}

	cmd4.Stdin = pipe3

	err = cmd3.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	pipe4, err := cmd4.Output()
	if err != nil {
		fmt.Println(err.Error())
		//		fmt.Println("Got through 3rd pipe")
	}
	return (string(pipe4))
}

func returnNumProcesses(process string) int {
	//cmd1 := exec.Command("ps", "-eaf", process, "-o", "user,pid,comm")
	cmd1 := exec.Command("ps", "-eaf")
	cmd2 := exec.Command("grep", process)
	cmd3 := exec.Command("grep", "-v", "grep")
	cmd4 := exec.Command("wc", "-l")

	pipe1, err := cmd1.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
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

	err = cmd2.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	pipe3, err := cmd3.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
		//		fmt.Println("Got through 2nd pipe")
	}

	cmd4.Stdin = pipe3

	err = cmd3.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	pipe4, err := cmd4.Output()
	if err != nil {
		fmt.Println(err.Error())
		//		fmt.Println("Got through 3rd pipe")
	}

	var process_count = (string(pipe4))

	fmt.Println("Number of processes for process", process, "is", process_count)

	process_count_string := strings.TrimSuffix(process_count, "\n")
	process_count_int, err := strconv.Atoi(process_count_string)
	if err != nil {
		fmt.Println(err)
	}

	if process_count_int == 0 {
		//fmt.Println("Process count is", process_count_int, "for user", user, "; need to start service")
		return 0
	} else {
		//fmt.Println("Process count is more than 1, it's", process_count_int, "for user", user, "; need to stop all services and start up once")
		// killProcesses(process_count_string)
		return 1
	}
}

func killProcesses(pids string) {
	scanner := bufio.NewScanner(strings.NewReader(pids))
	for scanner.Scan() {
		pid_string := strings.TrimSuffix(scanner.Text(), "\n")
		pid_int, err := strconv.Atoi(pid_string)
		//		pipe, err := cmd1.Output()
		if err != nil {
			log.Fatal(err)
		}
		//	exec.Command("kill", scanner.Text())
		fmt.Println("Killing this PID:", pid_int)
		syscall.Kill(pid_int, syscall.SIGKILL)
	}
}

/*..........SendEmail function ............
type DefaultEmailSetting struct {
	from    string
	to      []string
	pass    string
	cc      []string
	bcc     []string
	subject string
	body    string
}

func SendEmail(email DefaultEmailSetting) {
	from := email.from
	pass := email.pass
	to := email.to
	cc := email.cc

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Cc", cc...)
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "Hello <b>Michael</b> and <i>Zhiwei</i>!")
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer("smtp.gmail.com", 587, from, pass)

	// Send the email to mutiple receipts.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	} else {
		log.Print("Email Sent!!!!")
	}
}
......End SentEmail function.........*/
