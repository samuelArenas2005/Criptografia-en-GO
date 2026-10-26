# SKILL SPECIFICATION: CLI de Protección de Logs Estructurados en Reposo (Go)

## 1. Visión General del Dominio
- **Institución:** Universidad del Valle
- **Contexto:** Sistema de gestión académica enfocado en operaciones sensibles (registro y modificación de notas).
- **Problema de Negocio:** Garantizar la confidencialidad, integridad, trazabilidad y no repudio de registros académicos en almacenamiento secundario (datos en reposo).
- **Objetivo Técnico:** Desarrollar una herramienta interactiva CLI en **Go** que simule eventos de notas, construya logs estructurados en JSON, aplique hashing a múltiples niveles y ejecute cifrado/descifrado simétrico autenticado para validar la preservación bit a bit del contenido original.

---

## 2. Restricciones Técnicas Mandatorias
1. **Lenguaje:** Go (`golang`) en versión reciente.
2. **Criptografía y Hashing:** Uso exclusivo de bibliotecas nativas de la biblioteca estándar de Go:
   - Hashing: `crypto/sha256` (SHA-256).
   - Cifrado Simétrico: `crypto/aes` y `crypto/cipher` (preferiblemente **AES-GCM** para proveer confidencialidad e integridad autenticada).
   - Generación de aleatoriedad criptográfica: `crypto/rand`.
3. **Interfaz de Usuario:** CLI interactiva guiada por menús, basada en bibliotecas de terceros para terminal (ej. `github.com/manifoldco/promptui` o `github.com/charmbracelet/bubbletea` / `huh`).
4. **Artefactos Separados:** Hashes globales, claves de cifrado y archivos binarios/cifrados deben almacenarse desacoplados del log original en texto claro.

---

## 3. Modelo de Datos y Estructura JSON

### 3.1. Tipos de Eventos Soportados
- `REGISTRO_NOTA`: Primer ingreso de calificación para un curso y estudiante.
- `MODIFICACION_NOTA`: Alteración justificada de una nota existente (debe registrar `nota_anterior` y `nota_nueva`).

### 3.2. Reglas del Modelo
- Uso de identificadores universales o pseudónimos (UUID / UID) para `actor_id` (docente/administrativo), `estudiante_id`, `curso_id` y `evento_id`.
- Calificaciones: Distribución uniforme entre $0.0$ y $5.0$. Formato numérico flotante estándar.
- Marcas de tiempo: ISO 8601 / RFC 3339 en UTC.

### 3.3. Estructura de Evento y Prevención de Circularidad en Hash
El payload debe ser una sub-estructura independiente. El hash del evento se calcula serializando **únicamente el payload** a bytes antes de inyectar el campo `payload_hash` al objeto principal del log, evitando dependencias circulares.

```json
{
  "id_evento": "uuid-v4-string",
  "tipo": "MODIFICACION_NOTA",
  "timestamp": "2026-10-26T14:30:00Z",
  "payload": {
    "docente_id": "usr-prof-8812",
    "estudiante_id": "usr-est-1049",
    "curso_id": "cur-mat-201",
    "nota_anterior": 3.2,
    "nota_nueva": 4.5,
    "motivo": "Correccion de evaluacion parcial"
  },
  "payload_hash": "sha256(payload_bytes_canonical)"
}
```

---

## 4. Flujo Operativo Paso a Paso (Los 10 Pasos)

| Paso | Nombre | Descripción Técnica | Artefacto de Salida |
| :--- | :--- | :--- | :--- |
| **1** | **Definición de volumen** | Solicitar al usuario vía CLI el número entero $N$ ($N \ge 1$) de eventos a simular. | Parámetro en memoria |
| **2** | **Generación aleatoria** | Generar $N$ eventos aleatorios alternando entre `REGISTRO_NOTA` y `MODIFICACION_NOTA`, con notas aleatorias en $[0.0, 5.0]$ y pseudónimos/UUIDs. | Colección de structs en memoria |
| **3** | **Hash por evento (Payload)** | Serializar de forma determinística (canónica) el sub-objeto `payload` de cada evento, aplicar `SHA-256`, y asignar el hash resultante en el campo `payload_hash`. | Structs enriquecidos |
| **4** | **Consolidación del Log** | Serializar la colección completa de eventos a un JSON estructurado legible y escribirlo a disco. | `logs.json` |
| **5** | **Hash del Log en Claro** | Calcular el hash SHA-256 del contenido completo de `logs.json` en texto plano y guardarlo por separado. | `logs_original.sha256` |
| **6** | **Cifrado en Reposo** | Cifrar `logs.json` usando un algoritmo simétrico nativo (AES-256-GCM). Generar clave criptográfica segura (32 bytes) y nonce (12 bytes) vía `crypto/rand`. Almacenar el archivo cifrado y los metadatos criptográficos en archivos separados. | `logs.enc` y `secret.key` / `crypto_metadata.json` |
| **7** | **Hash del Log Cifrado** | Calcular el hash SHA-256 del binario resultante `logs.enc` y guardarlo por separado. | `logs_cifrado.sha256` |
| **8** | **Descifrado del Log** | Leer `logs.enc`, cargar la clave/nonce y ejecutar el descifrado simétrico para reconstruir el archivo en texto claro. | `logs_descifrado.json` |
| **9** | **Hash del Log Descifrado** | Calcular el hash SHA-256 de `logs_descifrado.json` y comparar bit a bit con el hash almacenado en el Paso 5. | Valor hash y booleano de comparación |
| **10** | **Presentación de Resultados** | Imprimir en la interfaz CLI un resumen estructurado: total de eventos, desglose por tipo, hashes generados en cada paso y veredicto formal de integridad (`MATCH` / `MISMATCH`). | Salida en consola CLI interactiva |

---

## 5. Diseño Criptográfico y Modelo de Seguridad

### 5.1. Mecanismos Empleados
- **Integridad Local (Nivel Evento):** `SHA-256` sobre el payload canónico para prevenir manipulación individual no detectada y certificar qué datos fueron firmados internamente.
- **Integridad Global (Nivel Archivo):** `SHA-256` sobre el archivo consolidado para verificar el estado integral del volumen completo.
- **Confidencialidad y Autenticación en Reposo:** `AES-256-GCM` (Galois/Counter Mode).
  - Ventaja: Cifrado autenticado (AEAD) que previene tanto ataques de lectura espía como ataques de modificación de texto cifrado (bit-flipping), pues valida un tag de autenticación (GMAC) de 16 bytes.

### 5.2. Aislamiento y Artefactos Separados
El taller simula la separación de dominios de seguridad:
- **Plano de Datos:** Archivo cifrado `logs.enc`.
- **Plano de Verificación:** Hashes externos (`logs_original.sha256`, `logs_cifrado.sha256`).
- **Plano de Llaves:** Insumos criptográficos (`crypto_metadata.json`), que en producción residirían en un HSM o KMS.

---

## 6. Arquitectura de Software Sugerida (Go)

```text
├── cmd/
│   └── cli/
│       └── main.go          # Punto de entrada y orquestador del menú interactivo CLI
├── internal/
│   ├── crypto/
│   │   ├── cipher.go        # Funciones de cifrado y descifrado AES-GCM (crypto/aes, crypto/cipher)
│   │   └── hasher.go        # Funciones utilitarias para cálculo de SHA-256
│   ├── generator/
│   │   └── event_gen.go     # Lógica de generación sintética de notas y UIDs
│   ├── models/
│   │   └── event.go         # Structs Go de eventos, payloads y esquemas JSON
│   └── storage/
│       └── file_manager.go  # Lectura/escritura de artefactos JSON, binarios y hashes
├── docs/
│   ├── structural_diagram.png    # Diagrama de estructura (paquetes/componentes)
│   └── behavioral_diagram.png    # Diagrama de comportamiento (secuencia/actividad)
├── go.mod
└── go.sum
```

---

## 7. Criterios de Evaluación y Entregables
1. **Código Fuente:** Repositorio estructurado en Go respetando buenas prácticas (`idiomatic Go`), uso de bibliotecas nativas para criptografía y CLI interactiva.
2. **Diagramas de Diseño:**
   - 1 Diagrama de Estructura (Arquitectura de paquetes / Clases / Componentes).
   - 1 Diagrama de Comportamiento (Diagrama de secuencia o de flujo de los 10 pasos).
3. **Presentación (PDF):** Diapositivas justificando las decisiones de diseño, algoritmos criptográficos seleccionados, campos del payload y diapositiva obligatoria de uso responsable de Inteligencia Artificial.
4. **Sustentación:** Video demostrativo donde participen todos los miembros del equipo ejecutando la CLI y explicando la validación del proceso de preservación de información.