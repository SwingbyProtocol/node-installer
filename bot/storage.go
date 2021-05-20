package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) sendLogFile(path string) error {
	logPath := fmt.Sprintf("%s/data/logs.log", path)
	msg := tgbotapi.NewDocumentUpload(b.ID, logPath)
	_, err := b.bot.Send(msg)
	return err
}

func getVersion() (*Version, error) {
	ver := &Version{}
	path := fmt.Sprintf(".version.json")
	str, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(str, ver)
	if err != nil {
		return nil, err
	}
	return ver, nil
}

func generateHostsfile(nodeIP string, target string) error {
	ipAddr := net.ParseIP(nodeIP)
	if ipAddr == nil {
		return errors.New("IP addr error")
	}
	text := fmt.Sprintf("[%s]\n%s", target, nodeIP)
	path := fmt.Sprintf("%s/hosts", DataPath)
	err := ioutil.WriteFile(path, []byte(text), 0666)
	if err != nil {
		return err
	}
	return nil
}

func getFileHostfile() (string, error) {
	path := fmt.Sprintf("%s/hosts", DataPath)
	str, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	strs := strings.Split(string(str), "]")
	ipAddr := net.ParseIP(strs[1][1:])
	if ipAddr == nil {
		return "", errors.New("IP addr parse error")
	}
	return ipAddr.String(), nil
}

func getDirSizeFromFile() (int, error) {
	path := fmt.Sprintf("/tmp/dir_size")
	str, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	strs := strings.Split(string(str), "\t")
	//log.Info(strs)
	intNum, _ := strconv.Atoi(strs[0])
	return intNum, nil
}

func getDiskSpaceFromFile() (int, error) {
	path := fmt.Sprintf("/tmp/var_size")
	str, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	strs := strings.Split(string(str), "\t")
	//log.Info(strs)
	intNum, _ := strconv.Atoi(strs[0])
	return intNum, nil
}

func getAvailableDiskSpaceFromFile() (int, error) {
	path := fmt.Sprintf("/tmp/var_available_size")
	str, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	strs := strings.Split(string(str), "\t")
	//log.Info(strs)
	intNum, _ := strconv.Atoi(strs[0])
	return intNum, nil
}

func getFileSSHKeyfie() (string, error) {
	path := fmt.Sprintf("%s/ssh_key", DataPath)
	str, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

func storeSSHKeyfile(key string) error {
	last := key[len(key)-1:]
	text := fmt.Sprintf("%s", key)
	if last != "\n" {
		text = fmt.Sprintf("%s%s", text, "\n")
	}
	path := fmt.Sprintf("%s/ssh_key", DataPath)
	err := ioutil.WriteFile(path, []byte(text), 0600)
	if err != nil {
		return err
	}
	return nil
}
