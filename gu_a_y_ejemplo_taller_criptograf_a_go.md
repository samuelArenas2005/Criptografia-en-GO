# Guía y Ejemplo: Taller de Protección de Logs Estructurados en Reposo (Go)

Para este taller de la Universidad del Valle, el objetivo es construir una herramienta de línea de comandos (CLI) en Go que simule, proteja y verifique logs académicos de notas en reposo.

A continuación se explica paso a paso qué busca cada uno de los 10 pasos descritos en la guía y se presenta un ejemplo concreto del resultado que debe producir la aplicación.

---

## Explicación Paso a Paso de los 10 Pasos

### 1. Paso 1: Definición del volumen de eventos
- **Qué se busca:** A través del menú CLI interactivo, el sistema debe pedir al usuario que ingrese cuántos eventos de log desea generar (por ejemplo, 3, 5 o 100).

### 2. Paso 2: Generación aleatoria de eventos
- **Qué se busca:** Crear sintéticamente los eventos solicitados divididos entre dos tipos: registro de nota y modificación de nota. Deben usarse identificadores únicos/pseudónimos (como UUIDs) para profesor, estudiante, curso, etc. Las notas deben estar en un rango uniforme de $0.0$ a $5.0$, y en las modificaciones debe evidenciarse la nota anterior y la nueva.

### 3. Paso 3: Hash del payload de cada operación
- **Qué se busca:** Garantizar la integridad a nivel de registro individual. Se calcula un hash criptográfico (por ejemplo, con `crypto/sha256` nativo de Go) sobre los datos del evento (payload) y se incluye dentro del mismo JSON. Se debe cuidar evitar la circularidad (el campo del hash no puede incluirse dentro de los datos usados para calcularse a sí mismo).

### 4. Paso 4: Construcción del log estructurado completo
- **Qué se busca:** Consolidar todos los eventos individuales con sus respectivos hashes en un único archivo o estructura JSON final (por ejemplo, `logs.json`).

### 5. Paso 5: Hash del log completo en claro / texto plano
- **Qué se busca:** Obtener una "huella digital" del archivo original completo antes de cualquier alteración o cifrado. Este hash se almacena en un artefacto separado (por ejemplo, `logs_original.hash`) para futuras verificaciones.

### 6. Paso 6: Cifrado del log en reposo
- **Qué se busca:** Proteger la confidencialidad de los datos frente a accesos no autorizados. Usando librerías estándar de Go (como AES-GCM o AES-CBC con `crypto/aes` y `crypto/cipher`), se cifra el archivo `logs.json` y se guarda en un archivo separado (por ejemplo, `logs.enc`), junto con los insumos necesarios (clave, nonce/IV).

### 7. Paso 7: Hash del log cifrado
- **Qué se busca:** Calcular y almacenar por separado el hash del archivo resultante cifrado (`logs_cifrado.hash`) para comprobar la integridad del artefacto protegido.

### 8. Paso 8: Descifrado del log
- **Qué se busca:** Tomar el artefacto cifrado (`logs.enc`), aplicar la operación inversa con la clave/nonce y reconstruir el contenido del log original (por ejemplo, generando `logs_descifrado.json`).

### 9. Paso 9: Hash del log descifrado
- **Qué se busca:** Calcular el hash del archivo recuperado en el paso 8 y compararlo estrictamente con el hash del log original guardado en el paso 5. Deben ser idénticos bit a bit.

### 10. Paso 10: Presentación de resultados
- **Qué se busca:** Mostrar en la terminal un resumen detallado y visual que liste la cantidad de eventos, tipos generados, los hashes de cada etapa y la conclusión formal de si la información original fue preservada.

---

## Ejemplo Concreto del Resultado

### 1. Archivo generado: `logs.json` (Pasos 2, 3 y 4)

```json
[
  {
    "id_evento": "evt-b1c4e7a1-89d2",
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
    "id_evento": "evt-d5f8a2e9-11c3",
    "tipo": "MODIFICACION_NOTA",
    "timestamp": "2026-10-26T14:07:30Z",
    "payload": {
      "docente_id": "usr-prof-9831",
      "estudiante_id": "usr-est-4421",
      "curso_id": "mat-comp-101",
      "nota_anterior": 4.5,
      "nota_nueva": 4.8,
      "motivo": "Correccion reclamo punto 3"
    },
    "payload_hash": "a17f8d689b09bc5397444b0e8ef384d7a7ff25927c3d25dfb064c12ef0202d29"
  }
]
```

---

### 2. Salida en la Terminal CLI (Pasos 1 al 10)

Al ejecutar el flujo completo en la CLI con una librería de terminal interactiva, el reporte final del Paso 10 debe verse similar a esto:

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
- Archivo consolidado generado: logs.json

[+] Paso 5: Hash del log original (en claro):
    -> SHA-256: 7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9069

[+] Paso 6: Cifrado en reposo (AES-256-GCM):
    -> Archivo protegido generado: logs.enc
    -> Clave y Nonce almacenados en: keystore.json

[+] Paso 7: Hash del log cifrado:
    -> SHA-256: 3b10e74e444851219b1b68b75f56860d5b6e2d141e573e8633785be31b0e9c8a

[+] Paso 8: Descifrado completado:
    -> Archivo recuperado: logs_descifrado.json

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