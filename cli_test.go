package gocli

import (
	"os"
	"strings"
	"testing"
)

type Arguments struct {
	All  bool   `args:"alias=a,argument=all,description=All values"`
	Text string `args:"alias=t,argument=text,description=Simple text"`
	Port int    `args:"argument=port,description=Port number"`
	Host string `args:"alias=h,description=Host name"`
}

func Test_cli_Parse(t *testing.T) {
	t.Run("printing help", func(t *testing.T) {
		os.Args = strings.Split("test.exe -a --text hello --port 8080 -h localhost", " ")
		result, err := LoadArguments(&Arguments{})
		if err != nil {
			t.Errorf("Parse error %s", err)
		} else {
			if result.Host != "localhost" {
				t.Error("Error parsing string with through alias")
			}
			if result.Port != 8080 {
				t.Error("Error parsing int")
			}
			if !result.All {
				t.Error("Error parsing boolean")
			}
		}
	})
}

func Test_cli_Print(t *testing.T) {
	t.Run("printing help", func(t *testing.T) {
		os.Args = []string{"-a --text hello --port 8080 -h localhost"}
		err := PrintHelp("testing", "Collect parameters", &Arguments{})
		if err != nil {
			t.Error("Print error")
		}
	})
}
