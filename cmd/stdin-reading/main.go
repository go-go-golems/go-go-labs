package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	// Reading piped data (if any)
	pipedData, _ := ioutil.ReadAll(os.Stdin)
	fmt.Printf("Received piped data: %s\n", pipedData)

	// Now, we want to read user input from the terminal

	// Open the terminal for reading
	tty, err := os.Open("/dev/tty")
	if err != nil {
		fmt.Println("Failed to open terminal:", err)
		return
	}
	defer func(tty *os.File) {
		_ = tty.Close()
	}(tty)

	reader := bufio.NewReader(tty)

	for {
		// Ask the user if they want to continue
		fmt.Print("Do you want to continue in chat? [y/n]: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read input:", err)
			return
		}

		switch input[0] {
		case 'y', 'Y':
			chat()
		case 'n', 'N':
			return
		default:
			fmt.Println("Invalid input. Please enter 'y' or 'n'.")
		}
	}
}

func chat() {
	fmt.Println("You're now in chat! (For the sake of this example, we'll simply return to the main loop after this message.)")
}
