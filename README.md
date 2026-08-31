# SophosQuick 🛡️🚀

> Conector gráfico ultra-rápido, moderno y seguro para **Sophos Connect VPN** escrito en Go.

![SophosQuick Dark Slate UI](https://img.shields.io/badge/UI-Dark%20Slate-0F172A?style=for-the-badge&logo=appveyor)
![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=for-the-badge&logo=go)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-blue?style=for-the-badge&logo=windows)
![License](https://img.shields.io/badge/License-Proprietary-red?style=for-the-badge)

---

## 📋 Índice
- [Descripción General](#-descripción-general)
- [Características Destacadas](#-características-destacadas)
- [Diseño y Experiencia de Usuario](#-diseño-y-experiencia-de-usuario)
- [Estructura y Arquitectura](#-estructura-y-arquitectura)
- [Seguridad y Criptografía (DPAPI)](#-seguridad-y-criptografía-dpapi)
- [Guía de Uso Rápido](#-guía-de-uso-rápido)
  - [1. Configuración de Contraseña Base (Única vez)](#1-configuración-de-contraseña-base-única-vez)
  - [2. Conexión Diaria con MFA / TOTP](#2-conexión-diaria-con-mfa--totp)
- [Compilación y Empaquetado](#-compilación-y-empaquetado)
  - [Compilación en Windows](#compilación-en-windows)
  - [Compilación Cruzada desde Linux](#compilación-cruzada-desde-linux)
- [Automatización CI/CD (GitHub Actions)](#-automatización-cicd-github-actions)
- [Integración EDR y Antivirus](#-integración-edr-y-antivirus)

---

## 🌟 Descripción General

**SophosQuick** es una aplicación de escritorio compacta desarrollada en **Go** que interactúa de manera nativa con el motor de línea de comandos de Sophos (`sccli.exe`). Diseñada para erradicar la fricción en la conexión diaria a redes corporativas protegidas con autenticación de dos factores (2FA / TOTP), combina una interfaz gráfica minimalista inspirada en Raycast/Linear con almacenamiento de credenciales de nivel empresarial.

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

La interfaz está organizada en una tarjeta moderna y equilibrada:

```
┌──────────────────────────────────────────────────┐
│  SophosQuick                 🟢 Desconectado     │
│  Secure VPN Launcher                             │
├──────────────────────────────────────────────────┤
│  Perfil de Conexión:                             │
│  [ vpn.company.com                         ▼ ] 🔄 │
├──────────────────────────────────────────────────┤
│  Código MFA / TOTP:                              │
│  [ 123456                                      ] │
│                                                  │
│  [ 🚀 Conectar VPN                             ] │
│  [ 🛑 Desconectar                              ] │
├──────────────────────────────────────────────────┤
│  [ 🔑 Configurar Contraseña Base               ] │
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
│       └── versioninfo.json     # Metadatos del binario (versión, autor, copyright)
├── cmd/
│   └── sophosquick/
│       └── main.go              # Punto de entrada principal
├── internal/
│   ├── config/
│   │   └── config.go            # Manejo de preferencias y perfiles por defecto
│   ├── crypto/
│   │   └── crypto.go            # Bóveda cifrada con Windows DPAPI
│   ├── sophos/
│   │   ├── client.go            # Cliente y lógica sccli.exe
│   │   ├── exec_windows.go      # Flags de ejecución silenciosa en Windows
│   │   └── exec_other.go        # Compatibilidad multiplataforma de desarrollo
│   └── ui/
│       ├── app.go               # Interfaz gráfica y controladores de eventos
│       └── theme.go             # Tema visual personalizado Dark Slate
├── go.mod                       # Módulo Go
└── README.md                    # Documentación del proyecto
```

---

## 🔒 Seguridad y Criptografía (DPAPI)

1. **Sin Contraseñas en Texto Plano:** La contraseña de red/Active Directory nunca se guarda en archivos desprotegidos ni en variables de entorno fijas.
2. **Atada a la Identidad del Usuario:** Emplea Windows **DPAPI (Data Protection API)** a través del formato seguro de PowerShell (`Export-Clixml`). La clave de descifrado está ligada a la sesión y hash del usuario en `%LOCALAPPDATA%\SophosVPN_Cred.xml`.
3. **Imposible de Extraer por Terceros:** Ni administradores locales de la máquina ni otros perfiles pueden descifrar las credenciales almacenadas.
4. **Limpieza en Memoria:** La concatenación de `ContraseñaBase + TOTP` solo existe durante la fracción de segundo requerida por `sccli.exe`.

---

## 🚀 Guía de Uso Rápido

### 1. Configuración de Contraseña Base (Única vez)
1. Ejecuta `sophosquick.exe`.
2. Haz clic en el botón inferior: **`🔑 Configurar Contraseña Base`**.
3. Ingresa tu contraseña de red/empresa (**SIN** el código MFA/TOTP) y presiona **Guardar**.
4. Recibirás una confirmación visual de que la contraseña ha sido cifrada y almacenada.

### 2. Conexión Diaria con MFA / TOTP
1. Abre `sophosquick.exe` (el cursor se ubicará de inmediato en el campo de código MFA).
2. Abre tu app de autenticación (Google Authenticator / Microsoft Authenticator) en tu teléfono.
3. Escribe los **6 dígitos** del código TOTP y presiona <kbd>Enter</kbd> (o haz clic en **`🚀 Conectar VPN`**).
4. El botón indicará `⌛ Conectando...` y el badge cambiará a verde cuando la orden sea procesada por Sophos Connect.

---

## 🛠️ Compilación y Empaquetado

### Compilación en Windows

#### Prerrequisitos:
* [Go 1.21+](https://go.dev/dl/)
* Compilador C (opcional para Fyne: [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) o `gcc` via MinGW64)
* Herramienta de recursos de versión:
  ```powershell
  go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
  ```

#### Pasos de Compilación:
```powershell
# 1. Clonar el repositorio
git clone https://github.com/tu-organizacion/sophosquick.git
cd sophosquick

# 2. Descargar dependencias
go mod tidy

# 3. Compilar automáticamente con el script incluido
.\build\windows\build.ps1 -Version "1.0.0"

# O compilar manualmente:
cd cmd\sophosquick
goversioninfo -manifest=../../build/windows/app.manifest ../../build/windows/versioninfo.json
go build -ldflags="-H=windowsgui -s -w" -o ../../dist/sophosquick.exe .
```

### Compilación Cruzada desde Linux

Para compilar el binario de Windows desde Linux usando `mingw-w64`:

```bash
# Instalar toolchain mingw en Debian/Ubuntu
sudo apt update && sudo apt install -y gcc-mingw-w64

# Compilar para Windows 64-bit
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
  go build -ldflags="-H=windowsgui -s -w" -o dist/sophosquick.exe ./cmd/sophosquick
```

---

## 🤖 Automatización CI/CD (GitHub Actions)

El repositorio incluye un flujo automatizado en [`.github/workflows/release.yml`](.github/workflows/release.yml).

### Cómo publicar una nueva versión:
1. Asegúrate de que todos los cambios estén commiteados en `main`.
2. Crea un tag de versión y súbelo a GitHub:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. GitHub Actions automáticamente:
   * Compilará el binario nativo de Windows en un runner limpio de `windows-latest`.
   * Incrustará el manifiesto de alta resolución y metadatos de versión.
   * Calculará el hash **SHA-256**.
   * Creará una nueva **GitHub Release** pública o privada adjuntando `sophosquick.exe` y `sophosquick.exe.sha256`.

---

## 🛡️ Integración EDR y Antivirus

| Característica | SophosQuick (Go Nativo) | Soluciones con PyInstaller / Python |
| :--- | :--- | :--- |
| **Arquitectura** | Binario PE x64 nativo directo | Auto-extractor en carpeta temporal (%TEMP%) |
| **Falsos Positivos** | 🟢 Mínimos / Nulos | 🔴 Frecuentes bloqueos heurísticos |
| **Whitelisting** | Hash SHA-256 único e inmutable | Difícil debido a DLLs dinámicas en runtime |
| **Rendimiento** | Inicio instantáneo (<50ms) | 2 a 5 segundos de descompresión |
| **Consola Oculta** | Subsystem Windows GUI nativo | Flags de proceso propensas a parpadeos |

---

## 📄 Licencia

Desarrollado para entornos corporativos y de infraestructura segura. Todos los derechos reservados.
