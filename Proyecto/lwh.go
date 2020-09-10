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
		posBloque1 := int64(1)
		posBloque2 := int64(2)
		tbInodo.AptBloques[0] = posBloque1
		tbInodo.AptBloques[1] = posBloque2
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
		fmt.Println("*** Se ha escrito SuperBoot  ***")
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
		fmt.Println("[ SuperBoot leido exitosamente ]")
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

//--- LOGIN & LOGOUT -------------------------------------------------------------------------

// Login : inicio de sesión del usuario
func Login(cmd CommandS) {

}

//--- FUNCIONES DE ESCRITURA DE ESTRUCTURAS --------------------------------------------------

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
		fmt.Println("*** Se ha escrito AVD  ***")
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
		fmt.Println("[ AVD leido exitosamente ]")
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
		fmt.Println("*** Se ha escrito Detalle Directorio  ***")
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
		fmt.Println("[ Detalle Directorio leido exitosamente ]")
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
		fmt.Println("*** Se ha escrito Tabla Inodo  ***")
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
		fmt.Println("*** Se ha escrito Bloque de Datos  ***")
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
		fmt.Println("[ Bloque de datos leído exitosamente ]")
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
		fmt.Println("*** Se ha escrito en BitMap  ***")
	} else {
		fmt.Println("[!] Ha ocurrido un error al escribir en BitMap.")
	}
}
