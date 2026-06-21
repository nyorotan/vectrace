package main

import (
	"bytes"
	"embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	_ "golang.org/x/image/bmp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Config struct {
	Flags string `json:"flags"`
}

//go:embed DD.bmp
var embeddedFiles embed.FS

func loadConfig() string {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return "-C" // デフォルト値
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "-C"
	}
	return cfg.Flags
}

func saveConfig(flags string) {
	cfg := Config{Flags: flags}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile("config.json", data, 0644)
}

func findExecutable(name string, baseDir string) (string, error) {
	if baseDir == "" {
		baseDir = "."
	}

	candidates := []string{
		name,
		filepath.Join(baseDir, name),
		filepath.Join(filepath.Dir(os.Args[0]), name),
	}

	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), name))
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			if abs, err := filepath.Abs(candidate); err == nil {
				return abs, nil
			}
			return candidate, nil
		}
	}

	return "", fmt.Errorf("executable %q not found in candidates: %s", name, strings.Join(candidates, ", "))
}

var mainGoroutineID uint64

func init() {
	mainGoroutineID = currentGoroutineID()
}

func currentGoroutineID() uint64 {
	var buf [30]byte
	runtime.Stack(buf[:], false)
	for i := 10; buf[i] != ' '; i++ {
		if i >= len(buf)-1 {
			break
		}
	}
	var id uint64
	for i := 10; i < len(buf) && buf[i] != ' '; i++ {
		if buf[i] < '0' || buf[i] > '9' {
			break
		}
		id = id*10 + uint64(buf[i]-'0')
	}
	return id
}

func runOnUIThread(fn func()) {
	if currentGoroutineID() == mainGoroutineID {
		fn()
		return
	}
	fyne.DoAndWait(fn)
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("vectrace_gui")
	window.Resize(fyne.NewSize(300, 380))
	window.SetFixedSize(true)

	initialFlags := loadConfig()

	bmpMagic := []byte{0x42, 0x4D}
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	var logoImg *canvas.Image
	if data, err := embeddedFiles.ReadFile("DD.bmp"); err == nil {
		if img, _, err := image.Decode(bytes.NewReader(data)); err == nil && img != nil {
			logoImg = canvas.NewImageFromImage(img)
			logoImg.FillMode = canvas.ImageFillContain
			logoImg.SetMinSize(fyne.NewSize(280, 200))
		}
	}

	label := widget.NewLabel("[Adding flags and values] \nexample: -C -z 5 -a 0.5")
	entry := widget.NewEntry()
	entry.SetText(initialFlags)
	progress := widget.NewProgressBarInfinite()
	progress.Hide() // 初期状態は隠しておく

	window.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		var validFiles []string

		for _, uri := range uris {
			filePath := uri.Path()
			// Windows環境でパスの先頭にスラッシュが付く場合の補正
			if len(filePath) > 3 && filePath[0] == '/' && filePath[2] == ':' {
				filePath = filePath[1:]
			}

			file, err := os.Open(filePath)
			if err != nil {
				continue
			}

			header := make([]byte, 30)
			n, err := io.ReadFull(file, header)
			file.Close()
			if err != nil && err != io.ErrUnexpectedEOF {
				continue
			}

			if n >= 30 && bytes.Equal(header[:2], bmpMagic) {
				bpp := binary.LittleEndian.Uint16(header[28:30])
				if bpp <= 8 {
					validFiles = append(validFiles, filePath)
				}
			} else if n >= 26 && bytes.Equal(header[:8], pngMagic) {
				bitDepth := header[24]
				colorType := header[25]
				if colorType == 3 || (colorType == 0 && bitDepth <= 8) {
					validFiles = append(validFiles, filePath)
				}
			}
		}

		if len(validFiles) > 0 {
			flags := strings.Fields(entry.Text)
			args := append(flags, validFiles...)

			runOnUIThread(func() {
				progress.Show()
				window.Content().Refresh()
			})

			go func() {
				cmdPath, err := findExecutable(vectraceCmd, ".")
				if err != nil {
					fmt.Fprintf(os.Stderr, "vectrace executable resolution failed: %v\n", err)
					runOnUIThread(func() {
						progress.Hide()
						window.Content().Refresh()
					})
					return
				}

				cmd := exec.Command(cmdPath, args...)
				hideConsoleWindow(cmd)
				if err := cmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "vectrace execution failed: %v\n", err)
				}

				runOnUIThread(func() {
					progress.Hide()
					window.Content().Refresh()
				})
			}()
		}
	})

	content := container.NewVBox()
	if logoImg != nil {
		content.Add(logoImg)
	}
	content.Add(label)
	content.Add(entry)
	content.Add(progress)

	window.SetContent(content)

	// 終了時の処理
	myApp.SetIcon(nil) // 必要に応じてアイコン設定
	window.ShowAndRun()

	// Run終了後に保存
	saveConfig(entry.Text)
}
