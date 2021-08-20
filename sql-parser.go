package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var inputPath string = "./input"
var outputPath string = "./output"
var typeMapper map[string]string = make(map[string]string)

func main() {
	fillTypeMapper()

	inputFiles, err := ioutil.ReadDir(inputPath)

	if err != nil {
		log.Fatal("Can't read input files...")
	}

	if len(inputFiles) == 0 {
		log.Fatal("No files found to process...")
	}

	for _, inputFile := range inputFiles {
		fileExtension := filepath.Ext(inputFile.Name())

		if fileExtension != ".sql" {
			log.Println("Skipping " + inputFile.Name() + " as it is " + fileExtension + " file...")

			continue
		}

		fileName := inputFile.Name()[0 : len(inputFile.Name())-len(fileExtension)]

		log.Println(fileName + " started processing...")

		inputFileFullPath := inputPath + "/" + fileName + ".sql"

		dDLCommandAllText, readCommandError := ioutil.ReadFile(inputFileFullPath)
		valueSchemaString := formatDDLToValueSchemaFormattedString(readCommandError, inputFileFullPath, dDLCommandAllText, typeMapper, fileName)

		outputFileFullPath := outputPath + "/" + fileName + ".json"

		ioutil.WriteFile(outputFileFullPath, []byte(valueSchemaString), 0644)

		log.Println(fileName + " write complete...")
	}
}

func fillTypeMapper() {
	typeMapper["[bigint]"] = "INT64"
	typeMapper["[datetime]"] = "STRING"
	typeMapper["[int]"] = "INT64"
	typeMapper["[numeric](13,"] = "FLOAT64"
	typeMapper["[decimal](29,"] = "FLOAT64"

	for i := 1; i <= 255; i = i + 1 {
		typeMapper["[varchar]("+strconv.Itoa(i)+")"] = "STRING"
	}

	for i := 1; i <= 4; i = i + 1 {
		typeMapper["[char]("+strconv.Itoa(i)+")"] = "STRING"
	}
}

func formatDDLToValueSchemaFormattedString(readCommandError error, fileFullPath string, dDLCommandAllText []byte, typeMapper map[string]string, fileName string) string {
	if readCommandError != nil {
		log.Fatal("Unable to read the input file "+fileFullPath, readCommandError)
	}

	dDLCommands := splitByEmptyNewline(string(dDLCommandAllText))

	var fieldSchemasStringSlice []string

	for _, dDLCommand := range dDLCommands {
		if caseInsensitiveContains(dDLCommand, "create") {
			dDLCommandTrimmed, _ := getStringInBetweenTwoString(dDLCommand, "(", ") ON")

			dDLCommandColumns := strings.Split(dDLCommandTrimmed, ",\n")

			for index, dDLCommandColumn := range dDLCommandColumns {
				dDLCommandColumnTrimmed := strings.TrimSpace(dDLCommandColumn)

				dDLCommandColumnParts := strings.Split(dDLCommandColumnTrimmed, " ")

				nullable := "true"
				if strings.Contains(dDLCommandColumnTrimmed, "NOT NULL") {
					nullable = "false"
				}

				var fieldSchemasStringBuilder strings.Builder

				if index > 0 {
					fieldSchemasStringBuilder.WriteString(",")
				}

				fieldSchemasStringBuilder.WriteString("\"")
				columnName, _ := getStringInBetweenTwoString(dDLCommandColumnParts[0], "[", "]")
				fieldSchemasStringBuilder.WriteString(columnName)
				fieldSchemasStringBuilder.WriteString("\":{\"type\":\"")

				if typeMapper[dDLCommandColumnParts[1]] == "" {
					log.Fatal("Can't find type mapper for type " + dDLCommandColumnParts[1])
				}

				fieldSchemasStringBuilder.WriteString(typeMapper[dDLCommandColumnParts[1]])
				fieldSchemasStringBuilder.WriteString("\",\"isOptional\":")
				fieldSchemasStringBuilder.WriteString(nullable)
				fieldSchemasStringBuilder.WriteString("}")

				fieldSchemasStringSlice = append(fieldSchemasStringSlice, fieldSchemasStringBuilder.String())
			}
		}
	}

	valueSchemaStringBuilder := buildValueSchemaString(fileName, fieldSchemasStringSlice)
	valueSchemaString := strings.ReplaceAll(valueSchemaStringBuilder.String(), "\"", "\\\"")
	return valueSchemaString
}

func buildValueSchemaString(fileName string, fieldSchemasStringSlice []string) strings.Builder {
	var valueSchemaStringBuilder strings.Builder

	valueSchemaStringBuilder.WriteString("{\"name\":\"")
	valueSchemaStringBuilder.WriteString(fileName)
	valueSchemaStringBuilder.WriteString("\",\"type\":\"STRUCT\",\"isOptional\": false,\"fieldSchemas\":{")

	for _, fieldSchemasStringSliceElement := range fieldSchemasStringSlice {
		valueSchemaStringBuilder.WriteString(fieldSchemasStringSliceElement)
	}

	valueSchemaStringBuilder.WriteString("}}")

	return valueSchemaStringBuilder
}

func splitByEmptyNewline(str string) []string {
	strNormalized := regexp.
		MustCompile("\r\n").
		ReplaceAllString(str, "\n")

	return regexp.
		MustCompile(`\n\s*\n`).
		Split(strNormalized, -1)
}

func caseInsensitiveContains(a string, b string) bool {
	return strings.Contains(
		strings.ToLower(a),
		strings.ToLower(b),
	)
}

func getStringInBetweenTwoString(str string, startS string, endS string) (result string, found bool) {
	s := strings.Index(str, startS)
	if s == -1 {
		return result, false
	}
	newS := str[s+len(startS):]
	e := strings.Index(newS, endS)
	if e == -1 {
		return result, false
	}
	result = newS[:e]
	return result, true
}
