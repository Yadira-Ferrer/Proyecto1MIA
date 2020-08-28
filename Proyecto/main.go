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

//EBR estructura para el Extended Boot Record
type EBR struct {
	PartStatus byte
	PartFit    byte
	PartStart  int64
	PartSize   int64
	PartNext   int64
	PartName   [16]byte
}

//InfoPartitions información sobre particiones
type InfoPartitions struct {
	free      bool
	primaries int
	extended  bool
}

//Mounted información sobre particiones a montar
type Mounted struct {
	Path   string
	Name   string
	Part   Partition
	Number int64
	Letter byte
}

var sliceMP []Mounted

func main() {
	var comando string = ""
	entrada := bufio.NewScanner(os.Stdin)

	sliceMP = make([]Mounted, 0)

	for {
		fmt.Printf("[Ingrese Comando]: ")
		entrada.Scan()
		comando = entrada.Text()
		//comando = strings.ToLower(comando)
		if comando == "salir" {
			break
		}
		arrayCmd := analizar(comando)
		//fmt.Println(arrayCmd)
		execCommands(arrayCmd)
	}
	printDiskInfo("/home/yadira/PruebaDisco/Disco1.dsk")
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
				//Tiene los parametros path y name
				x = x + 1 // Alcanzo el primer parametro
				cmd := CommandS{"mount", make([]Parameter, 0, 0)}
				if x < cmdsLen {
					for cmds[x].name != "comando" && cmds[x].name != "comentario" {
						cmd.Params = append(cmd.Params, Parameter{cmds[x].value, cmds[x+1].value})
						x = x + 2
						if x >= cmdsLen {
							break
						}
					}
				}
				//fmt.Println(cmd)
				MountPartition(cmd)
			case "unmount":
			case "rep":
				//Reportes tiene los atributos nombre, path(salida), ruta(entrada), id(particion montada)
				x = x + 1 // Alcanzo el primer parametro
				cmd := CommandS{"rep", make([]Parameter, 0, 4)}
				if x < cmdsLen {
					for cmds[x].name != "comando" && cmds[x].name != "comentario" {
						cmd.Params = append(cmd.Params, Parameter{cmds[x].value, cmds[x+1].value})
						x = x + 2
						if x >= cmdsLen {
							break
						}
					}
				}
				//fmt.Println(cmd)
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
				size = add
			} else {
				fmt.Println("[!] El valor del parametro 'add' no es un numero.")
			}
		case "delete":
			if strings.ToLower(prm.Value) == "fast" {
				delete = 'a'
				size = 1
			} else if strings.ToLower(prm.Value) == "full" {
				delete = 'u'
				size = 1
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
				typep = 'p'
			} else if strings.ToLower(prm.Value) == "e" {
				typep = 'e'
			} else if strings.ToLower(prm.Value) == "l" {
				typep = 'l'
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
	if size != 0 && path != "" && name != "" {
		fmt.Println("\n===== FORMATEO DE DISCO ========================================")
		if add != 0 { // Se va a agregar espacio
			fixDiskAdd(path, add, unit, name)
		} else if delete == 'a' { // Se realizará un formateo 'fast'
			fmt.Println("Path: ", path)
			fmt.Println("Name: ", name)
			fmt.Println("Type: ", string(delete))
			fixDiskDelete(path, name, delete)
		} else if delete == 'u' { // Se realizará un formateo 'full'
			fmt.Println("Path: ", path)
			fmt.Println("Name: ", name)
			fmt.Println("Type: ", string(delete))
			fixDiskDelete(path, name, delete)
		} else { // Se creará una partición
			var bname [16]byte
			var partCreated bool
			copy(bname[:], name)
			recmbr := readMBR(path)
			sizePart := getSize(size, unit)
			newpartition := Partition{PartStatus: 1, PartType: typep, PartFit: fit, PartStart: 0, PartSize: sizePart, PartName: bname}
			if (MBR{}) != recmbr {
				partCreated, recmbr = createPartition(recmbr, newpartition, path)
				if partCreated {
					writeMBR(path, recmbr)
				}
			} else {
				fmt.Println("Ha ocurrido un error al recuperar el MBR del disco.")
			}
		}
		//readMBR(path)
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
	//------------------------ MBR sizeOf = 204 bytes ------------------------

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
	var mbrSize int64 = int64(binary.Size(recMbr))
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
	//printMBR(recMbr)
	return recMbr
}

func readEBR(path string, position int64) EBR {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	// Posicion del puntero
	//currentPosition, err := file.Seek(position, 1)
	file.Seek(position, 1)
	//fmt.Println("Posicion 'Seek' ReadEbr: ", currentPosition)
	// Se declara EBR contenedor
	recEbr := EBR{}
	// Se obtiene el tamaño del EBR
	var ebrSize int64 = int64(binary.Size(recEbr))
	// Lectura los bytes determinados por ebrSize
	data := readBytes(file, ebrSize)
	// Convierte data en un buffer, necesario para decodificar binario
	buffer := bytes.NewBuffer(data)
	// Se decodifica y guarda en recMbr
	err = binary.Read(buffer, binary.BigEndian, &recEbr)
	if err != nil {
		log.Fatal("Fallo binary.Read ", err)
	}

	// Si todo sale bien se imprimirán los valores del MBR recuperado
	return recEbr
}

func writeMBR(path string, mbr MBR) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("[!] Error al abrir disco. ", err)
		return
	}
	defer file.Close()
	//Posición al inicio del archivo
	file.Seek(0, 0)
	dsk := &mbr
	//Escritura del struct MBR
	var binthree bytes.Buffer
	binary.Write(&binthree, binary.BigEndian, dsk)
	state := writeBytes(file, binthree.Bytes())
	if state {
		fmt.Println("*** MBR escrito exitosamente ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al escribir el MBR en el disco.")
	}

}

func writeByteArray(path string, position int64, size int64) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("[!] Error al abrir disco. ", err)
		return
	}
	defer file.Close()
	//Posición al inicio del archivo
	file.Seek(position, 1)
	zeroByte := make([]byte, size)
	zb := &zeroByte
	//Escritura del struct MBR
	var bin bytes.Buffer
	binary.Write(&bin, binary.BigEndian, zb)
	state := writeBytes(file, bin.Bytes())
	if state {
		fmt.Println("*** Escritura de []bytes exitosa  ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al escribir el []bytes.")
	}
}

func writeEBR(path string, ebr EBR, position int64) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("[!] Error al abrir disco. ", err)
		return
	}
	defer file.Close()
	// .Seek(position,mode) mode:
	// 0 = Inicio, 1 = Desde donde esta el puntero, 2 = Del final al inicio
	// Posición al inicio del archivo
	//currentPosition, err := file.Seek(position, 1)
	file.Seek(position, 1)
	//fmt.Println("Posicion 'Seek' WriteEbr: ", currentPosition)
	refebr := &ebr
	//Escritura del struct EBR
	var binthree bytes.Buffer
	binary.Write(&binthree, binary.BigEndian, refebr)
	state := writeBytes(file, binthree.Bytes())
	if state {
		fmt.Println("*** EBR escrito exitosamente ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al escribir EBR en el disco.")
	}
}

func writeBytes(file *os.File, bytes []byte) bool {
	_, err := file.Write(bytes)

	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func readBytes(file *os.File, number int64) []byte {
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
		if p.PartType == 'e' && p.PartStatus == 1 {
			infoParts.extended = true
		} else if p.PartType == 'p' && p.PartStatus == 1 {
			infoParts.primaries++
		} else if p.PartStatus == 0 {
			infoParts.free = true
		}
	}
	return infoParts
}

func createPartition(mbr MBR, p Partition, path string) (bool, MBR) {
	// Las particiones siempre se crean con el primer ajuste
	infparts := getPartitionsInfo(mbr.MbrPartitions)
	flgCreated := false // Toma valor verdadero si la particion se ha creado
	sizeOfMbr := int64(binary.Size(mbr))
	offset := sizeOfMbr // Desplazamiento de la pos de la particion
	switch p.PartType {
	case 'p':
		fmt.Println("[*] Se creara un particion Primaria...")
		if infparts.primaries < 3 && infparts.free {
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
							if !nameAlreadyExist(path, mbr.MbrPartitions, p.PartName) {
								p.PartStart = offset + 1
								mbr.MbrPartitions[i] = p
								flgCreated = true
								fmt.Println("*** Particion 'p' creada exitosamente ***")
								//nonSpace = true
								break
							} else {
								fmt.Println("[!] Ya existe una particion con el nombre: ", string(p.PartName[:]))
								break
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
							if !nameAlreadyExist(path, mbr.MbrPartitions, p.PartName) {
								p.PartStart = offset
								mbr.MbrPartitions[i] = p
								flgCreated = true
								fmt.Println("*** Particion 'p' creada exitosamente ***")
								//nonSpace = true
								break
							} else {
								fmt.Println("[!] Ya existe una particion con el nombre: ", string(p.PartName[:]))
								break
							}
						} else {
							fmt.Println("[!] No hay espacio disponible para crear la particion.")
							//nonSpace = true
							break
						}
					}
				} else {
					offset = cp.PartStart + cp.PartSize
				}
			}
		} else {
			fmt.Println("[!] Ya existen 3 particiones primarias...")
		}
	case 'e':
		fmt.Println("[*] Se creara un particion Extendida...")
		if !infparts.extended {
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
						if p.PartSize <= gap {
							if !nameAlreadyExist(path, mbr.MbrPartitions, p.PartName) {
								p.PartStart = offset + 1
								mbr.MbrPartitions[i] = p
								// Se crea un EBR
								ebr := EBR{PartStatus: 1, PartFit: p.PartFit, PartStart: (offset + 1), PartSize: 0, PartNext: 0, PartName: p.PartName}
								// Se escribe el EBR en el disco
								writeEBR(path, ebr, p.PartStart)
								flgCreated = true
								fmt.Println("*** Particion creada exitosamente ***")
								break
							} else {
								fmt.Println("[!] Ya existe una particion con el nombre: ", string(p.PartName[:]))
								break
							}
						} else {
							fmt.Println("[!] No hay espacio disponible para crear la particion.")
							break
						}
					} else { // No hay siguiente...
						gap := mbr.MbrSize - offset
						if p.PartSize <= gap {
							if !nameAlreadyExist(path, mbr.MbrPartitions, p.PartName) {
								p.PartStart = offset + 1
								mbr.MbrPartitions[i] = p
								// Se crea un EBR
								ebr := EBR{PartStatus: 1, PartFit: p.PartFit, PartStart: (offset + 1), PartSize: 0, PartNext: 0, PartName: p.PartName}
								// Se escribe el EBR en el disco
								writeEBR(path, ebr, p.PartStart)
								flgCreated = true
								fmt.Println("*** Particion creada exitosamente ***")
								break
							} else {
								fmt.Println("[!] Ya existe una particion con el nombre: ", string(p.PartName[:]))
								break
							}
						} else {
							fmt.Println("[!] No hay espacio disponible para crear la particion.")
							break
						}
					}
				} else {
					offset = cp.PartStart + cp.PartSize
				}
			} // Fin del For
		} else {
			fmt.Println("[!] Ya existe una partición extendida.")
		}
	case 'l':
		fmt.Println("[*] Se creara un particion Logica...")
		// Comprobar si existe una extendida
		if infparts.extended {
			var indexExt int64
			// Comprobar que el nombre no exista...
			if !nameAlreadyExist(path, mbr.MbrPartitions, p.PartName) {
				// Obtener el índice de la partición extendida
				for i, cp := range mbr.MbrPartitions {
					if cp.PartStatus == 1 && cp.PartType == 'e' {
						indexExt = int64(i)
						break
					}
				}
				// Guardar la partición extendida
				partExt := mbr.MbrPartitions[indexExt]
				// Leo el primer EBR
				firstEbr := readEBR(path, partExt.PartStart)
				auxEbr := EBR{}
				sizOfEbr := int64(binary.Size(auxEbr))
				// Desplazamiento = inicio del primer EBR + tamaño del EBR
				offset := firstEbr.PartStart + sizOfEbr
				// Se guarda en auxEBR el primer EBR
				auxEbr = firstEbr
				// Comprobación del espacio...
				endPart := partExt.PartStart + partExt.PartSize
				// Buscando espacio en el disco. Desplazamiento < Tamaño de la partición
				for offset < endPart {
					// Sino hay una partición siguiente ...
					if auxEbr.PartNext == 0 {
						// Diferencia entre el final de la particion y el desplazamiento
						gap := endPart - offset
						if p.PartSize <= gap {
							auxEbr.PartStatus = 1
							auxEbr.PartFit = p.PartFit
							auxEbr.PartSize = p.PartSize
							auxEbr.PartNext = offset + p.PartSize + 1
							auxEbr.PartName = p.PartName
							// Escribir el EBR
							writeEBR(path, auxEbr, auxEbr.PartStart)
							// Crear un nuevo EBR
							nuevoEBR := EBR{PartStatus: 0, PartFit: p.PartFit, PartStart: (offset + p.PartSize + 1), PartSize: 0, PartNext: 0, PartName: p.PartName}
							writeEBR(path, nuevoEBR, nuevoEBR.PartStart)
							fmt.Println("*** Particion L creada exitosamente ***")
							flgCreated = true
							break
						} else {
							fmt.Println("[!] No se puede crear la particion, no hay espacio.")
							flgCreated = false
							break
						}
					} else {
						posNext := auxEbr.PartNext
						offset = sizOfEbr + posNext
						auxEbr = readEBR(path, posNext)
					}
				} // Fin de For
			}
		} else {
			fmt.Println("[!] No es posible crear la particion logica, no existe una extendida.")
			flgCreated = false
		}
	}
	return flgCreated, mbr
}

func nameAlreadyExist(path string, partitions [4]Partition, name [16]byte) bool {
	for _, p := range partitions {
		if p.PartType == 'p' {
			if p.PartName == name {
				return true
			}
		} else if p.PartType == 'e' {
			if p.PartName == name {
				return true
			}
			// Revisar el nombre de las particiones logicas
			// Leer el primer EBR
			ebr := readEBR(path, p.PartStart)
			// Posicion en que finaliza la particion (sig. ebr)
			endPosition := p.PartStart + p.PartSize
			sizeOfEbr := int64(binary.Size(ebr))
			offset := sizeOfEbr + ebr.PartStart
			for offset < endPosition {
				if ebr.PartName == name {
					return true
				}
				if ebr.PartNext != 0 {
					offset = sizeOfEbr + ebr.PartNext
					ebr = readEBR(path, ebr.PartNext)
				} else {
					break
				}
			} // Finalizar ~ For
		} // Else de que no tiene asignada particion...
	}
	return false
}

func printPart(p Partition) {
	fmt.Println("   - Status: 1")
	fmt.Println("   - Type: ", string(p.PartType))
	fmt.Println("   - Fit: ", string(p.PartFit))
	fmt.Println("   - Start: ", p.PartStart)
	fmt.Println("   - Size: ", p.PartSize)
	fmt.Println("   - Name: ", string(p.PartName[:]))
}

func printEBR(e EBR) {
	fmt.Println("> EBR INFO:")
	if e.PartStatus == 0 {
		fmt.Println("   Estatus: 0")
	} else {
		fmt.Println("   Estatus: 1")
	}
	fmt.Println("   Ajuste: ", string(e.PartFit))
	fmt.Println("   Inicio: ", e.PartStart)
	fmt.Println("   Tamaño: ", e.PartSize)
	fmt.Println("   Siguiente: ", e.PartNext)
	fmt.Println("   Nombre: ", string(e.PartName[:]))
}

func printMBR(m MBR) {
	t := m.MbrTime
	fmt.Println("\n---MBR--------------------------------")
	fmt.Println("   Tamaño: ", m.MbrSize)
	fmt.Println("   Firma: ", m.MbrDiskSignature)
	fmt.Println("   F/H: ", t.Day, "/", t.Month, "/", t.Year, " ", t.Hour, ":", t.Minute, ":", t.Seconds)
	fmt.Println("--------------------------------------")
}

func printDiskInfo(path string) {
	m := readMBR(path)
	printMBR(m)
	fmt.Println("   Particiones:")
	for i, p := range m.MbrPartitions {
		if p.PartStatus == 1 {
			fmt.Println("   P[", i, "]")
			printPart(p)
			if p.PartType == 'e' {
				position := p.PartStart
				for true {
					ebr := readEBR(path, position)
					if ebr.PartStatus == 0 {
						break
					} else {
						printEBR(ebr)
						position = ebr.PartNext
					}
				}
			}
		}
	}
}

func fixDiskAdd(path string, size int64, unit byte, name string) {
	mbr := readMBR(path)
	space := getSize(size, unit)
	position := 0
	flgFound := false
	var bname [16]byte
	copy(bname[:], name)
	var p Partition
	// Encontrar la partición a la cual se le agregara/quitara espacio
	for i, cp := range mbr.MbrPartitions {
		if cp.PartName == bname && cp.PartStatus == 1 {
			flgFound = true
			position = i
			p = cp
			break
		}
	}
	if flgFound {
		// Si espacio es menor a cero se quitara espacio
		if space < 0 {
			// Si el espacio a eliminar es menor
			if space < p.PartSize {
				p.PartSize = p.PartSize + space // se supone que es un número negativo...
				mbr.MbrPartitions[position] = p
				writeMBR(path, mbr)
			} else {
				fmt.Println("[!] El espacio a quitar es mayor al tamaño de la particion.")
				return
			}
		} else { // El espacio es mayor 0, por lo que se agregará espacio
			// Si es la ultima particion...
			if position == 3 {
				gap := mbr.MbrSize - (p.PartStart + p.PartSize)
				if gap >= space {
					p.PartSize = p.PartSize + space
					mbr.MbrPartitions[position] = p
					writeMBR(path, mbr)
				} else {
					fmt.Println("[!] El espacio a agregar es mayor al disponible.")
					return
				}
			} else {
				var nextPartition Partition
				var extra int64
				var nextPos = position + 1 // Indice de la siguiente particion

				for nextPos < 5 {
					fmt.Println("Position: ", position)
					fmt.Println("1 NextPos: ", nextPos)
					if nextPos == 4 {
						extra = mbr.MbrSize - (p.PartStart + p.PartSize)
						fmt.Println("1 Extra: ", extra)
						break
					} else {
						fmt.Println("2 NextPos: ", nextPos)
						nextPartition = mbr.MbrPartitions[nextPos]
						nextPos = nextPos + 1
						if nextPartition.PartStatus == 1 {
							extra = nextPartition.PartStart - (p.PartStart + p.PartSize)
							fmt.Println("1 Extra: ", extra)
							break
						}
					}
				}
				//
				if extra >= space {
					p.PartSize = p.PartSize + space
					mbr.MbrPartitions[position] = p
					writeMBR(path, mbr)
				} else {
					fmt.Println("[!] El tamaño que se desea agregar excede al tamaño disponible en el disco.")
					return
				}
			}
			fmt.Println("** Se ha modificado el tamaño de la partición con exito. ***")
		}
	} else {
		fmt.Println("[!] No se ha encontrado la particion: ", name)
	}
}

func fixDiskDelete(path string, name string, typeDel byte) {
	mbr := readMBR(path)
	flgfound := false
	position := int64(0)
	var bname [16]byte
	copy(bname[:], name)
	// Recorrer las particiones...
	for i, p := range mbr.MbrPartitions {
		if p.PartName == bname && p.PartStatus == 1 {
			flgfound = true
			position = int64(i)
			break
		}
	}
	// Fast Delete
	if typeDel == 'a' {
		// Si la particion existe
		if flgfound {
			if mbr.MbrPartitions[position].PartType == 'p' {
				// Se marca la partición como partición inactiva...
				mbr.MbrPartitions[position].PartStatus = 0
				writeMBR(path, mbr)
				fmt.Println("*** Se ha eliminado la partición ", name, " *** ")
			} else if mbr.MbrPartitions[position].PartType == 'e' {
				// Se marca la partición como partición inactiva...
				mbr.MbrPartitions[position].PartStatus = 0
				writeMBR(path, mbr)
				fmt.Println("*** Se ha eliminado la partición ", name, " *** ")
			}
		} else {
			fmt.Println("[!] No se encontro la particion '", name, "'.")
		}
		// Full delete
	} else if typeDel == 'u' {
		// Si la particion existe...
		if flgfound {
			if mbr.MbrPartitions[position].PartType == 'p' {
				psize := mbr.MbrPartitions[position].PartSize
				pstart := mbr.MbrPartitions[position].PartStart
				// Se llena de ceros la partición...
				writeByteArray(path, pstart, psize)
				// Resetear la partición del mbr
				mbr.MbrPartitions[position] = Partition{}
				// Escribir MBR con la información actualizada
				writeMBR(path, mbr)
				fmt.Println("*** Se ha eliminado la partición ", name, " *** ")
			} else if mbr.MbrPartitions[position].PartType == 'e' {
				psize := mbr.MbrPartitions[position].PartSize
				pstart := mbr.MbrPartitions[position].PartStart
				// Se llena de ceros la partición...
				writeByteArray(path, pstart, psize)
				// Resetear la partición del mbr
				mbr.MbrPartitions[position] = Partition{}
				// Escribir MBR con la información actualizada
				writeMBR(path, mbr)
				fmt.Println("*** Se ha eliminado la partición ", name, " *** ")
				return
			}
		} else {
			fmt.Println("[!] No se encontro la particion '", name, "'.")
		}
	}
}

//MountPartition : se montarán las particoines
func MountPartition(cmd CommandS) {
	var path string = ""
	var name string = ""
	//Si no vienen parametros, se deben imprimir las particiones montadas
	fmt.Println("\n===== MONTAR PARTCION ==========================================")
	if len(cmd.Params) == 0 {
		if len(sliceMP) == 0 {
			fmt.Println("[*] No se han montando particiones...")
		} else {
			fmt.Println("Particiones montadas:")
			for _, mp := range sliceMP {
				id := "vd" + string(mp.Letter) + strconv.FormatInt(mp.Number, 10)
				fmt.Println("> id:", id, "| path:", mp.Path, "| name:", mp.Name)
			}
		}
	} else {
		for _, param := range cmd.Params {
			if strings.ToLower(param.Name) == "path" {
				path = delQuotationMark(param.Value)
			} else if strings.ToLower(param.Name) == "name" {
				name = delQuotationMark(param.Value)
			} else {
				fmt.Println("[!] Parametro para el comando 'mount' invalido...")
			}
		}
		if path != "" && name != "" {
			var bname [16]byte
			var num int64 = 1
			var letter byte = 'a'
			copy(bname[:], name)
			found := false
			flgext := false
			mbr := readMBR(path)
			if mbr != (MBR{}) {
				mount := Mounted{}
				for _, p := range mbr.MbrPartitions {
					if p.PartName == bname && p.PartStatus == 1 {
						if p.PartType == 'p' {
							mount.Part = p
							found = true
							break
						} else {
							flgext = true
							break
						}
					}
				}
				// Si se ha encontrado la partición
				if found {
					// Verificar que la partición se encuentre montada
					if !IsMounted(path, name) {
						if len(sliceMP) != 0 {
							letter, num = GetPartitionNum(path)
						}
						mount.Name = name
						mount.Path = path
						mount.Number = num
						mount.Letter = letter
						sliceMP = append(sliceMP, mount)
						fmt.Println("*** Partición Montada ***")
					} else {
						fmt.Println("[*] La partición '", name, "' ya se encuentra montada...")
					}
				} else {
					if flgext {
						fmt.Println("[!] Ha intentado montar una partición que no es primaria...")
					} else {
						fmt.Println("[!] No se ha encontrado la partición '", name, "'...")
					}
				}
			} else {
				fmt.Println("[!] Ha ocurrio un error puede que el disco no exista...")
			}
		} else {
			fmt.Println("[!] Faltan parametros obligatorios 'mount'...")
		}
	}
	fmt.Println("================================================================")
}

//GetPartitionNum retorna el número de la partición asignada para ese disco
func GetPartitionNum(path string) (byte, int64) {
	var num int64 = 1
	var letter byte = 'a'
	for _, m := range sliceMP {
		num = m.Number
		letter = m.Letter
		if m.Path == path {
			return letter, num + 1
		}
	}
	return letter + 1, 1
}

//IsMounted retorna verdadero si la partición ya esta montada
func IsMounted(path string, name string) bool {
	for _, mp := range sliceMP {
		if mp.Path == path && mp.Name == name {
			return true
		}
	}
	return false
}
