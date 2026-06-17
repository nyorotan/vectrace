🌐 **English** | [日本語](README.ja.md)

# Vectrace

Vectrace is a bitmap tracing tool based on [Potrace](https://potrace.sourceforge.net), a vectorization library developed by Peter Selinger. It narrows the scope (input: indexed color images only, output: SVG only) and adds functionality (support for color layers in SVG).

Under the hood, it is built upon a direct machine translation (transpilation) of potrace into Go using [cxgo](https://github.com/gotranspile/cxgo), and is written in pure Go with no C dependencies.

Processing is accelerated by multi-threading the trace operation for each color layer.



## Supported Input Formats (Restrictions)

To maintain processing efficiency and color separation accuracy, input images are subject to the following restrictions. **Full-color (24-bit/32-bit) images cannot be loaded directly.** Please reduce the number of colors (convert to indexed color) using an image editor beforehand.

- **BMP**: 1-bit / 4-bit / 8-bit (indexed color or grayscale)
- **PNG**: Indexed color (palette format) or grayscale with 8-bit or fewer

Note: An error will occur if the number of unique colors used internally exceeds 256.



## Option Flags

The option flags are only partially compatible with Potrace's general options and algorithm options.

### Compatible Flags

General options:

| Flag           | Description          |
| ------------- | ----------- |
| -h, --help    | Display help message |
| -v, --version | Display version information  |
| -l, --license | Display license information  |

Algorithm options:

| Flag                      | Description                                                              |
| ------------------------ | --------------------------------------------------------------- |
| -a, --alphamax float     | Threshold for controlling corner smoothness (default 1)                                       |
| -n, --no-opticurve       | Disable curve optimization                                                     |
| -O, --opttolerance float | Tolerance for curve optimization (default 0.2)                                       |
| -o, --output string      | Output file path (only valid for a single input source)                                        |
| -t, --turdsize int       | Remove noise (speckles) smaller than the specified pixel count (default 2)                           |
| -z, --turnpolicy int     | Policy for resolving path decomposition ambiguities<br>(0:black, 1:white, 2:left, 3:right, 4:minority, 5:majority, 6:random) |
| -k, --blacklevel float   | Binarization threshold for the image (default 0.5)                                          |

Modified and added options:

| Flag                   | Description                                      |
|:--------------------- | --------------------------------------- |
| -b, --bg-dilation int | Number of additional dilation passes to create an outline on a white background (default -1) |
| -C, --color           | Split a color image into multiple layers                      |
| -K, --force-black     | Force the base layer (outline) to black                      |

- bg-dilation controls the dilation size of the white mask image data placed at the back layer to prevent gaps from being transparent. Increasing this value causes objects to have a white outline on the outside.

- force-black creates a black outline layer by extracting dark areas from the image based on the "-k (threshold) flag" to prevent unintended gaps or white lines between adjacent colors. Normally, the layer is created using the average color of the extracted region.

- color splits a color image into multiple layers and traces each one. When this flag is set, a color SVG is output. Since the number of layers increases with the number of colors, images with many colors will significantly increase the file size and conversion time. (This is why the color count is limited to 256 or fewer.)



## Vectrace-gui

A Windows-only GUI frontend.
After launching, you can process supported images by drag & drop.
Option flags can also be specified by entering them in the text box. (The "-C" flag is required for color output.)

Note: `vectrace.exe` must be located in the same directory as `vectrace-gui.exe` to run. Additionally, the GUI checks file headers and ignores unsupported formats (such as full-color images).



## Examples

**Original Image**

![Original](https://github.com/nyorotan/vectrace/raijin-zu.bmp)

**SVG Image**

![Vectorized](https://github.com/nyorotan/vectrace//raijin-zu.svg)



## Usage

Convert a PNG image to a color SVG:

- Without the "-C" flag, a monochrome SVG will be generated.

```
vectrace -C -o ./testdata/test.svg ./testdata/test.png
```

Convert multiple images to color SVGs at once:

- Note: The "-o" flag cannot be used when multiple images are specified.

```
vectrace -C ./testdata/test1.png ./testdata/test2.png ...
```



## Build Instructions

Vectrace is written in pure Go and does not require CGO or gcc.

### Windows

Build `vectrace.exe` and `vectrace-gui.exe`:

```
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags="-s -w" -o vectrace ./cmd/vectrace/.
go build -trimpath -ldflags="-H windowsgui -s -w" -o vectrace-gui.exe ./cmd/vectrace-gui/.
```
Alternatively, use the provided batch file:
```batch
build_windows.bat
```

### Linux

Build `vectrace` only. (`vectrace-gui` is Windows-only.)

```
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -trimpath -ldflags="-s -w" -o vectrace ./cmd/vectrace/
```
Alternatively, use the provided shell script:
```bash
chmod +x build_linux.sh
./build_linux.sh
```
## License and Trademarks
 
- License: This project is released under the GNU General Public License v2.0 or later, inheriting the license of the original Potrace. This is a copyleft license; any derivative works of Vectrace must also be released under the same license terms, ensuring the source code remains available and the freedom to modify is preserved for all users.
- Trademark: "Potrace" is a trademark of Peter Selinger. To avoid confusion with the official Potrace, this project is named Vectrace and is explicitly identified as an unofficial derivative. See the [Potrace Trademark Policy](https://potrace.sourceforge.net/#trademarks) for details.

---
Copyright
- Copyright (C) 2001-2019 Peter Selinger (Original Potrace) 
- Modified by nyorotan 2026
