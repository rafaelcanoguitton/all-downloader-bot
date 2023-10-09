package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
  "github.com/joho/godotenv"


	bt "github.com/SakoDroid/telego/v2"
	cfg "github.com/SakoDroid/telego/v2/configs"
	objs "github.com/SakoDroid/telego/v2/objects"
)

type TikTokRedirectBody struct {
	redirect string
	follow   int
}

func main() {
  err := godotenv.Load()
  if err != nil {
    fmt.Println("Error loading .env file")
  }
  token := os.Getenv("TELEGRAM_TOKEN")
	bot, err := bt.NewBot(cfg.Default(token))
	if err != nil {
		panic(err)
	}

	// The general update channel.
	updateChannel := *(bot.GetUpdateChannel())

	// Adding a handler. Everytime the bot receives message "tiktok" in a private chat, it will wait for a link
	patternTiktok := `^https:\/\/(?:www\.)?tiktok\.com\/@[^/]+\/video\/\d+|https:\/\/vm\.tiktok\.com\/[^/]+$`
	patternReddit := `^https:\/\/reddit\.com\/r\/[A-Za-z0-9_]+\/s\/[A-Za-z0-9_]+$`

	bot.AddHandler(patternTiktok, func(u *objs.Update) {
		patternMobile := `^https:\/\/vm\.tiktok\.com\/[A-Za-z0-9_]+$`
		match, _ := regexp.MatchString(patternMobile, u.Message.Text)
		if match {
			u.Message.Text = redirectFromTikTokMobile(u.Message.Text)
		}

		// filePath := downloadTikTokContent(getTikTokVideoId(u.Message.Text))
		filePaths, fileType := downloadTikTokContent(getTikTokVideoId(u.Message.Text))
		if fileType == "video" {
			mediaSender := bot.SendVideo(u.Message.Chat.Id, 0, "", "", false)
			mediaSender.SendByFileIdOrUrl(filePaths[0], false, false)
		} else {
			// since telegram can only send 10 images at a time, we'll do several media groups
			mediaGroups := []*bt.MediaGroup{}
			for idx, fileUrl := range filePaths {
				if idx%10 == 0 {
					mediaGroups = append(mediaGroups, bot.CreateAlbum(0))
				}
				// if the index is the last one then skip it as we'll send it as a song
				if idx == len(filePaths)-1 {
					continue
				}
				pic, _ := mediaGroups[len(mediaGroups)-1].AddPhoto("", "", false, nil)
				pic.AddByFileIdOrURL(fileUrl)
			}
			// download the song, ffmpeg it to m4a and then send it
			songPath := filePaths[len(filePaths)-1]
			if err != nil {
				fmt.Println(err)
			}
			// download the song
			songRequest, err := http.NewRequest("GET", songPath, nil)
			songRequest.Header.Add("User-Agent", "TikTok 26.2.0 rv:262018 (iPhone; iOS 14.4.2; en_US) Cronet")
			songResponse, _ := http.DefaultClient.Do(songRequest)
			defer songResponse.Body.Close()
      //save it in a file to feed it to ffmpeg
      outputFileName := strconv.Itoa(rand.Intn(100000))
      outputFile, err := os.Create("./downloads/" + outputFileName + ".mp4")
      if err != nil {
        fmt.Println(err)
      }
      _, err = io.Copy(outputFile, songResponse.Body)
      defer outputFile.Close()
			// ffmpeg it to m4a
			cmd := exec.Command("ffmpeg", "-i", "./downloads/"+outputFileName + ".mp4", "-vn", "-acodec", "copy", "./downloads/"+outputFileName+".m4a")

      out, err := cmd.Output()
      if err != nil {
        fmt.Println(err)
      }
      fmt.Println(string(out))

			outputFileConverted, err := os.Open("./downloads/" + outputFileName + ".m4a")
      defer outputFileConverted.Close()
      // send the song
      songSender := bot.SendAudio(u.Message.Chat.Id, 0, "", "")
			if err != nil {
				fmt.Println(err)
			}
			for _, mediaGroup := range mediaGroups {
				_, err := mediaGroup.Send(u.Message.Chat.Id, false, false)
				if err != nil {
					fmt.Println(err)
				}
			}
			songSender.SendByFile(outputFileConverted, false, false)
      err = os.Remove("./downloads/" + outputFileName + ".mp4")
      if err != nil {
        fmt.Println(err)
      }
      err = os.Remove("./downloads/" + outputFileName + ".m4a")
      if err != nil {
        fmt.Println(err)
      }
		}

		// Register channel for receiving messages from this chat.
		// cc, _ := bot.AdvancedMode().RegisterChannel(strconv.Itoa(u.Message.Chat.Id), "message")
		// _, err := bot.SendMessage(u.Message.Chat.Id, "you sent a tiktok link", "", u.Message.MessageId, false, false)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// 	up := <-*cc
		// 	fmt.Println(up.Message.Text)
	}, "private")

	bot.AddHandler(patternReddit, func(u *objs.Update) {
		_, err := bot.SendMessage(u.Message.Chat.Id, "you sent a reddit link", "", u.Message.MessageId, false, false)
		if err != nil {
			fmt.Println(err)
		}
	}, "private")

	// any other message aka url

	// Monitores any other update. (Updates that don't contain text message "hi" in a private chat)
	go func() {
		for {
			update := <-updateChannel
			fmt.Println(update.Update_id)

			// Some processing on the update
		}
	}()

	bot.Run(true)
}

func getTikTokVideoId(url string) string {
	index := strings.Index(url, "/video/")
	var idVideo string
	if index != -1 {
		idVideo = url[index+7:]
		// purge query params if there are any
		index = strings.Index(idVideo, "?")
		if index != -1 {
			idVideo = idVideo[:index]
		}
		return idVideo
	}
	fmt.Println("'/video/' not found in the URL")
	return ""
}

// on the downloadTikTokContent function we'll return an array of files and a string detaling if it's a video or images
func downloadTikTokContent(id string) ([]string, string) {
	api_url := "https://api16-normal-c-useast1a.tiktokv.com/aweme/v1/feed/?aweme_id=" + id
	req, _ := http.NewRequest("GET", api_url, nil)
	req.Header.Add("User-Agent", "TikTok 26.2.0 rv:262018 (iPhone; iOS 14.4.2; en_US) Cronet")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var result map[string]interface{}
	json.Unmarshal([]byte(body), &result)
	isImage := nil != result["aweme_list"].([]interface{})[0].(map[string]interface{})["image_post_info"]
	fileUrls := []string{}
	if isImage {
		images := result["aweme_list"].([]interface{})[0].(map[string]interface{})["image_post_info"].(map[string]interface{})["images"].([]interface{})
		for _, image := range images {
			imageURI := image.(map[string]interface{})["display_image"].(map[string]interface{})["url_list"].([]interface{})[0]
			imageURIString, _ := imageURI.(string)
			fileUrls = append(fileUrls, imageURIString)
		}
		song_url := result["aweme_list"].([]interface{})[0].(map[string]interface{})["video"].(map[string]interface{})["play_addr"].(map[string]interface{})["url_list"].([]interface{})[0]
		songURLStr, _ := song_url.(string)
		fileUrls = append(fileUrls, songURLStr)
		return fileUrls, "image"
	}

	video_url := result["aweme_list"].([]interface{})[0].(map[string]interface{})["video"].(map[string]interface{})["play_addr"].(map[string]interface{})["url_list"].([]interface{})[0]
	videoURLStr, _ := video_url.(string)
	return []string{videoURLStr}, "video"
}

func redirectFromTikTokMobile(url string) string {
	fmt.Println("redirecting from mobile url: " + url)
	body := TikTokRedirectBody{
		redirect: "follow",
		follow:   10,
	}
	bodyMarshalled, _ := json.Marshal(body)
	req, _ := http.NewRequest("GET", url, bytes.NewReader(bodyMarshalled))
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	return res.Request.URL.String()
}
