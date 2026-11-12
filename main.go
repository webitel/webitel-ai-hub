// Command webitel-ai-hub is the AI Hub process.
//
// It serves traffic and, as a separate subcommand, applies database migrations.
package main

import (
	"fmt"
	"os"

	"github.com/webitel/webitel-ai-hub/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
