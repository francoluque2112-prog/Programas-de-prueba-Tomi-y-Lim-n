# 🐾 Tomi el Perro y la TV Antigua (Juego en Go)

Juego interactivo de escritorio desarrollado en **Go puro** con ventana transparente sobre Windows, soporte de **Drag & Drop real de archivos**, y el evento especial de transformación a **Hijack Tomi** (`Htomi.png`).

---

## 🎮 Mecánicas del Juego

### 1. 📺 Modo Normal
- **La TV Antigua (`Assets/TV antiguo.png`)**: Tomi duerme dentro de la TV. Al acercar el mouse, se despierta (`¡GUAU! (!)`), ladra y salta a la pantalla.
- **Persecución con 5s de retraso**: Tomi persigue el recorrido de tu mouse con 5 segundos de retraso. Si te detienes, él se detiene.
- **5 Carpetas en el Escritorio**: Arrastra los 5 archivos `Huesito_X_para_Tomi.txt` a la **Cuna (`Assets/Cuna.png`)** para ganar.

---

### 2. 🍋 El Evento Especial: `LimonWhiskas.exe` e Hijack Tomi
1. **Arrastra `LimonWhiskas.exe` a la TV**:
   - Si arrastras un archivo llamado `LimonWhiskas.exe` y lo sueltas sobre la **TV antigua**.
2. **Retroceso de Tomi con `Audios/Intro.mp3`**:
   - Empieza a escucharse **`Intro.mp3`**.
   - Tomi retrocede caminando de espaldas hasta meterse dentro de la pantalla del televisor.
3. **Reproducción del Video `Videos/Whua.mp4`**:
   - Una vez dentro, se reproduce el video **`Whua.mp4`** dentro de la pantalla de la TV antigua.
4. **Despertar de Hijack Tomi (`Assets/Htomi.png`) con `Audios/Chase.mp3`**:
   - Al finalizar el video, sale **Hijack Tomi** con ojos oscuros y pelaje brillante.
   - Empieza a sonar la música de persecución **`Chase.mp3`**.
5. **Comportamiento Feroz de Hijack Tomi**:
   - Reproduce el recorrido anterior del mouse con retraso, pero **mucho más rápido**.
   - **NO se detiene cuando dejas el mouse quieto**: sigue avanzando por el camino grabado hasta alcanzarte.
   - Tiene **pausas cortas y aleatorias** (glitch) y luego **cambia bruscamente de velocidad**.
   - Si te atrapa:
     > *"no me diste los huesitos por eso me comi tu mouse"*

---

## 🕹️ Controles

- **Arrastrar archivos de huesitos**: Llévalos a la Cuna.
- **Arrastrar `LimonWhiskas.exe`**: Suéltalo en la TV para iniciar la transformación.
- **Tecla [O]**: Abre rápidamente la carpeta del Escritorio en el Explorador de Windows.
- **Tecla [R]**: Reinicia la partida en cualquier momento.
- **Tecla [ESC]**: Cierra el juego limpiamente.

---

## 🚀 Cómo Jugar

Haz doble clic en [`jugar.bat`](file:///c:/work/T.tomi/jugar.bat) o ejecuta:

```powershell
.\tomi_game.exe
```
