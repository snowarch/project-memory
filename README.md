# Project Memory Bank

CLI profesional para tracking y gestión de proyectos locales con análisis IA integrado.

## Stack Técnico

- **Go 1.21+** - Binario standalone compilado
- **SQLite3** - Base de datos local sin servidor
- **Cobra** - Framework CLI profesional
- **Gum** - TUI interactivo para listados
- **Groq API** - Análisis IA con `moonshotai/kimi-k2-instruct`

## Características

### Detección Automática
- ✅ **Node.js** - package.json, frameworks (React, Next.js, Vue, Angular, Express)
- ✅ **Python** - requirements.txt
- ✅ **Go** - go.mod + extracción de versión
- ✅ **Rust** - Cargo.toml + edition
- ✅ **Git** - remote, branch, información del repositorio

### Gestión de Proyectos
- Estados: `active`, `paused`, `archived`, `completed`
- Progreso automático (0-100%)
- Búsqueda por nombre, descripción, path
- Filtros por status
- Timestamps automáticos

### Análisis IA
- Evaluación de estado actual
- Estimación de completitud
- Próximos pasos recomendados
- Identificación de blockers técnicos
- Consumo de tokens optimizado

## Instalación

```bash
# Clonar repositorio
git clone https://github.com/snowarch/project-memory.git
cd project-memory

# Compilar
make build

# Instalar globalmente (opcional)
make install

# Inicializar base de datos
pmem init
```

## Configuración

### API Key de Groq

```bash
# Variable de entorno (recomendado)
export GROQ_API_KEY='your-api-key-here'

# O usar flag
pmem analyze project-name --api-key='your-api-key'
```

### Base de Datos

Por defecto: `~/.local/share/pmem/projects.db`

Personalizar con flag `--db`:
```bash
pmem scan /path --db=/custom/path/projects.db
```

## Uso

### Comandos Disponibles

```bash
# Inicializar base de datos
pmem init

# Escanear directorio en busca de proyectos
pmem scan /path/to/projects
pmem scan /path/to/projects -v  # verbose

# Listar proyectos (TUI interactivo con Gum)
pmem list
pmem list --status active
pmem list --status completed --limit 10

# Ver/cambiar status de proyecto
pmem status project-name
pmem status project-name paused
pmem status project-name completed

# Análisis IA con Groq
pmem analyze project-name
pmem analyze project-name --api-key='...'
```

### Flags Globales

- `-v, --verbose` - Output detallado con logs DEBUG
- `-q, --quiet` - Silenciar output excepto errores
- `--db <path>` - Ruta custom a base de datos
- `-p, --path <path>` - Root path para scan

## Ejemplos

### Flujo Típico

```bash
# 1. Inicializar
pmem init

# 2. Escanear proyectos
export GROQ_API_KEY='gsk_...'
pmem scan ~/CascadeProjects -v

# 3. Ver lista interactiva
pmem list

# 4. Analizar proyecto específico
pmem analyze my-project

# 5. Actualizar estado
pmem status my-project completed
```

### Búsqueda y Filtros

```bash
# Listar solo proyectos activos
pmem list --status active

# Proyectos pausados
pmem list --status paused

# Primeros 5 proyectos
pmem list --limit 5
```

## Estructura del Proyecto

```
project-memory/
├── cmd/pmem/           # Entry point
├── internal/
│   ├── ai/             # Cliente Groq API
│   ├── commands/       # Comandos Cobra
│   ├── database/       # SQLite setup
│   ├── logger/         # Sistema de logging
│   ├── models/         # Estructuras de datos
│   ├── repository/     # Capa de datos
│   └── scanner/        # Detección de proyectos
├── go.mod
├── Makefile
└── README.md
```

## Tests

```bash
# Ejecutar todos los tests
go test -v ./...

# Test específico
go test -v ./internal/scanner
go test -v ./internal/repository

# Coverage
go test -cover ./...
```

**Test Status:** ✅ 23/23 tests passing

## Base de Datos

### Schema

- `projects` - Información de proyectos
- `technologies` - Stack tecnológico detectado
- `project_files` - Metadata de archivos
- `todos` - TODOs extraídos
- `ai_analyses` - Historial de análisis IA
- `activity_log` - Log de cambios
- `config` - Configuración del sistema

Ver [`internal/database/schema.sql`](internal/database/schema.sql) para detalles.

## Modelo IA

**Modelo actual:** `moonshotai/kimi-k2-instruct`

- Context window: 131,072 tokens
- Max completion: 16,384 tokens
- Análisis optimizado para proyectos de software

## Licencia

MIT License - Ver [LICENSE](LICENSE)

## Autor

**snowarch** - [GitHub](https://github.com/snowarch)
