package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
)

const TEPLATE_DIR = "templates"
const ROOT_TEPLATE_FILE = "layout.template"

func createCommonTemplate(filenames ...string) (*template.Template, error) {
	args := make([]string, 1, len(filenames)+1)
	args[0] = ROOT_TEPLATE_FILE
	args = append(args, filenames...)
	return createTemplateForFiles(args...)
}

func createTemplateForFiles(filenames ...string) (*template.Template, error) {

	fullFileNames := make([]string, 0, len(filenames))

	for _, element := range filenames {
		fullFileName := path.Join(TEPLATE_DIR, element)

		_, err := os.Stat(fullFileName)
		if err != nil {
			if os.IsNotExist(err) {
				log.Panicln(fullFileName + " does not exist")
			}
		}

		fullFileNames = append(fullFileNames, fullFileName)
	}

	temp, tempErr := template.ParseFiles(fullFileNames...)
	if tempErr != nil {
		fmt.Println(tempErr)
	}
	return temp, tempErr
}

func RenderInCommonTemplate(w http.ResponseWriter, dto interface{}, templateName string) error {
	template, err := createCommonTemplate(templateName)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := template.ExecuteTemplate(w, "layout", dto); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
