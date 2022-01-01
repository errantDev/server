package main

import (
	"fmt"
	"os"
	poker "server"
)

const dbFileName = "game.db.json"

func main() {
	fmt.Println("Let's play poker")
	fmt.Println("Type {Name} wins to record a win")
	store, closeFunc, err := poker.FileSystemStoreFromFile(dbFileName)
	game := poker.NewCLI(store, os.Stdin)
	game.PlayPoker()
}
