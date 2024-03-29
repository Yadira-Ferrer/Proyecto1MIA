package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// SizeStructures : almacena el tamaño de cada estructura del sistema LWH
type SizeStructures struct {
	sizeSB    int64
	sizeAV    int64
	sizeDDir  int64
	sizeInodo int64
	sizeBD    int64
	sizeLog   int64
}

//Mkfs : valida el comando
func Mkfs(cmd CommandS) {
	var idp string
	var typef byte = 'a'
	var add int64
	var unit byte = 'k'
	for _, prm := range cmd.Params {
		switch strings.ToLower(prm.Name) {
		case "id":
			idp = delQuotationMark(prm.Value)
		case "type":
			if strings.ToLower(prm.Value) == "fast" {
				typef = 'a'
			} else if strings.ToLower(prm.Value) == "fast" {
				typef = 'e'
			} else {
				fmt.Println("[!~MKFS] Valor invalido para el parametro 'type'...")
			}
		case "add":
			if n, err := strconv.Atoi(prm.Value); err == nil {
				add = int64(n)
			} else {
				fmt.Println("[!~MKFS] El valor del parametro 'add' no es un numero.")
			}
		case "unit":
			if strings.ToLower(prm.Value) == "b" {
				unit = 'b'
			} else if strings.ToLower(prm.Value) == "m" {
				unit = 'm'
			} else {
				unit = 'k'
			}
		}
	}
	fmt.Println("\n===== CREACION DEL SISTEMA DE ARCHIVOS =========================")
	fmt.Println("-id:", idp)
	fmt.Println("-type:", string(typef))
	fmt.Println("-add:", add)
	fmt.Println("-unit:", string(unit))
	if idp != "" {
		MakeFileSystem(idp, typef)
	} else {
		fmt.Println("[!] Hace falta el parametro obligatorio 'id'...")
	}
	fmt.Println("================================================================")
}

//MakeFileSystem : crea el sistema de archivos LWH
func MakeFileSystem(idp string, typef byte) {
	var flgfound bool = false
	mpartition := Mounted{}
	for _, mp := range sliceMP {
		idm := "vd" + string(mp.Letter) + strconv.FormatInt(mp.Number, 10)
		if idp == idm {
			flgfound = true
			mpartition = mp
			break
		}
	}
	if flgfound {
		var bname [16]byte
		partition := mpartition.Part
		// Se realiza el formateo de la partición
		if typef == 'u' {
			writeByteArray(mpartition.Path, partition.PartStart, partition.PartSize)
		}
		// Current Position Disk Partition
		var cpd int64
		// Se obtiene el tamaño de las estructuras y la cantidad (#Estructuras)
		sStrc, cStrc := GetNumberOfStructures(partition.PartSize)
		// Se creará el Super Boot
		newSB := SuperBoot{}
		// Nombre HD
		copy(bname[:], mpartition.Name)
		newSB.NombreHd = bname
		newSB.FechaCreacion = getCurrentTime()
		newSB.FechaUltimoMontaje = mpartition.TMount
		newSB.ConteoMontajes = 1
		// Cantidad de estructuras creadas
		newSB.CantArbolVirtual = 1
		newSB.CantDetalleDirectorio = 1
		newSB.CantidadInodos = 1
		newSB.CantidadBloques = 2
		// Cantidad de estructuras ocupadas...
		newSB.ArbolesVirtualesLibres = cStrc - 1
		newSB.DetallesDirectorioLibres = cStrc - 1
		newSB.InodosLibres = (cStrc * 5) - 1
		newSB.BloquesLibres = (cStrc * 20) - 2 // Por los dos bloques del archivo user.txt
		// Inicio BMap AVD = Inicio_Particion + SizeSB
		cpd = partition.PartStart + sStrc.sizeSB
		newSB.AptBmapArbolDirectorio = cpd
		// Inicio AVD = Inicio BitMap AVD + #Estructuras
		cpd = cpd + cStrc
		newSB.AptArbolDirectorio = cpd
		// Inicio BMap DDir = Inicio AVD + (sizeAVD*#Estructuras)
		cpd = cpd + (sStrc.sizeAV * cStrc)
		newSB.AptBmapDetalleDirectorio = cpd
		// Inicio DDir = Inicio BMap DDir + #Estructuras
		cpd = cpd + cStrc
		newSB.AptDetalleDirectorio = cpd
		// Inicio BMap Inodo = Inicio DDir + (sizeDDir * #Estructuras)
		cpd = cpd + (sStrc.sizeDDir * cStrc)
		newSB.AptBmapTablaInodo = cpd
		// Inicio Inodos = Inicio BMap Inodo + (5 * sizeInodo)
		cpd = cpd + (5 * cStrc)
		newSB.AptTablaInodo = cpd
		// Inicio BMap Bloque = Inicio Inodos + (5 * sizeInodo * #Estructuras)
		cpd = cpd + (5 * sStrc.sizeInodo * cStrc)
		newSB.AptBmapBloques = cpd
		// Inicio Bloque = Inicio Inodo + (20 * #Estructuras)
		cpd = cpd + (20 * cStrc)
		newSB.AptBloques = cpd
		// Inicio Bitacora (Log) = Inicio Bloque + (20 * sizeBloque * #Estructuras)
		cpd = cpd + (20 * sStrc.sizeBD * cStrc)
		newSB.AptLog = cpd
		// Inicio Copia SB = Inicio Bitacora + (sizeLog * #Estructuras)
		cpd = cpd + (sStrc.sizeLog * cStrc)
		//--- Se guarda el tamaño de las estructuras ------------------------------------
		newSB.TamStrcArbolDirectorio = sStrc.sizeAV
		newSB.TamStrcDetalleDirectorio = sStrc.sizeDDir
		newSB.TamStrcInodo = sStrc.sizeInodo
		newSB.TamStrcBloque = sStrc.sizeBD
		//--- Se guarda el primer bit vacio del bitmap de cada estructura ---------------
		newSB.PrimerBitLibreArbolDir = 2
		newSB.PrimerBitLibreDetalleDir = 2
		newSB.PrimerBitLibreTablaInodo = 2
		newSB.PrimerBitLibreBloques = 3
		//--- Numero Magico -------------------------------------------------------------
		newSB.NumeroMagico = 201503442
		//--- Escribir SB en Disco ------------------------------------------------------
		WriteSuperBoot(mpartition.Path, newSB, partition.PartStart)
		//--- Escritura de la Copia de SB -----------------------------------------------
		WriteSuperBoot(mpartition.Path, newSB, cpd)
		//--- (1) Crear un AVD : root "/" -----------------------------------------------
		avdRoot := ArbolVirtualDir{}
		avdRoot.FechaCreacion = getCurrentTime()
		copy(avdRoot.NombreDirectorio[:], "/")
		copy(avdRoot.AvdPropietario[:], "root")
		copy(avdRoot.AvdGID[:], "root")
		avdRoot.AvdPermisos = 777
		avdRoot.AptDetalleDirectorio = 1
		WriteAVD(mpartition.Path, avdRoot, newSB.AptArbolDirectorio)
		//--- (2) Crear un Detalle de Directorio ----------------------------------------
		detalleDir := DetalleDirectorio{}
		archivoInf := InfoArchivo{}
		archivoInf.FechaCreacion = getCurrentTime()
		archivoInf.FechaModifiacion = getCurrentTime()
		copy(archivoInf.FileName[:], "user.txt")
		archivoInf.ApInodo = 1
		detalleDir.InfoFile[0] = archivoInf
		WriteDetalleDir(mpartition.Path, detalleDir, newSB.AptDetalleDirectorio)
		//--- (3) Crear una Tabla de Inodo ----------------------------------------------
		strAux := "1,G,root\n1,U,root,201503442\n"
		tbInodo := TablaInodo{}
		tbInodo.NumeroInodo = 1 // Primer Inodo creado
		tbInodo.SizeArchivo = int64(len(strAux))
		tbInodo.CantBloquesAsignados = 2
		tbInodo.AptBloques[0] = int64(1)
		tbInodo.AptBloques[1] = int64(2)
		copy(tbInodo.IDPropietario[:], "root")
		copy(tbInodo.IDUGrupo[:], "root")
		tbInodo.IPermisos = 777
		WriteTInodo(mpartition.Path, tbInodo, newSB.AptTablaInodo)
		//--- (4) Creación de los Bloques de datos --------------------------------------
		bloque1 := BloqueDeDatos{}
		copy(bloque1.Data[:], strAux[0:25])
		WriteBloqueD(mpartition.Path, bloque1, newSB.AptBloques)
		bloque2 := BloqueDeDatos{}
		copy(bloque2.Data[:], strAux[25:len(strAux)])
		WriteBloqueD(mpartition.Path, bloque2, newSB.AptBloques+newSB.TamStrcBloque)
		//--- (5) Escribir en BitMap ----------------------------------------------------
		auxBytes := []byte{1}
		WriteBitMap(mpartition.Path, auxBytes, newSB.AptBmapArbolDirectorio)
		WriteBitMap(mpartition.Path, auxBytes, newSB.AptBmapDetalleDirectorio)
		WriteBitMap(mpartition.Path, auxBytes, newSB.AptBmapTablaInodo)
		auxBytes = append(auxBytes, 1)
		WriteBitMap(mpartition.Path, auxBytes, newSB.AptBmapBloques)
	} else {
		fmt.Println("[!] La particion", idp, " no se encuentra montada...")
	}
}

//GetNumberOfStructures : obtiene la cantidad de estructuras segun la formula...
func GetNumberOfStructures(sizePartition int64) (SizeStructures, int64) {
	sos := SizeStructures{}
	sos.sizeSB = int64(binary.Size(SuperBoot{}))
	sos.sizeAV = int64(binary.Size(ArbolVirtualDir{}))
	sos.sizeDDir = int64(binary.Size(DetalleDirectorio{}))
	sos.sizeInodo = int64(binary.Size(TablaInodo{}))
	sos.sizeBD = int64(binary.Size(BloqueDeDatos{}))
	sos.sizeLog = int64(binary.Size(Log{}))
	// Referencia
	/* fmt.Println("Tamaño Particion:", sizePartition)
	fmt.Println("Tamaño SB:", sos.sizeSB)
	fmt.Println("Tamaño AVD:", sos.sizeAV)
	fmt.Println("Tamaño Detalle Dir:", sos.sizeDDir)
	fmt.Println("Tamaño Inodo:", sos.sizeInodo)
	fmt.Println("Tamaño Bloque de Datos:", sos.sizeBD)
	fmt.Println("Tamaño Bitacora (Log):", sos.sizeLog) */
	// Calculo del numero de estructuras
	numStructures := int64((sizePartition - (2 * sos.sizeSB)) / (27 + sos.sizeAV + sos.sizeDDir + (5*sos.sizeInodo + (20 * sos.sizeBD) + sos.sizeLog)))
	//fmt.Println("Cantidad de estructuras:", numStructures)
	return sos, numStructures
}

// WriteSuperBoot : escribe el super boot en la particion
func WriteSuperBoot(path string, sboot SuperBoot, position int64) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("[!] Error al abrir disco. ", err)
		return
	}
	defer file.Close()
	//Posición donde se inicia a escribir
	file.Seek(position, 1)
	sb := &sboot
	//Escritura del struct MBR
	var bin bytes.Buffer
	binary.Write(&bin, binary.BigEndian, sb)
	state := writeBytes(file, bin.Bytes())
	if state {
		//fmt.Println("*** Se ha escrito SuperBoot  ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al escribir el SuperBoot.")
	}
}

// ReadSuperBoot : recupera el superboot de la particion
func ReadSuperBoot(path string, position int64) SuperBoot {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	// Posicion del puntero
	//currentPosition, err := file.Seek(position, 1)
	file.Seek(position, 1)
	//fmt.Println("Posicion 'Seek' ReadEbr: ", currentPosition)
	// Se declara SB contenedor
	recSB := SuperBoot{}
	// Se obtiene el tamaño del EBR
	var sizeSB int64 = int64(binary.Size(recSB))
	// Lectura los bytes determinados por ebrSize
	data := readBytes(file, sizeSB)
	// Convierte data en un buffer, necesario para decodificar binario
	buffer := bytes.NewBuffer(data)
	// Se decodifica y guarda en recMbr
	err = binary.Read(buffer, binary.BigEndian, &recSB)
	if err != nil {
		log.Fatal("[!] Fallo la lectura del Super Boot", err)
	} else {
		//fmt.Println("[ SuperBoot leido exitosamente ]")
	}
	return recSB
}

// PrintSuperBoot : imprime el contenido del SB
func PrintSuperBoot(path string, position int64) {
	sb := ReadSuperBoot(path, position)
	fmt.Println("Nombre HD", string(sb.NombreHd[:]))
	fmt.Println("ArbolesVirtualesLibres", sb.ArbolesVirtualesLibres)
	fmt.Println("DetallesDirectorioLibres", sb.DetallesDirectorioLibres)
	fmt.Println("InodosLibres", sb.InodosLibres)
	fmt.Println("ArbolesVirtualesLibres", sb.BloquesLibres)
}

//--- LOGIN & LOGOUT --------------------------------------------------------------------

//--- MAKE DIRECTORY --------------------------------------------------------------------

// Mkdir : valida los parametros del comando MakeDir
func Mkdir(cmd CommandS) {
	isP := false
	idp := ""
	dirs := ""
	cm := Mounted{}
	for _, param := range cmd.Params {
		switch strings.ToLower(param.Name) {
		case "id":
			idp = delQuotationMark(param.Value)
		case "path":
			if strings.Contains(param.Value, "\"") {
				dirs = strings.Replace(param.Value, "\"", "", -1)
				continue
			}
			dirs = param.Value
		case "p":
			isP = true
		}
	}
	// Creación de Directorios
	fmt.Println("\n===== CREAR DIRECTORIOS ========================================")
	fmt.Println("- Id Particion:", idp)
	fmt.Println("- Dirs:", dirs)
	fmt.Println("- P:", isP)
	// Obtener la particion
	flgfound := false
	for _, mp := range sliceMP {
		idm := "vd" + string(mp.Letter) + strconv.FormatInt(mp.Number, 10)
		if idp == idm {
			flgfound = true
			cm = mp
			break
		}
	}
	// Si la partición se encuentra montada
	if flgfound {
		MakeDirs(cm, dirs, isP)
	} else {
		fmt.Println("[!] La particion", idp, " no se encuentra montada...")
	}
	fmt.Println("================================================================")
}

//MakeDirs : Crea los direcotorios
func MakeDirs(cm Mounted, dirs string, isP bool) int {
	sb := ReadSuperBoot(cm.Path, cm.Part.PartStart)
	//fmt.Println(cm.Path)
	// Se elimina el primer '/'
	sliceDirs := GetDirsNames(dirs[1:len(dirs)])
	// Obtener el directorio "/"
	cDir := ReadAVD(cm.Path, sb.AptArbolDirectorio)
	// Empezar a buscar si existen los directorios
	auxDir := cDir
	posAuxDir := sb.AptArbolDirectorio
	for x, name := range sliceDirs {
		posAuxDir, auxDir = GetSubDir(name, cm.Path, sb, auxDir, posAuxDir)
		fmt.Println("> Current Dir Name:", string(auxDir.NombreDirectorio[:]))
		var bname [16]byte
		copy(bname[:], name)
		// Si el directorio actual es diferente al directorio que busco -> NO EXISTE...
		fmt.Println("BName:", string(bname[:]), "NameDir:", string(auxDir.NombreDirectorio[:]))
		if bname == auxDir.NombreDirectorio {
			// Me quedo solo con las carpetas que hacen falta crear...
			if x+1 < len(sliceDirs) {
				sliceDirs = sliceDirs[(x + 1):len(sliceDirs)]
				continue
			} else {
				fmt.Println("[*] El directorio '", dirs, "', ya existe...")
				return 2
			}
		}
		break
	}
	// Corroborar Directorio Obtenido...
	fmt.Println("PosAuxDir: ", posAuxDir)
	fmt.Println("Directorio: ", string(auxDir.NombreDirectorio[:]))
	// Verificar si será necesario crear un aptIndirecto
	flgAptInd, indxApt := createAptInd(auxDir.AptArregloSubDir)
	if flgAptInd {
		// Se agrega al slice de directorios a crear, el directorio indirecto
		nameDir := GetString(auxDir.NombreDirectorio)
		sliceDirs = append([]string{nameDir}, sliceDirs...)
	}
	fmt.Println("Dirs a Crear:", sliceDirs)
	// Recorrer el BitMap hasta encontrar la cantidad de espacios continuos libres
	cantIndx := int64(len(sliceDirs))
	posBmap := sb.AptBmapDetalleDirectorio
	sizeBmap := sb.AptDetalleDirectorio - posBmap
	indxBm := GetBmPositions(cantIndx, cm.Path, sizeBmap, posBmap, sb.PrimerBitLibreArbolDir)
	fmt.Println("IndxBmp:", indxBm)
	fbit := indxBm[0]
	lbit := indxBm[cantIndx-1]
	//fmt.Println("FB:", fbit, "LB:", lbit, "IndexApt:", indxApt, "Flag-Ind:", flgAptInd, "CantIndx:", cantIndx)
	// ==============================================================================
	// Si viene el parametro 'p', se crean carpetas recursivas
	if isP {
		// Si es necesario crear un aptIndirecto
		if flgAptInd {
			// Se actualiza el AptIndirecto de la carpeta Actual
			auxDir.AptIndirecto = indxBm[0]
			// Se escribe el Directorio Padre
			WriteAVD(cm.Path, auxDir, posAuxDir)
			// Creación del directorio Indirecto
			dirIndirecto := ArbolVirtualDir{}
			dirIndirecto.FechaCreacion = getCurrentTime()
			dirIndirecto.NombreDirectorio = auxDir.NombreDirectorio
			dirIndirecto.AptArregloSubDir[0] = indxBm[1]
			copy(dirIndirecto.AvdPropietario[:], "root")
			copy(dirIndirecto.AvdGID[:], "root")
			dirIndirecto.AvdPermisos = 664
			// Actualizo AuxDir y posAuxDir con el nuevo AVD Indirecto
			position := sb.AptArbolDirectorio + (sb.TamStrcArbolDirectorio * (indxBm[0] - 1))
			auxDir = dirIndirecto
			// Se escribe el directorio Indirecto
			WriteAVD(cm.Path, dirIndirecto, position)
			sliceDirs = sliceDirs[1:len(sliceDirs)]
			indxBm = indxBm[1:len(indxBm)]
		} else {
			// Se actualiza el Apuntador de la carpeta padre hacia la carpeta a crear...
			auxDir.AptArregloSubDir[indxApt] = indxBm[0]
			// Se escribe el Directorio Padre
			WriteAVD(cm.Path, auxDir, posAuxDir)
		}
		// Creacion de los directorios especificados...
		for i, cdn := range sliceDirs {
			newDir := ArbolVirtualDir{}
			newDir.FechaCreacion = getCurrentTime()
			if (i + 1) < len(indxBm) {
				newDir.AptArregloSubDir[0] = indxBm[i+1]
			}
			copy(newDir.NombreDirectorio[:], cdn)
			copy(newDir.AvdPropietario[:], "root")
			copy(newDir.AvdGID[:], "root")
			newDir.AvdPermisos = 664
			// Se escribe el directorio Indirecto
			position := sb.AptArbolDirectorio + (sb.TamStrcArbolDirectorio * (indxBm[i] - 1))
			WriteAVD(cm.Path, newDir, position)
		}
	} else {
		if len(sliceDirs) == 1 && flgAptInd == false {
			// Se actualiza el Apuntador de la carpeta padre hacia la carpeta a crear...
			auxDir.AptArregloSubDir[indxApt] = indxBm[0]
			// Se escribe el Directorio Padre
			WriteAVD(cm.Path, auxDir, posAuxDir)
			// Se crea el directorio hijo
			newDir := ArbolVirtualDir{}
			newDir.FechaCreacion = getCurrentTime()
			copy(newDir.NombreDirectorio[:], sliceDirs[0])
			copy(newDir.AvdPropietario[:], "root")
			copy(newDir.AvdGID[:], "root")
			newDir.AvdPermisos = 664
			// Calcular la posicion de escritura
			position := sb.AptArbolDirectorio + (sb.TamStrcArbolDirectorio * (indxBm[0] - 1))
			// Escribir el Directorio en el disco
			WriteAVD(cm.Path, newDir, position)
			// En caso se deba crear un apt Indirecto y el directorio solicitado...
		} else if len(sliceDirs) == 2 && flgAptInd == true {
			// Se actualiza el AptIndirecto de la carpeta Actual
			auxDir.AptIndirecto = indxBm[0]
			// Se escribe el Directorio Padre
			WriteAVD(cm.Path, auxDir, posAuxDir)
			// Se crear el AVD Indirecto
			dirIndirecto := ArbolVirtualDir{}
			dirIndirecto.FechaCreacion = getCurrentTime()
			dirIndirecto.NombreDirectorio = auxDir.NombreDirectorio
			dirIndirecto.AptArregloSubDir[0] = indxBm[1]
			copy(dirIndirecto.AvdPropietario[:], "root")
			copy(dirIndirecto.AvdGID[:], "root")
			dirIndirecto.AvdPermisos = 664
			// Se escribe el directorio Indirecto
			position := sb.AptArbolDirectorio + (sb.TamStrcArbolDirectorio * (indxBm[0] - 1))
			WriteAVD(cm.Path, dirIndirecto, position)
			// Se crea el directorio especificado
			newDir := ArbolVirtualDir{}
			newDir.FechaCreacion = getCurrentTime()
			copy(newDir.NombreDirectorio[:], sliceDirs[1])
			copy(newDir.AvdPropietario[:], "root")
			copy(newDir.AvdGID[:], "root")
			newDir.AvdPermisos = 664
			// Se escribe el directorio Especificado
			position = sb.AptArbolDirectorio + (sb.TamStrcArbolDirectorio * (indxBm[1] - 1))
			WriteAVD(cm.Path, newDir, position)
		} else {
			fmt.Println("[!] Se requiere crear mas de un directorio\n    pero no fue definido el parametro '-p'...")
			return 0
		}
	}
	// Escritura del BitMap de Directorios
	bArray := make([]byte, 0)
	for int64(len(bArray)) < cantIndx {
		bArray = append(bArray, 1)
	}
	WriteBitMap(cm.Path, bArray, sb.AptBmapArbolDirectorio+(fbit-1))
	// Actualización de SB
	sb.CantArbolVirtual = sb.CantArbolVirtual + cantIndx
	sb.ArbolesVirtualesLibres = sb.ArbolesVirtualesLibres - cantIndx
	sb.PrimerBitLibreArbolDir = lbit + 1
	WriteSuperBoot(cm.Path, sb, cm.Part.PartStart)
	return 1
}

// GetDirsNames : obtiene el nombre de las carpetas a ser creadas...
func GetDirsNames(dir string) []string {
	namesDir := strings.Split(dir, "/")
	return namesDir
}

// GetSubDir : obtiene el subdirectorio si este existe...
func GetSubDir(name string, pathdsk string, sb SuperBoot, currentDir ArbolVirtualDir, position int64) (int64, ArbolVirtualDir) {
	aptSubDirs := currentDir.AptArregloSubDir
	for _, apt := range aptSubDirs {
		if apt > 0 {
			// posicion = Inicio AVD + (sizeAVD * numero_de_estructura)
			posCurrent := sb.AptArbolDirectorio + (sb.TamStrcArbolDirectorio * (apt - 1))
			auxSubDir := ReadAVD(pathdsk, posCurrent)
			var bname [16]byte
			copy(bname[:], name)
			//fmt.Println("> NextAVD:", position, " (", apt, ")")
			if bname == auxSubDir.NombreDirectorio {
				return posCurrent, auxSubDir
			}
		}
	}
	// Sino se encontro en los subdirectorios, se busca en el indirecto...
	// Si existe un apuntador indirecto...
	aptInd := currentDir.AptIndirecto
	if aptInd > 0 {
		posCurrent := sb.AptArbolDirectorio + (sb.TamStrcArbolDirectorio * (aptInd - 1))
		auxSubDir := ReadAVD(pathdsk, position)
		//fmt.Println("> NextAVD:", position, " (", aptInd, ")")
		return GetSubDir(name, pathdsk, sb, auxSubDir, posCurrent)
	}
	return position, currentDir
}

func createAptInd(aptDir [6]int64) (bool, int64) {
	for i, apt := range aptDir {
		if apt == 0 {
			return false, int64(i)
		}
	}
	return true, -1
}

// GetBmPositions : obtiene las posiciones del bitmap
func GetBmPositions(cant int64, pathdsk string, sizeBm int64, posBmap int64, ffree int64) []int64 {
	fpos := make([]int64, 0)
	// Recuperar BitMap AVD
	state, bmap := GetByteArray(pathdsk, sizeBm, posBmap)
	if state {
		for i := int64(ffree - 1); i < sizeBm; i++ {
			for x := int64(0); x < cant; x++ {
				if bmap[i] == 0 {
					fpos = append(fpos, (i + 1))
				} else {
					fpos = make([]int64, 0)
				}
				i++
			}
			if int64(len(fpos)) == cant {
				break
			}
		}
	} else {
		fmt.Println("[!] Error al leer BitMap...")
	}
	return fpos
}

//--- MAKE FILE -------------------------------------------------------------------------

//Mkfile : creación de archivos
func Mkfile(cmd CommandS) {
	idp := ""            // [*] Obligatorio
	pathFile := ""       // [*] Obligatorio
	isP := false         // [-] Opcional
	sizeFile := int64(0) // [-] Opcional
	contFile := ""       // [-] Opcional
	cm := Mounted{}
	for _, param := range cmd.Params {
		switch strings.ToLower(param.Name) {
		case "id":
			idp = delQuotationMark(param.Value)
		case "path":
			if strings.Contains(param.Value, "\"") {
				pathFile = strings.Replace(param.Value, "\"", "", -1)
				continue
			}
			pathFile = param.Value
		case "p":
			isP = true
		case "size":
			if n, err := strconv.Atoi(param.Value); err == nil {
				sizeFile = int64(n)
			} else {
				fmt.Println("[!] El valor del parametro 'size' no es un numero.")
			}
		case "cont":
			if strings.Contains(param.Value, "\"") {
				contFile = strings.Replace(param.Value, "\"", "", -1)
				continue
			}
			contFile = param.Value
		}
	}
	if pathFile != "" && idp != "" {
		// Creación de Directorios
		fmt.Println("\n===== CREAR ARCHIVOS ===========================================")
		fmt.Println("- ID:", idp)
		fmt.Println("- Path File:", pathFile)
		fmt.Println("- Size:", sizeFile)
		fmt.Println("- Contenido:", contFile)
		fmt.Println("- P:", isP)
		// Obtener la particion
		flgfound := false
		for _, mp := range sliceMP {
			idm := "vd" + string(mp.Letter) + strconv.FormatInt(mp.Number, 10)
			if idp == idm {
				flgfound = true
				cm = mp
				break
			}
		}
		// Si la particion ha sido montada
		if flgfound {
			sb := ReadSuperBoot(cm.Path, cm.Part.PartStart)
			// Obtener el directorio "/"
			cDir := ReadAVD(cm.Path, sb.AptArbolDirectorio)
			// Empezar a buscar si existen los directorios
			auxDir := cDir
			posAuxDir := sb.AptArbolDirectorio
			// File Path, File Name, File Extension
			p, n, e := SeparatePath(pathFile)
			fileName := n + "." + e
			p = p[0:(len(p) - 1)]
			dirsOk := true
			if isP {
				if len(p) > 0 { // Directorio Raiz
					stateDirs := MakeDirs(cm, p, isP)
					// Actualizo SuperBoot
					sb = ReadSuperBoot(cm.Path, cm.Part.PartStart)
					if stateDirs == 0 {
						dirsOk = false
					} else {
						sliceDirs := GetDirsNames(p[1:len(p)])
						//fmt.Println(sliceDirs)
						// Busco el directorio
						for _, name := range sliceDirs {
							// Obtengo el directorio y su posicion...
							posAuxDir, auxDir = GetSubDir(name, cm.Path, sb, auxDir, posAuxDir)
						}
					}
				}
			} else { // ACÁ EL DIRECTORIO YA DEBE EXISTIR
				if len(p) > 0 { // Directorio Raiz
					sliceDirs := GetDirsNames(p[1:len(p)])
					//fmt.Println(sliceDirs)
					// Busco el directorio
					for _, name := range sliceDirs {
						// Obtengo el directorio y su posicion...
						posAuxDir, auxDir = GetSubDir(name, cm.Path, sb, auxDir, posAuxDir)
						var bname [16]byte
						copy(bname[:], name)
						if bname != auxDir.NombreDirectorio {
							fmt.Println("[!] No se puede crear el archivo", pathFile, "\n    el directorio no existe y el parametro -p no fue definido.")
							dirsOk = false
						}
					}
				}

			}
			// El directorio Existe...
			if dirsOk {
				fmt.Println("> Current Dir Name:", string(auxDir.NombreDirectorio[:]))
				fmt.Println("> Current Dir Pos:", posAuxDir)
				var posdd int64
				var ddir DetalleDirectorio
				var aptInfo int
				//-------------------------------------------------------------------
				// ¿Existe un detalle directorio en la carpeta?
				if auxDir.AptDetalleDirectorio > 0 {
					// Obtener posicion del detalle directorio
					posdd = sb.AptDetalleDirectorio + (sb.TamStrcDetalleDirectorio * (auxDir.AptDetalleDirectorio - 1))
					// Obtener detalle directorio
					posdd, ddir, aptInfo = GetDetDir(cm, sb, posdd)
					// Actualizo el SuperBoot
					sb = ReadSuperBoot(cm.Path, cm.Part.PartStart)
				} else {
					numDetDir := sb.PrimerBitLibreDetalleDir
					auxDir.AptDetalleDirectorio = numDetDir
					WriteAVD(cm.Path, auxDir, posAuxDir)
					// Creo un nuevo detalle directorio
					ddir = DetalleDirectorio{}
					posdd = sb.AptDetalleDirectorio + (sb.TamStrcDetalleDirectorio * (numDetDir - 1))
					aptInfo = 0
					WriteDetalleDir(cm.Path, ddir, posdd)
					// Escribir el Bitmap DetDir
					auxBytes := []byte{1}
					realposBmap := sb.AptBmapDetalleDirectorio + (numDetDir - 1)
					WriteBitMap(cm.Path, auxBytes, realposBmap)
					// Actulizo los valores en el Super Boot
					sb.CantDetalleDirectorio = sb.CantDetalleDirectorio + 1
					sb.PrimerBitLibreDetalleDir = numDetDir + 1
					sb.DetallesDirectorioLibres = sb.DetallesDirectorioLibres - 1
				}
				// Validaciones para crear archivo...
				if contFile != "" {
					// Crear Bloque e Inodos
					bloques := CreateDataBlock(contFile, cm, sb)
					inodos := CreateInode(cm, sb, bloques, int64(len(contFile)))
					// Guardar Referencias en DetalleDir
					ddir.InfoFile[aptInfo].ApInodo = inodos[0]
					ddir.InfoFile[aptInfo].FechaCreacion = getCurrentTime()
					ddir.InfoFile[aptInfo].FechaModifiacion = getCurrentTime()
					copy(ddir.InfoFile[aptInfo].FileName[:], fileName)
					// Escribo el Detalle Directorio actualizado
					WriteDetalleDir(cm.Path, ddir, posdd)
					// Actualizacion de SuperBoot
					sb.PrimerBitLibreBloques = sb.PrimerBitLibreBloques + int64(len(bloques))
					sb.CantidadBloques = sb.CantidadBloques + int64(len(bloques))
					sb.BloquesLibres = sb.BloquesLibres - int64(len(bloques))
					//-----------------------------------------------------------
					sb.PrimerBitLibreTablaInodo = sb.PrimerBitLibreTablaInodo + int64(len(inodos))
					sb.CantidadInodos = sb.CantidadInodos + int64(len(inodos))
					sb.InodosLibres = sb.InodosLibres - int64(len(inodos))
					// Escritura del SuperBoot
					WriteSuperBoot(cm.Path, sb, cm.Part.PartStart)
					// Actualizar BitMaps
					bArray := make([]byte, 0)
					for int64(len(bArray)) < int64(len(bloques)) {
						bArray = append(bArray, 1)
					}
					WriteBitMap(cm.Path, bArray, sb.AptBmapBloques+(bloques[0]-1))
					//-----------------------------------------------------------
					bArray = make([]byte, 0)
					for int64(len(bArray)) < int64(len(inodos)) {
						bArray = append(bArray, 1)
					}
					WriteBitMap(cm.Path, bArray, sb.AptBmapTablaInodo+(inodos[0]-1))
				} else if sizeFile > 0 {
					contFile = GenerateContent(sizeFile)
					// Crear Bloque e Inodos
					bloques := CreateDataBlock(contFile, cm, sb)
					inodos := CreateInode(cm, sb, bloques, int64(len(contFile)))
					// Guardar Referencias en DetalleDir
					ddir.InfoFile[aptInfo].ApInodo = inodos[0]
					ddir.InfoFile[aptInfo].FechaCreacion = getCurrentTime()
					ddir.InfoFile[aptInfo].FechaModifiacion = getCurrentTime()
					copy(ddir.InfoFile[aptInfo].FileName[:], fileName)
					// Escribo el Detalle Directorio actualizado
					WriteDetalleDir(cm.Path, ddir, posdd)
					// Actualizacion de SuperBoot
					sb.PrimerBitLibreBloques = sb.PrimerBitLibreBloques + int64(len(bloques))
					sb.CantidadBloques = sb.CantidadBloques + int64(len(bloques))
					sb.BloquesLibres = sb.BloquesLibres - int64(len(bloques))
					//-----------------------------------------------------------
					sb.PrimerBitLibreTablaInodo = sb.PrimerBitLibreTablaInodo + int64(len(inodos))
					sb.CantidadInodos = sb.CantidadInodos + int64(len(inodos))
					sb.InodosLibres = sb.InodosLibres - int64(len(inodos))
					// Escritura del SuperBoot
					WriteSuperBoot(cm.Path, sb, cm.Part.PartStart)
					// Actualizar BitMaps
					bArray := make([]byte, 0)
					for int64(len(bArray)) < int64(len(bloques)) {
						bArray = append(bArray, 1)
					}
					WriteBitMap(cm.Path, bArray, sb.AptBmapBloques+(bloques[0]-1))
					//-----------------------------------------------------------
					bArray = make([]byte, 0)
					for int64(len(bArray)) < int64(len(inodos)) {
						bArray = append(bArray, 1)
					}
					WriteBitMap(cm.Path, bArray, sb.AptBmapTablaInodo+(inodos[0]-1))
				} else if sizeFile == 0 {
					// Creo un Bloque de datos vacio...
					bvacio := BloqueDeDatos{}
					numBloque := sb.PrimerBitLibreBloques
					posBD := sb.AptBloques + (sb.TamStrcBloque * (numBloque - 1))
					WriteBloqueD(cm.Path, bvacio, posBD)
					// Creo al Inodo
					numInodo := sb.PrimerBitLibreTablaInodo
					inodo := TablaInodo{}
					inodo.NumeroInodo = numInodo
					inodo.CantBloquesAsignados = 1
					inodo.AptBloques[0] = numBloque
					copy(inodo.IDPropietario[:], "root")
					copy(inodo.IDUGrupo[:], "root")
					inodo.IPermisos = 664
					// Escribir Inodo
					posInodo := sb.AptTablaInodo + (sb.TamStrcInodo * (numInodo - 1))
					WriteTInodo(cm.Path, inodo, posInodo)
					// Guardar Referencias en DetalleDir
					ddir.InfoFile[aptInfo].ApInodo = numInodo
					ddir.InfoFile[aptInfo].FechaCreacion = getCurrentTime()
					ddir.InfoFile[aptInfo].FechaModifiacion = getCurrentTime()
					copy(ddir.InfoFile[aptInfo].FileName[:], fileName)
					// Escribo el Detalle Directorio actualizado
					WriteDetalleDir(cm.Path, ddir, posdd)
					// Actualizacion de SuperBoot
					sb.PrimerBitLibreBloques = sb.PrimerBitLibreBloques + 1
					sb.CantidadBloques = sb.CantidadBloques + 1
					sb.BloquesLibres = sb.BloquesLibres - 1
					//-----------------------------------------------------------
					sb.PrimerBitLibreTablaInodo = sb.PrimerBitLibreTablaInodo + 1
					sb.CantidadInodos = sb.CantidadInodos + 1
					sb.InodosLibres = sb.InodosLibres - 1
					// Escritura del SuperBoot
					WriteSuperBoot(cm.Path, sb, cm.Part.PartStart)
					// Actualizar BitMaps
					auxBytes := []byte{1}
					WriteBitMap(cm.Path, auxBytes, sb.AptBmapBloques+(numBloque-1))
					//-----------------------------------------------------------
					WriteBitMap(cm.Path, auxBytes, sb.AptBmapTablaInodo+(numInodo-1))
				} else {
					fmt.Println("[!] El tamaño de archivo no puede ser un valor negativo...")
					fmt.Println("================================================================")
					return
				}
			}
		} else {
			fmt.Println("[!] La particion", idp, " no se encuentra montada...")
		}
	} else {
		fmt.Println("[!~MKFILE] Hacen falta parametros obligatorios...")
	}
	fmt.Println("================================================================")
}

// GetDetDir : obtener el detalle directorio...
func GetDetDir(cm Mounted, sb SuperBoot, position int64) (int64, DetalleDirectorio, int) {
	detdir := ReadDetalleDir(cm.Path, position)
	for i, infile := range detdir.InfoFile {
		if infile == (InfoArchivo{}) {
			return position, detdir, i
		}
	}
	// Sino se encontro espacio en el detalle directorio, se busca en el indirecto...
	// Si existe un apuntador indirecto...
	if detdir.ApDetalleDirectorio > 0 {
		// Calcular la posicion del detalle directorio indirecto
		posdd := sb.AptDetalleDirectorio + (sb.TamStrcDetalleDirectorio * (detdir.ApDetalleDirectorio - 1))
		return GetDetDir(cm, sb, posdd)
	}
	// Se creara un indirecto...
	// Se guarda el Apuntador Indirecto al nuevo detalle directorio...
	detdir.ApDetalleDirectorio = sb.PrimerBitLibreDetalleDir
	WriteDetalleDir(cm.Path, detdir, position)
	// Se crea el Detalle Directorio
	indDetDir, posIndDetDir := CreateDetDir(cm, sb)
	return posIndDetDir, indDetDir, 0
}

// CreateDetDir : crea un detalle directorio
func CreateDetDir(cm Mounted, sb SuperBoot) (DetalleDirectorio, int64) {
	numDetDir := sb.PrimerBitLibreDetalleDir
	// Creo la estructura de detalle de Directorio
	newDetDir := DetalleDirectorio{}
	// Calcular la posicion de escritura
	posNewDetDir := sb.AptDetalleDirectorio + (sb.TamStrcDetalleDirectorio * (numDetDir - 1))
	// Escribir nuevo detalle de directorio
	WriteDetalleDir(cm.Path, newDetDir, posNewDetDir)
	// Actualizar Info SB
	sb.CantDetalleDirectorio = sb.CantDetalleDirectorio + 1
	sb.PrimerBitLibreDetalleDir = numDetDir + 1
	sb.DetallesDirectorioLibres = sb.DetallesDirectorioLibres - 1
	WriteSuperBoot(cm.Path, sb, cm.Part.PartStart)
	// Escribir el Bitmap DetDir
	auxBytes := []byte{1}
	realposBmap := sb.AptBmapDetalleDirectorio + (numDetDir - 1)
	WriteBitMap(cm.Path, auxBytes, realposBmap)
	// Retorno el nuevo detalle directorio
	return newDetDir, posNewDetDir
}

// CreateDataBlock : crea los bloques de datos
func CreateDataBlock(cont string, cm Mounted, sb SuperBoot) []int64 {
	// Obtener el tamaño del contenido
	lenCont := int64(len(cont))
	fmt.Println("LC:", lenCont)
	// Calculo del Residuo
	residuo := lenCont % 25
	fmt.Println("Residuo:", residuo)
	// Cantidad de Bloques
	cantBloques := int64(0)
	if residuo == 0 {
		cantBloques = lenCont / 25
	} else {
		cantBloques = (lenCont / 25) + 1
	}
	fmt.Println("Bloques:", cantBloques)
	// Obtener arreglo de Posiciones
	posBmapBloque := make([]int64, 0)
	primetBitLibre := sb.PrimerBitLibreBloques
	for i := int64(0); i < cantBloques; i++ {
		posBmapBloque = append(posBmapBloque, (primetBitLibre + i))
	}
	fmt.Println("Posiciones:", posBmapBloque)
	// Creacion de los bloques de datos
	for _, posBloque := range posBmapBloque {
		newBloque := BloqueDeDatos{}
		if lenCont >= 25 {
			s := cont[0:25]
			copy(newBloque.Data[:], s)
			cont = cont[25:lenCont]
		} else {
			copy(newBloque.Data[:], cont)
		}
		lenCont = int64(len(cont))
		// Escribir el bloque de datos en el disco
		posBD := sb.AptBloques + (sb.TamStrcBloque * (posBloque - 1))
		WriteBloqueD(cm.Path, newBloque, posBD)
	}
	return posBmapBloque
}

// CreateInode : crea los Inodos necesarios
func CreateInode(cm Mounted, sb SuperBoot, posBloques []int64, fileSize int64) []int64 {
	// Obtener cantidad de Inodos
	cantInodos := int64(0)
	if len(posBloques)%4 == 0 {
		cantInodos = int64((len(posBloques) / 4))
	} else {
		cantInodos = int64((len(posBloques) / 4) + 1)
	}
	// Obtener posicion de Inodos
	posInode := make([]int64, 0)
	for i := int64(0); i < cantInodos; i++ {
		posInode = append(posInode, (sb.PrimerBitLibreTablaInodo + i))
	}
	// Crear los Inodos
	Inodo := TablaInodo{}
	Inodo.SizeArchivo = fileSize
	Inodo.CantBloquesAsignados = int64(len(posBloques))
	copy(Inodo.IDPropietario[:], "root")
	copy(Inodo.IDUGrupo[:], "root")
	Inodo.IPermisos = 664
	// Completar Inodo
	auxBloques := posBloques
	for x, inode := range posInode {
		newInodo := TablaInodo{}
		newInodo = Inodo
		newInodo.NumeroInodo = inode
		if len(auxBloques) > 4 {
			i := 0
			for i < 4 {
				newInodo.AptBloques[i] = auxBloques[i]
				i++
			}
			auxBloques = auxBloques[4:len(auxBloques)]
		} else {
			i := 0
			for i < len(auxBloques) {
				newInodo.AptBloques[i] = auxBloques[i]
				i++
			}
		}
		// Si hay un inodo despues
		if (x + 1) < len(posInode) {
			newInodo.AptIndirecto = posInode[x+1]
		}
		position := int64(sb.AptTablaInodo + (sb.TamStrcInodo * (inode - 1)))
		WriteTInodo(cm.Path, newInodo, position)
	}
	return posInode
}

// GenerateContent : genera contenido para un archivo de tamaño 'n'
func GenerateContent(size int64) string {
	var b byte = 65
	content := ""
	for i := int64(0); i < size; i++ {
		if b == 91 {
			b = 65
		}
		content += string(b)
		b++
	}
	return content
}

//--- FUNCIONES DE ESCRITURA DE ESTRUCTURAS ---------------------------------------------

// WriteAVD : escribe avd en disco
func WriteAVD(path string, avd ArbolVirtualDir, position int64) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("[!] Error al abrir disco. ", err)
		return
	}
	defer file.Close()
	//Posición donde se inicia a escribir
	file.Seek(position, 1)
	avdW := &avd
	//Escritura del struct MBR
	var bin bytes.Buffer
	binary.Write(&bin, binary.BigEndian, avdW)
	state := writeBytes(file, bin.Bytes())
	if state {
		//fmt.Println("*** Se ha escrito AVD  ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al escribir el AVD.")
	}
}

// ReadAVD : lee avd en disco
func ReadAVD(path string, position int64) ArbolVirtualDir {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	// Posicion del puntero
	//currentPosition, err := file.Seek(position, 1)
	file.Seek(position, 1)
	//fmt.Println("Posicion 'Seek' ReadEbr: ", currentPosition)
	// Se declara AVD contenedor
	recAVD := ArbolVirtualDir{}
	// Se obtiene el tamaño del EBR
	var sizeSB int64 = int64(binary.Size(recAVD))
	// Lectura los bytes determinados por ebrSize
	data := readBytes(file, sizeSB)
	// Convierte data en un buffer, necesario para decodificar binario
	buffer := bytes.NewBuffer(data)
	// Se decodifica y guarda en recMbr
	err = binary.Read(buffer, binary.BigEndian, &recAVD)
	if err != nil {
		log.Fatal("[!] Fallo la lectura del AVD", err)
	} else {
		//fmt.Println("[ AVD leido exitosamente ]")
	}
	return recAVD
}

// WriteDetalleDir : escribe DetalleDir en disco
func WriteDetalleDir(path string, ddir DetalleDirectorio, position int64) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("[!] Error al abrir disco. ", err)
		return
	}
	defer file.Close()
	//Posición donde se inicia a escribir
	file.Seek(position, 1)
	rec := &ddir
	//Escritura del struct MBR
	var bin bytes.Buffer
	binary.Write(&bin, binary.BigEndian, rec)
	state := writeBytes(file, bin.Bytes())
	if state {
		//fmt.Println("*** Se ha escrito Detalle Directorio  ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al escribir el Detalle Directorio.")
	}
}

// ReadDetalleDir : recupera DetalleDir del disco
func ReadDetalleDir(path string, position int64) DetalleDirectorio {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	// Posicion del puntero
	//currentPosition, err := file.Seek(position, 1)
	file.Seek(position, 1)
	//fmt.Println("Posicion 'Seek' ReadEbr: ", currentPosition)
	// Se declara AVD contenedor
	rec := DetalleDirectorio{}
	// Se obtiene el tamaño del EBR
	var size int64 = int64(binary.Size(rec))
	// Lectura los bytes determinados por size
	data := readBytes(file, size)
	// Convierte data en un buffer, necesario para decodificar binario
	buffer := bytes.NewBuffer(data)
	// Se decodifica y guarda en recMbr
	err = binary.Read(buffer, binary.BigEndian, &rec)
	if err != nil {
		log.Fatal("[!] Fallo la lectura del Detalle Directorio", err)
	} else {
		//fmt.Println("[ Detalle Directorio leido exitosamente ]")
	}
	return rec
}

// WriteTInodo : escribe Tabla Indo en disco
func WriteTInodo(path string, tinodo TablaInodo, position int64) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("[!] Error al abrir disco. ", err)
		return
	}
	defer file.Close()
	//Posición donde se inicia a escribir
	file.Seek(position, 1)
	rec := &tinodo
	//Escritura del struct MBR
	var bin bytes.Buffer
	binary.Write(&bin, binary.BigEndian, rec)
	state := writeBytes(file, bin.Bytes())
	if state {
		//fmt.Println("*** Se ha escrito Tabla Inodo  ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al escribir Tabla Inodo.")
	}
}

// ReadTInodo : recupera DetalleDir del disco
func ReadTInodo(path string, position int64) TablaInodo {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	// Posicion del puntero
	//currentPosition, err := file.Seek(position, 1)
	file.Seek(position, 1)
	//fmt.Println("Posicion 'Seek' ReadEbr: ", currentPosition)
	// Se declara AVD contenedor
	rec := TablaInodo{}
	// Se obtiene el tamaño del EBR
	var size int64 = int64(binary.Size(rec))
	// Lectura los bytes determinados por ebrSize
	data := readBytes(file, size)
	// Convierte data en un buffer, necesario para decodificar binario
	buffer := bytes.NewBuffer(data)
	// Se decodifica y guarda en recMbr
	err = binary.Read(buffer, binary.BigEndian, &rec)
	if err != nil {
		log.Fatal("[!] Fallo la lectura de Tabla Inodo", err)
	} else {
		fmt.Println("[ Tabla Inodo leída exitosamente ]")
	}
	return rec
}

// WriteBloqueD : escribe bloque de datos en disco
func WriteBloqueD(path string, bd BloqueDeDatos, position int64) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("[!] Error al abrir disco. ", err)
		return
	}
	defer file.Close()
	//Posición donde se inicia a escribir
	file.Seek(position, 1)
	rec := &bd
	//Escritura del struct MBR
	var bin bytes.Buffer
	binary.Write(&bin, binary.BigEndian, rec)
	state := writeBytes(file, bin.Bytes())
	if state {
		//fmt.Println("*** Se ha escrito Bloque de Datos  ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al Bloque de Datos.")
	}
}

// ReadBloqueD : recupera bloque de datos del disco
func ReadBloqueD(path string, position int64) BloqueDeDatos {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	// Posicion del puntero
	//currentPosition, err := file.Seek(position, 1)
	file.Seek(position, 1)
	//fmt.Println("Posicion 'Seek' ReadEbr: ", currentPosition)
	// Se declara Bloque de Datos contenedor
	rec := BloqueDeDatos{}
	// Se obtiene el tamaño del EBR
	var size int64 = int64(binary.Size(rec))
	// Lectura los bytes determinados por ebrSize
	data := readBytes(file, size)
	// Convierte data en un buffer, necesario para decodificar binario
	buffer := bytes.NewBuffer(data)
	// Se decodifica y guarda en recMbr
	err = binary.Read(buffer, binary.BigEndian, &rec)
	if err != nil {
		log.Fatal("[!] Fallo la lectura de Bloque de Datos", err)
	} else {
		//fmt.Println("[ Bloque de datos leído exitosamente ]")
	}
	return rec
}

// WriteBitMap escribir en Bitmap
func WriteBitMap(path string, btes []byte, position int64) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("[!] Error al abrir disco. ", err)
		return
	}
	defer file.Close()
	//Posición donde se inicia a escribir
	file.Seek(position, 1)
	bts := &btes
	//Escritura del struct MBR
	var bin bytes.Buffer
	binary.Write(&bin, binary.BigEndian, bts)
	state := writeBytes(file, bin.Bytes())
	if state {
		//fmt.Println("*** Se ha escrito en BitMap  ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al escribir en BitMap.")
	}
}
