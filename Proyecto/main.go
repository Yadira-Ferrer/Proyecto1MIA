package main

import (
	"bufio"
	"fmt"
	"log"
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
	//comando = strings.ToLower(comando)
	/* if comando == "salir" {
		break
	} */
	arrayCmd := analizar(comando)
	execCommands(arrayCmd)
	//}
}

func execCommands(cmds []Token) {
	cmdsLen := len(cmds)
	for x := 0; x < cmdsLen; x++ {
		switch cmds[x].name {
		case "comando":
			switch cmds[x].value {
			case "exec":
				exec(cmds[x+2].value)
				x += 2
			case "pause":

			case "mkdisk":
			case "rmdisk":
			case "fdisk":
			case "mount":
			case "unmount":
			case "rep":
			}
		case "parametro":
		case "numero":
		case "comentario":
			fmt.Println(cmds[x].value)
		case "cadena":
		case "path":
		case "id":
		}
	}
}

func exec(path string) {
	var content string = ""
	// Eliminación de las comillas en el path
	if strings.Contains(path, "\"") {
		path = strings.Replace(path, "\"", "", -1)
		fmt.Println("El nuevo path es: ", path)
	}
	// Se abre el archivo especificado por el path
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	} else {
		defer file.Close()
		// Se lee el contenido del archivo si este ha sido abierto
		scanner := bufio.NewScanner(file)
		// Se concatena el contenido del archivo en un string
		for scanner.Scan() {
			content += scanner.Text() + "\n"
		}
		// Se reemplazan los caracteres de salto de linea '\*' definidos por el lenguaje
		content = strings.Replace(content, "\\*\n", "", -1)
		fmt.Println(content)
		// Se analiza la entrada
	}

}
