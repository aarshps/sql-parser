package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

var inputPath string = "./input"

func main() {
	typeMapper := make(map[string]string)
	typeMapper["[varchar](50)"] = "STRING"
	typeMapper["[bigint]"] = "INT"
	typeMapper["[datetime]"] = "STRING"
	typeMapper["[int]"] = "INT"
	typeMapper["[char](4)"] = "STRING"
	typeMapper["[char](1)"] = "STRING"
	typeMapper["[varchar](1)"] = "STRING"
	typeMapper["[varchar](5)"] = "STRING"
	typeMapper["[varchar](2)"] = "STRING"
	typeMapper["[varchar](3)"] = "STRING"
	typeMapper["[varchar](4)"] = "STRING"
	typeMapper["[varchar](30)"] = "STRING"
	typeMapper["[varchar](8)"] = "STRING"
	typeMapper["[varchar](200)"] = "STRING"
	typeMapper["[varchar](10)"] = "STRING"
	typeMapper["[varchar](255)"] = "STRING"
	typeMapper["[varchar](12)"] = "STRING"
	typeMapper["[varchar](11)"] = "STRING"
	typeMapper["[numeric](13,"] = "INT"

	fileName := "ABC.MedicalReplimental"

	fileFullPath := inputPath + "/" + fileName + ".sql"

	dDLCommandAllText, readCommandError := ioutil.ReadFile(fileFullPath)
	valueSchemaString := formatDDLToValueSchemaFormattedString(readCommandError, fileFullPath, dDLCommandAllText, typeMapper, fileName)

	fmt.Println(valueSchemaString)
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
