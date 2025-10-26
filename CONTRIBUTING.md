# Contributing to Project Memory Bank

Gracias por tu interés en contribuir a Project Memory Bank.

## Proceso de Desarrollo

### 1. Fork y Clone

```bash
# Fork en GitHub, luego:
git clone https://github.com/tu-usuario/project-memory.git
cd project-memory
```

### 2. Setup

```bash
# Instalar dependencias
go mod download

# Compilar
make build

# Ejecutar tests
go test -v ./...
```

### 3. Crear Branch

```bash
git checkout -b feature/tu-feature
# o
git checkout -b fix/tu-bug-fix
```

### 4. Desarrollo

- Escribe código limpio y testeado
- Sigue convenciones de Go (gofmt, golint)
- Añade tests para nuevas funcionalidades
- Actualiza documentación si es necesario

### 5. Tests

```bash
# Tests unitarios
go test -v ./...

# Tests con coverage
go test -cover ./...

# Race detector
go test -race ./...

# Lint
golangci-lint run
```

### 6. Commit

Usa [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: añadir soporte para Java projects
fix: corregir detección de versiones en Go
docs: actualizar README con ejemplos
test: añadir tests para scanner
refactor: simplificar lógica de repositories
```

### 7. Push y PR

```bash
git push origin feature/tu-feature
```

Crea Pull Request en GitHub con:
- Título descriptivo
- Descripción clara de cambios
- Screenshots si aplica
- Referencias a issues relacionados

## Guías de Código

### Estructura

```go
// Comentarios para funciones públicas
func PublicFunction(arg string) error {
    // Validación primero
    if arg == "" {
        return fmt.Errorf("arg cannot be empty")
    }
    
    // Lógica
    result := processArg(arg)
    
    return nil
}
```

### Testing

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "test",
            wantErr:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
            }
            
            if result != tt.expected {
                t.Errorf("expected: %v, got: %v", tt.expected, result)
            }
        })
    }
}
```

### Logging

```go
// Usar logger en lugar de fmt
logger.Debug("Processing project: %s", projectName)
logger.Info("Scan complete: %d projects", count)
logger.Warn("Failed to detect tech: %v", err)
logger.Error("Critical error: %v", err)
```

## Áreas de Contribución

### Prioridad Alta
- [ ] Soporte para más lenguajes (Java, C++, C#)
- [ ] Extracción de TODOs de código fuente
- [ ] Cache de análisis IA
- [ ] Rate limiting para API calls

### Prioridad Media
- [ ] Comando `search` para búsqueda avanzada
- [ ] Export a JSON/CSV
- [ ] Git commit history analysis
- [ ] Dependency vulnerability scanning

### Prioridad Baja
- [ ] Web dashboard (opcional)
- [ ] Plugin system
- [ ] Custom templates
- [ ] Multi-language support en UI

## Reportar Bugs

Usa [GitHub Issues](https://github.com/snowarch/project-memory/issues) con:

- **Título claro**: "Error al escanear proyectos Rust"
- **Descripción**: Qué esperabas vs qué ocurrió
- **Steps to reproduce**: Pasos exactos
- **Environment**: OS, Go version, pmem version
- **Logs**: Output con `--verbose` si aplica

## Preguntas

Para preguntas generales, usa [GitHub Discussions](https://github.com/snowarch/project-memory/discussions).

## Código de Conducta

- Sé respetuoso y constructivo
- Acepta feedback con apertura
- Enfócate en lo mejor para el proyecto
- Ayuda a otros contributors

## Licencia

Al contribuir, aceptas que tu código se licencia bajo MIT License.
