# Taller Cripto Go: Protección de Logs Estructurados en Reposo

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Security](https://img.shields.io/badge/Cryptography-Native_Go_crypto-green.svg)](#mecanismos-criptográficos)

Herramienta de línea de comandos (CLI) interactiva desarrollada en **Go** para simular, proteger y verificar la confidencialidad e integridad de registros académicos estructurados en almacenamiento secundario (datos en reposo).

Proyecto desarrollado en el marco académico de la **Universidad del Valle**.

---

## Tabla de Contenidos

1. [Visión General del Proyecto](#visión-general-del-proyecto)
2. [Flujo Operativo de los 10 Pasos](#flujo-operativo-de-los-10-pasos)
3. [Organización del Proyecto (Arquitectura)](#organización-del-proyecto-arquitectura)
4. [Especificación de la Estructura de Datos JSON](#especificación-de-la-estructura-de-datos-json)
5. [Modelo Relacional de Entidades](#modelo-relacional-de-entidades)
6. [Mecanismos Criptográficos y Seguridad](#mecanismos-criptográficos-y-seguridad)
7. [Guía de Instalación y Ejecución](#guía-de-instalación-y-ejecución)
8. [Licencia MIT](#licencia-mit)

---

## Visión General del Proyecto

- **Institución:** Universidad del Valle
- **Contexto:** Sistema de gestión académica enfocado en operaciones sensibles (registro y modificación de calificaciones).
- **Problema de Negocio:** Garantizar la **confidencialidad**, **integridad**, **trazabilidad** y **no repudio** de calificaciones académicas almacenadas en disco frente a amenazas de manipulación no autorizada, robo de información o corrupción en reposo.
- **Objetivo Técnico:** Diseñar e implementar una arquitectura limpia en Go que permita:
  1. Generar eventos sintéticos estructurados en formato JSON.
  2. Implementar un esquema de integridad multinivel con **SHA-256** (a nivel de registro individual y a nivel de archivo consolidado).
  3. Cifrar los registros en reposo utilizando **AES-256-GCM** (cifrado simétrico autenticado AEAD).
  4. Desacoplar los planos de seguridad (datos cifrados, artefactos de verificación y llaves).
  5. Descifrar y validar bit a bit que la información recuperada sea exactamente idéntica a la original.

---

## Flujo Operativo de los 10 Pasos

El ciclo de vida de la ejecución está compuesto por 10 pasos secuenciales:

| Paso | Nombre | Descripción Técnica | Artefacto de Salida |
| :---: | :--- | :--- | :--- |
| **1** | **Definición de volumen** | Solicitar al usuario vía CLI el número entero $N$ ($N \ge 1$) de eventos a simular. | Parámetro en memoria |
| **2** | **Generación aleatoria** | Generar $N$ eventos aleatorios alternando entre `REGISTRO_NOTA` y `MODIFICACION_NOTA`, con notas en $[0.0, 5.0]$ y pseudónimos/UUIDs. | Colección de structs en memoria |
| **3** | **Hash por evento (Payload)** | Serializar de forma determinística (canónica) el sub-objeto `payload` de cada evento, aplicar `SHA-256`, y asignar el hash resultante en el campo `payload_hash`. | Structs enriquecidos |
| **4** | **Consolidación del Log** | Serializar la colección completa de eventos a un JSON legible y escribirlo a disco. | `artifacts/logs.json` |
| **5** | **Hash del Log en Claro** | Calcular el hash SHA-256 del contenido completo de `logs.json` en texto plano y guardarlo por separado. | `artifacts/logs_original.hash` |
| **6** | **Cifrado en Reposo** | Cifrar `logs.json` usando `AES-256-GCM`. Generar clave criptográfica segura (32 bytes) y nonce (12 bytes) vía `crypto/rand`. Almacenar el archivo cifrado y los metadatos de las claves de forma desacoplada. | `artifacts/logs.enc`<br>`artifacts/keystore.json` |
| **7** | **Hash del Log Cifrado** | Calcular el hash SHA-256 del binario resultante `logs.enc` y guardarlo por separado. | `artifacts/logs_cifrado.hash` |
| **8** | **Descifrado del Log** | Leer `logs.enc`, cargar la clave/nonce y ejecutar el descifrado simétrico para reconstruir el archivo en texto claro. | `artifacts/logs_descifrado.json` |
| **9** | **Hash del Log Descifrado** | Calcular el hash SHA-256 de `logs_descifrado.json` y comparar bit a bit con el hash almacenado en el Paso 5. | Valor hash y booleano de comparación |
| **10** | **Presentación de Resultados** | Imprimir en la interfaz CLI un resumen estructurado: total de eventos, desglose por tipo, hashes generados en cada etapa y veredicto formal de integridad (`MATCH` / `MISMATCH`). | Salida en consola interactiva |

---

## Organización del Proyecto (Arquitectura)

Este repositorio sigue las directrices estándar de organización de proyectos en Go (*Standard Go Project Layout*), estructurado bajo el patrón de arquitectura limpia / por capas:

```text
taller-cripto-go/
├── cmd/
│   └── cli/
│       └── main.go                 # Punto de entrada de la aplicación CLI
├── internal/
│   ├── domain/                     # Capa de Dominio (Modelos e interfaces puras)
│   │   ├── event.go                # Structs de Registro, Modificación, Payload y Enums
│   │   └── repository.go           # Interfaces de persistencia y servicios
│   ├── generator/                  # Capa de Simulación / Negocio
│   │   └── generator.go            # Generación aleatoria de eventos (Paso 1 y 2)
│   ├── crypto/                     # Capa Criptográfica (Nativa de Go)
│   │   ├── hash.go                 # Implementación SHA-256 (Paso 3, 5, 7 y 9)
│   │   └── cipher.go               # Cifrado/Descifrado AES-256-GCM (Paso 6 y 8)
│   ├── storage/                    # Capa de Infraestructura / Persistencia en Reposo
│   │   └── file_storage.go         # Lectura y escritura de artefactos (.json, .enc, .hash)
│   └── ui/                         # Capa de Presentación / Orquestación CLI
│       ├── menu.go                 # Menú interactivo CLI con librería externa
│       └── presenter.go            # Formateo y renderizado de resultados (Paso 10)
├── artifacts/                      # Directorio de salida de archivos generados
│   ├── .gitkeep
│   ├── logs.json                   # Log consolidado en texto plano (Paso 4)
│   ├── logs_original.hash          # Hash del log en claro (Paso 5)
│   ├── logs.enc                    # Log cifrado en reposo (Paso 6)
│   ├── keystore.json               # Insumos criptográficos (Clave y Nonce) (Paso 6)
│   ├── logs_cifrado.hash           # Hash del artefacto cifrado (Paso 7)
│   └── logs_descifrado.json        # Log recuperado tras descifrar (Paso 8)
├── go.mod                          # Definición del módulo y dependencias
├── go.sum                          # Checksums de dependencias externas
└── README.md                       # Documentación técnica de ejecución y diseño
```

### Responsabilidad de cada Capa

1. **`cmd/cli/`**:
   - `main.go`: Contiene la función `main()`. Inicializa las dependencias, orquesta el ciclo de vida de la aplicación y transfiere el control a la capa de interfaz de usuario.
2. **`internal/domain/`**:
   - `event.go`: Define las estructuras de datos fundamentales (`LogEvent`, `Payload`, tipos de evento `REGISTRO_NOTA` y `MODIFICACION_NOTA`) sin dependencias externas.
   - `repository.go`: Establece los contratos e interfaces (`StorageRepository`, `CryptoService`, `GeneratorService`) para asegurar un bajo acoplamiento e inversión de dependencias (DIP).
3. **`internal/generator/`**:
   - `generator.go`: Implementa la simulación de notas aleatorias (distribución uniforme entre $0.0$ y $5.0$), asignación de UUIDs / UIDs para docentes, estudiantes y cursos, y la alternancia de eventos.
4. **`internal/crypto/`**:
   - `hash.go`: Implementa el cálculo de hashes SHA-256 tanto para cadenas y bytes de payloads como para flujos de lectura de archivos en disco (`crypto/sha256`).
   - `cipher.go`: Implementa el cifrado simétrico autenticado AES-GCM de 256 bits (`crypto/aes`, `crypto/cipher`) y generación de entropía segura (`crypto/rand`).
5. **`internal/storage/`**:
   - `file_storage.go`: Implementa la persistencia de datos en el sistema de archivos (`artifacts/`), gestionando la lectura y escritura de archivos JSON, binarios cifrados y archivos de hashes.
6. **`internal/ui/`**:
   - `menu.go`: Provee la interfaz interactiva basada en consola (utilizando bibliotecas modernas de terminal como `promptui` o `bubbletea`).
   - `presenter.go`: Renderiza el reporte final del Paso 10 con tablas estilizadas, resumen de eventos, trazabilidad de hashes y veredicto de integridad.
7. **`artifacts/`**:
   - Carpeta destinada exclusivamente al aislamiento y almacenamiento de los datos en reposo y sus artefactos de auditoría criptográfica.

---

## Especificación de la Estructura de Datos JSON

La especificación técnica del esquema de datos en reposo se basa en el documento de especificación formal (`estrucutra_json.pdf`):

### 1. Prevención de Circularidad en el Hash
Para garantizar la integridad individual sin caer en dependencias circulares:
- El campo `payload_hash` es un sello criptográfico que se calcula **estrictamente sobre los bytes serializados en formato canónico del sub-objeto `payload`**.
- Los metadatos de nivel superior (`id_evento`, `tipo`, `timestamp`) y el propio campo `payload_hash` **no forman parte** de la entrada del hash del payload.

### 2. Tabla de Campos del Evento de Log

| Campo / Ruta JSON | Tipo de Dato | Obligatorio | Descripción y Propósito en el Taller |
| :--- | :--- | :---: | :--- |
| `id_evento` | `string` (UUID v4) | Sí | Identificador único del evento de log para trazabilidad global. |
| `tipo` | `string` | Sí | Clasificación del evento: `"REGISTRO_NOTA"` o `"MODIFICACION_NOTA"`. |
| `timestamp` | `string` (ISO 8601 UTC) | Sí | Marca temporal de la operación (e.g., `"2026-10-26T14:30:00Z"`). |
| `payload` | `object` | Sí | Objeto canónico de negocio: agrupa los datos sensibles sobre los cuales se calcula el hash individual. |
| `payload.docente_id` | `string` (UUID / UID) | Sí | Identificador único o pseudónimo del profesor que ejecuta la acción. |
| `payload.estudiante_id` | `string` (UUID / UID) | Sí | Identificador único o pseudónimo del estudiante evaluado. |
| `payload.curso_id` | `string` | Sí | Identificador de la asignatura o grupo académico. |
| `payload.nota` | `float64` | Condicional | Calificación asignada ($0.0$ a $5.0$). Solo se usa si `tipo == "REGISTRO_NOTA"`. |
| `payload.nota_anterior` | `float64` | Condicional | Valor previo de la calificación ($0.0$ a $5.0$). Solo se usa si `tipo == "MODIFICACION_NOTA"`. |
| `payload.nota_nueva` | `float64` | Condicional | Nuevo valor de la calificación ($0.0$ a $5.0$). Solo se usa si `tipo == "MODIFICACION_NOTA"`. |
| `payload.motivo` | `string` | Opcional | Justificación del cambio de calificación (usado en modificaciones). |
| `payload_hash` | `string` (Hex SHA-256) | Sí | Sello de integridad individual: hash calculado estrictamente sobre los bytes serializados de `payload`. |

### 3. Ejemplo Concreto: Archivo Consolidado (`artifacts/logs.json`)

```json
[
  {
    "id_evento": "evt-7c91a4b2-e190-4c31-9981-d2f10b784a11",
    "tipo": "REGISTRO_NOTA",
    "timestamp": "2026-10-26T14:05:00Z",
    "payload": {
      "docente_id": "usr-prof-9831",
      "estudiante_id": "usr-est-4421",
      "curso_id": "mat-comp-101",
      "nota": 4.5
    },
    "payload_hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  },
  {
    "id_evento": "evt-9a81e3d4-f281-4b12-a124-c5a92d831e54",
    "tipo": "MODIFICACION_NOTA",
    "timestamp": "2026-10-26T14:07:30Z",
    "payload": {
      "docente_id": "usr-prof-9831",
      "estudiante_id": "usr-est-4421",
      "curso_id": "mat-comp-101",
      "nota_anterior": 4.5,
      "nota_nueva": 4.8,
      "motivo": "Correccion de evaluacion parcial"
    },
    "payload_hash": "a17f8d689b09bc5397444b0e8ef384d7a7ff25927c3d25dfb064c12ef0202d29"
  }
]
```

---

## Modelo Relacional de Entidades

La estructura de auditoría modela la interacción entre las entidades principales del sistema académico:

```text
  +--------------------------------+       +--------------------------------+       +--------------------------------+
  |            DOCENTE             |       |           ESTUDIANTE           |       |             CURSO              |
  +--------------------------------+       +--------------------------------+       +--------------------------------+
  | docente_id (PK, string / UUID) |       | estudiante_id(PK, string/UUID) |       | curso_id (PK, string / Código) |
  | nombre (string / seudónimo)    |       | nombre (string / seudónimo)    |       | nombre_materia (string)        |
  +--------------------------------+       +--------------------------------+       +--------------------------------+
                  |                                        |                                        |
                  | registra_o_modifica                    | sujeto_de                              | corresponde_a
                  \-----------------------------------\    |    /-----------------------------------/
                                                       v   v   v
                                             +--------------------+
                                             |     LOG_EVENTO     |
                                             +--------------------+
```

- **Docente:** Identifica a la autoridad académica responsable de registrar o alterar la calificación.
- **Estudiante:** Identifica al alumno evaluado al que pertenece el registro académico.
- **Curso:** Contextualiza la materia o asignatura donde se asienta la nota.
- **Log Evento:** Registra la transacción inmutable con su respectivo sello criptográfico.

---

## Mecanismos Criptográficos y Seguridad

1. **Integridad Individual (SHA-256):**
   - Garantiza que los campos del `payload` de una nota no hayan sido alterados de forma aislada.
2. **Integridad Global (SHA-256):**
   - Almacenado en `artifacts/logs_original.hash` para certificar el estado íntegro del consolidado antes del cifrado.
3. **Confidencialidad y Autenticación en Reposo (AES-256-GCM):**
   - Modo AEAD (*Authenticated Encryption with Associated Data*).
   - Generación de clave de 32 bytes (256 bits) y Nonce de 12 bytes mediante `crypto/rand`.
   - Incorpora una etiqueta de autenticación GMAC de 16 bytes que previene ataques de *bit-flipping* o manipulación de texto cifrado.
4. **Aislamiento de Dominios de Seguridad:**
   - **Plano de Datos:** Archivo binario cifrado (`logs.enc`).
   - **Plano de Verificación:** Hashes externos (`logs_original.hash`, `logs_cifrado.hash`).
   - **Plano de Gestión de Llaves:** Almacén de claves (`keystore.json`), simulando un KMS o HSM.

---

## Guía de Instalación y Ejecución

### Prerrequisitos

- **Go:** Versión 1.22 o superior instalada ([Descargar Go](https://go.dev/dl/)).
- Terminal con soporte para secuencias ANSI (PowerShell, Bash, Zsh, Windows Terminal).

### 1. Inicialización y Descarga de Dependencias

Desde la raíz del proyecto `taller-cripto-go/`:

```bash
# Descargar y sincronizar dependencias del proyecto
go mod tidy
```

### 2. Ejecución Directa

Para ejecutar la aplicación CLI interactiva sin compilar un binario permanente:

```bash
go run ./cmd/cli/main.go
```

### 3. Compilación a Binario Ejecutable

Si deseas compilar la aplicación para generar un binario independiente:

**En Linux / macOS:**
```bash
go build -o bin/taller-cripto-go ./cmd/cli
./bin/taller-cripto-go
```

**En Windows:**
```powershell
go build -o bin/taller-cripto-go.exe ./cmd/cli
.\bin\taller-cripto-go.exe
```

### 4. Salida Esperada en Consola (Paso 10)

Al finalizar el flujo interactivo de 10 pasos, el presentador de la CLI mostrará un reporte similar al siguiente:

```text
======================================================================
               REPORTE FINAL DE AUDITORÍA Y SEGURIDAD                 
======================================================================

[+] Paso 1: Volumen configurado: 2 eventos generados.
[+] Paso 2: Resumen de eventos:
    - 1 evento(s) de REGISTRO_NOTA
    - 1 evento(s) de MODIFICACION_NOTA

[+] Paso 3 y 4: Integridad de eventos y log consolidado:
    - Evento 1 Hash (SHA-256): e3b0c44298fc1c149afbf4c8996fb924...
    - Evento 2 Hash (SHA-256): a17f8d689b09bc5397444b0e8ef384d7...
    - Archivo consolidado generado: artifacts/logs.json

[+] Paso 5: Hash del log original (en claro):
    -> SHA-256: 7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9069
    -> Guardado en: artifacts/logs_original.hash

[+] Paso 6: Cifrado en reposo (AES-256-GCM):
    -> Archivo protegido generado: artifacts/logs.enc
    -> Clave y Nonce almacenados en: artifacts/keystore.json

[+] Paso 7: Hash del log cifrado:
    -> SHA-256: 3b10e74e444851219b1b68b75f56860d5b6e2d141e573e8633785be31b0e9c8a
    -> Guardado en: artifacts/logs_cifrado.hash

[+] Paso 8: Descifrado completado:
    -> Archivo recuperado: artifacts/logs_descifrado.json

[+] Paso 9: Comparación de integridad:
    - Hash Log Original:   7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9069
    - Hash Log Descifrado: 7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9069

----------------------------------------------------------------------
[RESULTADO]: COINCIDENCIA EXACTA (MATCH)
[CONCLUSIÓN]: El proceso de cifrado y descifrado preservó la 
              integridad y confidencialidad de la información original 
              sin alteraciones en reposo.
======================================================================
```

---

## Licencia MIT

Este proyecto se encuentra bajo los términos y condiciones de la **Licencia MIT**.

```text
MIT License

Copyright (c) 2026 Universidad del Valle

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
