package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	var comando string = ""
	entrada := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("[Ingrese Comando]: ")
		entrada.Scan()
		comando = entrada.Text()
		if strings.ToLower(comando) == "salir" {
			break
		}
		analizar(comando)
	}
}
