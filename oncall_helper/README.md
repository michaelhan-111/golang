Written by: 
- Michael Han
- Zhiwei Wang

This script was created for the following reasons: 
- Help reduce on-call noise; starting with having to delete files to get disk space within normal range(s)
- Designed to run all the time, schedule defined by cron 
- Designed to be hands off, until files have been deleted. If there's still a disk space issue, the SME for the respective service would need to look into it
- As a way to learn Go but find a work problem to solve to expedite the learning
- We will look for more problems to solve around on-call noise and add them to this script/program

Pre-Requisites: 
- services.yaml needs to be in the same directory; this is the config file for the script
- services.yaml needs to have your service there. You can follow HDS as an example. Make sure your service is not in there more than once
- The wrapper script was created since we have the services.yaml in the same directory as the script
- Splunk log setup has to be done on each individual service; HDS can be used as an example if needed 

Use/rules: 
- This script is designed to be easy to onboard; only the yaml file needs to be updated. The script should never need to be modified
- It will start looking for files 15 days or older and keep deleting until it gets down to 3 days; if disk space is is still an issue, an SME with need to be involved 
- Just add the cookbook to your runlist to install, "recipe[oncall_helper]". This will do the following things: 
    - create /usr/bin/oncall_helper directory
    - create the binary to execute
    - create the yaml config file
    - create the shell wrapper script
    - create the cron, currently set to run every hour, at the 4th minute
- Make sure that you verify your service is not already there; do not add duplicate services in the YAML file 
- Make sure to follow the formatting; HDS can we used as an example. If you add another service, make sure to add your service to the DEFAULTSERVICES stanza as well 
- We log the output to a log (via cron job) and send the log to Splunk. From there, we can add monitoring and email/notify recipients as needed
- Explanation of fields: 
    - extensions
        - This specifies which file extensions to delete under each directory specified in services.yaml. If not specified, it will use the following default log file extensions: ".log" and ".gz"
        - Example: if ".log" is defined, it will delete everything with ".log" extension under each directory specified  
    - extensions_ignore
        - This specifies which files to not delete; they need/have to be exact log names
    - SERVICESTORESTART
        - This specifies which services/processes to restart, based on service/process OWNER, not service/process name. 
            - Example: splunk process is started by splunk user. So script will look for processes started by the splunk user and kill them before restarting 
- For any changes to the services.yaml, they will need to be propagated from the Automations repo --> chef-repo-qa; /Automations/utilities/oncall_helper/services.yaml --> /chef-repo-qa/cookbooks/oncall_helper/files/default/services.yaml
- If you need to rebuild a binary after updating the source code: env GOOS=linux GOARCH=amd64 go build oncall_helper.go. 
    - For other platforms, refer to this: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04