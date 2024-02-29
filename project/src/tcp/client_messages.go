package tcp

import (
	"fmt"
	"net"
	"strings"
)

func SendMessage(conn net.Conn, message string) error {
	// Ensure the message ends with a newline character, which may be needed depending on the server's reading logic.
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Printf("Failed to send message: %s\n", err)
		return err
	}
	return nil
}
