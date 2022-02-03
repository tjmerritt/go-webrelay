package main

import (
	"fmt"
	"github.com/tjmerritt/go-webrelay"
)

func main()  {
	client, err := webrelay.New("192.168.120.15", "", "")
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
	}
	client.SetUserAgent("curl/7.77.0")
	state, err := client.GetState()
	if err != nil {
		fmt.Printf("Error getting state: %v\n", err)
	}
	fmt.Printf("%s: [%s]\n", client.ModelName(), formatState(state))
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
		str += fmt.Sprintf("Relay%d: %s", i + 1, v)
	}
	return str
}
