# Modelado UML: Diagramas de Estructura y Comportamiento

> [!NOTE]
> **Fundamentación y Estándar de Modelado UML:**
> Este documento consolida la arquitectura del sistema de protección criptográfica de logs en reposo según los estándares de modelado UML y los requerimientos del taller de la Universidad del Valle.
> 
> - **Diagramas de Estructura (Estáticos):** Modelan la organización de paquetes, componentes modulares, interfaces, estructuras de datos (`structs`) y la segregación de planos de seguridad en almacenamiento secundario.
> - **Diagramas de Comportamiento (Dinámicos):** Modelan la interacción temporal y el flujo secuencial de los 10 pasos operativos mediante líneas de vida (*lifelines*), focos de activación y fragmentos combinados (`loop`, `alt / else`).

---

# PARTE 1: DIAGRAMAS DE ESTRUCTURA (ARQUITECTURA ESTÁTICA)

## 1.1. Diagrama de Componentes y Paquetes (Package & Component Diagram)

**Propósito:** Ilustrar la descomposición modular del sistema siguiendo las directrices de Arquitectura Limpia (*Clean Architecture*) e Inversión de Dependencias. La capa de presentación (`ui`) orquesta las capas de `generator`, `crypto` y `storage`, mientras que la capa `domain` contiene los modelos e interfaces desacopladas.

```mermaid
flowchart TD
    subgraph CMD["cmd/cli (Punto de Entrada)"]
        Main["main.go\n(Función main)"]
    end

    subgraph UI["internal/ui (Presentación e Interacción CLI)"]
        Menu["menu.go\n(Orquestador del Flujo de 10 Pasos)"]
        Presenter["presenter.go\n(Renderizador del Reporte y Explorer)"]
    end

    subgraph DOMAIN["internal/domain (Capa de Dominio Puro)"]
        direction TB
        LogEvent["Struct LogEvent"]
        Payload["Struct Payload"]
        CryptoService["Interface CryptoService"]
        StorageRepo["Interface StorageRepository"]
        GenService["Interface GeneratorService"]
    end

    subgraph CRYPTO["internal/crypto (Capa Criptográfica)"]
        Hash["hash.go\n(SHA-256 Payload & Archivos)"]
        Cipher["cipher.go\n(AES-256-GCM Cifrado/Descifrado)"]
        KeystoreStruct["Struct Keystore"]
    end

    subgraph GENERATOR["internal/generator (Capa de Simulación)"]
        Gen["generator.go\n(Generación Sintética de Notas)"]
    end

    subgraph STORAGE["internal/storage (Capa de Persistencia)"]
        FileStore["file_storage.go\n(E/S Archivos en Reposo)"]
    end

    subgraph ARTIFACTS["artifacts/ (Almacenamiento Físico Secundario)"]
        LogsJSON["logs.json"]
        LogsOrigHash["logs_original.hash"]
        LogsEnc["logs.enc"]
        KeystoreJSON["keystore.json"]
        LogsCifradoHash["logs_cifrado.hash"]
        LogsDecJSON["logs_descifrado.json"]
    end

    Main --> Menu
    Menu --> Presenter
    Menu --> Gen
    Menu --> Hash
    Menu --> Cipher
    Menu --> FileStore

    Gen -.->|Produce| LogEvent
    Hash -.->|Calcula payload_hash| Payload
    FileStore -->|Persiste / Lee| ARTIFACTS
    Cipher -->|Cifra en / Descifra de| ARTIFACTS
    Hash -->|Lee en streaming| ARTIFACTS

    CryptoService -.->|Implementado por| CRYPTO
    StorageRepo -.->|Implementado por| STORAGE
    GenService -.->|Implementado por| GENERATOR
```

---

## 1.2. Diagrama de Clases y Estructuras de Datos (Class & Struct Diagram)

**Propósito:** Modelar las estructuras de datos nativas en Go, los contratos de interfaz del dominio y las relaciones de composición y agregación entre los tipos del sistema.

```mermaid
classDiagram
    direction TB

    class LogEvent {
        +string IDEvento
        +string Tipo
        +string Timestamp
        +Payload Payload
        +string PayloadHash
    }

    class Payload {
        +string DocenteID
        +string EstudianteID
        +string CursoID
        +float64 Nota
        +float64 NotaAnterior
        +float64 NotaNueva
        +string Motivo
    }

    class Keystore {
        +string Key
        +string Nonce
    }

    class ExecutionResult {
        +int TotalEvents
        +int RegistroCount
        +int ModificacionCount
        +LogEvent[] Events
        +string OriginalLogPath
        +string OriginalHash
        +string EncryptedLogPath
        +string EncryptedHash
        +string KeystorePath
        +string DecryptedLogPath
        +string DecryptedHash
        +bool IsIntegrityValid
    }

    class CryptoService {
        <<interface>>
        +ComputePayloadHashes(events: LogEvent[]) LogEvent[]
        +ComputeFileHash(path: string) string
        +EncryptLog(plainPath: string, encPath: string, keystorePath: string) error
        +DecryptLog(encPath: string, keystorePath: string, decPath: string) error
        +CompareHashes(h1: string, h2: string) bool
    }

    class StorageRepository {
        <<interface>>
        +SaveLogsJSON(path: string, events: LogEvent[]) error
        +SaveHash(path: string, hash: string) error
    }

    class GeneratorService {
        <<interface>>
        +GenerateEvents(n: int) LogEvent[]
    }

    LogEvent *-- Payload : Composición (1 a 1)
    ExecutionResult o-- LogEvent : Agregación de eventos
    CryptoService ..> LogEvent : Calcula integridad de payload
    CryptoService ..> Keystore : Serializa credenciales
    StorageRepository ..> LogEvent : Serializa a disco
```

---

## 1.3. Diagrama de Estructura de Seguridad: Aislamiento de Planos en Reposo

**Propósito:** Representar el desacoplamiento arquitectónico de los tres dominios de seguridad en almacenamiento secundario (`artifacts/`): plano de datos, plano de verificación (hashes) y plano de gestión de llaves criptográficas (simulando un KMS/HSM).

```mermaid
flowchart LR
    subgraph PLAIN["Datos en Claro (Auditoría Inicial y Final)"]
        L1["artifacts/logs.json\n(Log Original en Claro)"]
        L2["artifacts/logs_descifrado.json\n(Log Recuperado post-descifrado)"]
    end

    subgraph DATA_PLANE["1. Plano de Datos Cifrados (Confidencialidad)"]
        ENC["artifacts/logs.enc\n(Ciphertext binario protegido con AES-256-GCM)"]
    end

    subgraph VERIFICATION_PLANE["2. Plano de Verificación (Integridad Multinivel)"]
        H1["artifacts/logs_original.hash\n(Huella SHA-256 original)"]
        H2["artifacts/logs_cifrado.hash\n(Huella SHA-256 del ciphertext)"]
    end

    subgraph KEY_PLANE["3. Plano de Gestión de Claves (Almacén Criptográfico)"]
        KEY["artifacts/keystore.json\n(Clave 256-bit y Nonce 96-bit en Hex)"]
    end

    L1 -->|AES-256-GCM Encrypt| ENC
    L1 -->|SHA-256 Streaming| H1
    ENC -->|SHA-256 Streaming| H2
    KEY -.->|Provee Clave y Nonce| ENC
    ENC -->|AES-256-GCM Decrypt + GMAC Verify| L2
    H1 -.->|Comparación bit a bit| L2
```

---

# PARTE 2: DIAGRAMAS DE COMPORTAMIENTO (SECUENCIA UML)

## 2.1. Proceso 1: Generación de Datos e Integridad en Claro (Pasos 1 al 5)

**Propósito:** Orquestar la solicitud del volumen $N$, la simulación sintética de eventos de calificaciones académicas, el cálculo del hash individual SHA-256 por cada `payload` (evitando circularidad criptográfica) y la persistencia del log consolidado junto con su huella digital original.

```mermaid
sequenceDiagram
    autonumber
    actor U as Usuario
    participant UI as Menu (UI)
    participant G as Generador
    participant H as Hash SHA-256 (Crypto)
    participant S as Almacenamiento (Storage)

    Note over U,S: Proceso 1 · Generación e integridad (Pasos 1-5)

    Note over U,G: Pasos 1-2 · Volumen y generación
    U->>+UI: ingresarCantidad(N)
    UI->>+G: generarEventos(N)
    loop [i = 1..N]
        G->>G: crearEvento(tipo, UIDs, nota)
    end
    G-->>-UI: eventos[]

    Note over UI,H: Paso 3 · Hash del payload
    loop [por cada evento]
        UI->>+H: hashPayload(payload)
        H-->>-UI: payload_hash
    end

    Note over UI,S: Pasos 4-5 · Log consolidado y hash en claro
    UI->>+S: guardarLog(eventos[])
    S-->>-UI: logs.json
    UI->>+H: hashArchivo(logs.json)
    H-->>-UI: hash_original
    UI->>+S: guardarHash(hash_original)
    S-->>-UI: logs_original.hash
    UI-->>-U: log generado y sellado
```

---

## 2.2. Proceso 2: Protección y Cifrado en Reposo (Pasos 6 y 7)

**Propósito:** Proteger la confidencialidad de los datos en disco mediante cifrado simétrico autenticado **AES-256-GCM**. Se genera entropía criptográfica para la clave y el nonce, desacoplando el plano de datos (`logs.enc`) del plano de gestión de llaves (`keystore.json`) y sellando la huella del archivo cifrado (`logs_cifrado.hash`).

```mermaid
sequenceDiagram
    autonumber
    participant UI as Menu (UI)
    participant S as Almacenamiento (Storage)
    participant C as Cifrador AES-256-GCM (Crypto)
    participant H as Hash SHA-256 (Crypto)

    Note over UI,H: Proceso 2 · Protección en reposo (Pasos 6-7)

    Note over UI,C: Paso 6 · Cifrado
    UI->>+S: leerLog(logs.json)
    S-->>-UI: texto_claro
    UI->>+C: cifrar(texto_claro)
    C->>C: generarClaveYNonce()
    C-->>-UI: logs.enc, clave, nonce
    UI->>+S: guardarCifrado(logs.enc)
    S-->>-UI: ok
    UI->>+S: guardarKeystore(clave, nonce)
    S-->>-UI: keystore.json

    Note over UI,H: Paso 7 · Hash del log cifrado
    UI->>+H: hashArchivo(logs.enc)
    H-->>-UI: hash_cifrado
    UI->>+S: guardarHash(hash_cifrado)
    S-->>-UI: logs_cifrado.hash
```

---

## 2.3. Proceso 3: Recuperación, Verificación y Auditoría Final (Pasos 8 al 10)

**Propósito:** Recuperar el contenido original aplicando descifrado autenticado. Mediante un fragmento condicional alternativo (`alt / else`), se verifica la etiqueta GMAC: si los datos o la clave fueron adulterados, el proceso aborta; si son válidos, se calcula el hash del archivo descifrado y se compara bit a bit contra el hash original, presentando el veredicto al usuario.

```mermaid
sequenceDiagram
    autonumber
    actor U as Usuario
    participant UI as Menu (UI)
    participant S as Almacenamiento (Storage)
    participant C as Cifrador AES-256-GCM (Crypto)
    participant H as Hash SHA-256 (Crypto)
    participant P as Presentador (UI)

    Note over U,P: Proceso 3 · Recuperación y verificación (Pasos 8-10)

    Note over UI,C: Paso 8 · Descifrado
    UI->>+S: leerCifradoYKeystore()
    S-->>-UI: logs.enc, clave, nonce
    UI->>C: descifrar(logs.enc, clave, nonce)

    alt [tag GCM inválido / datos adulterados]
        C-->>UI: error de autenticación
        UI->>P: mostrarError()
        P-->>U: archivo alterado o clave incorrecta
    else [tag GCM válido / autenticado]
        C-->>UI: texto_claro
        UI->>+S: guardarDescifrado(texto_claro)
        S-->>-UI: logs_descifrado.json

        Note over UI,H: Paso 9 · Verificación de integridad
        UI->>+H: hashArchivo(logs_descifrado.json)
        H-->>-UI: hash_descifrado
        UI->>UI: comparar(hash_descifrado, hash_original)

        Note over UI,U: Paso 10 · Resultados
        UI->>+P: mostrarResultados(N, tipos, hashes, veredicto)
        P-->>-U: reporte final (MATCH / MISMATCH)
    end
```
