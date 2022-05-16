package utils

import (
	"io/ioutil"
)

func ReadFileAsset(path string) string {
	if fileData, err := ioutil.ReadFile(path); err == nil {
		return string(fileData[:])
	} else {
		panic(err.Error())
	}
}
