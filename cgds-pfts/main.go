package main

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
	"compress/gzip"

	"golang.org/x/crypto/ssh"
)

//Declare variables here

func main() {
	yesterdays_date_full := time.Now().AddDate(0, 0, -1)
	yesterdays_date_only := yesterdays_date_full.Format("01-02-2006")
	yesterdays_year := yesterdays_date_full.Year()
	//year := strconv.Itoa(yesterdays_year)
	yesterdays_month := int(yesterdays_date_full.Month())
	month := strconv.Itoa(yesterdays_month)
	yesterdays_day := yesterdays_date_full.Day()
	day := strconv.Itoa(yesterdays_day)
	dir_base := "/data/reporting/biz/"

	if yesterdays_month < 10 {
		month = "0" + month
	}

	if yesterdays_day < 10 {
		day = "0" + day
	}

	full_path := (dir_base + strconv.Itoa(yesterdays_year) + "/" + month + "/" + day + "/")
	file_name := (yesterdays_date_only + ".tar.gz")
	//fmt.Println("Yesterday's date in full:", yesterdays_date_full)
	//fmt.Println("Yesterday's date only:", yesterdays_date_only)
	fmt.Println("Yesterday's year only:", yesterdays_year)
	fmt.Println("Yesterday's month only:", month)
	fmt.Println("Yesterday's day only:", day)
	fmt.Println("Full path to tar files:", full_path)
	fmt.Println("File name:", file_name)
	//	cmd_cd := exec.Command("cd", full_path)
	//	cmd_tar := exec.Command("tar", "-zcvf", file_name, full_path)
	os.Chdir(full_path)
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println("error dir: ", mydir)
	}

	dir, err := os.Open(full_path)
	checkerror(err)
	defer dir.Close()

	files, err := ioutil.ReadDir(full_path)
	if err != nil {
		log.Fatal(err)
	}

	tarfile, err := os.Create(file_name)
	checkerror(err)
	defer tarfile.Close()

	gw := gzip.NewWriter(tarfile)
	defer gw.Close()

	//var fileWriter io.WriteCloser = tarfile

	tarfileWriter := tar.NewWriter(gw)
	defer tarfileWriter.Close()

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}
		//file, err := os.Open(dir.Name() + string(filepath.Separator) + fileInfo.Name())
		file, err := os.Open(fileInfo.Name())

		checkerror(err)
		defer file.Close()

		header := new(tar.Header)
		header.Name = file.Name()
		header.Size = fileInfo.Size()
		header.Mode = int64(fileInfo.Mode())
		header.ModTime = fileInfo.ModTime()

		err = tarfileWriter.WriteHeader(header)
		checkerror(err)
		_, err = io.Copy(tarfileWriter, file)
		checkerror(err)
	}

	gw.Close()
	tarfileWriter.Close()
	fmt.Println("finish tar part")

	//move the create tar to dir_base
	//shell script picks up from dir_base

	new_file_name := dir_base + file_name
	err = os.Rename(file_name, new_file_name)
	if err != nil {
		log.Fatal(err)
	}

	shell_script_path := "/data/reporting/biz/cgds_script_push_file.sh"
	cmd := exec.Command("/bin/sh", shell_script_path)
	_, err = cmd.Output()
	if err != nil {
		println(err.Error())
		return
	}

	//clean up
	cmd_rm := exec.Command("rm", "-f", new_file_name)
	_, err = cmd_rm.Output()
	if err != nil {
		println(err.Error())
		return
	}

	/* SAVE FOR USE LATER:
		fmt.Println("Here are the files under", full_path, ":")
		for _, file := range files {
			fmt.Println(file.Name())
		}

	// Block to push the file to PFTS; decided to do it inside main b/c I wanted to use variables set inside main()

	user := "[username]"
	remote := "[endpoint]"
	port := "[port]"
	keypath := "[path of key]"

	// get host public key
	// hostKey := getHostKey(remote)

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			PublicKeyFile(keypath),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	// connect
	conn, err := ssh.Dial("tcp", remote+port, config)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// create new SFTP client
	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	srcPath := full_path + file_name
	dstPath := "/incoming"

	// create destination file
	dstFile, err := client.Create(dstPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	// create source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		log.Fatal(err)
	}

	// copy source file to destination file
	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d bytes copied\n", bytes)

	*/

	/* Need to figure out the following:
	1. tar the directory:
		   1. use os.exec
		   2. create tarball file, add files to existing file: https://ispycode.com/Blog/golang/2016-10/Archive-directory-with-tar
	2. send tarball to S3 bucket: either in this script or setup diff cron
	*/
}

func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func checkerror(err error) {
	if err != nil {
		fmt.Println("error is: ", err)
		os.Exit(1)
	}
}
