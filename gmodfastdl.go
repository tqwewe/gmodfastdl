package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const inputPrompt string = ": "

func main() {
	var (
		err               error
		gmodPath          string = "garrysmod"
		fastdl            []string
		workshopContent   string
		downloadFileTypes []string = []string{
			"mdl",
			"vmt",
			"vtf",
			"wav",
			"mp3",
		}
	)

	workshop := flag.Int("workshop", 0, "Workshop Collection ID to add to FastDL")
	flag.Parse()

	if *workshop == 0 {
		fmt.Println("No '-workshop' ID specified, would you like to continue anyway? (Y/n)")
		if strings.ToLower(Scanln()) != "y" {
			os.Exit(0)
		}
	} else {
		fmt.Print("Scanning workshop content")
		workshopContent, err = GetWorkshopAddons(*workshop)
		if err != nil {
			PrintFail()
		} else {
			PrintSuccess()
		}
	}

	path, err := os.Stat(gmodPath)
	if os.IsNotExist(err) || !path.IsDir() {
		fmt.Println("Could not find garrysmod folder. Please specify the path of your garrysmod folder relative to this program.")
		for {
			path := Scanln()
			if fi, err := os.Stat(ConcatDir(path, "addons")); err == nil {
				if fi.IsDir() {
					gmodPath = path
					break
				} else {
					fmt.Printf("%v is not a directory.\n", ConcatDir(path, "addons"))
					continue
				}
			} else {
				fmt.Printf("Unable to locate %v.\n", ConcatDir(path, "addons"))
				continue
			}
		}
	}

	fmt.Print("Scanning reclusively for fastdl files in " + gmodPath + "/addons/*")
	if err := filepath.Walk(ConcatDir(gmodPath, "addons"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		for _, fileType := range downloadFileTypes {
			if match, _ := filepath.Match("*."+fileType, info.Name()); match {
				fastdl = append(fastdl, path)
				return nil
			}
		}
		return nil
	}); err != nil {
		PrintFail()
	} else {
		PrintSuccess()
	}

	var fastdlString string = "if ( SERVER ) then\n"
	for _, path := range fastdl {
		path = strings.Replace(path, string(os.PathSeparator), "/", -1)
		path = strings.Replace(path, `\`, "/", -1)
		path = strings.TrimPrefix(path, gmodPath + "/addons/")
		if !regexp.MustCompile(`.*_\d{6,}\/.*`).MatchString(path) {
			sm := regexp.MustCompile(`.*?\/(.*)`).FindAllStringSubmatch(path, -1)
			if len(sm) > 0 {
				if len(sm[0]) > 1 {
					fastdlString += "    resource.AddFile('" + sm[0][1] + "')\n"
					continue
				}
			}
		}
	}
	fastdlString += "end"

	fmt.Print("Scanning for directory ", gmodPath, "/lua/autorun/server")
	if _, err := os.Stat(ConcatDir(gmodPath, "lua", "autorun", "server")); os.IsNotExist(err) {
		PrintFail()

		fmt.Print("Attempting to create directory ", gmodPath, "/lua/autorun/server")
		os.Mkdir(ConcatDir(gmodPath, "lua"), os.ModePerm)
		os.Mkdir(ConcatDir(gmodPath, "lua", "autorun"), os.ModePerm)
		os.Mkdir(ConcatDir(gmodPath, "lua", "autorun", "server"), os.ModePerm)

		if _, err := os.Stat(ConcatDir(gmodPath, "lua", "autorun", "server")); os.IsNotExist(err) {
			PrintFail()

			fmt.Print("Writing to temporary file in ", gmodPath+"/fastdl.lua")
			err = ioutil.WriteFile(ConcatDir(gmodPath, "fastdl.lua"), []byte(workshopContent + fastdlString), 755)
			if err != nil {
				PrintFail()
				fmt.Print("FAILED TO WRITE FASTDL TO FILE")
				PrintWarning()
				fmt.Print("Echoing output to terminal in 5 seconds")
				PrintWarning()

				fmt.Print(workshopContent + fastdlString)
			} else {
				PrintSuccess()
			}

			fmt.Print("YOU MUST MOVE THE 'fastdl.lua' FILE TO YOUR 'lua/autorun/server' FOLDER IN YOUR GMOD SERVER BEFORE YOUR FASTDL WILL WORK")
			PrintWarning()
		} else {
			PrintSuccess()

			fmt.Print("Writing file 'fastdl.lua'")
			err = ioutil.WriteFile(ConcatDir(gmodPath, "lua", "autorun", "server", "fastdl.lua"), []byte(workshopContent + fastdlString), 755)
			if err != nil {
				PrintFail()

				fmt.Print("FAILED TO WRITE FASTDL TO FILE")
				PrintWarning()

				fmt.Print("Echoing output to terminal in 5 seconds")
				PrintWarning()

				fmt.Print(workshopContent + fastdlString)
			} else {
				PrintSuccess()
			}
		}
	} else {
		PrintSuccess()

		fmt.Print("Writing file 'fastdl.lua'")
		err = ioutil.WriteFile(ConcatDir(gmodPath, "lua", "autorun", "server", "fastdl.lua"), []byte(workshopContent + fastdlString), 755)
		if err != nil {
			PrintFail()

			fmt.Print("FAILED TO WRITE FASTDL TO FILE")
			PrintWarning()

			fmt.Print("Echoing output to terminal in 5 seconds")
			PrintWarning()

			fmt.Print(workshopContent + fastdlString)
		} else {
			PrintSuccess()
		}
	}

	fmt.Println("Bye Bye!")
}

func ConcatDir(dirs ...string) string {
	var finalPath string
	for i, dir := range dirs {
		finalPath += dir
		if i < len(dirs)-1 {
			finalPath += string(os.PathSeparator)
		}
	}
	return finalPath
}

func Scanln() string {
	color.New(color.Bold, color.FgBlue).PrintFunc()(inputPrompt)
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return ""
	}
	input = strings.Replace(input, "\n", "", -1)
	input = strings.Replace(input, "\r", "", -1)
	return input
}

func PrintSuccess() {
	color.New(color.Bold, color.FgHiWhite).PrintFunc()(" [")
	color.New(color.Bold, color.FgGreen).PrintFunc()("Success")
	color.New(color.Bold, color.FgHiWhite).PrintFunc()("]\n")
}

func PrintFail() {
	color.New(color.Bold, color.FgHiWhite).PrintFunc()(" [")
	color.New(color.Bold, color.FgRed).PrintFunc()("Fail")
	color.New(color.Bold, color.FgHiWhite).PrintFunc()("]\n")
}

func PrintWarning() {
	color.New(color.Bold, color.FgHiWhite).PrintFunc()(" [")
	color.New(color.Bold, color.FgYellow).PrintFunc()("Warning")
	color.New(color.Bold, color.FgHiWhite).PrintFunc()("]\n")
}

func GetWorkshopAddons(id int) (string, error) {
	resp, err := http.Get("http://steamcommunity.com/sharedfiles/filedetails/?id=" + strconv.FormatInt(int64(id), 10))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	addons := regexp.MustCompile(`<a href="http:\/\/steamcommunity.com\/sharedfiles\/filedetails\/\?id=(\d+)"><div class="workshopItemTitle">(.*)<\/div><\/a>`).FindAllStringSubmatch(string(body), -1)
	var addonString string
	for _, addon := range addons {
		if len(addon) >= 1 {
			addonString += "resource.AddFile('" + addon[1] + "')"
		}
		if len(addon) >= 2 {
			addonString += " -- " + addon[2]
		}
		addonString += "\n"
	}

	return addonString, nil
}
