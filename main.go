package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	instagramUrl = "https://www.instagram.com"
	userAgent    = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.92 Safari/537.36"
	queryMedia   = instagramUrl + "/graphql/query/?query_hash=%s&variables={\"id\":\"%s\",\"first\":%s,\"after\":\"%s\"}"
)

var transport *http.Transport
var wg sync.WaitGroup

var username = flag.String("username", "", "a instagram username")
var sessionId = flag.String("sessionId", "", "a valid instagram session id")
var proxy = flag.String("proxy", "", "network proxy")
var downloadSize = flag.String("downloadSize", "4", "parallel download size")
var parseSize = flag.String("parseSize", "12", "page size of media")

func init() {
	flag.Parse()
	err := setupNetwork()
	if err != nil {
		panic(err)
	}
}

func main() {
	err := parseUser(*username, time.Second*10)
	if err != nil {
		panic(err)
	}
}

func parseUser(username string, timeout time.Duration) error {
	err := os.Mkdir(username, 0777)
	if err != nil {
		return err
	}

	downloadUrl := make(chan string)
	size, err := strconv.Atoi(*downloadSize)
	if err != nil {
		return err
	}
	for i := 0; i < size; i++ {
		go downloadFile(username, downloadUrl, timeout)
		wg.Add(1)
	}
	userPage, err := getUserPage(username, timeout)
	if err != nil {
		return err
	}
	queryId, err := getQueryId(userPage, timeout)
	if err != nil {
		return err
	}
	targetUserId, err := getUserId(userPage)
	if err != nil {
		return err
	}

	var endCursor string
	for {
		queryUrl := fmt.Sprintf(queryMedia, queryId, targetUserId, *parseSize, endCursor)
		content, err := readContent(queryUrl, timeout)
		if err != nil {
			return err
		}

		var userMedia UserMedia
		json.Unmarshal(content, &userMedia)
		if userMedia.Status != "ok" {
			return errors.New("cannot get user media")
		}

		media := userMedia.Data.User.EdgeOwnerToTimelineMedia
		endCursor = media.PageInfo.EndCursor

		for _, m := range media.Edges {
			if m.Node.EdgeSidecarToChildren.Edges != nil {
				for _, n := range m.Node.EdgeSidecarToChildren.Edges {
					downloadUrl <- n.Node.DisplayURL
				}
			} else {
				if m.Node.IsVideo {
					downloadUrl <- m.Node.VideoURL
				} else {
					downloadUrl <- m.Node.DisplayURL
				}
			}
		}

		if !media.PageInfo.HasNextPage {
			break
		}
	}
	close(downloadUrl)
	wg.Wait()

	return nil
}

func setupNetwork() error {
	if *proxy != "" {
		proxyUrl, err := url.Parse(*proxy)
		if err != nil {
			return err
		}

		transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	} else {
		transport = &http.Transport{}
	}

	return nil
}

func readContent(url string, timeout time.Duration) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{Name: "sessionid", Value: *sessionId})
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Transport: transport, Timeout: timeout}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func getUserPage(username string, timeout time.Duration) ([]byte, error) {
	userPage, err := readContent(instagramUrl+"/"+username, timeout)
	if err != nil {
		return nil, err
	}

	return userPage, nil
}

func getUserId(userPage []byte) (string, error) {
	userIdCompile, err := regexp.Compile("<script type=\"text/javascript\">window._sharedData = (.+?);</script>")
	if err != nil {
		return "", err
	}

	userIdByte := userIdCompile.FindSubmatch(userPage)[1]
	var userMainPage UserMainPage
	json.Unmarshal(userIdByte, &userMainPage)
	userId := userMainPage.EntryData.ProfilePage[0].Graphql.User.ID

	return userId, err
}

func getQueryId(userPage []byte, timeout time.Duration) (string, error) {

	preLoadJSCompile, err := regexp.Compile("<link rel=\"preload\" href=\"(.+?)\"")
	if err != nil {
		return "", err
	}

	jsLinkByte := preLoadJSCompile.FindSubmatch(userPage)[1]
	jsLink := instagramUrl + string(jsLinkByte)

	profilePageContainer, err := readContent(jsLink, timeout)
	if err != nil {
		return "", err
	}

	queryIdCompile, err := regexp.Compile("queryId:\"(.+?)\"")
	if err != nil {
		return "", err
	}

	queryIdArray := queryIdCompile.FindAllSubmatch(profilePageContainer, -1)
	queryId := string(queryIdArray[2][1])

	return queryId, nil
}

func downloadFile(filepath string, url chan string, timeout time.Duration) {
	client := &http.Client{Transport: transport, Timeout: timeout}

	for downloadUrl := range url {
		resp, err := client.Get(downloadUrl)
		if err != nil {
			log.Println(err)
		}

		if resp.StatusCode != http.StatusOK {
			log.Println(errors.New(downloadUrl + " " + resp.Status))
		}

		downloadFileName := strings.Split(path.Base(downloadUrl), "?")[0]
		downloadFileName = path.Join(filepath, path.Base(downloadFileName))
		output, err := os.Create(downloadFileName)
		if err != nil {
			log.Println(err)
		}

		_, err = io.Copy(output, resp.Body)
		if err != nil {
			log.Println(err)
		}

		output.Close()
		resp.Body.Close()
	}

	wg.Done()
}
