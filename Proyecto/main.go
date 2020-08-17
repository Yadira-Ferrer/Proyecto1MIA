package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

//Parameter para el manejo de los parametros de los comandos
type Parameter struct {
	Name  string
	Value string
}

//CommandS para el manejo de los comandos
type CommandS struct {
	Name   string
	Params []Parameter
}

//Time para el manejo de la fecha y la hora
type Time struct {
	Day     int64
	Month   int64
	Year    int64
	Hour    int64
	Minute  int64
	Seconds int64
}

//Partition para el manejo de las particiones
type Partition struct {
	PartStatus byte
	PartType   byte
	PartFit    byte
	PartStar   int64
	PartSize   int64
	PartName   [16]byte
}

//MBR estructura para el Master Boot Record
type MBR struct {
	MbrSize          int64
	MbrTime          Time
	MbrDiskSignature int64
	Martitions       [4]Partition
}

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
				fmt.Println("Presione <ENTER> para continuar...")
				input := bufio.NewScanner(os.Stdin)
				input.Scan()
			case "mkdisk":
				cmd := CommandS{"mkdisk", make([]Parameter, 0, 4)}
				x = x + 1
				for cmds[x].name != "comando" {
					fmt.Println(">>> ", cmds[x].value, cmds[x+1].value)
					cmd.Params = append(cmd.Params, Parameter{cmds[x].value, cmds[x+1].value})
					x = x + 2
					if x >= cmdsLen {
						break
					}
				}
				mkdisk(cmd)
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
	path = delQuotationMark(path)
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

func mkdisk(comando CommandS) {
	var size int64 = 0   // Obligatorio
	var path string = "" // Obligatorio
	var name string = "" // Obligatorio
	var unit byte = 'm'  // Opcional

	plen := len(comando.Params)
	//fmt.Println("LEN: ", plen)
	for i := 0; i < plen; i++ {
		p := comando.Params[i]
		//fmt.Println(">> ", p.name, " : ", p.value)
		switch strings.ToLower(p.Name) {
		case "path":
			path = delQuotationMark(p.Value)
		case "size":
			if n, err := strconv.Atoi(p.Value); err == nil {
				size = int64(n)
			} else {
				fmt.Println("[!] El valor del parametro 'size' no es un numero.")
			}
		case "name":
			name = delQuotationMark(p.Value)
		case "unit":
			if strings.ToLower(p.Value) == "k" {
				unit = 'k'
			} else if strings.ToLower(p.Value) == "m" {
				unit = 'm'
			} else {
				fmt.Println("[!] El valor del parametro 'unit' es invalido.")
			}
		default:
			fmt.Println("[!] Error con los parametros de 'mkdisk'.")
			return
		}
	}
	// Verificación de los parametros obligatorios
	if size > 0 && path != "" && name != "" {
		fmt.Println("\n===== DISCO A CREAR ============================================")
		fmt.Println("Size: ", size)
		fmt.Println("Path: ", path)
		fmt.Println("Name: ", name)
		fmt.Println("Unit: ", string(unit))
		makeDisk(size, path, name, unit)
		readFile(path + name)
		fmt.Println("================================================================")
	} else {
		fmt.Println("[!~MKDISK] Faltan parametros obligatorios.")
	}
}

func makeDisk(size int64, path string, name string, unit byte) {
	// Creacion del MBR
	var mbrDisk MBR
	// Se obtiene la fecha y la hora
	mbrDisk.MbrTime = getCurrentTime()
	// Se genera el mbr signature
	mbrDisk.MbrDiskSignature = getSignature(mbrDisk.MbrTime)
	// Se calcula el tamaño del disco
	realSize := getSize(size, unit)
	mbrDisk.MbrSize = realSize
	// Se crea el directorio si este no existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
		fmt.Println("Se ha creado el directorio: ", path)
	}
	// Se crea el archivo binario ~ alias disco
	diskpath := path + name
	file, err := os.Create(diskpath)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	var ebit int8 = 0
	s := &ebit
	//fmt.Println(unsafe.Sizeof(ebit))
	var bin bytes.Buffer
	binary.Write(&bin, binary.BigEndian, s)
	writeBytes(file, bin.Bytes())
	// .Seek(position,mode) mode:
	// 0 = Inicio, 1 = Desde donde esta el puntero, 2 = Del final al inicio
	file.Seek(realSize-1, 0)
	var bintwo bytes.Buffer
	binary.Write(&bintwo, binary.BigEndian, s)
	writeBytes(file, bintwo.Bytes())
	//------------------------------------------------------------------------
	//-------- Acá inicia la escritrua del struct del MBR en el disco --------
	//------------------------ MBR sizeOf = 224 bytes ------------------------

	//Posición al inicio del archivo
	file.Seek(0, 0)
	dsk := &mbrDisk
	//Escritura del struct MBR
	var binthree bytes.Buffer
	binary.Write(&binthree, binary.BigEndian, dsk)
	writeBytes(file, binthree.Bytes())
}

func getCurrentTime() Time {
	ctime := time.Now()
	return Time{Day: int64(ctime.Day()), Month: int64(ctime.Month()), Year: int64(ctime.Year()), Hour: int64(ctime.Hour()), Minute: int64(ctime.Minute()), Seconds: int64(ctime.Second())}
}

func getSignature(t Time) int64 {
	return t.Year - t.Day - t.Month - t.Hour - t.Minute - t.Seconds
}

func getSize(size int64, unit byte) int64 {
	if unit == 'k' {
		return size * 1024
	} else if unit == 'm' {
		return size * 1024 * 1024
	} else {
		return size
	}
}

func readFile(path string) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Println(err)
	}
	// Se declara MBR contenedor
	recMbr := MBR{}
	// Se obtiene el tamaño del MBR
	var mbrSize int = int(unsafe.Sizeof(recMbr))
	// Lectura los bytes determinados por mbrSize
	data := readBytes(file, mbrSize)
	// Convierte data en un buffer, necesario para decodificar binario
	buffer := bytes.NewBuffer(data)
	// Se decodifica y guarda en recMbr
	err = binary.Read(buffer, binary.BigEndian, &recMbr)
	if err != nil {
		log.Fatal("Fallo binay.Read", err)
	}

	// Si todo sale bien se imprimirán los valores del MBR recuperado
	fmt.Println("\nSe recupero el siguiente MBR:")
	fmt.Println("RecMBR size: ", recMbr.MbrSize)
	fmt.Println("RecMBR signature: ", recMbr.MbrDiskSignature)
	t := recMbr.MbrTime
	fmt.Println("RecMBR fyh: ", t.Day, "/", t.Month, "/", t.Year, " ", t.Hour, ":", t.Minute, ":", t.Seconds)
	fmt.Println(t)
}

func writeBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)

	if err != nil {
		log.Println(err)
	}
}

func readBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number) //array de bytes

	_, err := file.Read(bytes) // Leido -> bytes
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func delQuotationMark(s string) string {
	if strings.Contains(s, "\"") {
		s = strings.Replace(s, "\"", "", -1)
	}
	user, _ := user.Current()
	s = strings.Replace(s, "/home/", user.HomeDir+"/", 1)
	return s
}
