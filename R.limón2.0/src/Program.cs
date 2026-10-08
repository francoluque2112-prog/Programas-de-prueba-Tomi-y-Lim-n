using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Drawing2D;
using System.IO;
using System.Media;
using System.Reflection;
using System.Runtime.InteropServices;
using System.Text;
using System.Windows.Forms;
using Microsoft.Win32;

namespace LimonWhiskasGame
{
    static class Program
    {
        [STAThread]
        static void Main()
        {
            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);
            
            EmbeddedResourceManager.ExtractAll();
            WallpaperManager.BackupOriginalWallpaper();
            Application.ApplicationExit += (s, e) => 
            { 
                WallpaperManager.RestoreOriginalWallpaper();
                AudioEngine.CloseAll(); 
            };
            
            // Iniciar directamente con la fase del Gato Cazador saliendo de la papelera
            Application.Run(new HunterForm());
        }
    }

    public static class EmbeddedResourceManager
    {
        private static string tempResourceDir = null;

        public static string GetTempDirectory()
        {
            if (tempResourceDir == null)
            {
                tempResourceDir = Path.Combine(Path.GetTempPath(), "LimonWhiskasResources");
                if (!Directory.Exists(tempResourceDir))
                {
                    try { Directory.CreateDirectory(tempResourceDir); } catch { }
                }
            }
            return tempResourceDir;
        }

        private static string Normalize(string s)
        {
            if (string.IsNullOrEmpty(s)) return "";
            return s.Replace("á", "a").Replace("é", "e").Replace("í", "i").Replace("ó", "o").Replace("ú", "u")
                    .Replace("Á", "a").Replace("É", "e").Replace("Í", "i").Replace("Ó", "o").Replace("Ú", "u")
                    .ToLowerInvariant().Trim();
        }

        public static Stream GetResourceStream(string name)
        {
            try
            {
                Assembly asm = Assembly.GetExecutingAssembly();
                string[] names = asm.GetManifestResourceNames();
                string targetNorm = Normalize(name);

                foreach (string n in names)
                {
                    string nNorm = Normalize(n);
                    if (nNorm == targetNorm || nNorm.EndsWith("." + targetNorm))
                    {
                        return asm.GetManifestResourceStream(n);
                    }
                }
            }
            catch { }
            return null;
        }

        public static string GetExtractedFilePath(string fileName)
        {
            // 1. Si existe en la carpeta local de assets o Audios, usar esa
            string[] searchDirs = new string[]
            {
                Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "Audios"),
                Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "assets"),
                Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "audios"),
                Path.Combine(Directory.GetCurrentDirectory(), "Audios"),
                Path.Combine(Directory.GetCurrentDirectory(), "assets"),
                AppDomain.CurrentDomain.BaseDirectory,
                Directory.GetCurrentDirectory()
            };

            foreach (string dir in searchDirs)
            {
                try
                {
                    if (Directory.Exists(dir))
                    {
                        string p = Path.Combine(dir, fileName);
                        if (File.Exists(p) && new FileInfo(p).Length > 0) return Path.GetFullPath(p);
                    }
                }
                catch { }
            }

            // 2. Extraer desde el recurso incrustado en el propio EXE al directorio temporal
            try
            {
                string targetDir = GetTempDirectory();
                string targetPath = Path.Combine(targetDir, fileName);

                using (Stream stream = GetResourceStream(fileName))
                {
                    if (stream != null)
                    {
                        if (File.Exists(targetPath) && new FileInfo(targetPath).Length == stream.Length && stream.Length > 0)
                        {
                            return targetPath;
                        }

                        using (FileStream fs = new FileStream(targetPath, FileMode.Create, FileAccess.Write, FileShare.ReadWrite))
                        {
                            byte[] buffer = new byte[65536];
                            int read;
                            while ((read = stream.Read(buffer, 0, buffer.Length)) > 0)
                            {
                                fs.Write(buffer, 0, read);
                            }
                        }
                        return targetPath;
                    }
                }

                if (File.Exists(targetPath) && new FileInfo(targetPath).Length > 0)
                {
                    return targetPath;
                }
            }
            catch { }

            return null;
        }

        public static Image GetImage(string fileName)
        {
            // 1. Intentar directamente desde la memoria interna del EXE
            try
            {
                using (Stream stream = GetResourceStream(fileName))
                {
                    if (stream != null)
                    {
                        using (Image img = Image.FromStream(stream))
                        {
                            return new Bitmap(img);
                        }
                    }
                }
            }
            catch { }

            // 2. Fallback desde archivo en disco o extraído
            string path = GetExtractedFilePath(fileName);
            if (!string.IsNullOrEmpty(path) && File.Exists(path))
            {
                try
                {
                    using (Image img = Image.FromFile(path))
                    {
                        return new Bitmap(img);
                    }
                }
                catch { }
            }

            return null;
        }

        public static Icon GetIcon(string fileName)
        {
            // 1. Intentar directamente desde la memoria interna del EXE
            try
            {
                using (Stream stream = GetResourceStream(fileName))
                {
                    if (stream != null)
                    {
                        return new Icon(stream);
                    }
                }
            }
            catch { }

            // 2. Fallback desde archivo en disco o extraído
            string path = GetExtractedFilePath(fileName);
            if (!string.IsNullOrEmpty(path) && File.Exists(path))
            {
                try
                {
                    return new Icon(path);
                }
                catch { }
            }

            return null;
        }

        public static void ExtractAll()
        {
            string[] allResources = new string[]
            {
                "cazador.jpg", "glitch.jpg", "error.jpg",
                "Limon.png", "LimonF.png", "Whiskas.ico", "Whiskas.png",
                "CD-Limon.ico", "Whua.mp3", "F1.mp3", "F2.mp3", "Limon.A.mp3"
            };

            foreach (string res in allResources)
            {
                try
                {
                    GetExtractedFilePath(res);
                }
                catch { }
            }
        }
    }

    public static class WallpaperManager
    {
        [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
        private static extern int SystemParametersInfo(int uAction, int uParam, string lpvParam, int fuWinIni);

        [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
        private static extern int SystemParametersInfo(int uAction, int uParam, StringBuilder lpvParam, int fuWinIni);

        private const int SPI_SETDESKWALLPAPER = 20;
        private const int SPI_GETDESKWALLPAPER = 0x0073;
        private const int SPIF_UPDATEINIFILE = 0x01;
        private const int SPIF_SENDCHANGE = 0x02;

        private static string originalWallpaperPath = null;
        private static string originalWallpaperStyle = "2";
        private static string originalTileWallpaper = "0";
        private static bool isWallpaperChanged = false;
        private static string tempConvertedWallpaper = null;

        public static void BackupOriginalWallpaper()
        {
            if (originalWallpaperPath != null) return;

            try
            {
                using (RegistryKey key = Registry.CurrentUser.OpenSubKey(@"Control Panel\Desktop", false))
                {
                    if (key != null)
                    {
                        object val = key.GetValue("WallPaper");
                        if (val != null) originalWallpaperPath = val.ToString();

                        object style = key.GetValue("WallpaperStyle");
                        if (style != null) originalWallpaperStyle = style.ToString();

                        object tile = key.GetValue("TileWallpaper");
                        if (tile != null) originalTileWallpaper = tile.ToString();
                    }
                }
            }
            catch { }

            if (string.IsNullOrEmpty(originalWallpaperPath))
            {
                try
                {
                    StringBuilder sb = new StringBuilder(500);
                    if (SystemParametersInfo(SPI_GETDESKWALLPAPER, sb.Capacity, sb, 0) > 0)
                    {
                        originalWallpaperPath = sb.ToString();
                    }
                }
                catch { }
            }

            if (originalWallpaperPath == null) originalWallpaperPath = "";
        }

        public static void SetLimonWallpaper()
        {
            try
            {
                BackupOriginalWallpaper();

                string[] candidateNames = new string[] { "LimónF.png", "LimonF.png", "LimónF.jpg", "LimonF.jpg", "LimonF.bmp" };
                string foundPath = null;

                string[] searchDirs = new string[]
                {
                    Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "assets"),
                    Path.Combine(Directory.GetCurrentDirectory(), "assets"),
                    AppDomain.CurrentDomain.BaseDirectory,
                    Directory.GetCurrentDirectory()
                };

                foreach (string dir in searchDirs)
                {
                    if (Directory.Exists(dir))
                    {
                        foreach (string name in candidateNames)
                        {
                            string p = Path.Combine(dir, name);
                            if (File.Exists(p)) { foundPath = Path.GetFullPath(p); break; }
                        }
                    }
                    if (foundPath != null) break;
                }

                if (foundPath != null && File.Exists(foundPath))
                {
                    string tempDir = Path.GetTempPath();
                    tempConvertedWallpaper = Path.Combine(tempDir, "limon_wallpaper_temp.bmp");

                    using (Image img = Image.FromFile(foundPath))
                    {
                        img.Save(tempConvertedWallpaper, System.Drawing.Imaging.ImageFormat.Bmp);
                    }

                    try
                    {
                        using (RegistryKey key = Registry.CurrentUser.OpenSubKey(@"Control Panel\Desktop", true))
                        {
                            if (key != null)
                            {
                                key.SetValue("WallpaperStyle", "2"); // Stretch / Fill
                                key.SetValue("TileWallpaper", "0");
                            }
                        }
                    }
                    catch { }

                    SystemParametersInfo(SPI_SETDESKWALLPAPER, 0, tempConvertedWallpaper, SPIF_UPDATEINIFILE | SPIF_SENDCHANGE);
                    isWallpaperChanged = true;
                }
            }
            catch { }
        }

        public static void RestoreOriginalWallpaper()
        {
            if (!isWallpaperChanged) return;

            try
            {
                try
                {
                    using (RegistryKey key = Registry.CurrentUser.OpenSubKey(@"Control Panel\Desktop", true))
                    {
                        if (key != null)
                        {
                            if (!string.IsNullOrEmpty(originalWallpaperStyle))
                                key.SetValue("WallpaperStyle", originalWallpaperStyle);
                            if (!string.IsNullOrEmpty(originalTileWallpaper))
                                key.SetValue("TileWallpaper", originalTileWallpaper);
                        }
                    }
                }
                catch { }

                string pathToRestore = originalWallpaperPath ?? "";
                SystemParametersInfo(SPI_SETDESKWALLPAPER, 0, pathToRestore, SPIF_UPDATEINIFILE | SPIF_SENDCHANGE);
                isWallpaperChanged = false;
            }
            catch { }

            try
            {
                if (!string.IsNullOrEmpty(tempConvertedWallpaper) && File.Exists(tempConvertedWallpaper))
                {
                    File.Delete(tempConvertedWallpaper);
                }
            }
            catch { }
        }
    }

    public static class AudioEngine
    {
        [DllImport("winmm.dll", CharSet = CharSet.Auto)]
        private static extern long mciSendString(string strCommand, StringBuilder strReturn, int iReturnLength, IntPtr hwndCallback);

        [DllImport("kernel32.dll", CharSet = CharSet.Auto, SetLastError = true)]
        private static extern uint GetShortPathName(string lpszLongPath, StringBuilder lpszShortPath, uint cchBuffer);

        private static string pathWhua = null;
        private static string pathF1 = null;
        private static string pathF2 = null;
        private static string pathFA = null;

        private static string currentBgm = null;
        private static bool isFAActive = false;
        private static Timer loopTimer = null;

        private static string GetShortPath(string path)
        {
            try
            {
                if (string.IsNullOrEmpty(path)) return path;
                StringBuilder shortPath = new StringBuilder(260);
                if (GetShortPathName(path, shortPath, (uint)shortPath.Capacity) > 0)
                {
                    return shortPath.ToString();
                }
            }
            catch { }
            return path;
        }

        public static string FindAudio(string fileName)
        {
            string[] searchDirs = new string[]
            {
                Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "Audios"),
                Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "audios"),
                Path.Combine(Directory.GetCurrentDirectory(), "Audios"),
                Path.Combine(Directory.GetCurrentDirectory(), "audios"),
                AppDomain.CurrentDomain.BaseDirectory,
                Directory.GetCurrentDirectory()
            };

            foreach (string dir in searchDirs)
            {
                if (Directory.Exists(dir))
                {
                    string fullPath = Path.Combine(dir, fileName);
                    if (File.Exists(fullPath)) return Path.GetFullPath(fullPath);
                }
            }
            return null;
        }

        public static void InitAll()
        {
            CloseAll();

            pathWhua = FindAudio("Whua.mp3");
            pathF1 = FindAudio("F1.mp3");
            pathF2 = FindAudio("F2.mp3");
            pathFA = FindAudio("Limón.A.mp3");
            if (string.IsNullOrEmpty(pathFA)) pathFA = FindAudio("Limon.A.mp3");

            // Precargar el canal de salto Whua
            if (!string.IsNullOrEmpty(pathWhua) && File.Exists(pathWhua))
            {
                try
                {
                    string sWhua = GetShortPath(pathWhua);
                    mciSendString("close sfx_jump", null, 0, IntPtr.Zero);
                    mciSendString(string.Format("open \"{0}\" type mpegvideo alias sfx_jump", sWhua), null, 0, IntPtr.Zero);
                }
                catch { }
            }

            // Iniciar temporizador de bucle automático indestructible
            if (loopTimer == null)
            {
                loopTimer = new Timer();
                loopTimer.Interval = 200; // Monitorea cada 200ms
                loopTimer.Tick += (s, e) => { CheckAndLoopBgm(); };
            }
            loopTimer.Start();
        }

        public static void PlayJump()
        {
            try
            {
                if (string.IsNullOrEmpty(pathWhua) || !File.Exists(pathWhua))
                {
                    pathWhua = FindAudio("Whua.mp3");
                }

                if (!string.IsNullOrEmpty(pathWhua) && File.Exists(pathWhua))
                {
                    StringBuilder status = new StringBuilder(128);
                    mciSendString("status sfx_jump mode", status, status.Capacity, IntPtr.Zero);
                    if (string.IsNullOrEmpty(status.ToString()))
                    {
                        string sWhua = GetShortPath(pathWhua);
                        mciSendString(string.Format("open \"{0}\" type mpegvideo alias sfx_jump", sWhua), null, 0, IntPtr.Zero);
                    }
                    mciSendString("play sfx_jump from 0", null, 0, IntPtr.Zero);
                }
            }
            catch { }
        }

        public static void PlayF1()
        {
            if (currentBgm == "F1") return;
            currentBgm = "F1";
            isFAActive = false;

            try
            {
                mciSendString("stop bgm_a", null, 0, IntPtr.Zero);
                mciSendString("close bgm_a", null, 0, IntPtr.Zero);
                mciSendString("stop bgm", null, 0, IntPtr.Zero);
                mciSendString("close bgm", null, 0, IntPtr.Zero);

                if (string.IsNullOrEmpty(pathF1) || !File.Exists(pathF1))
                {
                    pathF1 = FindAudio("F1.mp3");
                }

                if (!string.IsNullOrEmpty(pathF1) && File.Exists(pathF1))
                {
                    string sF1 = GetShortPath(pathF1);
                    mciSendString(string.Format("open \"{0}\" type mpegvideo alias bgm", sF1), null, 0, IntPtr.Zero);
                    mciSendString("play bgm from 0", null, 0, IntPtr.Zero);
                }
            }
            catch { }
        }

        public static void PlayF2()
        {
            if (currentBgm == "F2" && !isFAActive) return;
            currentBgm = "F2";
            isFAActive = false;

            try
            {
                mciSendString("stop bgm_a", null, 0, IntPtr.Zero);
                mciSendString("close bgm_a", null, 0, IntPtr.Zero);
                mciSendString("stop bgm", null, 0, IntPtr.Zero);
                mciSendString("close bgm", null, 0, IntPtr.Zero);

                if (string.IsNullOrEmpty(pathF2) || !File.Exists(pathF2))
                {
                    pathF2 = FindAudio("F2.mp3");
                }

                if (!string.IsNullOrEmpty(pathF2) && File.Exists(pathF2))
                {
                    string sF2 = GetShortPath(pathF2);
                    mciSendString(string.Format("open \"{0}\" type mpegvideo alias bgm", sF2), null, 0, IntPtr.Zero);
                    mciSendString("play bgm from 0", null, 0, IntPtr.Zero);
                }
            }
            catch { }
        }

        public static void PlayF2WithA()
        {
            if (currentBgm != "F2")
            {
                PlayF2();
            }

            if (!isFAActive)
            {
                isFAActive = true;
                try
                {
                    mciSendString("stop bgm_a", null, 0, IntPtr.Zero);
                    mciSendString("close bgm_a", null, 0, IntPtr.Zero);

                    if (string.IsNullOrEmpty(pathFA) || !File.Exists(pathFA))
                    {
                        pathFA = FindAudio("Limon.A.mp3");
                        if (string.IsNullOrEmpty(pathFA)) pathFA = FindAudio("Limón.A.mp3");
                    }

                    if (!string.IsNullOrEmpty(pathFA) && File.Exists(pathFA))
                    {
                        string sFA = GetShortPath(pathFA);
                        mciSendString(string.Format("open \"{0}\" type mpegvideo alias bgm_a", sFA), null, 0, IntPtr.Zero);
                        mciSendString("play bgm_a from 0", null, 0, IntPtr.Zero);
                    }
                }
                catch { }
            }
        }

        private static void CheckAndLoopBgm()
        {
            if (string.IsNullOrEmpty(currentBgm)) return;

            // Revisar si la pista BGM principal terminó
            try
            {
                StringBuilder sb = new StringBuilder(128);
                mciSendString("status bgm mode", sb, sb.Capacity, IntPtr.Zero);
                string mode = sb.ToString().Trim().ToLowerInvariant();

                if (mode == "stopped")
                {
                    mciSendString("play bgm from 0", null, 0, IntPtr.Zero);
                }
            }
            catch { }

            // Revisar si la pista secundaria A terminó
            if (isFAActive)
            {
                try
                {
                    StringBuilder sb = new StringBuilder(128);
                    mciSendString("status bgm_a mode", sb, sb.Capacity, IntPtr.Zero);
                    string mode = sb.ToString().Trim().ToLowerInvariant();

                    if (mode == "stopped")
                    {
                        mciSendString("play bgm_a from 0", null, 0, IntPtr.Zero);
                    }
                }
                catch { }
            }
        }

        public static void StopBgm()
        {
            currentBgm = null;
            isFAActive = false;
            try
            {
                mciSendString("stop bgm", null, 0, IntPtr.Zero);
                mciSendString("close bgm", null, 0, IntPtr.Zero);
                mciSendString("stop bgm_a", null, 0, IntPtr.Zero);
                mciSendString("close bgm_a", null, 0, IntPtr.Zero);
            }
            catch { }
        }

        public static void CloseAll()
        {
            currentBgm = null;
            isFAActive = false;
            try
            {
                if (loopTimer != null) loopTimer.Stop();
                mciSendString("close all", null, 0, IntPtr.Zero);
            }
            catch { }
        }
    }

    // ==========================================
    // FASE 1: GATO CAZADOR EN EL ESCRITORIO
    // ==========================================
    public class HunterForm : Form
    {
        private Image cazadorImg;
        private PictureBox picCat;
        private Form hudForm;
        private Label lblHud;

        private int remainingSeconds = 60; // 1 minuto para esquivar
        private Timer chaseTimer;
        private Timer physicsTimer;

        private double posX, posY;
        private double startJumpX, startJumpY;
        private double targetX, targetY;
        private double jumpHeight = 120.0;
        private int totalJumpFrames = 20;
        private int currentJumpFrame = 0;

        private int baseSize = 200;
        private int currentWidth = 200;
        private int currentHeight = 200;

        private enum JumpState { Preparing, Leaping, Landing }
        private JumpState currentState = JumpState.Preparing;
        private int stateTicks = 0;
        private Random rnd = new Random();

        public static List<string> DesktopTrashFiles = new List<string>();

        public HunterForm()
        {
            this.FormBorderStyle = FormBorderStyle.None;
            this.ShowInTaskbar = false;
            this.TopMost = true;
            this.Size = new Size(baseSize, baseSize);
            this.StartPosition = FormStartPosition.Manual;
            this.BackColor = Color.Magenta;
            this.TransparencyKey = Color.Magenta;

            AudioEngine.InitAll();

            LoadCazadorImage();
            ScatterRecycleBinTrashOnDesktop();
            CreateHudBanner();

            Rectangle screen = Screen.PrimaryScreen.WorkingArea;
            posX = screen.Left + 40;
            posY = screen.Top + 60;
            this.Location = new Point((int)posX, (int)posY);

            picCat = new PictureBox();
            picCat.Dock = DockStyle.Fill;
            picCat.SizeMode = PictureBoxSizeMode.Zoom;
            picCat.Image = cazadorImg;
            picCat.BackColor = Color.Transparent;
            this.Controls.Add(picCat);

            chaseTimer = new Timer();
            chaseTimer.Interval = 1000;
            chaseTimer.Tick += ChaseTimer_Tick;

            physicsTimer = new Timer();
            physicsTimer.Interval = 20; // 50 FPS para fluidez en el salto
            physicsTimer.Tick += PhysicsTimer_Tick;

            this.Load += (s, e) =>
            {
                hudForm.Show();
                chaseTimer.Start();
                physicsTimer.Start();
                try { SystemSounds.Asterisk.Play(); } catch { }
            };

            this.FormClosing += (s, e) =>
            {
                chaseTimer.Stop();
                physicsTimer.Stop();
                AudioEngine.CloseAll();
                if (hudForm != null && !hudForm.IsDisposed)
                {
                    hudForm.Close();
                }
            };
        }

        private void LoadCazadorImage()
        {
            cazadorImg = EmbeddedResourceManager.GetImage("cazador.jpg");
            if (cazadorImg != null) return;

            Bitmap bmp = new Bitmap(200, 200);
            using (Graphics g = Graphics.FromImage(bmp))
            {
                g.Clear(Color.DarkOrange);
                using (Font f = new Font("Segoe UI", 16F, FontStyle.Bold))
                {
                    g.DrawString("🐱 CAZADOR", f, Brushes.White, 15, 80);
                }
            }
            cazadorImg = bmp;
        }

        private void ScatterRecycleBinTrashOnDesktop()
        {
            try
            {
                string desktop = Environment.GetFolderPath(Environment.SpecialFolder.Desktop);
                if (!Directory.Exists(desktop)) return;

                string[] trashNames = new string[]
                {
                    "🗑️_papel_arrugado.txt",
                    "🗑️_cascara_de_platano.txt",
                    "🗑️_espina_de_pescado.txt",
                    "🗑️_lata_de_atun_vacia.txt",
                    "🗑️_calcetin_viejo.txt",
                    "🗑️_caja_de_pizza_vacia.txt",
                    "🗑️_basura_tirada_por_limon.txt"
                };

                string[] trashContents = new string[]
                {
                    "🐱 ¡Limón vació la papelera de reciclaje y tiró esto en tu escritorio mientras saltaba con todo el cuerpo!",
                    "🍌 ¡Una cáscara de plátano que estaba en la papelera!",
                    "🐟 ¡Limón buscó pescado en la papelera y desordenó todo!",
                    "🥫 Lata vacía... ¡Limón quiere whiskas de verdad, no latas vacías de la basura!",
                    "🧦 Un calcetín que estaba en la papelera.",
                    "🍕 Restos de pizza sacados de la papelera por Limón.",
                    "🗑️ ¡Todo el escritorio quedó desordenado porque Limón salió de la papelera a cazar tu mouse!"
                };

                for (int i = 0; i < trashNames.Length; i++)
                {
                    string filePath = Path.Combine(desktop, trashNames[i]);
                    File.WriteAllText(filePath, trashContents[i]);
                    DesktopTrashFiles.Add(filePath);
                }
            }
            catch { }
        }

        public static void CleanupDesktopTrash()
        {
            foreach (string file in DesktopTrashFiles)
            {
                try
                {
                    if (File.Exists(file))
                    {
                        File.Delete(file);
                    }
                }
                catch { }
            }
            DesktopTrashFiles.Clear();
        }

        private void CreateHudBanner()
        {
            hudForm = new Form();
            hudForm.FormBorderStyle = FormBorderStyle.None;
            hudForm.ShowInTaskbar = false;
            hudForm.TopMost = true;
            hudForm.BackColor = Color.FromArgb(30, 30, 30);
            hudForm.Size = new Size(580, 50);
            hudForm.StartPosition = FormStartPosition.Manual;

            Rectangle screen = Screen.PrimaryScreen.WorkingArea;
            hudForm.Location = new Point(screen.Left + (screen.Width - hudForm.Width) / 2, screen.Top + 15);

            lblHud = new Label();
            lblHud.Text = "🏃 ¡ESQUIVA A LIMÓN CAZADOR! ⏱️ Tiempo restante: 01:00";
            lblHud.Font = new Font("Segoe UI", 13F, FontStyle.Bold);
            lblHud.ForeColor = Color.FromArgb(255, 204, 0);
            lblHud.Dock = DockStyle.Fill;
            lblHud.TextAlign = ContentAlignment.MiddleCenter;
            hudForm.Controls.Add(lblHud);
        }

        private void ChaseTimer_Tick(object sender, EventArgs e)
        {
            remainingSeconds--;
            int mins = remainingSeconds / 60;
            int secs = remainingSeconds % 60;
            lblHud.Text = string.Format("🏃 ¡ESQUIVA A LIMÓN CAZADOR! ⏱️ Tiempo restante: {0:D2}:{1:D2}", mins, secs);

            if (remainingSeconds <= 10)
            {
                lblHud.ForeColor = Color.FromArgb(255, 75, 75);
            }

            if (remainingSeconds <= 0)
            {
                chaseTimer.Stop();
                physicsTimer.Stop();
                AudioEngine.CloseAll();
                hudForm.Hide();
                this.Hide();

                MessageBox.Show(
                    "¡Increíble! 🏆\n\n¡Esquivaste los saltos feroces de Limón durante 1 minuto completo!\nSe cansó de saltar, limpió su desorden y se fue a dormir. 🐱💤",
                    "🐱 ¡Limón se fue a dormir!",
                    MessageBoxButtons.OK,
                    MessageBoxIcon.Information
                );

                CleanupDesktopTrash();
                AudioEngine.CloseAll();
                this.Close();
                Environment.Exit(0);
            }
        }

        private void PhysicsTimer_Tick(object sender, EventArgs e)
        {
            Point mouse = Cursor.Position;
            stateTicks++;

            Rectangle catHitbox = new Rectangle((int)posX + 25, (int)posY + 25, Math.Max(10, currentWidth - 50), Math.Max(10, currentHeight - 50));
            if (catHitbox.Contains(mouse))
            {
                OnCatCaughtMouse();
                return;
            }

            switch (currentState)
            {
                case JumpState.Preparing:
                    currentWidth = baseSize + 25;
                    currentHeight = baseSize - 35;

                    if (stateTicks > rnd.Next(7, 14)) // 140 - 280 ms
                    {
                        currentState = JumpState.Leaping;
                        stateTicks = 0;
                        currentJumpFrame = 0;

                        // Reproducir Whua.mp3 instantáneamente en CADA salto
                        AudioEngine.PlayJump();

                        startJumpX = posX;
                        startJumpY = posY;

                        targetX = mouse.X - (baseSize / 2);
                        targetY = mouse.Y - (baseSize / 2);

                        double dx = targetX - startJumpX;
                        double dy = targetY - startJumpY;
                        double dist = Math.Sqrt(dx * dx + dy * dy);

                        totalJumpFrames = Math.Max(14, Math.Min(26, (int)(dist / 28)));
                        jumpHeight = Math.Max(80.0, Math.Min(220.0, dist * 0.35));
                    }
                    break;

                case JumpState.Leaping:
                    currentJumpFrame++;
                    double progress = (double)currentJumpFrame / (double)totalJumpFrames;
                    if (progress > 1.0) progress = 1.0;

                    double currentBaseX = startJumpX + (targetX - startJumpX) * progress;
                    double currentBaseY = startJumpY + (targetY - startJumpY) * progress;

                    double arc = Math.Sin(progress * Math.PI) * jumpHeight;
                    posX = currentBaseX;
                    posY = currentBaseY - arc;

                    double scaleFactor = 1.0 + Math.Sin(progress * Math.PI) * 0.55;
                    currentWidth = (int)(baseSize * scaleFactor);
                    currentHeight = (int)(baseSize * scaleFactor);

                    if (currentJumpFrame >= totalJumpFrames)
                    {
                        currentState = JumpState.Landing;
                        stateTicks = 0;
                    }
                    break;

                case JumpState.Landing:
                    currentWidth = baseSize + 35;
                    currentHeight = baseSize - 25;

                    if (stateTicks > rnd.Next(4, 8)) // 80 - 160 ms
                    {
                        currentState = JumpState.Preparing;
                        stateTicks = 0;
                    }
                    break;
            }

            if (this.Size.Width != currentWidth || this.Size.Height != currentHeight)
            {
                this.Size = new Size(currentWidth, currentHeight);
            }

            Rectangle screen = Screen.PrimaryScreen.WorkingArea;
            posX = Math.Max(screen.Left - 50, Math.Min(screen.Right - currentWidth + 50, posX));
            posY = Math.Max(screen.Top - 50, Math.Min(screen.Bottom - currentHeight + 50, posY));

            this.Location = new Point((int)posX, (int)posY);
        }

        private void OnCatCaughtMouse()
        {
            chaseTimer.Stop();
            physicsTimer.Stop();
            hudForm.Hide();
            this.Hide();

            // Cambiar fondo de pantalla al asset LimónF.png
            WallpaperManager.SetLimonWallpaper();

            try { SystemSounds.Hand.Play(); } catch { }

            MessageBox.Show(
                "¡LIMÓN TE ATRAPÓ EN PLENO SALTO! 🐱💥🐾\n\n¡Saltó con todo su cuerpo y no pudiste esquivarlo!\n¡AHORA DAME MIS 3 WHISKAS ANTES DE QUE SE ACABE EL TIEMPO!",
                "🐱 ¡ATRAPADO POR LIMÓN!",
                MessageBoxButtons.OK,
                MessageBoxIcon.Warning
            );

            // Iniciar la Fase 2 con su propio ciclo
            using (MainForm mainGame = new MainForm())
            {
                mainGame.ShowDialog();
            }

            WallpaperManager.RestoreOriginalWallpaper();
            CleanupDesktopTrash();
            AudioEngine.CloseAll();
            this.Close();
            Environment.Exit(0);
        }
    }

    // ==========================================
    // FASE 2: GESTOR DE VENTANAS GLITCH / ERROR
    // ==========================================
    public static class GlitchManager
    {
        public static List<Image> GlitchImages = new List<Image>();
        public static List<GlitchForm> ActiveGlitches = new List<GlitchForm>();
        public static MainForm MainRef = null;
        private static Random rnd = new Random();
        public const int MaxConcurrentGlitches = 12;

        public static void LoadGlitchImages()
        {
            GlitchImages.Clear();
            string[] imageNames = new string[] { "glitch.jpg", "error.jpg", "glitch.png", "error.png" };
            string[] searchDirs = new string[]
            {
                Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "assets"),
                Path.Combine(Directory.GetCurrentDirectory(), "assets"),
                AppDomain.CurrentDomain.BaseDirectory,
                Directory.GetCurrentDirectory()
            };

            foreach (string dir in searchDirs)
            {
                if (Directory.Exists(dir))
                {
                    foreach (string name in imageNames)
                    {
                        string path = Path.Combine(dir, name);
                        if (File.Exists(path))
                        {
                            try
                            {
                                using (var img = Image.FromFile(path))
                                {
                                    GlitchImages.Add(new Bitmap(img));
                                }
                            }
                            catch { }
                        }
                    }
                }
            }

            if (GlitchImages.Count == 0)
            {
                GlitchImages.Add(CreateFallbackGlitchImage("GLITCH 0x000F9", Color.Black, Color.Magenta));
                GlitchImages.Add(CreateFallbackGlitchImage("FATAL ERROR: NO WHISKAS", Color.DarkRed, Color.White));
            }
        }

        private static Image CreateFallbackGlitchImage(string text, Color bg, Color fg)
        {
            Bitmap bmp = new Bitmap(260, 180);
            using (Graphics g = Graphics.FromImage(bmp))
            {
                g.Clear(bg);
                using (Font f = new Font("Consolas", 12F, FontStyle.Bold))
                using (Brush b = new SolidBrush(fg))
                {
                    g.DrawString(text, f, b, 15, 60);
                    g.DrawString("🐱 LIMÓN CORRUPTED", f, Brushes.Yellow, 15, 90);
                }
            }
            return bmp;
        }

        public static void SpawnGlitch()
        {
            if (MainRef == null || MainRef.IsDisposed || !MainRef.IsGameActive) return;
            if (ActiveGlitches.Count >= MaxConcurrentGlitches) return;
            if (GlitchImages.Count == 0) return;

            try
            {
                Image chosenImg = GlitchImages[rnd.Next(GlitchImages.Count)];
                GlitchForm gf = new GlitchForm(chosenImg);
                ActiveGlitches.Add(gf);
                gf.Show();
            }
            catch { }
        }

        public static void RemoveGlitch(GlitchForm gf)
        {
            if (ActiveGlitches.Contains(gf))
            {
                ActiveGlitches.Remove(gf);
            }
        }

        public static void CloseAllGlitches()
        {
            var list = new List<GlitchForm>(ActiveGlitches);
            ActiveGlitches.Clear();
            foreach (var gf in list)
            {
                try
                {
                    if (!gf.IsDisposed)
                    {
                        gf.Close();
                        gf.Dispose();
                    }
                }
                catch { }
            }
        }
    }

    public class GlitchForm : Form
    {
        private Point anchorLocation;
        private Timer shakeTimer;
        private Timer lifetimeTimer;
        private Timer duplicateTimer;
        private static Random rnd = new Random();
        private PictureBox picBox;

        public GlitchForm(Image img)
        {
            this.FormBorderStyle = FormBorderStyle.FixedToolWindow;
            this.ShowInTaskbar = false;
            this.TopMost = false;
            this.AllowDrop = true;
            this.BackColor = Color.Black;

            string[] titles = new string[] { "⚠️ GLITCH_LIMÓN.EXE", "❌ FATAL_ERROR", "🐱 WHISKAS_NOT_FOUND", "⚠️ CORRUPCIÓN", "⚡ LIMÓN GLITCH", "ERROR: 0x999WHISKAS" };
            this.Text = titles[rnd.Next(titles.Length)];

            int w = rnd.Next(220, 340);
            int h = (int)(w * 0.72);
            this.Size = new Size(w, h);
            this.StartPosition = FormStartPosition.Manual;

            Rectangle screenBounds = Screen.PrimaryScreen.WorkingArea;
            int maxX = Math.Max(0, screenBounds.Width - w);
            int maxY = Math.Max(0, screenBounds.Height - h);
            anchorLocation = new Point(screenBounds.Left + rnd.Next(0, maxX), screenBounds.Top + rnd.Next(0, maxY));
            this.Location = anchorLocation;

            picBox = new PictureBox();
            picBox.Dock = DockStyle.Fill;
            picBox.SizeMode = PictureBoxSizeMode.Zoom;
            picBox.Image = img;
            picBox.BackColor = Color.Black;
            picBox.AllowDrop = true;
            this.Controls.Add(picBox);

            this.DragEnter += Glitch_DragEnter;
            this.DragDrop += Glitch_DragDrop;
            picBox.DragEnter += Glitch_DragEnter;
            picBox.DragDrop += Glitch_DragDrop;

            shakeTimer = new Timer();
            shakeTimer.Interval = 35;
            shakeTimer.Tick += (s, e) =>
            {
                if (this.IsDisposed) return;
                int ox = rnd.Next(-5, 6);
                int oy = rnd.Next(-5, 6);
                this.Location = new Point(anchorLocation.X + ox, anchorLocation.Y + oy);
            };
            shakeTimer.Start();

            lifetimeTimer = new Timer();
            lifetimeTimer.Interval = rnd.Next(1200, 4500);
            lifetimeTimer.Tick += (s, e) =>
            {
                lifetimeTimer.Stop();
                this.Close();
            };
            lifetimeTimer.Start();

            duplicateTimer = new Timer();
            duplicateTimer.Interval = rnd.Next(900, 3000);
            duplicateTimer.Tick += (s, e) =>
            {
                duplicateTimer.Stop();
                if (rnd.Next(0, 100) < 65)
                {
                    GlitchManager.SpawnGlitch();
                }
            };
            duplicateTimer.Start();

            this.FormClosed += (s, e) =>
            {
                shakeTimer.Stop();
                shakeTimer.Dispose();
                lifetimeTimer.Stop();
                lifetimeTimer.Dispose();
                duplicateTimer.Stop();
                duplicateTimer.Dispose();
                GlitchManager.RemoveGlitch(this);
            };
        }

        private void Glitch_DragEnter(object sender, DragEventArgs e)
        {
            if (GlitchManager.MainRef != null && !GlitchManager.MainRef.IsDisposed)
            {
                if (e.Data.GetDataPresent(DataFormats.FileDrop))
                {
                    e.Effect = DragDropEffects.Copy;
                    return;
                }
            }
            e.Effect = DragDropEffects.None;
        }

        private void Glitch_DragDrop(object sender, DragEventArgs e)
        {
            if (GlitchManager.MainRef != null && !GlitchManager.MainRef.IsDisposed)
            {
                GlitchManager.MainRef.ProcessDroppedFiles(e);
            }
        }
    }

    // ==========================================
    // FASE 2: VENTANA PRINCIPAL DE WHISKAS
    // ==========================================
    public class MainForm : Form
    {
        private int remainingSeconds = 120; // 2 minutos
        private int deliveredWhiskas = 0;
        private const int RequiredWhiskas = 3;
        private bool gameCompleted = false;
        private bool gameOver = false;

        public bool IsGameActive { get { return !gameCompleted && !gameOver; } }

        private Timer gameTimer;
        private Timer shakeTimer;
        private Timer jumpTimer;
        private Timer glitchSpawnTimer;

        private Point baseAnchorLocation;
        private Random rnd = new Random();
        private int shakeIntensity = 4;

        private Label lblTitle;
        private Label lblTimer;
        private Label lblCounter;
        private Label lblStatus;
        private PictureBox picLimon;
        private Panel dropZonePanel;
        private Label lblDropHint;
        private Panel slotsPanel;
        private Image whiskasImage = null;

        private string whiskasIcoPath = null;
        private string whiskasSourceWebp = null;

        private List<string> createdWhiskasFiles = new List<string>();
        private List<string> chosenFolderNames = new List<string>();
        private HashSet<string> deliveredFilePaths = new HashSet<string>(StringComparer.OrdinalIgnoreCase);

        public MainForm()
        {
            GlitchManager.MainRef = this;
            InitializeComponents();
            InitAssets();
            GlitchManager.LoadGlitchImages();
            SpawnWhiskasFiles();
            StartGame();
        }

        private void InitAssets()
        {
            whiskasIcoPath = EmbeddedResourceManager.GetExtractedFilePath("Whiskas.ico");
            whiskasSourceWebp = EmbeddedResourceManager.GetExtractedFilePath("Whiskas.png");
            whiskasImage = EmbeddedResourceManager.GetImage("Whiskas.png");
        }

        private void InitializeComponents()
        {
            this.Text = "🔒 Tus Archivos han sido encriptados";
            this.Size = new Size(820, 680);
            this.StartPosition = FormStartPosition.CenterScreen;
            this.FormBorderStyle = FormBorderStyle.FixedDialog;
            this.MaximizeBox = false;
            this.MinimizeBox = true;
            this.BackColor = Color.FromArgb(255, 250, 245);
            this.AllowDrop = true;
            this.Font = new Font("Segoe UI", 10F, FontStyle.Regular);

            try
            {
                Icon appIcon = EmbeddedResourceManager.GetIcon("CD-Limón.ico");
                if (appIcon == null) appIcon = EmbeddedResourceManager.GetIcon("CD-Limon.ico");
                if (appIcon == null) appIcon = EmbeddedResourceManager.GetIcon("Whiskas.ico");
                if (appIcon != null) this.Icon = appIcon;
            }
            catch { }

            this.Load += (s, e) =>
            {
                baseAnchorLocation = this.Location;
            };

            this.DragEnter += MainForm_DragEnter;
            this.DragDrop += MainForm_DragDrop;
            this.FormClosing += MainForm_FormClosing;

            Panel headerPanel = new Panel();
            headerPanel.Dock = DockStyle.Top;
            headerPanel.Height = 118;
            headerPanel.BackColor = Color.FromArgb(255, 235, 235);
            headerPanel.BorderStyle = BorderStyle.FixedSingle;
            headerPanel.Padding = new Padding(12, 6, 12, 8);
            headerPanel.AllowDrop = true;
            headerPanel.DragEnter += MainForm_DragEnter;
            headerPanel.DragDrop += MainForm_DragDrop;

            TableLayoutPanel headerTable = new TableLayoutPanel();
            headerTable.Dock = DockStyle.Fill;
            headerTable.RowCount = 2;
            headerTable.ColumnCount = 1;
            headerTable.RowStyles.Add(new RowStyle(SizeType.Absolute, 38F));
            headerTable.RowStyles.Add(new RowStyle(SizeType.Percent, 100F));
            headerTable.AllowDrop = true;
            headerTable.DragEnter += MainForm_DragEnter;
            headerTable.DragDrop += MainForm_DragDrop;

            lblTitle = new Label();
            lblTitle.Text = "🔒 ¡Tus Archivos han sido encriptados! 🔒";
            lblTitle.Font = new Font("Segoe UI", 16F, FontStyle.Bold);
            lblTitle.ForeColor = Color.FromArgb(192, 41, 43);
            lblTitle.TextAlign = ContentAlignment.MiddleCenter;
            lblTitle.Dock = DockStyle.Fill;
            lblTitle.Margin = new Padding(0);
            lblTitle.AllowDrop = true;
            lblTitle.DragEnter += MainForm_DragEnter;
            lblTitle.DragDrop += MainForm_DragDrop;
            headerTable.Controls.Add(lblTitle, 0, 0);

            Panel descContainer = new Panel();
            descContainer.Dock = DockStyle.Fill;
            descContainer.BackColor = Color.White;
            descContainer.BorderStyle = BorderStyle.FixedSingle;
            descContainer.Margin = new Padding(4, 2, 4, 0);
            descContainer.Padding = new Padding(10, 4, 10, 4);
            descContainer.AllowDrop = true;
            descContainer.DragEnter += MainForm_DragEnter;
            descContainer.DragDrop += MainForm_DragDrop;

            Label lblDescription = new Label();
            lblDescription.Text = "📄 Descripción: Si no me das los whiskas hasta que se acabe el tiempo tus archivos seran irrecuperables atravez de cualquier medio.";
            lblDescription.Font = new Font("Segoe UI", 10F, FontStyle.Bold);
            lblDescription.ForeColor = Color.FromArgb(140, 20, 20);
            lblDescription.TextAlign = ContentAlignment.MiddleCenter;
            lblDescription.Dock = DockStyle.Fill;
            lblDescription.AllowDrop = true;
            lblDescription.DragEnter += MainForm_DragEnter;
            lblDescription.DragDrop += MainForm_DragDrop;
            descContainer.Controls.Add(lblDescription);

            headerTable.Controls.Add(descContainer, 0, 1);
            headerPanel.Controls.Add(headerTable);

            this.Controls.Add(headerPanel);

            TableLayoutPanel mainLayout = new TableLayoutPanel();
            mainLayout.Dock = DockStyle.Fill;
            mainLayout.ColumnCount = 2;
            mainLayout.RowCount = 1;
            mainLayout.ColumnStyles.Add(new ColumnStyle(SizeType.Percent, 46F));
            mainLayout.ColumnStyles.Add(new ColumnStyle(SizeType.Percent, 54F));
            mainLayout.Padding = new Padding(15, 5, 15, 15);
            mainLayout.AllowDrop = true;
            mainLayout.DragEnter += MainForm_DragEnter;
            mainLayout.DragDrop += MainForm_DragDrop;
            this.Controls.Add(mainLayout);

            Panel leftPanel = new Panel();
            leftPanel.Dock = DockStyle.Fill;
            leftPanel.Padding = new Padding(5);
            leftPanel.AllowDrop = true;
            leftPanel.DragEnter += MainForm_DragEnter;
            leftPanel.DragDrop += MainForm_DragDrop;

            picLimon = new PictureBox();
            picLimon.Dock = DockStyle.Fill;
            picLimon.SizeMode = PictureBoxSizeMode.Zoom;
            picLimon.BackColor = Color.FromArgb(250, 240, 230);
            picLimon.BorderStyle = BorderStyle.FixedSingle;
            picLimon.AllowDrop = true;
            picLimon.DragEnter += MainForm_DragEnter;
            picLimon.DragDrop += MainForm_DragDrop;
            LoadLimonImage();

            leftPanel.Controls.Add(picLimon);
            mainLayout.Controls.Add(leftPanel, 0, 0);

            Panel rightPanel = new Panel();
            rightPanel.Dock = DockStyle.Fill;
            rightPanel.Padding = new Padding(10, 0, 10, 0);
            rightPanel.AllowDrop = true;
            rightPanel.DragEnter += MainForm_DragEnter;
            rightPanel.DragDrop += MainForm_DragDrop;

            lblTimer = new Label();
            lblTimer.Text = "⏳ Cuenta Regresiva: 02:00";
            lblTimer.Font = new Font("Segoe UI", 16F, FontStyle.Bold);
            lblTimer.ForeColor = Color.FromArgb(192, 41, 43);
            lblTimer.AutoSize = false;
            lblTimer.Height = 35;
            lblTimer.Dock = DockStyle.Top;
            lblTimer.TextAlign = ContentAlignment.MiddleCenter;
            rightPanel.Controls.Add(lblTimer);

            lblCounter = new Label();
            lblCounter.Text = "Whiskas entregados: 0 / 3";
            lblCounter.Font = new Font("Segoe UI", 13F, FontStyle.Bold);
            lblCounter.ForeColor = Color.FromArgb(44, 62, 80);
            lblCounter.AutoSize = false;
            lblCounter.Height = 30;
            lblCounter.Dock = DockStyle.Top;
            lblCounter.TextAlign = ContentAlignment.MiddleCenter;
            rightPanel.Controls.Add(lblCounter);

            slotsPanel = new Panel();
            slotsPanel.Dock = DockStyle.Top;
            slotsPanel.Height = 50;
            slotsPanel.Padding = new Padding(10, 5, 10, 5);
            slotsPanel.AllowDrop = true;
            slotsPanel.DragEnter += MainForm_DragEnter;
            slotsPanel.DragDrop += MainForm_DragDrop;

            TableLayoutPanel slotsTable = new TableLayoutPanel();
            slotsTable.Dock = DockStyle.Fill;
            slotsTable.ColumnCount = 3;
            slotsTable.RowCount = 1;
            slotsTable.ColumnStyles.Add(new ColumnStyle(SizeType.Percent, 33.33F));
            slotsTable.ColumnStyles.Add(new ColumnStyle(SizeType.Percent, 33.33F));
            slotsTable.ColumnStyles.Add(new ColumnStyle(SizeType.Percent, 33.33F));
            slotsTable.AllowDrop = true;
            slotsTable.DragEnter += MainForm_DragEnter;
            slotsTable.DragDrop += MainForm_DragDrop;

            for (int i = 0; i < 3; i++)
            {
                Label slotLbl = new Label();
                slotLbl.Text = "🥫 (Vacío)";
                slotLbl.Font = new Font("Segoe UI", 9.5F, FontStyle.Regular);
                slotLbl.ForeColor = Color.Gray;
                slotLbl.TextAlign = ContentAlignment.MiddleCenter;
                slotLbl.Dock = DockStyle.Fill;
                slotLbl.BorderStyle = BorderStyle.FixedSingle;
                slotLbl.BackColor = Color.White;
                slotLbl.Margin = new Padding(3);
                slotLbl.Tag = i;
                slotLbl.AllowDrop = true;
                slotLbl.DragEnter += MainForm_DragEnter;
                slotLbl.DragDrop += MainForm_DragDrop;
                slotsTable.Controls.Add(slotLbl, i, 0);
            }
            slotsPanel.Controls.Add(slotsTable);
            rightPanel.Controls.Add(slotsPanel);

            dropZonePanel = new Panel();
            dropZonePanel.Dock = DockStyle.Top;
            dropZonePanel.Height = 125;
            dropZonePanel.Margin = new Padding(0, 8, 0, 8);
            dropZonePanel.BackColor = Color.FromArgb(255, 245, 230);
            dropZonePanel.BorderStyle = BorderStyle.FixedSingle;
            dropZonePanel.AllowDrop = true;
            dropZonePanel.DragEnter += MainForm_DragEnter;
            dropZonePanel.DragDrop += MainForm_DragDrop;

            lblDropHint = new Label();
            lblDropHint.Text = "📥 ¡ARRASTRA Y SUELTA AQUÍ\nLOS ARCHIVOS WHISKAS!";
            lblDropHint.Font = new Font("Segoe UI", 12F, FontStyle.Bold);
            lblDropHint.ForeColor = Color.FromArgb(217, 83, 30);
            lblDropHint.TextAlign = ContentAlignment.MiddleCenter;
            lblDropHint.Dock = DockStyle.Fill;
            lblDropHint.AllowDrop = true;
            lblDropHint.DragEnter += MainForm_DragEnter;
            lblDropHint.DragDrop += MainForm_DragDrop;
            dropZonePanel.Controls.Add(lblDropHint);
            rightPanel.Controls.Add(dropZonePanel);

            lblStatus = new Label();
            lblStatus.Text = "🔍 Buscando whiskas en tus carpetas...";
            lblStatus.Font = new Font("Segoe UI", 9.5F, FontStyle.Regular);
            lblStatus.ForeColor = Color.FromArgb(60, 60, 60);
            lblStatus.Dock = DockStyle.Fill;
            lblStatus.TextAlign = ContentAlignment.MiddleCenter;
            lblStatus.AllowDrop = true;
            lblStatus.DragEnter += MainForm_DragEnter;
            lblStatus.DragDrop += MainForm_DragDrop;
            rightPanel.Controls.Add(lblStatus);

            mainLayout.Controls.Add(rightPanel, 1, 0);

            gameTimer = new Timer();
            gameTimer.Interval = 1000;
            gameTimer.Tick += GameTimer_Tick;

            shakeTimer = new Timer();
            shakeTimer.Interval = 35;
            shakeTimer.Tick += (s, e) =>
            {
                if (!IsGameActive) return;
                int ox = rnd.Next(-shakeIntensity, shakeIntensity + 1);
                int oy = rnd.Next(-shakeIntensity, shakeIntensity + 1);
                this.Location = new Point(baseAnchorLocation.X + ox, baseAnchorLocation.Y + oy);
            };

            jumpTimer = new Timer();
            jumpTimer.Interval = 2500;
            jumpTimer.Tick += (s, e) =>
            {
                if (!IsGameActive) return;
                JumpToRandomScreenPosition();
                jumpTimer.Interval = rnd.Next(1800, 3800);
            };

            glitchSpawnTimer = new Timer();
            glitchSpawnTimer.Interval = 1500;
            glitchSpawnTimer.Tick += (s, e) =>
            {
                if (!IsGameActive) return;
                GlitchManager.SpawnGlitch();
                glitchSpawnTimer.Interval = rnd.Next(900, 2200);
            };
        }

        private void JumpToRandomScreenPosition()
        {
            Rectangle screen = Screen.PrimaryScreen.WorkingArea;
            int maxX = Math.Max(0, screen.Width - this.Width);
            int maxY = Math.Max(0, screen.Height - this.Height);
            baseAnchorLocation = new Point(screen.Left + rnd.Next(0, maxX), screen.Top + rnd.Next(0, maxY));
        }

        private void LoadLimonImage()
        {
            Image img = EmbeddedResourceManager.GetImage("Limon.png");
            if (img != null)
            {
                picLimon.Image = img;
                return;
            }

            Bitmap bmp = new Bitmap(280, 280);
            using (Graphics g = Graphics.FromImage(bmp))
            {
                g.Clear(Color.FromArgb(255, 230, 180));
                using (Font f = new Font("Segoe UI", 16F, FontStyle.Bold))
                using (Brush b = new SolidBrush(Color.FromArgb(217, 83, 30)))
                {
                    g.DrawString("🐱 Limón", f, b, 75, 115);
                }
            }
            picLimon.Image = bmp;
        }

        private void CreateShortcutWithIcon(string shortcutPath, string targetPath, string iconPath)
        {
            try
            {
                StringBuilder sb = new StringBuilder();
                sb.AppendLine("[InternetShortcut]");
                sb.AppendLine("URL=file:///" + targetPath.Replace('\\', '/'));
                if (!string.IsNullOrEmpty(iconPath) && File.Exists(iconPath))
                {
                    sb.AppendLine("IconFile=" + iconPath);
                    sb.AppendLine("IconIndex=0");
                }
                File.WriteAllText(shortcutPath, sb.ToString());
            }
            catch { }
        }

        private void SpawnWhiskasFiles()
        {
            CleanupFiles();
            createdWhiskasFiles.Clear();
            chosenFolderNames.Clear();

            try
            {
                List<KeyValuePair<string, string>> candidateDirs = new List<KeyValuePair<string, string>>();

                string desktop = Environment.GetFolderPath(Environment.SpecialFolder.Desktop);
                if (Directory.Exists(desktop)) candidateDirs.Add(new KeyValuePair<string, string>("Escritorio", desktop));

                string docs = Environment.GetFolderPath(Environment.SpecialFolder.MyDocuments);
                if (Directory.Exists(docs)) candidateDirs.Add(new KeyValuePair<string, string>("Documentos", docs));

                string pics = Environment.GetFolderPath(Environment.SpecialFolder.MyPictures);
                if (Directory.Exists(pics)) candidateDirs.Add(new KeyValuePair<string, string>("Imágenes", pics));

                string userProfile = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);
                string downloads = Path.Combine(userProfile, "Downloads");
                if (Directory.Exists(downloads)) candidateDirs.Add(new KeyValuePair<string, string>("Descargas", downloads));

                string videos = Path.Combine(userProfile, "Videos");
                if (Directory.Exists(videos)) candidateDirs.Add(new KeyValuePair<string, string>("Videos", videos));

                if (candidateDirs.Count == 0)
                {
                    candidateDirs.Add(new KeyValuePair<string, string>("Carpeta Local", AppDomain.CurrentDomain.BaseDirectory));
                }

                for (int i = candidateDirs.Count - 1; i > 0; i--)
                {
                    int j = rnd.Next(i + 1);
                    var tmp = candidateDirs[i];
                    candidateDirs[i] = candidateDirs[j];
                    candidateDirs[j] = tmp;
                }

                string[] fileNames = new string[] { "whiskas_1.txt", "whiskas_2.txt", "whiskas_3.txt" };
                string[] shortcutNames = new string[] { "whiskas_1.url", "whiskas_2.url", "whiskas_3.url" };
                string[] pngNames = new string[] { "whiskas_1.png", "whiskas_2.png", "whiskas_3.png" };
                string[] webpNames = new string[] { "whiskas_1.webp", "whiskas_2.webp", "whiskas_3.webp" };

                string[] fileContents = new string[]
                {
                    "🐱 ¡Este es el Whiskas #1 para Limón!\n🥫 Arrástrame hasta la ventanita para alimentarlo.",
                    "🐱 ¡Este es el Whiskas #2 para Limón!\n🥫 ¡Sabor Salmón! Arrástrame a la ventanita.",
                    "🐱 ¡Este es el Whiskas #3 para Limón!\n🥫 ¡Sabor Atún! Arrástrame a la ventanita antes de que acabe el tiempo."
                };

                for (int i = 0; i < 3; i++)
                {
                    var folder = candidateDirs[i % candidateDirs.Count];
                    string filePath = Path.Combine(folder.Value, fileNames[i]);
                    string urlPath = Path.Combine(folder.Value, shortcutNames[i]);
                    string pngPath = Path.Combine(folder.Value, pngNames[i]);
                    string webpPath = Path.Combine(folder.Value, webpNames[i]);

                    try
                    {
                        File.WriteAllText(filePath, fileContents[i]);
                        createdWhiskasFiles.Add(filePath);

                        CreateShortcutWithIcon(urlPath, filePath, whiskasIcoPath);
                        createdWhiskasFiles.Add(urlPath);

                        if (!string.IsNullOrEmpty(whiskasSourceWebp) && File.Exists(whiskasSourceWebp))
                        {
                            try
                            {
                                File.Copy(whiskasSourceWebp, pngPath, true);
                                createdWhiskasFiles.Add(pngPath);
                            }
                            catch { }
                            try
                            {
                                File.Copy(whiskasSourceWebp, webpPath, true);
                                createdWhiskasFiles.Add(webpPath);
                            }
                            catch { }
                        }

                        chosenFolderNames.Add(folder.Key);
                    }
                    catch
                    {
                        filePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, fileNames[i]);
                        File.WriteAllText(filePath, fileContents[i]);
                        createdWhiskasFiles.Add(filePath);
                        chosenFolderNames.Add("Carpeta Local");
                    }
                }
            }
            catch
            {
                for (int i = 0; i < 3; i++)
                {
                    string filePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, string.Format("whiskas_{0}.txt", i + 1));
                    File.WriteAllText(filePath, "🐱 Whiskas para Limón!");
                    createdWhiskasFiles.Add(filePath);
                    chosenFolderNames.Add("Carpeta de la App");
                }
            }

            lblStatus.Text = string.Format(
                "🔍 Limón escondió 3 whiskas con icono en:\n📁 {0}\n¡Búscalos y arrástralos aquí dentro del tiempo!",
                string.Join(", ", chosenFolderNames.ToArray())
            );
        }

        private void StartGame()
        {
            remainingSeconds = 120;
            deliveredWhiskas = 0;
            gameCompleted = false;
            gameOver = false;
            shakeIntensity = 4;
            deliveredFilePaths.Clear();

            UpdateTimerDisplay();
            UpdateCounterDisplay();
            UpdateAudioTrack();

            gameTimer.Start();
            shakeTimer.Start();
            jumpTimer.Start();
            glitchSpawnTimer.Start();
        }

        private void StopChaosEffects()
        {
            shakeTimer.Stop();
            jumpTimer.Stop();
            glitchSpawnTimer.Stop();
            GlitchManager.CloseAllGlitches();

            Rectangle screen = Screen.PrimaryScreen.WorkingArea;
            this.Location = new Point(
                screen.Left + (screen.Width - this.Width) / 2,
                screen.Top + (screen.Height - this.Height) / 2
            );
        }

        private void UpdateAudioTrack()
        {
            if (gameCompleted || gameOver)
            {
                AudioEngine.StopBgm();
                return;
            }

            // Antes de los 60 segundos (02:00 a 01:01): Suena F1.mp3 en bucle constante
            if (remainingSeconds > 60)
            {
                AudioEngine.PlayF1();
            }
            // Últimos 60 segundos (01:00 a 00:06): Suena F2.mp3 en bucle
            else if (remainingSeconds > 5)
            {
                AudioEngine.PlayF2();
            }
            // Últimos 5 segundos (00:05 a 00:01): Suena F2.mp3 con Limón.A.mp3
            else if (remainingSeconds > 0)
            {
                AudioEngine.PlayF2WithA();
            }
        }

        private void GameTimer_Tick(object sender, EventArgs e)
        {
            remainingSeconds--;
            UpdateTimerDisplay();
            UpdateAudioTrack();

            if (remainingSeconds <= 10)
            {
                shakeIntensity = 10;
            }
            else if (remainingSeconds <= 60)
            {
                shakeIntensity = 7;
            }
            else
            {
                shakeIntensity = 4;
            }

            if (remainingSeconds <= 0)
            {
                gameTimer.Stop();
                gameOver = true;
                StopChaosEffects();
                AudioEngine.StopBgm();
                WallpaperManager.RestoreOriginalWallpaper();

                lblTimer.Text = "⏰ ¡TIEMPO AGOTADO!";
                lblTimer.ForeColor = Color.Red;
                lblStatus.Text = "😹 Era Broma no te encriptamos nada te la creiste no? y Limón no recibio sus 3 whiskas a tiempo.";

                DialogResult res = MessageBox.Show(
                    "Era Broma no te encriptamos nada te la creiste no? y Limón no recibio sus 3 whiskas a tiempo 😿\n\n¿Quieres volver a intentarlo?",
                    "¡Tiempo Agotado!",
                    MessageBoxButtons.YesNo,
                    MessageBoxIcon.Information
                );

                if (res == DialogResult.Yes)
                {
                    WallpaperManager.SetLimonWallpaper();
                    SpawnWhiskasFiles();
                    StartGame();
                }
                else
                {
                    gameCompleted = true;
                    WallpaperManager.RestoreOriginalWallpaper();
                    CleanupFiles();
                    HunterForm.CleanupDesktopTrash();
                    AudioEngine.CloseAll();
                    this.Close();
                }
            }
        }

        private void UpdateTimerDisplay()
        {
            int mins = remainingSeconds / 60;
            int secs = remainingSeconds % 60;
            lblTimer.Text = string.Format("⏳ Cuenta Regresiva: {0:D2}:{1:D2}", mins, secs);

            if (remainingSeconds <= 5)
            {
                lblTimer.ForeColor = Color.DarkRed;
                lblStatus.Text = "🚨 ¡ÚLTIMOS 5 SEGUNDOS! 🐱💥 ¡DAME EL WHISKAS YA!";
            }
            else if (remainingSeconds <= 60)
            {
                lblTimer.ForeColor = Color.FromArgb(231, 76, 60);
                if (remainingSeconds == 60)
                {
                    lblStatus.Text = "⚠️ ¡ÚLTIMOS 60 SEGUNDOS! 🐱 ¡Rápido con los whiskas!";
                }
            }
            else
            {
                lblTimer.ForeColor = Color.FromArgb(192, 41, 43);
            }
        }

        private void UpdateCounterDisplay()
        {
            lblCounter.Text = string.Format("Whiskas entregados: {0} / {1}", deliveredWhiskas, RequiredWhiskas);

            TableLayoutPanel slotsTable = slotsPanel.Controls[0] as TableLayoutPanel;
            if (slotsTable != null)
            {
                for (int i = 0; i < 3; i++)
                {
                    Label lbl = slotsTable.GetControlFromPosition(i, 0) as Label;
                    if (lbl != null)
                    {
                        if (i < deliveredWhiskas)
                        {
                            lbl.Text = "✨ 🥫 ¡Listo!";
                            lbl.BackColor = Color.FromArgb(212, 239, 223);
                            lbl.ForeColor = Color.FromArgb(39, 174, 96);
                            lbl.Font = new Font("Segoe UI", 9.5F, FontStyle.Bold);
                        }
                        else
                        {
                            lbl.Text = "🥫 (Vacío)";
                            lbl.BackColor = Color.White;
                            lbl.ForeColor = Color.Gray;
                            lbl.Font = new Font("Segoe UI", 9.5F, FontStyle.Regular);
                        }
                    }
                }
            }
        }

        private void MainForm_DragEnter(object sender, DragEventArgs e)
        {
            if (gameCompleted || gameOver)
            {
                e.Effect = DragDropEffects.None;
                return;
            }

            if (e.Data.GetDataPresent(DataFormats.FileDrop))
            {
                e.Effect = DragDropEffects.Copy;
            }
            else
            {
                e.Effect = DragDropEffects.None;
            }
        }

        private void MainForm_DragDrop(object sender, DragEventArgs e)
        {
            ProcessDroppedFiles(e);
        }

        public void ProcessDroppedFiles(DragEventArgs e)
        {
            if (gameCompleted || gameOver) return;

            if (e.Data.GetDataPresent(DataFormats.FileDrop))
            {
                string[] files = (string[])e.Data.GetData(DataFormats.FileDrop);
                bool acceptedAny = false;

                foreach (string file in files)
                {
                    string fileName = Path.GetFileName(file).ToLowerInvariant();

                    bool isWhiskas = createdWhiskasFiles.Contains(file) || fileName.Contains("whiskas");

                    if (isWhiskas)
                    {
                        string baseKey = fileName;
                        if (fileName.StartsWith("whiskas_1")) baseKey = "whiskas_1";
                        else if (fileName.StartsWith("whiskas_2")) baseKey = "whiskas_2";
                        else if (fileName.StartsWith("whiskas_3")) baseKey = "whiskas_3";

                        if (!deliveredFilePaths.Contains(baseKey) && !deliveredFilePaths.Contains(file))
                        {
                            deliveredFilePaths.Add(baseKey);
                            deliveredFilePaths.Add(file);
                            deliveredWhiskas++;
                            acceptedAny = true;

                            try
                            {
                                if (File.Exists(file)) File.Delete(file);
                                string dir = Path.GetDirectoryName(file);
                                if (!string.IsNullOrEmpty(dir))
                                {
                                    string relatedUrl = Path.Combine(dir, baseKey + ".url");
                                    string relatedPng = Path.Combine(dir, baseKey + ".png");
                                    string relatedWebp = Path.Combine(dir, baseKey + ".webp");
                                    string relatedTxt = Path.Combine(dir, baseKey + ".txt");
                                    if (File.Exists(relatedUrl)) File.Delete(relatedUrl);
                                    if (File.Exists(relatedPng)) File.Delete(relatedPng);
                                    if (File.Exists(relatedWebp)) File.Delete(relatedWebp);
                                    if (File.Exists(relatedTxt)) File.Delete(relatedTxt);
                                }
                            }
                            catch { }

                            try { SystemSounds.Asterisk.Play(); } catch { }

                            UpdateCounterDisplay();

                            if (deliveredWhiskas >= RequiredWhiskas)
                            {
                                OnAllWhiskasDelivered();
                                return;
                            }
                        }
                        else
                        {
                            lblStatus.Text = "⚠️ ¡Ese whiskas ya fue entregado!";
                        }
                    }
                    else
                    {
                        lblStatus.Text = string.Format("❌ \"{0}\" no es un Whiskas. ¡Limón solo come Whiskas! 😺", Path.GetFileName(file));
                    }
                }

                if (acceptedAny && deliveredWhiskas < RequiredWhiskas)
                {
                    lblStatus.Text = string.Format("😻 ¡Ñam! Limón se comió un whiskas. Faltan {0}.", RequiredWhiskas - deliveredWhiskas);
                }
            }
        }

        private void OnAllWhiskasDelivered()
        {
            gameTimer.Stop();
            gameCompleted = true;
            StopChaosEffects();
            AudioEngine.CloseAll();
            WallpaperManager.RestoreOriginalWallpaper();
            CleanupFiles();
            HunterForm.CleanupDesktopTrash();

            try { SystemSounds.Exclamation.Play(); } catch { }

            using (Form thanksForm = new Form())
            {
                thanksForm.Text = "🐱 ¡Gracias!";
                thanksForm.Size = new Size(420, 230);
                thanksForm.StartPosition = FormStartPosition.CenterParent;
                thanksForm.FormBorderStyle = FormBorderStyle.FixedDialog;
                thanksForm.MaximizeBox = false;
                thanksForm.MinimizeBox = false;
                thanksForm.BackColor = Color.FromArgb(255, 252, 245);

                Label lblThanks = new Label();
                lblThanks.Text = "✨ ¡Gracias! ✨\n\n¡Limón recibió sus 3 whiskas a tiempo y la corrupción se ha detenido! 🐱💖";
                lblThanks.Font = new Font("Segoe UI", 12F, FontStyle.Bold);
                lblThanks.ForeColor = Color.FromArgb(39, 174, 96);
                lblThanks.TextAlign = ContentAlignment.MiddleCenter;
                lblThanks.Dock = DockStyle.Top;
                lblThanks.Height = 110;
                lblThanks.Padding = new Padding(15);
                thanksForm.Controls.Add(lblThanks);

                Button btnAccept = new Button();
                btnAccept.Text = "Aceptar";
                btnAccept.Font = new Font("Segoe UI", 11F, FontStyle.Bold);
                btnAccept.BackColor = Color.FromArgb(46, 204, 113);
                btnAccept.ForeColor = Color.White;
                btnAccept.FlatStyle = FlatStyle.Flat;
                btnAccept.FlatAppearance.BorderSize = 0;
                btnAccept.Size = new Size(130, 38);
                btnAccept.Location = new Point((thanksForm.ClientSize.Width - btnAccept.Width) / 2, 130);
                btnAccept.Click += (s, ev) =>
                {
                    thanksForm.DialogResult = DialogResult.OK;
                    thanksForm.Close();
                };
                thanksForm.Controls.Add(btnAccept);

                thanksForm.ShowDialog(this);
            }

            WallpaperManager.RestoreOriginalWallpaper();
            AudioEngine.CloseAll();
            this.Close();
        }

        private void MainForm_FormClosing(object sender, FormClosingEventArgs e)
        {
            if (!gameCompleted)
            {
                e.Cancel = true;
                MessageBox.Show(
                    "¡No puedes cerrar la ventanita hasta que le des los 3 whiskas a Limón! 🐱⏳\n\nBusca los archivos de whiskas en tus carpetas y arrástralos a la ventanita.",
                    "🐱 Limón tiene hambre",
                    MessageBoxButtons.OK,
                    MessageBoxIcon.Warning
                );
            }
            else
            {
                StopChaosEffects();
                AudioEngine.CloseAll();
                WallpaperManager.RestoreOriginalWallpaper();
                CleanupFiles();
                HunterForm.CleanupDesktopTrash();
            }
        }

        private void CleanupFiles()
        {
            foreach (string file in createdWhiskasFiles)
            {
                try
                {
                    if (File.Exists(file))
                    {
                        File.Delete(file);
                    }
                }
                catch { }
            }
        }
    }
}
