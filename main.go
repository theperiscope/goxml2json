package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type stringArrayFlag []string

func (f *stringArrayFlag) String() string {
	return strings.Join(*f, ", ")
}

func (f *stringArrayFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var appSettingsSlotSettings, connectionStringsSlotSettings stringArrayFlag

type AppSetting struct {
	XMLName     xml.Name `xml:"add" json:"-"`
	Key         string   `xml:"key,attr" json:"name"`
	Value       string   `xml:"value,attr" json:"value"`
	SlotSetting bool     `xml:"-" json:"slotSetting"`
}

type AppSettings struct {
	XMLName      xml.Name     `xml:"appSettings" json:"-"`
	ConfigSource string       `xml:"configSource,attr" json:="-"`
	AppSettings  []AppSetting `xml:"add"`
}

type ConnectionString struct {
	XMLName          xml.Name `xml:"add" json:"-"`
	Name             string   `xml:"name,attr" json:"name"`
	ConnectionString string   `xml:"connectionString,attr" json:"value"`
	ProviderName     string   `xml:"providerName" json:"type"`
	SlotSetting      bool     `xml:"-" json:"slotSetting"`
}

type ConnectionStrings struct {
	XMLName           xml.Name           `xml:"connectionStrings" json:"-"`
	ConfigSource      string             `xml:"configSource,attr" json:="-"`
	ConnectionStrings []ConnectionString `xml:"add"`
}

type Configuration struct {
	XMLName           xml.Name          `xml:"configuration" json:"-"`
	AppSettings       AppSettings       `xml:"appSettings"`
	ConnectionStrings ConnectionStrings `xml:"connectionStrings"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected 'appSettings' or 'connectionStrings' subcommands")
		os.Exit(1)
	}

	appSettingsCmd := flag.NewFlagSet("appSettings", flag.ExitOnError)
	var appSettingsInputFileName, appSettingsOutputFileName string
	appSettingsCmd.StringVar(&appSettingsInputFileName, "i", "", "required, input file name")
	appSettingsCmd.StringVar(&appSettingsOutputFileName, "o", "", "required, output file name")
	appSettingsCmd.Var(&appSettingsSlotSettings, "slotSetting", "optional, multiple allowed, use to specify names of the slot settings")

	connectionStringsCmd := flag.NewFlagSet("connectionStrings", flag.ExitOnError)
	var connectionStringsInputFileName, connectionStringsOutputFileName string
	connectionStringsCmd.StringVar(&connectionStringsInputFileName, "i", "", "required, input file name")
	connectionStringsCmd.StringVar(&connectionStringsOutputFileName, "o", "", "required, output file name")
	connectionStringsCmd.Var(&connectionStringsSlotSettings, "slotSetting", "optional, multiple allowed, use to specify names of the slot settings")

	switch os.Args[1] {
	case "appSettings":
		appSettingsCmd.Parse(os.Args[2:])
		if appSettingsInputFileName == "" || appSettingsOutputFileName == "" {
			fmt.Fprintf(os.Stderr, "missing required -i and -o flags\n")
			os.Exit(2)
		}
		ProcessAppSettings(appSettingsInputFileName, appSettingsOutputFileName)
	case "connectionStrings":
		connectionStringsCmd.Parse(os.Args[2:])
		if connectionStringsInputFileName == "" || connectionStringsOutputFileName == "" {
			fmt.Fprintf(os.Stderr, "missing required -i and -o flags\n")
			os.Exit(2)
		}
		ProcessWebConnections(connectionStringsInputFileName, connectionStringsOutputFileName)
	default:
		fmt.Println("expected 'appSettings' or 'connectionStrings' subcommands")
		os.Exit(1)
	}
}

func NewConfigurationFromFile(fileName string) (*Configuration, error) {
	input, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer input.Close()

	byteValue, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, err
	}

	var configuration Configuration
	err = xml.Unmarshal(byteValue, &configuration)
	if err != nil {
		return nil, err
	}

	appSettings := configuration.AppSettings

	if len(appSettings.ConfigSource) > 0 {
		input, err := os.Open(appSettings.ConfigSource)
		if err != nil {
			return nil, err
		}
		defer input.Close()

		byteValue, err := ioutil.ReadAll(input)
		if err != nil {
			return nil, err
		}

		err = xml.Unmarshal(byteValue, &configuration.AppSettings)
		if err != nil {
			return nil, err
		}
	}

	connectionStrings := configuration.ConnectionStrings

	if len(connectionStrings.ConfigSource) > 0 {
		input, err := os.Open(connectionStrings.ConfigSource)
		if err != nil {
			log.Fatal(err)
		}
		defer input.Close()

		byteValue, err := ioutil.ReadAll(input)
		if err != nil {
			log.Fatal(err)
		}

		err = xml.Unmarshal(byteValue, &configuration.ConnectionStrings)
		if err != nil {
			log.Fatal(err)
		}
	}

	return &configuration, nil
}

func ProcessAppSettings(inputFileName, outputFileName string) {
	config, err := NewConfigurationFromFile(inputFileName)
	if err != nil {
		log.Fatal(err)
	}

	appSettings := config.AppSettings

	for i := range appSettings.AppSettings {
		for _, slotSetting := range appSettingsSlotSettings {
			if appSettings.AppSettings[i].Key == slotSetting {
				appSettings.AppSettings[i].SlotSetting = true
			}
		}
	}

	SerializeToJson(appSettings.AppSettings, outputFileName)
}

func ProcessWebConnections(inputFileName, outputFileName string) {
	configuration, err := NewConfigurationFromFile(inputFileName)
	if err != nil {
		log.Fatal(err)
	}

	connectionStrings := configuration.ConnectionStrings

	for i := range connectionStrings.ConnectionStrings {
		connectionStrings.ConnectionStrings[i].ProviderName = "SQLServer"
		for _, slotSetting := range connectionStringsSlotSettings {
			if connectionStrings.ConnectionStrings[i].Name == slotSetting {
				connectionStrings.ConnectionStrings[i].SlotSetting = true
			}
		}
	}

	SerializeToJson(connectionStrings.ConnectionStrings, outputFileName)
}

func SerializeToJson(data interface{}, outputFileName string) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	output, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer output.Close()

	w := bufio.NewWriter(output)
	w.Write(jsonBytes)
	w.Flush()

	return nil
}
