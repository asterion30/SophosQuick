import tkinter as tk
from tkinter import ttk, simpledialog, messagebox
import subprocess
import os
import sys

# Función para obtener la ruta correcta del script si está empaquetado como binario (.exe) con PyInstaller
def get_resource_path(relative_path):
    try:
        # PyInstaller crea una carpeta temporal y guarda la ruta en sys._MEIPASS
        base_path = sys._MEIPASS
    except Exception:
        base_path = os.path.dirname(os.path.abspath(__file__))
    return os.path.join(base_path, relative_path)

# Ruta absoluta al script de PowerShell (soporta ejecución normal y empaquetada)
SCRIPT_PATH = get_resource_path("Conectar-SophosVPN.ps1")

# Configuracion de conexiones (Puedes agregar más o cambiar los nombres)
CONEXIONES = [
    "vpn.company.com",
    "vpn_backup.company.com"
]

def ejecutar_script(argumentos):
    # Comando base con el bypass de ejecución y apuntando al archivo
    comando = [
        "powershell.exe", 
        "-ExecutionPolicy", "Bypass", 
        "-NoProfile", 
        "-File", SCRIPT_PATH,
        "-ConnectionName", combo_conexiones.get()
    ] + argumentos
    
    # 0x08000000 corresponde a CREATE_NO_WINDOW en Windows (oculta la consola negra)
    creationflags = 0x08000000 
    
    try:
        # Usamos subprocess.run para esperar a que finalice y ver si hubo errores
        result = subprocess.run(
            comando, 
            creationflags=creationflags,
            capture_output=True,
            text=True
        )
        if result.returncode != 0:
            return False, result.stderr
        
        return True, result.stdout
    except Exception as e:
        return False, str(e)

# Funciones de los botones
def conectar():
    # Aparece un dialogo nativo de tkinter para pedir el TOTP
    totp = simpledialog.askstring("Autenticación MFA", "🔑 Ingresa tu código TOTP:", parent=root)
    if not totp:
        return # El usuario canceló
    
    exito, salida = ejecutar_script(["-TotpCode", totp])
    if exito:
        messagebox.showinfo("VPN", "Conexión en proceso...\nRevisa tu Sophos para confirmar que estás conectado.")
    else:
        messagebox.showerror("Error al conectar", f"Ocurrió un error:\n{salida}")

def desconectar():
    exito, salida = ejecutar_script(["-Desconectar"])
    if exito:
        messagebox.showinfo("VPN", "Se envió el comando de desconexión.\nSalida:\n" + (salida.strip() if salida else "Desconectado."))
    else:
        messagebox.showerror("Error de Desconexión", f"Ocurrió un error:\n{salida}")

def cambiar_password():
    password = simpledialog.askstring("Cambiar Contraseña Base", "🔑 Ingresa tu nueva contraseña base (sin TOTP):", show="*", parent=root)
    if not password:
        return
        
    exito, salida = ejecutar_script(["-Configurar", "-NewPasswordBase", password])
    if exito:
        messagebox.showinfo("Éxito", "¡Contraseña actualizada y guardada correctamente!")
    else:
        messagebox.showerror("Error al cambiar", f"No se pudo guardar la contraseña:\n{salida}")


# Construcción de la Interfaz Gráfica
root = tk.Tk()
root.title("Gestor Sophos VPN")
root.geometry("320x250")
root.resizable(False, False)

# Forzamos que la ventana principal se ubique al frente
root.attributes("-topmost", True)
root.after_idle(root.attributes, '-topmost', False)

tk.Label(root, text="Conexión Sophos VPN", font=("Arial", 14, "bold")).pack(pady=10)

# Agregando el selector de conexión
frame_combo = tk.Frame(root)
frame_combo.pack(pady=5)
tk.Label(frame_combo, text="Perfil:").pack(side=tk.LEFT)
combo_conexiones = ttk.Combobox(frame_combo, values=CONEXIONES, state="readonly", width=22)
combo_conexiones.current(0) # Selecciona el primero por defecto
combo_conexiones.pack(side=tk.LEFT, padx=5)

tk.Button(root, text="🚀 Conectar VPN", width=25, height=2, command=conectar).pack(pady=5)
tk.Button(root, text="🛑 Desconectar", width=25, command=desconectar).pack(pady=5)
tk.Button(root, text="🔑 Cambiar Contraseña Base", width=25, command=cambiar_password).pack(pady=5)

root.mainloop()
