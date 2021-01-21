package bot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) sendKeyStoreFile(path string) {
	stakeKeyPath := fmt.Sprintf("%s/key_%s.json", path, b.nConf.Network)
	msg := tgbotapi.NewDocumentUpload(b.ID, stakeKeyPath)
	b.bot.Send(msg)
}

func generateHostsfile(nodeIP string, target string) error {
	ipAddr := net.ParseIP(nodeIP)
	if ipAddr == nil {
		return errors.New("IP addr error")
	}
	text := fmt.Sprintf("[%s]\n%s", target, nodeIP)
	path := fmt.Sprintf("%s/hosts", dataPath)
	err := ioutil.WriteFile(path, []byte(text), 0666)
	if err != nil {
		return err
	}
	return nil
}

func getFileHostfile() (string, error) {
	path := fmt.Sprintf("%s/hosts", dataPath)
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

func getFileSSHKeyfie() (string, error) {
	path := fmt.Sprintf("%s/ssh_key", dataPath)
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
	path := fmt.Sprintf("%s/ssh_key", dataPath)
	err := ioutil.WriteFile(path, []byte(text), 0600)
	if err != nil {
		return err
	}
	return nil
}
