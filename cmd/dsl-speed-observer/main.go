package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

type fullOutput struct {
	Timestamp outputTime        `json:"timestamp"`
	UserInfo  *speedtest.User   `json:"user_info"`
	Servers   speedtest.Servers `json:"servers"`
}
type outputTime time.Time

func main() {
	user, err := speedtest.FetchUserInfo()
	if err != nil {
		fmt.Println("Warning: Cannot fetch user information. http://www.speedtest.net/speedtest-config.php is temporarily unavailable.")
	}
	printUser(user)

	serverList, err := speedtest.FetchServerList(user)
	logError(err)
	printServerList(serverList)

	targets, err := serverList.FindServer([]int{})
	logError(err)
	startTest(targets, false, true)

	jsonBytes, err := json.Marshal(
		fullOutput{
			Timestamp: outputTime(time.Now()),
			UserInfo:  user,
			Servers:   targets,
		},
	)
	logError(err)

	fmt.Println(string(jsonBytes))
}

func startTest(servers speedtest.Servers, savingMode bool, jsonOutput bool) {
	for _, s := range servers {
		printServer(s)

		err := s.PingTest()
		logError(err)

		if jsonOutput {
			err := s.DownloadTest(savingMode)
			logError(err)

			err = s.UploadTest(savingMode)
			logError(err)

			continue
		}

		printLatencyResult(s)

		err = testDownload(s, savingMode)
		logError(err)
		err = testUpload(s, savingMode)
		logError(err)

		printServerResult(s)
	}

	if !jsonOutput && len(servers) > 1 {
		printAverageServerResult(servers)
	}
}

func testDownload(server *speedtest.Server, savingMode bool) error {
	quit := make(chan bool)
	fmt.Printf("Download Test: ")
	go dots(quit)
	err := server.DownloadTest(savingMode)
	quit <- true
	if err != nil {
		return err
	}
	fmt.Println()
	return err
}

func testUpload(server *speedtest.Server, savingMode bool) error {
	quit := make(chan bool)
	fmt.Printf("Upload Test: ")
	go dots(quit)
	err := server.UploadTest(savingMode)
	quit <- true
	if err != nil {
		return err
	}
	fmt.Println()
	return nil
}

func printServerResult(server *speedtest.Server) {
	fmt.Printf(" \n")

	fmt.Printf("Download: %5.2f Mbit/s\n", server.DLSpeed)
	fmt.Printf("Upload: %5.2f Mbit/s\n\n", server.ULSpeed)
	valid := server.CheckResultValid()
	if !valid {
		fmt.Println("Warning: Result seems to be wrong. Please speedtest again.")
	}
}

func printAverageServerResult(servers speedtest.Servers) {
	avgDL := 0.0
	avgUL := 0.0
	for _, s := range servers {
		avgDL = avgDL + s.DLSpeed
		avgUL = avgUL + s.ULSpeed
	}
	fmt.Printf("Download Avg: %5.2f Mbit/s\n", avgDL/float64(len(servers)))
	fmt.Printf("Upload Avg: %5.2f Mbit/s\n", avgUL/float64(len(servers)))
}

func printLatencyResult(server *speedtest.Server) {
	fmt.Println("Latency:", server.Latency)
}

func printUser(user *speedtest.User) {
	if user.IP != "" {
		fmt.Printf("Testing From IP: %s\n", user.String())
	}
}

func printServerList(serverList speedtest.ServerList) {
	for _, s := range serverList.Servers {
		fmt.Printf("[%4s] %8.2fkm ", s.ID, s.Distance)
		fmt.Printf(s.Name + " (" + s.Country + ") by " + s.Sponsor + "\n")
	}
}

func printServer(s *speedtest.Server) {
	fmt.Printf(" \n")
	fmt.Printf("Target Server: [%4s] %8.2fkm ", s.ID, s.Distance)
	fmt.Printf(s.Name + " (" + s.Country + ") by " + s.Sponsor + "\n")
}

func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func dots(quit chan bool) {
	for {
		select {
		case <-quit:
			return
		default:
			time.Sleep(time.Second)
			fmt.Print(".")
		}
	}
}
