# codeshelf

Pick a saved code by title in your terminal. Arrows to move, **Enter** to view, **Space** to copy.

## Install + run (Windows cmd)

```
curl -L -o "%LOCALAPPDATA%\Microsoft\WindowsApps\codes.exe" https://github.com/raviteja3737/codeshelf/releases/latest/download/codes-windows-amd64.exe && codes
```

After that, just type `codes`. Run the same line again to update.

If `WindowsApps` isn't on PATH (locked-down laptops), run it from the current folder instead:

```
curl -L -o codes.exe https://github.com/raviteja3737/codeshelf/releases/latest/download/codes-windows-amd64.exe && codes
```

## Mac / Linux (untested)

Replace `darwin-arm64` with `darwin-amd64`, `linux-amd64` or `linux-arm64` as needed.

```
sudo curl -L -o /usr/local/bin/codes https://github.com/raviteja3737/codeshelf/releases/latest/download/codes-darwin-arm64 && sudo chmod +x /usr/local/bin/codes && codes
```

Linux needs `xclip` or `wl-clipboard` for copying.

## Adding codes

Put one file per code in `snippets/` (file name = title), then rebuild and publish a release.
Personal extras can go in `~/.codes` on any machine and show up in the list too.
