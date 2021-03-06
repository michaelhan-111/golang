package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

/*	Declare variables here 	*/

/* create JSON struct to deal with Slack JSON */
type SlackJSONMessage struct {
	Clientmsgid string                  `json:"client_msg_id"`
	Slacktype   string                  `json:"type"`
	Text        string                  `json:"text"`
	User        string                  `json:"user"`
	Ts          string                  `json:"ts"`
	Purpose     string                  `json:"purpose"`
	Files       []SlackJSONMessageFiles `json:"files"`
}

/* create pointer to the array of Slack JSON objects */
type SlackJSONMessageResponse struct {
	Messages []SlackJSONMessage `json:"messages"`
}

type SlackJSONProfile struct {
	Title                 string      `json:"title"`
	Phone                 string      `json:"phone"`
	Skype                 string      `json:"skype"`
	RealName              string      `json:"real_name"`
	RealNameNormalized    string      `json:"real_name_normalized"`
	DisplayName           string      `json:"display_name"`
	DisplayNameNormalized string      `json:"display_name_normalized"`
	Fields                interface{} `json:"fields"`
	StatusText            string      `json:"status_text"`
	StatusEmoji           string      `json:"status_emoji"`
	StatusExpiration      int         `json:"status_expiration"`
	AvatarHash            string      `json:"avatar_hash"`
	Email                 string      `json:"email"`
	FirstName             string      `json:"first_name"`
	LastName              string      `json:"last_name"`
	Image24               string      `json:"image_24"`
	Image32               string      `json:"image_32"`
	Image48               string      `json:"image_48"`
	Image72               string      `json:"image_72"`
	Image192              string      `json:"image_192"`
	Image512              string      `json:"image_512"`
	StatusTextCanonical   string      `json:"status_text_canonical"`
}

type SlackJSONProfileResponse struct {
	Profile SlackJSONProfile `json:"profile"`
}

type SlackJSONMessageFiles struct {
	Id                 string `json:"id"`
	Created            int    `json:"created"`
	Timestamp          int    `json:"timestamp"`
	Name               string `json:"name"`
	Title              string `json:"title"`
	Mimetype           string `json:"mimetype"`
	Filetype           string `json:"filetype"`
	PrettyType         string `json:"pretty_type"`
	User               string `json:"user"`
	Editable           bool   `json:"editable"`
	Size               int    `json:"size"`
	Mode               string `json:"mode"`
	IsExternal         bool   `json:"is_external"`
	ExternalType       string `json:"external_type"`
	IsPublic           bool   `json:"is_public"`
	PublicUrlShared    bool   `json:"public_url_shared"`
	DisplayAsBot       bool   `json:"display_as_bot"`
	Username           string `json:"username"`
	UrlPrivate         string `json:"url_private"`
	UrlPrivateDownload string `json:"url_private_download"`
	Permalink          string `json:"permalink"`
	PermalinkPublic    string `json:"permalink_public"`
	EditLink           string `json:"edit_link"`
	Preview            string `json:"preview"`
	PreviewHighlight   string `json:"preview_highlight"`
	Lines              int    `json:"lines"`
	Lines_More         int    `json:"lines_more"`
	PreviewIsTruncated bool   `json:"preview_is_truncated"`
	IsStarred          bool   `json:"is_starred"`
	HasRichPreview     bool   `json:"has_rich_preview"`
}

//type SlackJSONMessageFilesResponse struct {
//	Files SlackJSONMessageFiles `json:"files"`
//}

type YamlConf struct {
	Slack_auth_token string `yaml:"slack_auth_token"`
	Slack_channel_id string `yaml:"slack_channel_id"`
}

var messageText string = "message"
var slackUserInfo = map[string]string{}
var configFile = "slack.yaml"

func main() {
	/* read output from yaml instead of hard coding the slack auth token and channel room inside the script/program */
	var yamlOutput YamlConf
	yamlOutput.readFromYaml()

	/* Example: url := "https://slack.com/api/channels.history?token=xoxp-267900653829-665768827328-663931256177-17b3511420c8ef1e1380b1144968a948&channel=CKK8J6JAY&pretty=1clear" */
	message_url := "https://slack.com/api/channels.history?token=" + yamlOutput.Slack_auth_token + "&channel=" + yamlOutput.Slack_channel_id + "&count=1000&pretty=1clear"
	req, err := http.NewRequest("GET", message_url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	//fmt.Println(string(body))

	/* Unmarshal the JSON output from Slack and then parse it */
	json_output := []byte(body)
	json_obj := SlackJSONMessageResponse{}
	err = json.Unmarshal(json_output, &json_obj)
	if err != nil {
		fmt.Println(err)
	}
	json_length := len(json_obj.Messages)
	//	fmt.Println("Length of json:", json_length)
	for i := 0; i < json_length; i++ {
		/* pull back only message types; not sure if there are other types when pulling back the messages data from Slack
		   epoch time conversion: string --> float64 --> int64 --> date; time package can't handle float64 */
		if json_obj.Messages[i].Slacktype == messageText && json_obj.Messages[i].Text != "" {
			x, err := strconv.ParseFloat(json_obj.Messages[i].Ts, 64)
			if err != nil {
				fmt.Println(err)
			}
			var nonfloatepoch int64 = int64(x)
			user_info_url := "https://slack.com/api/users.profile.get?token=" + yamlOutput.Slack_auth_token + "&user=" + json_obj.Messages[i].User + "&pretty=1"
			req_profile, err := http.NewRequest("GET", user_info_url, nil)
			if err != nil {
				fmt.Println("GET call failed, quitting")
				log.Fatal(err)
			}
			req_profile.Header.Add("Content-Type", "application/json")
			res_profile, err := http.DefaultClient.Do(req_profile)
			if err != nil {
				fmt.Println("Failed adding headers, quitting")
				log.Fatal(err)
			}
			body_profile, err := ioutil.ReadAll(res_profile.Body)
			if err != nil {
				fmt.Println("Failed to read JSON output before unmarshalling, quitting")
				log.Fatal(err)
			}
			defer res_profile.Body.Close()

			json_profile_output := []byte(body_profile)
			//			fmt.Println(string(json_profile_output))
			json_profile_obj := SlackJSONProfileResponse{}
			err = json.Unmarshal(json_profile_output, &json_profile_obj)
			if err != nil {
				fmt.Println("Failed to unmarshal the output, quitting")
				log.Fatal(err)
			}
			/* If the message text contains an "@" and 9 alphanumeric characters. Use that to search Slack for the user's profile
			   1. find message that contains "@"
			   2. within the message, pull out the text that contains @[A-Z]
			   3. take the text and use Slack's API to see if you can pull back a profile to get the user's real name
			   4. replace the display name with the real name in the message
			   5. print the updated message
			   This assumes the Slack User ID is a mix of capital letters and numbers, 9 in length
			   Pulling back the display name includes the @ --> strip the @ --> use this value to lookup the profile */
			//display_name_expression := regexp.MustCompile(`@([A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9])`)
			display_name_expression := regexp.MustCompile("@[A-Z0-9]{9}")
			//display_name_expression_formatted := regexp.MustCompile(`@[A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9][A-Z0-9]`)
			/* Need to add a check for messages that contain an @ symbol but is not a Slack ID */
			expression_match_check := display_name_expression.MatchString(json_obj.Messages[i].Text)
			//expression_match_check_formatted := display_name_expression_formatted.MatchString(json_obj.Messages[i].Text)
			if expression_match_check {
				find_slack_user_id := display_name_expression.FindString(json_obj.Messages[i].Text)
				slack_user_id_without_symbol := strings.Replace(find_slack_user_id, "@", "", -1)
				slack_user_id := retrieveSlackProfile(slack_user_id_without_symbol)
				//retrieveSlackProfile(slack_user_id_without_symbol)
				updated_message_text := strings.Replace(json_obj.Messages[i].Text, slack_user_id_without_symbol, slack_user_id, -1)
				//fmt.Println("find_slack_user_id:", find_slack_user_id)
				//fmt.Println("drop_symbol:", drop_symbol)
				//fmt.Println("slack_user_id:", slack_user_id)
				fmt.Println((time.Unix(nonfloatepoch, 0)), ":", json_profile_obj.Profile.RealName, ":", updated_message_text)
			} else {
				fmt.Println((time.Unix(nonfloatepoch, 0)), ":", json_profile_obj.Profile.RealName, ":", json_obj.Messages[i].Text)
			}
			//fmt.Println((time.Unix(nonfloatepoch, 0)), ":", updated_message_text, ":", json_profile_obj.Profile.RealName)
			//fmt.Println(json_profile_obj.Profile.Realname)
			//fmt.Println(string(body_profile))
			//json_profile_length := len(json_profile_obj.Profile)
			//fmt.Println((time.Unix(nonfloatepoch, 0)), ":", json_obj.Messages[i].Text, "BY:", json_obj.Messages[i].User)
			//fmt.Println(time.Unix(nonfloatepoch, 0))
			//fmt.Printf("%d\n", nonfloatepoch)
			//fmt.Println("Message:", json_obj.Messages[i].Text, "Timestamp:", json_obj.Messages[i].Ts)
		} else if json_obj.Messages[i].Slacktype == messageText && json_obj.Messages[i].Text == "" {
			//fmt.Println("Came across a message without text, hopefully a code snippet or attachment")
			fmt.Println("Code snippet or attachment:", json_obj.Messages[i].Files[0].Permalink)
		} else {
			fmt.Println("Came across a slack type that wasn't a message:", json_obj.Messages[i].Slacktype)
		}
	}
}

func (yamlInput *YamlConf) readFromYaml() {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Error reading yaml:", err)
	}
	err = yaml.Unmarshal(yamlFile, yamlInput)
	if err != nil {
		fmt.Println("Error unmarshalling:", err)
	}
}

func retrieveSlackProfile(slackID string) string {
	/* 1. create an empty map above
	   2. when pulling the values, add them to the map if not there already. key value pairs: Slack ID:Real Name
	   3. Each time, check the map to see if the key exists; if not, add. If the key does exist, pull back the Real Name from the map. This prevents unnecessary polling via the API. */
	_, exist := slackUserInfo[slackID]
	if exist {
		return slackUserInfo[slackID]
		//fmt.Println(slackUserInfo)
	} else {
		var yamlOutput YamlConf
		yamlOutput.readFromYaml()
		user_info_url := "https://slack.com/api/users.profile.get?token=" + yamlOutput.Slack_auth_token + "&user=" + slackID + "&pretty=1"
		req_profile, err := http.NewRequest("GET", user_info_url, nil)
		//fmt.Println(req_profile)
		if err != nil {
			fmt.Println("GET call failed, quitting")
			log.Fatal(err)
		}
		req_profile.Header.Add("Content-Type", "application/json")
		res_profile, err := http.DefaultClient.Do(req_profile)
		if err != nil {
			fmt.Println("Failed adding headers, quitting")
			log.Fatal(err)
		}
		body_profile, err := ioutil.ReadAll(res_profile.Body)
		if err != nil {
			fmt.Println("Failed to read JSON output before unmarshalling, quitting")
			log.Fatal(err)
		}
		defer res_profile.Body.Close()

		json_profile_output := []byte(body_profile)
		json_profile_obj := SlackJSONProfileResponse{}
		err = json.Unmarshal(json_profile_output, &json_profile_obj)
		if err != nil {
			fmt.Println("Failed to unmarshal the output, quitting")
			log.Fatal(err)
			//fmt.Println(err)
		}
		slackUserInfo[slackID] = json_profile_obj.Profile.RealName
		return slackUserInfo[slackID]
	}
}

/* USE FOR LATER

URL to pull back user info:
https://slack.com/api/users.profile.get?token=xoxp-267900653829-665768827328-663931256177-17b3511420c8ef1e1380b1144968a948&user=W8FHWMLNP&pretty=1

func translateSlackJSON(body []byte) (*SlackJSONResponse, error) {
	var s = new(SlackJSONResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("Error:", err)
	}
	return s, err
}

func checkerror(err error) {
	if err != nil {
		fmt.Println("error is: ", err)
		os.Exit(1)
	}
}

	//	fmt.Println("URL : ", url)
	// fmt.Println(string(body))

	json_length := len(x.Messages)
	fmt.Println(x.Messages)
	fmt.Println("Length of the json response:", json_length)
	for i := 0; i < json_length; i++ {
		fmt.Println(x.Messages[i])
	}

	var x interface{}
	err = json.Unmarshall(string(body), &x)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("JSON stuff : %+v", x)
*/
