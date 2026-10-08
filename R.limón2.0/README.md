# 🐱 Limón Whiskas - Mini Juego Completo

Aplicación ejecutable para Windows ([`LimonWhiskas.exe`](file:///c:/work/R.limón2.0/LimonWhiskas.exe)) con icono oficial [`assets/Whiskas.ico`](file:///c:/work/R.limón2.0/assets/Whiskas.ico).

---

## 🎮 Fases del Juego

### 🏃 FASE 1: Limón Cazador sale de la Papelera
1. **Salida de la Papelera y Desorden en el Escritorio:**
   - Al ejecutar el juego, el gato Limón sale de la papelera de reciclaje.
   - Desordena elementos de la papelera arrojándolos por todo tu escritorio (`🗑️_papel_arrugado.txt`, `🗑️_espina_de_pescado.txt`, etc.).
2. **Caza del Cursor con Saltos y Sonido [`Audios/Whua.mp3`](file:///c:/work/R.limón2.0/Audios/Whua.mp3):**
   - Limón ([`assets/cazador.jpg`](file:///c:/work/R.limón2.0/assets/cazador.jpg)) se agacha, vibra y se lanza por el aire en un gran arco parabólico hacia tu mouse, agrandándose en pleno vuelo.
   - **En cada salto se reproduce instantáneamente `Whua.mp3`** mediante canales precargados para asegurar que suene en todos y cada uno de los saltos sin excepciones.
   - En la parte superior verás la cuenta regresiva de **1 minuto (01:00)**.
3. **Dos finales posibles en Fase 1:**
   - 🏆 **Si logras esquivarlo durante 1 minuto:** Limón se cansa, limpia el desorden del escritorio y se va a dormir plácidamente (`🐱💤`).
   - 💥 **Si Limón te atrapa el mouse:** ¡Te atrapa! Salta el mensaje de advertencia y se desata la **Fase 2**.

---

### 🥫 FASE 2: La Maldición de los Whiskas (Caos y Glitches)
1. **Cuenta Regresiva de 2 Minutos (02:00):** Tienes 2 minutos para calmar a Limón entregándole 3 whiskas.
2. **Música Continua en Bucle Nativo (F1.mp3 y F2.mp3):**
   - **02:00 a 01:01 (Antes de los 60 segundos):** Suena y se repite en bucle constante [`Audios/F1.mp3`](file:///c:/work/R.limón2.0/Audios/F1.mp3).
   - **Últimos 60 segundos (01:00 a 00:00):** Transiciona a [`Audios/F2.mp3`](file:///c:/work/R.limón2.0/Audios/F2.mp3) y se mantiene en bucle infinito nativo hasta que termine el juego.
   - **Últimos 5 segundos:** Suena `Audios/Limón.A.mp3` en simultáneo.
3. **Temblor y Teletransporte de Ventanas:**
   - La ventana principal tiembla y salta de posición cada pocos segundos.
   - Ventanas secundarias de glitch (`assets/glitch.jpg`) y error (`assets/error.jpg`) aparecen por toda la pantalla, tiemblan en su propio lugar, se duplican y se cierran solas.
4. **Búsqueda de Whiskas con Icono Oficial ([`assets/Whiskas.ico`](file:///c:/work/R.limón2.0/assets/Whiskas.ico) y [`assets/Whiskas.png`](file:///c:/work/R.limón2.0/assets/Whiskas.png)):**
   - Limón esconde 3 whiskas con el icono `Whiskas.ico` repartidos al azar en:
     - 📁 **Descargas**
     - 📁 **Documentos**
     - 📁 **Imágenes**
     - 📁 **Videos**
     - 📁 **Escritorio**
5. **Drag & Drop (Arrastrar y Soltar):** Arrastra los 3 whiskas hacia la ventana o hacia cualquiera de las ventanas secundarias.
6. **Protección de Cierre:** No puedes cerrar la ventana hasta que le des los 3 whiskas.
7. **Victoria:** Al completar la entrega, se detiene el caos, se limpian los archivos temporales y aparece la ventana de **"¡Gracias!"** con el botón **"Aceptar"**.

---

## 🚀 Cómo ejecutar
- Haz doble clic sobre [`LimonWhiskas.exe`](file:///c:/work/R.limón2.0/LimonWhiskas.exe).
- Para volver a compilar: ejecuta [`compilar.ps1`](file:///c:/work/R.limón2.0/compilar.ps1) en PowerShell.
