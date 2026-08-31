param(
    [string]$ConnectionName = "vpn.vittal.com.ar", # Cambia esto por el nombre exacto de tu conexión en Sophos
    [string]$Username = $env:USERNAME,             # Usa el nombre del usuario de Windows actual
    [switch]$Configurar,                       # Usar este flag para guardar/cambiar la contraseña
    [switch]$Desconectar,                      # Usar este flag para desconectar
    [string]$TotpCode,                         # Permitir pasar el MFA como parametro
    [string]$NewPasswordBase                   # Permitir pasar la contraseña por parametro
)

$CredentialPath = "$env:LOCALAPPDATA\SophosVPN_Cred.xml"
$SccliPath = "C:\Program Files (x86)\Sophos\Connect\sccli.exe"

# Verificar si Sophos Connect CLI existe
if (-not (Test-Path $SccliPath)) {
    Write-Host "Error: No se encontró sccli.exe en $SccliPath. Verifica que Sophos Connect esté instalado." -ForegroundColor Red
    exit
}

# Opción para desconectar
if ($Desconectar) {
    Write-Host "Desconectando la VPN '$ConnectionName'..." -ForegroundColor Yellow
    # Utilizamos & para la ejecución directa sin romper el contexto de sesión de Python
    & $SccliPath disable -n $ConnectionName
    Write-Host "Comando finalizado."
    exit
}

# Opción para configurar/actualizar la contraseña
if ($Configurar -or -not (Test-Path $CredentialPath)) {
    Write-Host "=== Configuración de Credenciales ===" -ForegroundColor Cyan
    if ($NewPasswordBase) {
        $securePassword = ConvertTo-SecureString -String $NewPasswordBase -AsPlainText -Force
    } else {
        Write-Host "Se guardará tu contraseña base de forma encriptada (ligada a tu usuario de Windows)."
        $securePassword = Read-Host -Prompt "Ingresa tu contraseña base (SIN el código TOTP)" -AsSecureString
    }
    $securePassword | Export-Clixml -Path $CredentialPath
    Write-Host "Contraseña encriptada y guardada correctamente en $CredentialPath." -ForegroundColor Green
    
    if ($Configurar) {
        # Si solo queríamos configurar, terminamos aquí.
        exit
    }
}

# Leer la contraseña encriptada
try {
    $securePassword = Import-Clixml -Path $CredentialPath
    $bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePassword)
    $basePassword = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
}
finally {
    if ($bstr) {
        [System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
    }
}

# Solicitar el código TOTP al usuario
if ($TotpCode) {
    $totp = $TotpCode
} else {
    $totp = Read-Host -Prompt "Ingresa tu código TOTP (MFA)"
}

if ([string]::IsNullOrWhiteSpace($totp)) {
    Write-Host "El código TOTP no puede estar vacío. Cancelando..." -ForegroundColor Red
    exit
}

# Combinar la contraseña base con el código TOTP
$combinedPassword = "${basePassword}${totp}"

Write-Host "Conectando a la VPN '$ConnectionName' con el usuario '$Username'..." -ForegroundColor Yellow

# Ejecutar el comando de conexión de Sophos
# Ocultamos los errores comunes de PowerShell para que se vea la salida limpia del sccli
& $SccliPath enable -n $ConnectionName -u $Username -p $combinedPassword

# Limpieza en memoria por seguridad
$combinedPassword = $null
$basePassword = $null
$totp = $null

Write-Host "Proceso finalizado." -ForegroundColor Cyan
