# media-sorter

Watches a directory (e.g. an rclone-mounted Put.io folder) and automatically moves video files into an organised Plex/Unraid media library.

## How it works

1. Scans the source directory recursively for video files (`.mkv`, `.mp4`, `.avi`, etc.)
2. Waits for each file to be **stable** (size unchanged across two consecutive scans) before moving it — prevents moving partially-synced rclone files
3. Parses the filename to determine TV show or movie
4. **TV shows** → `TVSHOW_PATH/ShowName/Season_XX/ShowName.SXXEXX.quality.ext`
5. **Movies** → `MOVIES_PATH/Title.Year.Quality.ext`
6. Removes empty directories left behind in the source

## TV show filename formats supported

| Format | Example |
|--------|---------|
| `S01E02` | `The.Bear.S03E05.1080p.mkv` |
| `Season 02 - 01` | `[Fansub] Mob Psycho Season 02 - 01.mkv` |
| `Show - 01` | `[Group] Some Show - 05.mkv` (assumes Season 01) |

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SCAN_INTERVAL` | — | `300` | Seconds between scans |
| `DRY_RUN` | — | unset | Set to any value to log moves without executing them |

The container always reads from `/mnt/source` and writes to `/mnt/tvshows` and `/mnt/movies`. Configure those via volume mounts.

## Unraid — Community Applications (easiest)

1. In Unraid, open **Apps** (Community Applications)
2. Click **Add Container** (the icon in the top right)
3. Paste this URL into the template field:
   ```
   https://raw.githubusercontent.com/bnandaku/media-sorter/main/unraid/media-sorter.xml
   ```
4. Click **OK** — all variables and paths load pre-filled
5. Set your three host paths:
   - **Source Path** → your rclone Put.io mount (e.g. `/mnt/user/putio`)
   - **TV Shows Path** → your TV shows share (e.g. `/mnt/user/TV Shows`)
   - **Movies Path** → your movies share (e.g. `/mnt/user/Movies`)
6. Optionally set `DRY_RUN=true` for a test run first
7. Click **Apply**

## Docker (manual)

```bash
docker run -d \
  --name media-sorter \
  -e SCAN_INTERVAL=300 \
  -v /your/rclone/mount:/mnt/source \
  -v /your/tv/path:/mnt/tvshows \
  -v /your/movies/path:/mnt/movies \
  bharatram1/media-sorter:latest
```

### Test with dry run first

```bash
docker run --rm \
  -e DRY_RUN=true \
  -v /your/rclone/mount:/mnt/source \
  -v /your/tv/path:/mnt/tvshows \
  -v /your/movies/path:/mnt/movies \
  bharatram1/media-sorter:latest
```

## GitHub Actions setup

Add these secrets to your GitHub repository:

- `DOCKERHUB_USERNAME` — your Docker Hub username (`bharatram1`)
- `DOCKERHUB_TOKEN` — a Docker Hub access token (not your password)

The workflow triggers on every push to `main` and on version tags (`v1.0.0`, etc.).

## Building locally

```bash
go build -o media-sorter .
./media-sorter
```
