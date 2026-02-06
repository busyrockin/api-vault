package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("API Vault v0.1.0")
        fmt.Println("Usage: api-vault <command>")
        os.Exit(0)
    }

    command := os.Args[1]
    
    switch command {
    case "version":
        fmt.Println("API Vault v0.1.0-dev")
    default:
        fmt.Printf("Unknown command: %s\n", command)
        os.Exit(1)
    }
}
