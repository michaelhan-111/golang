Written by: 
- Michael Han

This script/program was created for the following reasons: 
- To automate the timeline creation after an incident/issue has occurred
- The script/program takes the activity in a slack room and creates a timeline from it, in chronological order based off when the message was entered

Pre-Requisites: 
- This assumes most people that use this program will be on a Windows (amd64)  or Mac (darwin, amd64); for any other OS, we'll need to create another binary

Use/rules: 
- The configuration file, slack.yaml, needs to be in the same directory as the binary and/or go code
  - any time the tool needs to be run, the configuration file needs to be updated to the right slack room ID
    - the slack room ID can be found by right clicking on the room ID (i.e. inc0912477, etc) and copying the link; the slack room ID
is everything after the last '/' in the URL
- This only works for public slack rooms; testing against private slack rooms has not been done yet 

To execute: 
1. download the respective binary corresponding to the OS of the system where this tool will run
2. Create/download the configuration file, slack.yaml, to the same location where the binary is and will be executed. The configuration field: slack_channel_id will need to be updated everytime before running. This identifies which slack room to create the timeline from. To obtain the slack channel ID, refer to the Use/rules section above
3. For any issues, please report them to me: michael_han@intuit.com 
