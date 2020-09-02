package main

//SuperBoot : contiene toda la info del sistema
type SuperBoot struct {
	nombreHd                 [16]byte
	cantArbolVirtual         int64
	cantDetalleDirectorio    int64
	cantidadInodos           int64
	cantidadBloques          int64
	arbolesVirtualesLibres   int64
	detallesDirectorioLibres int64
	inodosLibres             int64
	bloquesLibres            int64
	fechaCreacion            Time
	fechaUltimoMontaje       Time
	conteoMontajes           int64
	aptBmapArbolDirectorio   int64
	aptArbolDirectorio       int64
	aptBmapDetalleDirectorio int64
	aptDetalleDirectorio     int64
	aptBmapTablaInodo        int64
	aptTablaInodo            int64
	aptBmapBloques           int64
	aptBloques               int64
	aptLog                   int64
	tamStrcArbolDirectorio   int64
	tamStrcDetalleDirectorio int64
	tamStrcInodo             int64
	tamStrcBloque            int64
	primerBitLibreArbolDir   int64
	primerBitLibreDetalleDir int64
	primerBitLibreTablaInodo int64
	primerBitLibreBloques    int64
	numeroMagico             int64
}

//ArbolVirtualDir : para la creación de carpetas
type ArbolVirtualDir struct {
	fechaCreacion        Time
	nombreDirectorio     [16]byte
	aptArregloSubDir     [6]int64
	aptDetalleDirectorio int64
	aptArbolVirtualDir   int64
	avdPropietario       [16]byte
}

//DetalleDirectorio : son los i-nodos
type DetalleDirectorio struct {
	fileName1           [16]byte
	apInodo1            int64
	fechaCreacion1      Time
	fechaModifiacion1   Time
	fileName2           [16]byte
	apInodo2            int64
	fechaCreacion2      Time
	fechaModifiacion2   Time
	fileName3           [16]byte
	apInodo3            int64
	fechaCreacion3      Time
	fechaModifiacion3   Time
	fileName4           [16]byte
	apInodo4            int64
	fechaCreacion4      Time
	fechaModifiacion4   Time
	fileName5           [16]byte
	apInodo5            int64
	fechaCreacion5      Time
	fechaModifiacion5   Time
	apDetalleDirectorio int64 //Apuntador al siguiente detalle-directorio
}

//TablaInodo : para le manejo de archivos de directorio
type TablaInodo struct {
	conteoInodo          int64
	sizeArchivo          int64
	cantBloquesAsignados int64
	apIndirecto          int64
	idPropietario        [16]byte
}

//BloqueDeDatos : para la creación de archivos
type BloqueDeDatos struct {
	data [25]byte
}

//Log : Bitacora
type Log struct {
	tipoOperacion [16]byte
	tipo          byte // 0 = archivo, 1 = directorio
	nombre        [16]byte
	contenido     [50]byte
	fecha         Time
}
