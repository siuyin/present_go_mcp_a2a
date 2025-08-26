package main

import (
	"fmt"

	"github.com/siuyin/dflt"
)

func main() {
	prompt := dflt.EnvString("PROMPT", "Do you have the iPhone 14?")
	res := inventoryLookup(prompt)

	formatResponse(fmt.Sprintf(`user query: %s
	response from inventory lookup: %s
	Please answer the user query in a concise and professional manner`, prompt, res))
}
