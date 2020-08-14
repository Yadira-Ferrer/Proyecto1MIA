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

	//for {
	fmt.Printf("[Ingrese Comando]: ")
	entrada.Scan()
	comando = entrada.Text()
	comando = strings.ToLower(comando)
	/* if comando == "salir" {
		break
	} */
	analizar(comando)
	//}
}
