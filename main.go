package main

import (
	"github.com/latchmihay/rtapi"
	"fmt"
	"log"
	"strconv"
	"time"
	"os"
	"regexp"
	"strings"
	"path/filepath"
	"errors"
	"os/exec"
)

const (
	hourTime = 3600
	weekTime = 604800
	logFileLoc = "/tmp/ks-tor.log"
)

// check error function
func CheckErr(err error) {
	if err == nil {
		return
	}
	log.Fatal(err)
}

func checkFolder(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		msg := fmt.Sprintf("Folder %s was not found! Torrent needs extraction!", path)
		return errors.New(msg)
	}
	return nil
}

func extract(from string,to string) error {
	script := "/var/www/rutorrent/plugins/unpack/unrar_dir.sh"
	arg1 := "/usr/bin/unrar"
	out, err := exec.Command(script, arg1, from, to).Output()
	if err != nil {
		return err
	}
	log.Printf("\n%s\n", out)
	return nil
}

func main() {
	f, err := os.OpenFile(logFileLoc, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	// set logging format
	log.SetFlags(log.LstdFlags)
	log.SetOutput(f)

	rt, err := rtapi.NewRtorrent("127.0.0.1:51102") // Or /path/to/socket for "scgi_local".
	if err != nil {
		CheckErr(err)
	}

	// Get torrents
	torrents, err := rt.Torrents()
	if err != nil {
		CheckErr(err)
	}
	torrents.Sort(rtapi.ByAgeLoad)

	log.Println("Number of torrents:", len(torrents))

	for _, t := range torrents {
		// Run only on finished torrents
		if t.State == "Seeding" {
			ageToString := strconv.FormatUint(t.AgeLoad, 10)
			ageStringToInt, err := strconv.ParseInt(ageToString, 10,64)
			CheckErr(err)
			humanTime := time.Unix(ageStringToInt, 0)

			ageToInt, err := strconv.Atoi(ageToString)
			CheckErr(err)

			ageToIntWeeek := ageToInt + weekTime + hourTime

			now := time.Now()
			nowUnix := now.Unix()

			nowUnixString := strconv.FormatInt(nowUnix, 10)
			nowUnixInt, err := strconv.Atoi(nowUnixString)
			CheckErr(err)

			fmt.Printf("%s [%s] [%s] [%s] [%s]\n",t.Name,t.State,t.Percent, t.Path, humanTime)
			if nowUnixInt > ageToIntWeeek {
				// Torrent is older then a week, deleting it with its data
				log.Printf("Torrent %s is older than a Week and a hour, deleting it now!", t.Name)
				err = rt.Delete(true, t)
				CheckErr(err)

			}

			tvRegEx, err := regexp.Compile("(?i)s[0-9][0-9]e[0-9][0-9]")
			CheckErr(err)
			sesonRegEx, err := regexp.Compile("(?i)s[0-9][0-9]")
			CheckErr(err)
			movRegEx, err := regexp.Compile("(?i)(bluray)|(hdrip)|(brrip)|(web-dl)|(dvdrip)|(webrip)|(hdtv)|(bdrip)")
			CheckErr(err)
			sportRegEx, err := regexp.Compile("(?i)(motogp)|(formula)")
			CheckErr(err)

			parseTorName := strings.Split(t.Name, ".")
			loop:for _,v := range parseTorName {
				switch {
				case tvRegEx.MatchString(v):
					log.Printf("Torrent %s matched TV\n", t.Name)
					folderName := filepath.Base(t.Path)
					err = checkFolder("/TV/" + folderName)
					if err != nil {
						log.Println(err)
						extract(t.Path, "/TV/" )
					}
					break loop
				case sesonRegEx.MatchString(v):
					log.Printf("Torrent %s matched TV\n", t.Name)
					folderName := filepath.Base(t.Path)
					err = checkFolder("/TV/" + folderName)
					if err != nil {
						log.Println(err)
						extract(t.Path, "/TV/" )
					}
					break loop
				case sportRegEx.MatchString(v):
					log.Printf("Torrent %s matched Sport\n", t.Name)
					folderName := filepath.Base(t.Path)
					err = checkFolder("/Sport/" + folderName)
					if err != nil {
						log.Println(err)
						extract(t.Path, "/Sport/")
					}
					break loop
				case movRegEx.MatchString(v):
					log.Printf("Torrent %s matched Movie\n", t.Name)
					folderName := filepath.Base(t.Path)
					err = checkFolder("/Mov/" + folderName)
					if err != nil {
						log.Println(err)
						extract(t.Path, "/Mov/" )
					}
					break loop
				}
			}
		}
	}
}
