package main

//SuperBoot : contiene toda la info del sistema
type SuperBoot struct {
	NombreHd                 [16]byte
	CantArbolVirtual         int64
	CantDetalleDirectorio    int64
	CantidadInodos           int64
	CantidadBloques          int64
	ArbolesVirtualesLibres   int64
	DetallesDirectorioLibres int64
	InodosLibres             int64
	BloquesLibres            int64
	FechaCreacion            Time
	FechaUltimoMontaje       Time
	ConteoMontajes           int64
	AptBmapArbolDirectorio   int64
	AptArbolDirectorio       int64
	AptBmapDetalleDirectorio int64
	AptDetalleDirectorio     int64
	AptBmapTablaInodo        int64
	AptTablaInodo            int64
	AptBmapBloques           int64
	AptBloques               int64
	AptLog                   int64
	TamStrcArbolDirectorio   int64
	TamStrcDetalleDirectorio int64
	TamStrcInodo             int64
	TamStrcBloque            int64
	PrimerBitLibreArbolDir   int64
	PrimerBitLibreDetalleDir int64
	PrimerBitLibreTablaInodo int64
	PrimerBitLibreBloques    int64
	NumeroMagico             int64
}

//ArbolVirtualDir : para la creación de carpetas
type ArbolVirtualDir struct {
	FechaCreacion        Time
	NombreDirectorio     [16]byte
	AptArregloSubDir     [6]int64
	AptDetalleDirectorio int64
	AptArbolVirtualDir   int64
	AvdPropietario       [10]byte // Id del usuario propietario
	AvdGID               [10]byte // Id del grupo al que pertenece el usuario creador
	AvdPermisos          int64    // Codigo con el numero de permiso (777)
}

//DetalleDirectorio : son los i-nodos
type DetalleDirectorio struct {
	FileName1           [16]byte
	ApInodo1            int64
	FechaCreacion1      Time
	FechaModifiacion1   Time
	FileName2           [16]byte
	ApInodo2            int64
	FechaCreacion2      Time
	FechaModifiacion2   Time
	FileName3           [16]byte
	ApInodo3            int64
	FechaCreacion3      Time
	FechaModifiacion3   Time
	FileName4           [16]byte
	ApInodo4            int64
	FechaCreacion4      Time
	FechaModifiacion4   Time
	FileName5           [16]byte
	ApInodo5            int64
	FechaCreacion5      Time
	FechaModifiacion5   Time
	ApDetalleDirectorio int64 //Apuntador al siguiente detalle-directorio
}

//TablaInodo : para le manejo de archivos de directorio
type TablaInodo struct {
	ConteoInodo          int64
	SizeArchivo          int64
	CantBloquesAsignados int64
	ApIndirecto          int64
	IDPropietario        [10]byte
	IDUGrupo             [10]byte
	IPermisos            int64
}

//BloqueDeDatos : para la creación de archivos
type BloqueDeDatos struct {
	Data [25]byte
}

//Log : Bitacora
type Log struct {
	TipoOperacion [16]byte
	Tipo          byte // 0 = archivo, 1 = directorio
	Nombre        [16]byte
	Contenido     [50]byte
	Fecha         Time
}
