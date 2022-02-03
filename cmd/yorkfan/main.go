package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tjmerritt/go-webrelay"
)

// env:
//    YORKFAN_HOST=
//    YORKFAN_USER=
//    YORKFAN_PASSWORD=
// yorkfan [--host name] [--user name] [--password passwd] [--password-prompt] [low|med|medium|high]

func main() {
	host := "192.168.1.2"
	user := ""
	passwd := ""

	if env := os.Getenv("YORKFAN_HOST"); env != "" {
		host = env
	}

	if env := os.Getenv("YORKFAN_USER"); env != "" {
		user = env
	}

	if env := os.Getenv("YORKFAN_PASSWORD"); env != "" {
		passwd = env
	}

	flag.StringVar(&host, "host", host, "WebRelay host name or IP address")
	flag.StringVar(&user, "user", user, "WebRelay username")
	flag.StringVar(&passwd, "password", passwd, "WebRelay password")
	//flag.BoolVar(&prompt, "password-prompt", "", "Prompt for WebRelay password")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Printf("Usage: yorkfan low|med|high\n")
		os.Exit(1)
	}
	speed := args[0]
	if speed != "low" && speed != "med" && speed != "medium" && speed != "high" {
		fmt.Printf("Usage: yorkfan low|med|high\n")
		os.Exit(1)
	}

	//fmt.Printf("Host: %s, User: %s, Password: %s\n", host, user, passwd)

	client, err := webrelay.New(host, user, passwd)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	state, err := client.GetState()
	if err != nil {
		fmt.Printf("Error getting state: %v\n", err)
		os.Exit(1)
	}
	currentSpeed, err := checkState(state)
	if err != nil {
		fmt.Printf("Error: Fan in unknown state: %s: %v\n", currentSpeed, err)
	}
	enables, newSpeed, err := updateState(state, speed)
	if err != nil {
		fmt.Printf("Error updating fan speed: %v\n", err)
		os.Exit(1)
	}
	//fmt.Printf("Enables: %v, State: %v\n", enables, state)
	err = client.SetState(enables, state)
	if err != nil {
		fmt.Printf("Error updating fan speed: %v\n", err)
		os.Exit(1)
	}
	//checkNewState, _ := checkState(state)
	if currentSpeed == "" || currentSpeed == newSpeed {
		fmt.Printf("yorkfan: Fan speed set to %s\n", newSpeed)
		//fmt.Printf("yorkfan: Fan speed set to %s (%s)\n", newSpeed, checkNewState)
	} else {
		fmt.Printf("yorkfan: Fan speed changed from %s to %s\n", currentSpeed, newSpeed)
		//fmt.Printf("yorkfan: Fan speed changed from %s to %s (%s)\n", currentSpeed, newSpeed, checkNewState)
	}
}

func checkState(state []bool) (string, error) {
	if len(state) != 4 {
		return "", fmt.Errorf("Current state is invalid: relay count is not 4: currently %d\n", len(state))
	}
	if (state[0] && (state[1] || state[2])) || (state[1] && state[2]) {
		return "", fmt.Errorf("Current state is invalid: multiple speeds selected: %s\n", formatState(state))
	}
	if !state[0] && !state[1] && !state[2] {
		return "", fmt.Errorf("Current state is invalid: no speed selected: %s\n", formatState(state))
	}
	if state[0] {
		return "HIGH", nil
	}
	if state[1] {
		return "MED", nil
	}
	return "LOW", nil
}

func updateState(state []bool, speed string) ([]bool, string, error) {
	state[0] = false
	state[1] = false
	state[2] = false

	enables := []bool{
		true,
		true,
		true,
		false,
	}

	switch speed {
	case "low":
		state[2] = true
		return enables, "LOW", nil
	case "med", "medium":
		state[1] = true
		return enables, "MED", nil
	case "high":
		state[0] = true
		return enables, "HIGH", nil
	}

	return nil, "", fmt.Errorf("Unrecognized speed setting: %s\n", speed)
}

func formatState(state []bool) string {
	str := ""
	for i := range state {
		if i > 0 {
			str += ", "
		}
		v := "OFF"
		if state[i] {
			v = "ON"
		}
		str += fmt.Sprintf("Relay%d: %s", i+1, v)
	}
	return str
}
