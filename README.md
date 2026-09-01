# QuickConnect for Sophos VPN 🛡️🚀

> Conector gráfico ultra-rápido, moderno y seguro para **Sophos Connect VPN** escrito en Go.

![QuickConnect Dark Slate UI](https://img.shields.io/badge/UI-Dark%20Slate-0F172A?style=for-the-badge&logo=appveyor)
![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=for-the-badge&logo=go)
![Platform](https://img.shields.io/badge/Platform-Windows%2010%20%2F%2011-blue?style=for-the-badge&logo=windows)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

> ⚖️ **Aviso Legal / Trademark Disclaimer:**  
> *Sophos* y *Sophos Connect* son marcas comerciales registradas propiedad de **Sophos Ltd.**  
> Este proyecto (**QuickConnect for Sophos VPN**) es una herramienta independiente de código abierto y **no está afiliada, patrocinada, mantenida ni respaldada por Sophos Ltd.**

---

## 📋 Índice
- [Descripción General](#-descripción-general)
- [Características Destacadas](#-características-destacadas)
- [Diseño y Experiencia de Usuario](#-diseño-y-experiencia-de-usuario)
- [Estructura y Arquitectura](#-estructura-y-arquitectura)
- [Seguridad y Criptografía (DPAPI)](#-seguridad-y-criptografía-dpapi)
- [Guía de Uso Rápido](#-guía-de-uso-rápido)
  - [1. Configuración de Credenciales (Única vez)](#1-configuración-de-contraseña-base-única-vez)
  - [2. Conexión Diaria con MFA / TOTP](#2-conexión-diaria-con-mfa--totp)
- [Compilación y Empaquetado](#-compilación-y-empaquetado)
- [Automatización CI/CD (GitHub Actions)](#-automatización-cicd-github-actions)
- [Integración EDR y Antivirus](#-integración-edr-y-antivirus)
- [Aviso Legal](#-aviso-legal)

---

## 🌟 Descripción General

**QuickConnect for Sophos VPN** es una aplicación de escritorio nativa desarrollada en **Go** que interactúa con el motor de línea de comandos de Sophos (`sccli.exe`). Diseñada para erradicar la fricción en la conexión diaria a redes corporativas protegidas con autenticación de dos factores (2FA / TOTP), combina una interfaz gráfica minimalista con almacenamiento de credenciales de nivel empresarial mediante DPAPI de Windows.

---

## ✨ Características Destacadas

* **🎨 Estética Dark Slate Compacta:** Diseño estilo *"Floating Pill / Compact Modern Card"* (320 × 360 px) con paleta Slate (`#0F172A`, `#1E293B`) y acentos en Indigo (`#6366F1`) y Cyan (`#06B6D4`).
* **⚡ Flujo Ultra-Rápido con Tecla `Enter`:** Input MFA/TOTP enfocado automáticamente al iniciar; basta con escribir los 6 dígitos y presionar <kbd>Enter</kbd> para conectar.
* **🔄 Descubrimiento Dinámico de Perfiles:** Escaneo automático de conexiones configuradas en los directorios de Sophos Connect, con botón de recarga manual.
* **🔐 Bóveda Cifrada DPAPI:** Almacenamiento seguro de la contraseña corporativa ligado a la cuenta de usuario de Windows mediante la API criptográfica nativa del sistema operativo (`Export-Clixml` / DPAPI).
* **🔇 Cero Ventanas Negras:** Ejecución 100% silenciosa en segundo plano (`CREATE_NO_WINDOW` / `-H=windowsgui`).
* **📦 Binario Nativo Independiente:** Compilado a código máquina nativo en Go, sin intérpretes ni desempaquetadores temporales que disparen falsos positivos en EDR/antivirus.

---

## 🎨 Diseño y Experiencia de Usuario

La interfaz está organizada en una tarjeta moderna, estilizada y equilibrada:

```
┌──────────────────────────────────────────────────┐
│  QuickConnect                ⚪ Desconectado     │
│  for Sophos VPN                                  │
│  👤 Usuario: rortega (Personalizado)             │
├──────────────────────────────────────────────────┤
│  Perfil de Conexión:                             │
│  [ Mi-VPN-Corporativa                      ▼ ] 🔄 │
├──────────────────────────────────────────────────┤
│  Código MFA / TOTP (6 dígitos):                  │
│  [ 123456                                      ] │
│                                                  │
│  [ 🚀 Conectar VPN                             ] │
│  [ 🛑 Desconectar                              ] │
├──────────────────────────────────────────────────┤
│  [ 🔑 Configurar Credenciales / Usuario        ] │
│             Listo para conectar                  │
└──────────────────────────────────────────────────┘
```

---

## 🏗️ Estructura y Arquitectura

```
sophosquick/
├── .github/
│   └── workflows/
│       └── release.yml          # Pipeline CI/CD para compilación y releases
├── build/
│   └── windows/
│       ├── app.manifest         # Manifiesto Windows (DPI awareness, estilos modernos)
│       ├── build.ps1            # Script de compilación para PowerShell
│       ├── app.ico              # Icono nativo de la aplicación
│       └── versioninfo.json     # Metadatos del binario (versión, autor, copyright)
├── cmd/
│   └── sophosquick/
│       ├── main.go              # Punto de entrada principal
│       ├── app.ico              # Icono integrado
│       ├── app.manifest         # Manifiesto Win32
│       └── versioninfo.json     # Metadatos PE
├── internal/
│   ├── config/
│   │   └── config.go            # Manejo de preferencias y perfiles
│   ├── crypto/
│   │   ├── dpapi_windows.go     # Syscalls nativas a CryptProtectData (crypt32.dll)
│   │   └── credentials.go       # Bóveda cifrada en %LOCALAPPDATA%\SophosVPN_Cred.bin
│   ├── sophos/
│   │   ├── client.go            # Integración segura con sccli.exe
│   │   └── discovery.go         # Descubrimiento dinámico de perfiles
│   └── ui/
│       ├── app_windows.go       # GUI pura Win32 GDI/User32 (sin requerimiento OpenGL)
│       └── ui.go                # Interfaz agnóstica de UI
├── go.mod                       # Módulo Go
└── README.md                    # Documentación del proyecto
```

---

## 🔒 Seguridad y Criptografía (DPAPI)

1. **Sin Contraseñas en Texto Plano:** La contraseña base corporativa nunca se guarda en archivos desprotegidos ni variables de entorno.
2. **Atada a la Identidad del Usuario:** Emplea **Windows DPAPI (Data Protection API)** a través de llamadas nativas a `CryptProtectData` en `crypt32.dll`. La clave de descifrado está ligada a la sesión del usuario en `%LOCALAPPDATA%\SophosVPN_Cred.bin`.
3. **Imposible de Extraer por Terceros:** Ni administradores de red ni otros perfiles del equipo pueden descifrar las credenciales almacenadas.
4. **Sanitización en Memoria:** La concatenación `ContraseñaBase + TOTP` solo existe durante la fracción de segundo requerida por `sccli.exe` y las salidas en pantalla se filtran para no exponer claves.

---

## 🚀 Guía de Uso Rápido

### 1. Configuración de Credenciales (Única vez)
1. Ejecuta `quickconnect.exe`.
2. Haz clic en el botón inferior: **`🔑 Configurar Credenciales / Usuario`**.
3. Ingresa tu contraseña base corporativa (**SIN** el código MFA/TOTP).
4. *(Opcional)* Si tu usuario de la VPN difiere del usuario de Windows logueado, marca la casilla **"Personalizar usuario"** y escribe tu usuario corporativo.
5. Presiona **Guardar Credenciales**. Recibirás una confirmación visual de que han sido cifradas y almacenadas en la bóveda DPAPI.

### 2. Conexión Diaria con MFA / TOTP
1. Abre `quickconnect.exe` (el cursor se ubicará de inmediato en el campo de código MFA).
2. Abre tu app de autenticación (Microsoft Authenticator / Google Authenticator) en tu teléfono.
3. Escribe los **6 dígitos** del código TOTP y presiona <kbd>Enter</kbd> (o haz clic en **`🚀 Conectar VPN`**).
4. El botón indicará `⌛ Conectando...` y el estado cambiará a verde cuando la orden sea procesada por Sophos Connect.
5. Para desconectarte en cualquier momento, presiona **`🛑 Desconectar`**.

---

## 🛠️ Compilación y Empaquetado

### Compilación en Windows

#### Prerrequisitos:
* [Go 1.21+](https://go.dev/dl/)
* Herramienta de recursos de versión:
  ```powershell
  go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.4.0
  ```

#### Pasos de Compilación:
```powershell
# 1. Clonar el repositorio
git clone https://github.com/asterion30/SophosQuick.git
cd SophosQuick

# 2. Descargar dependencias
go mod tidy

# 3. Compilar automáticamente con el script incluido
.\build\windows\build.ps1 -Version "1.0.4"

# O compilar manualmente:
cd cmd\sophosquick
goversioninfo
go build -ldflags="-H=windowsgui -s -w" -o ../../dist/quickconnect.exe .
```

---

## 🤖 Automatización CI/CD (GitHub Actions)

El repositorio incluye un pipeline automatizado y seguro en [`.github/workflows/release.yml`](.github/workflows/release.yml).

### Cómo publicar una nueva versión:
1. Asegúrate de que todos los cambios estén commiteados en `main`.
2. Crea un tag de versión y súbelo a GitHub:
   ```bash
   git tag v1.0.4
   git push origin v1.0.4
   ```
3. GitHub Actions automáticamente:
   * Compilará el binario nativo optimizado `quickconnect.exe`.
   * Incrustará el manifiesto DPI-Aware v2, icono y metadatos de versión.
   * Generará el paquete comprimido `quickconnect-windows-amd64.zip`.
   * Calculará los hashes **SHA-256**.
   * Publicará una nueva **GitHub Release** con todos los activos listos para descarga.

---

## 🛡️ Integración EDR y Antivirus

| Característica | QuickConnect (Go Win32 Nativo) | Soluciones con Python / PyInstaller |
| :--- | :--- | :--- |
| **Arquitectura** | Binario PE x64 nativo directo | Auto-extractor en carpeta temporal (%TEMP%) |
| **Falsos Positivos** | 🟢 Mínimos / Nulos | 🔴 Frecuentes bloqueos heurísticos |
| **Distribución** | Paquete comprimido .ZIP con SHA-256 | Archivos ejecutables huérfanos |
| **Rendimiento** | Inicio instantáneo (<30ms) | 2 a 5 segundos de descompresión |
| **Requerimientos Gráficos** | 🟢 GDI puro (0% OpenGL/GPU) | 🔴 Requiere drivers acelerados |

---

## ⚖️ Aviso Legal / Trademark Disclaimer

*Sophos* y *Sophos Connect* son marcas comerciales registradas propiedad de **Sophos Ltd.**  
Este proyecto (**QuickConnect for Sophos VPN**) es una herramienta de software libre independiente y **no está afiliada, patrocinada, mantenida ni respaldada por Sophos Ltd.**

---

## 📄 Licencia

Distribuido bajo la Licencia **MIT**. Consulta el archivo [LICENSE](LICENSE) para más detalles.
