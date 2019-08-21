Written by: 
- Michael Han

This script/program was created for the following reason(s): 
- To automate notifications to slack and emails 

Pre-Requisites: 
- Email integration has been added to the slack room(s) that are being informed. Slack email integration steps: https://slackhq.com/email-meet-slack-slack-email

Use/rules: 
- The configuration file is in the same directory as the script (notification_config.yaml); the location of the configuration file is hardcoded in the script
- To update the yaml, follow the structure that's currently in place. Any modifications to the structure will require code updates, which can be done
- To execute: ./slack_email_notification -group=sre OR go run slack_email_notification.go -group=sre (if you have go installed)
  - The group name is CASE SENSITIVE; review the notification_config.yaml for group names. In the example, sre and SRE are different
  - To confirm the script/program executed successfully, look for the following log message (below is an example):
    - 2019/08/19 09:32:07 xxxx sent!
    - no default was set intentionally
    - the script/program will not fail if an invalid group name is supplied as the group; this is done intentionally
  - The group name used in the command line matches what's under recipients in the yaml file 
- The binary included works on Mac OS and may not work on linux distros; not sure about Windows (created using go build, without defining GOOS and/or GOARCH; created on a MAC OS X running High Sierra). A binary can be created for other OS' as needed. 

Gotchas/Notes: 
- Email is being sent via gmail mail servers; they're reliable so I decided to use them
