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
	"strings"

	_ "golang.org/x/image/bmp"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
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

func main() {
	var mw *walk.MainWindow
	var te *walk.TextEdit
	var iv *walk.ImageView
	var pb *walk.ProgressBar

	initialFlags := loadConfig()

	// BMPとPNGのマジックナンバーを定義
	// BMP: "BM" (0x42 0x4D)
	// PNG: 0x89 0x50 0x4E 0x47 0x0D 0x0A 0x1A 0x0A
	bmpMagic := []byte{0x42, 0x4D}
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	// 埋め込みファイルから画像を読み込む
	var logo walk.Image
	if data, err := embeddedFiles.ReadFile("DD.bmp"); err == nil {
		if img, _, err := image.Decode(bytes.NewReader(data)); err == nil && img != nil {
			// image.Image を walk.Bitmap に変換（エラー時は logo = nil のまま）
			logo, _ = walk.NewBitmapFromImage(img)
		}
	}

	if err := (MainWindow{
		AssignTo: &mw,
		Title:    "vectrace_gui",
		Size:     Size{Width: 280, Height: 280},
		Layout:   VBox{},
		OnDropFiles: func(files []string) {
			var validFiles []string
			for _, filePath := range files {
				func() {
					file, err := os.Open(filePath)
					if err != nil {
						return
					}
					defer file.Close()

					// ヘッダー情報を読み込む (BMPのBitBitCount取得に30バイト必要)
					header := make([]byte, 30)
					n, err := io.ReadFull(file, header)
					if err != nil && err != io.ErrUnexpectedEOF {
						return
					}

					// BMP: 28バイト目から2バイトがBits Per Pixel (256色以下は8以下)
					if n >= 30 && bytes.Equal(header[:2], bmpMagic) {
						bpp := binary.LittleEndian.Uint16(header[28:30])
						if bpp <= 8 {
							validFiles = append(validFiles, filePath)
						}
						return
					}

					// PNG: 24バイト目がBitDepth、25バイト目がColorType
					if n >= 26 && bytes.Equal(header[:8], pngMagic) {
						bitDepth := header[24]
						colorType := header[25]
						// ColorType 3 (Indexed) は常に256色以下
						// ColorType 0 (Grayscale) かつ 8bit以下は256階調以下
						if colorType == 3 || (colorType == 0 && bitDepth <= 8) {
							validFiles = append(validFiles, filePath)
						}
						return
					}
				}()
			}

			if len(validFiles) > 0 {
				// TextEditの内容をスペースで分割して引数として使用
				flags := strings.Fields(te.Text())
				args := append(flags, validFiles...)

				// プログレスバーを「処理中（アニメーション）」の状態にする
				pb.SetValue(0)
				pb.SetMarqueeMode(true)

				cmd := exec.Command("./vectrace", args...)
				go func() {
					_ = cmd.Run() // コマンドの終了を待機
					mw.Synchronize(func() {
						pb.SetMarqueeMode(false)
						pb.SetValue(100)
					})
				}()
			}
		},

		Children: []Widget{
			ImageView{
				AssignTo: &iv,
				Image:    logo,
			},
			TextLabel{
				Text: "[Adding flags and values] \nexample: -C -z 5 -a 0.5",
			},

			TextEdit{
				AssignTo: &te,
				Text:     initialFlags,
			},
			ProgressBar{
				AssignTo: &pb,
			},
		},
	}).Create(); err != nil {
		fmt.Fprintf(os.Stderr, "Window creation failed: %v\n", err)
		os.Exit(1)
	}

	// 終了時に設定を保存
	mw.Closing().Attach(func(canCancel *bool, reason walk.CloseReason) {
		saveConfig(te.Text())
	})

	hwnd := mw.Handle()
	style := win.GetWindowLong(hwnd, win.GWL_STYLE)
	style &^= win.WS_THICKFRAME | win.WS_MAXIMIZEBOX
	win.SetWindowLong(hwnd, win.GWL_STYLE, style)
	mw.Run()
}
