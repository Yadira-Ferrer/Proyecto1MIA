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
	PartStart  int64
	PartSize   int64
	PartName   [16]byte
}

//MBR estructura para el Master Boot Record
type MBR struct {
	MbrSize          int64
	MbrTime          Time
	MbrDiskSignature int64
	MbrPartitions    [4]Partition
}

// InfoPartitions información sobre particiones
type InfoPartitions struct {
	free      bool
	primaries int
	extended  bool
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
	//fmt.Println(arrayCmd)
	execCommands(arrayCmd)
	//}
}

func execCommands(cmds []Token) {
	cmdsLen := len(cmds)
	for x := 0; x < cmdsLen; x++ {
		switch strings.ToLower(cmds[x].name) {
		case "comando":
			switch strings.ToLower(cmds[x].value) {
			case "exec":
				exec(cmds[x+2].value)
				x += 2
			case "pause":
				fmt.Println("Presione <ENTER> para continuar...")
				input := bufio.NewScanner(os.Stdin)
				input.Scan()
			case "mkdisk":
				x = x + 1
				cmd := CommandS{"mkdisk", make([]Parameter, 0, 4)}
				for cmds[x].name != "comando" && cmds[x].name != "comentario" {
					//fmt.Println(">>> ", cmds[x].value, cmds[x+1].value)
					cmd.Params = append(cmd.Params, Parameter{cmds[x].value, cmds[x+1].value})
					x = x + 2
					if x >= cmdsLen {
						break
					}
				}
				mkdisk(cmd)
			case "rmdisk":
				x = x + 1 // Alcanzo el parametro
				if cmds[x].value == "path" {
					x = x + 1 // Alcanzo el valor del parametro
					path := delQuotationMark(cmds[x].value)
					fmt.Println("\n===== SE ELIMINARÁ EL DISCO ====================================")
					fmt.Println("Path: ", path)
					e := os.Remove(path)
					if e != nil {
						log.Fatal("[!~RMDIKS] ", e)
					}
					fmt.Println("*** El disco ha sido eliminado exitosamente ***")
					fmt.Println("================================================================")
				} else {
					fmt.Println("[!~RMDISK] Error con el parametro del comando.")
				}
			case "fdisk":
				// Las particiones se crean con el <primer ajuste>
				x = x + 1 // Alcanzo el primer parametro
				cmd := CommandS{"fdisk", make([]Parameter, 0, 8)}
				for cmds[x].name != "comando" && cmds[x].name != "comentario" {
					//fmt.Println(">>> ", cmds[x].value, cmds[x+1].value)
					cmd.Params = append(cmd.Params, Parameter{cmds[x].value, cmds[x+1].value})
					x = x + 2
					if x >= cmdsLen {
						break
					}
				}
				fdisk(cmd)
			case "mount":
			case "unmount":
			case "rep":
			}
		case "parametro":
		case "numero":
		case "comentario":
			fmt.Println("\n", cmds[x].value)
		case "cadena":
		case "path":
		case "id":
		}
	}
}

/*----- Comando Exec -----------------------------------------------------*/
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
		fmt.Println("\n", content)
		// Se analiza la entrada
		arrayCmd := analizar(content)
		execCommands(arrayCmd)
	}
}

/*----- Comando Mkdisk ---------------------------------------------------*/
func mkdisk(comando CommandS) {
	var size int64 = 0   // (*) Obligatorio
	var path string = "" // (*) Obligatorio
	var name string = "" // (*) Obligatorio
	var unit byte = 'm'  // (!) Opcional

	plen := len(comando.Params)
	//fmt.Println("LEN: ", plen)
	for i := 0; i < plen; i++ {
		p := comando.Params[i]
		//fmt.Println(">> ", p.Name, " : ", p.Value)
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
		fmt.Println("================================================================")
	} else {
		fmt.Println("[!~MKDISK] Faltan parametros obligatorios.")
	}
}

/*----- Comando Fdisk ----------------------------------------------------*/
func fdisk(comando CommandS) {
	fmt.Println(comando)
	var size int64        //(*) Obligatorio
	var path string       //(*) Obligatorio
	var name string       //(*) Obligatorio
	var add int64         //(!) Opcional
	var delete byte = 'x' //(!) Opcional
	var fit byte = 'w'    //(!) Opcional
	var typep byte = 'p'  //(!) Opcional
	var unit byte = 'k'   //(!) Opcional

	for _, prm := range comando.Params {
		switch strings.ToLower(prm.Name) {
		case "size":
			if n, err := strconv.Atoi(prm.Value); err == nil {
				size = int64(n)
			} else {
				fmt.Println("[!] El valor del parametro 'size' no es un numero.")
			}
		case "path":
			path = delQuotationMark(prm.Value)
		case "name":
			name = delQuotationMark(prm.Value)
		case "add":
			if n, err := strconv.Atoi(prm.Value); err == nil {
				add = int64(n)
			} else {
				fmt.Println("[!] El valor del parametro 'add' no es un numero.")
			}
		case "delete":
			if strings.ToLower(prm.Value) == "fast" {
				delete = 'a'
			} else if strings.ToLower(prm.Value) == "full" {
				delete = 'u'
			} else {
				fmt.Println("[!] El valor del parametro 'delete' es invalido.")
			}
		case "fit":
			if strings.ToLower(prm.Value) == "bf" {
				fit = 'b'
			} else if strings.ToLower(prm.Value) == "ff" {
				fit = 'f'
			} else if strings.ToLower(prm.Value) == "wf" {
				fit = 'w'
			} else {
				fmt.Println("[!] El valor del parametro 'fit' es invalido.")
			}
		case "type":
			if strings.ToLower(prm.Value) == "p" {
				fit = 'p'
			} else if strings.ToLower(prm.Value) == "e" {
				fit = 'e'
			} else if strings.ToLower(prm.Value) == "l" {
				fit = 'l'
			} else {
				fmt.Println("[!] El valor del parametro 'type' es invalido.")
			}
		case "unit":
			if strings.ToLower(prm.Value) == "b" {
				unit = 'b'
			} else if strings.ToLower(prm.Value) == "k" {
				unit = 'k'
			} else if strings.ToLower(prm.Value) == "m" {
				unit = 'm'
			} else {
				fmt.Println("[!] El valor del parametro 'unit' es invalido.")
			}
		default:
			fmt.Println("[!] Error con los parametros de 'fdisk'.")
			return
		}
	}
	// Verificación de los parametros obligatorios
	if size > 0 && path != "" && name != "" {
		fmt.Println("\n===== FORMATEO DE DISCO ========================================")
		if add > 0 { // Se va a agregar espacio

		} else if delete == 'a' { // Se realizará un formateo 'fast'

		} else if delete == 'u' { // Se realizará un formateo 'full'

		} else { // Se creará una partición
			var bname [16]byte
			copy(bname[:], name)
			recmbr := readMBR(path)
			sizePart := getSize(size, unit)
			newpartition := Partition{PartStatus: 1, PartType: typep, PartFit: fit, PartStart: 0, PartSize: sizePart, PartName: bname}
			if (MBR{}) != recmbr {
				recmbr = createPartition(recmbr, newpartition)
				for i, p := range recmbr.MbrPartitions {
					if p.PartStatus == 1 {
						fmt.Println("Particion [", i, "]")
						fmt.Println("   Status: 1")
						fmt.Println("   Type: ", string(p.PartType))
						fmt.Println("   Fit: ", string(p.PartFit))
						fmt.Println("   Start: ", p.PartStart)
						fmt.Println("   Size: ", p.PartSize)
						fmt.Println("   Name: ", string(p.PartName[:]))
					}
				}
			} else {
				fmt.Println("Ha ocurrido un error al recuperar el MBR del disco.")
			}
		}
		fmt.Println("================================================================")
	} else {
		fmt.Println("[!~FDISK] Faltan parametros obligatorios.")
	}
}

/*----- Función que crea el disco ----------------------------------------*/
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
	fmt.Println("*** Disco creado exitosamente ***")
}

/*----- Función que obtiene la hora y la fecha actual --------------------*/
func getCurrentTime() Time {
	ctime := time.Now()
	return Time{Day: int64(ctime.Day()), Month: int64(ctime.Month()), Year: int64(ctime.Year()), Hour: int64(ctime.Hour()), Minute: int64(ctime.Minute()), Seconds: int64(ctime.Second())}
}

/*----- Función que genera 'firma' del disco -----------------------------*/
func getSignature(t Time) int64 {
	return t.Year - t.Day - t.Month - t.Hour - t.Minute - t.Seconds
}

/*----- Función que obtiene el tamaño en bytes del disco -----------------*/
func getSize(size int64, unit byte) int64 {
	if unit == 'k' {
		return size * 1024
	} else if unit == 'm' {
		return size * 1024 * 1024
	} else {
		return size
	}
}

func readMBR(path string) MBR {
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
	t := recMbr.MbrTime
	fmt.Println("\n---MBR--------------------------------")
	fmt.Println("   Tamaño: ", recMbr.MbrSize)
	fmt.Println("   Firma: ", recMbr.MbrDiskSignature)
	fmt.Println("   F/H: ", t.Day, "/", t.Month, "/", t.Year, " ", t.Hour, ":", t.Minute, ":", t.Seconds)
	fmt.Println("--------------------------------------")
	return recMbr
}

func writeMBR(path string, mbr MBR) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Println(err)
		return
	}
	//Posición al inicio del archivo
	file.Seek(0, 0)
	dsk := &mbr
	//Escritura del struct MBR
	var binthree bytes.Buffer
	binary.Write(&binthree, binary.BigEndian, dsk)
	writeBytes(file, binthree.Bytes())
	fmt.Println("*** MBR escrito exitosamente ***")
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

func getPartitionsInfo(partitions [4]Partition) InfoPartitions {
	infoParts := InfoPartitions{}
	for _, p := range partitions {
		if p.PartType == 'e' {
			infoParts.extended = true
		} else if p.PartType == 'p' {
			infoParts.primaries++
		} else {
			infoParts.free = true
		}
	}
	return infoParts
}

func createPartition(mbr MBR, p Partition) MBR {
	// Las particiones siempre se crean con el primer ajuste
	sizeOfMbr := int64(unsafe.Sizeof(mbr))
	//spaceFree := false  // Bandera para verficar si la partición encaja en el espacio
	offset := sizeOfMbr // Desplazamiento de la pos de la particion
	//nonSpace := false   // Bandera para verificar sino hay espacio
	/* logicp := false
	positionl := 0 */
	switch p.PartType {
	case 'p':
		for i, cp := range mbr.MbrPartitions {
			if cp.PartStatus == 0 {
				flgPartition := false
				var nextPartPos int64 = 0
				for x := i + 1; x < 4; x++ {
					if mbr.MbrPartitions[x].PartStatus != 0 {
						flgPartition = true
						nextPartPos = mbr.MbrPartitions[x].PartStart
						break
					}
				}
				// Verficar las siguientes...
				if flgPartition {
					gap := nextPartPos - offset // inicio de la particion siguiente - desplazamiento
					// El tamaño de la partición a crear es menor o igual al espacio libre ¿?
					if p.PartSize <= gap {
						if !nameAlreadyExist(mbr.MbrPartitions, p.PartName) {
							p.PartStart = offset + 1
							mbr.MbrPartitions[i] = p
							fmt.Println("*** Particion creada exitosamente ***")
							//nonSpace = true
							break
						} else {
							fmt.Println("[!] Ya existe una particion con el nombre: ", string(p.PartName[:]))
						}
					} else {
						if i == 3 {
							fmt.Println("[!] No hay espacio disponible para crear la particion.")
							//nonSpace = true
							break
						}
					}
				} else {
					gap := mbr.MbrSize - offset
					//fmt.Println("smbr - offeset = ", gap)
					if p.PartSize <= gap {
						if !nameAlreadyExist(mbr.MbrPartitions, p.PartName) {
							p.PartStart = offset
							mbr.MbrPartitions[i] = p
							fmt.Println("*** Particion creada exitosamente ***")
							//nonSpace = true
							break
						} else {
							fmt.Println("[!] Ya existe una particion con el nombre: ", string(p.PartName[:]))
						}
					} else {
						fmt.Println("[!] No hay espacio disponible para crear la particion.")
						//nonSpace = true
					}
				}
			} else {
				offset = mbr.MbrPartitions[i].PartStart + mbr.MbrPartitions[i].PartSize
				/* 				if i == 3 {
					spaceFree = true
					nonSpace = true
				} */
			}
		}
	case 'e':
	case 'l':
	}
	return mbr
}

func nameAlreadyExist(partitions [4]Partition, name [16]byte) bool {
	for _, p := range partitions {
		if p.PartName == name {
			return true
		}
	}
	return false
}
