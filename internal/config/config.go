package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type WsInfo struct {
	Url string `json:"url"`
	Pid int    `json:"pid"`
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".browser-tools-go", "ws.json"), nil
}

func SaveWsInfo(url string, pid int) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	info := WsInfo{Url: url, Pid: pid}
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func LoadWsInfo() (*WsInfo, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var info WsInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func RemoveWsInfo() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	// Check if file exists before removing
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}