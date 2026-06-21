set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags="-s -w" -o vectrace.exe ./cmd/vectrace/
go build -trimpath -ldflags="-H windowsgui -s -w" -o vectrace-gui_win.exe ./vectrace-gui/windows/.
