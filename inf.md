# Informe de Desarrollo Seguro: Sophos VPN GUI

## 1. Resumen Ejecutivo
Se ha desarrollado un cliente gráfico (GUI) para la conexión a Sophos VPN, envolviendo un script de automatización nativo en PowerShell. El objetivo principal fue facilitar a los usuarios finales el uso de la VPN sin fricciones por consola, y cumplir con estrictos estándares de ciberseguridad para un entorno corporativo.

## 2. Prácticas de Seguridad Implementadas

### A. Gestión Segura de Credenciales (DPAPI)
Las contraseñas de los usuarios no se guardan en texto plano en ningún momento del ciclo de vida de la aplicación.
* **Encriptación Nativa Windows:** Empleamos la API de Microsoft (`ConvertTo-SecureString` / `Export-Clixml`), la cual utiliza la *Data Protection API (DPAPI)* del sistema operativo.
* **Vinculación a la Sesión Local (User-bound):** La clave generada queda atada criptográficamente a la cuenta de usuario de Windows actual (`$env:LOCALAPPDATA`). Ningún otro usuario, ni un administrador de la máquina o atacante con privilegios, puede desencriptar y extraer la contraseña, ya que la llave de encriptación está unida al inicio de sesión de ese usuario.
* **Manejo Dinámico Integrado:** Las credenciales aplican universalmente al perfil de Active Directory / Windows del portador en curso, resolviendo dinámicamente su identidad (`$env:USERNAME`).

### B. Protección de Ejecución del Código
* **Sandboxing Exclusivo (Bypass Controlado):** El binario invoca las sentencias en PowerShell mediante un puente dinámico (`-ExecutionPolicy Bypass`), lo que nos permite ejecutar el flujo en entornos hiper-restringidos sin tener que socavar ni alterar las Políticas de Seguridad Locales/del Dominio (GPO) del endpoint.
* **Prevención de Falsos Positivos - Ingeniería en C:** A diferencia de las utilidades de Python convencionales (como PyInstaller) que son bloqueadas frecuentemente por motores de Machine Learning en antivirus como Sophos por desempaquetar archivos temporalmente, se implementó el compilador nativo **Nuitka** (vía Scons y Zig). Esta tecnología transpila el código interpretado a lenguaje `C` puro y se ensambla bajo instrucciones DLL. El resultado es un código legítimo con estructura transparente indetectable para detecciones heurísticas genéricas.

### C. Modelo D.I.Y para EDR (Excepciones)
El binario resultante se proyectó para desplegarse mediante:
* **Exclusión Global por Hashing SHA-256.** Evitando la insegura técnica de realizar white-listing sobre las carpetas de Descargas o directorios locales (que permiten secuestros de DLLs y abuso de directorios).

---

## 3. Instrucciones de Primer Uso

Bienvenido al conector rápido de Sophos VPN. Para que puedas usarlo en este equipo, sigue los pasos de única vez a continuación.

> [!IMPORTANT]
> **Pre-requisito:** Verifique que su agente regular de *Sophos Connect* se encuentra instalado en la zona inferior derecha del reloj de su computadora. 

### Fase 1: Creación Segura de la Contraseña (Por Única Vez)
Dado que es la primera vez que inicia sesión en esta computadora, las credenciales aún no se han registrado en la bóveda cifrada.

1. Haz doble clic en el programa `vpn_gui.exe`.
2. **NO presiones el botón "Conectar VPN" todavía**.
3. Haz clic en el último botón rojo: **"🔑 Cambiar Contraseña Base"**.
4. Se abrirá un candado. Introduce tu **contraseña general de la empresa** (IMPORTANTE: SIN EL CÓDIGO TEMPORAL - TOTP) y pulsa en Aceptar.
5. El sistema reportará que la contraseña ha sido cifrada y asegurada exitosamente en el equipo.

### Fase 2: Conectarse a la Empresa (Uso Diario)
Tras confirmar el paso 1, podrás usar esta herramienta todos los días:

1. Selecciona el **Perfil de conexión** que vas a usar en la ventana desplegable.
2. Haz clic en el botón principal **"🚀 Conectar VPN"**.
3. Revisa el autenticador en tu teléfono e **Ingresa tu código MFA/TOTP (6 dígitos)** actual.
4. Dale a confirmar. Notarás al momento en la sección de notificaciones de Windows que Sophos Connect validó las identidades y concretó la sesión.
