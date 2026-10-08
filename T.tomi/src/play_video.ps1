param(
    [int]$x = 100,
    [int]$y = 100,
    [int]$w = 200,
    [int]$h = 150,
    [string]$videoPath = ""
)

Add-Type -AssemblyName PresentationCore, PresentationFramework, WindowsBase

if (-not (Test-Path $videoPath)) {
    exit 0
}

$window = New-Object System.Windows.Window
$window.WindowStyle = [System.Windows.WindowStyle]::None
$window.AllowsTransparency = $true
$window.Background = [System.Windows.Media.Brushes]::Black
$window.Left = $x
$window.Top = $y
$window.Width = $w
$window.Height = $h
$window.Topmost = $true
$window.ShowInTaskbar = $false
$window.ResizeMode = [System.Windows.ResizeMode]::NoResize

$media = New-Object System.Windows.Controls.MediaElement
$media.Source = New-Object System.Uri($videoPath)
$media.LoadedBehavior = [System.Windows.Controls.MediaState]::Play
$media.UnloadedBehavior = [System.Windows.Controls.MediaState]::Stop
$media.Stretch = [System.Windows.Media.Stretch]::UniformToFill

$media.Add_MediaEnded({
    $window.Close()
})

$media.Add_MediaFailed({
    $window.Close()
})

$window.Content = $media
$window.ShowDialog() | Out-Null
