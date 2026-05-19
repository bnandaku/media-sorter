# media-sorter

Watches a directory (e.g. an rclone-mounted Put.io folder) and automatically moves video files into an organised Plex/Unraid media library.

## How it works

1. Scans `SOURCE_PATH` recursively for video files (`.mkv`, `.mp4`, `.avi`, etc.)
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
| `SOURCE_PATH` | ✅ | — | Directory to watch (rclone mount) |
| `TVSHOW_PATH` | ✅ | — | Destination for TV shows |
| `MOVIES_PATH` | ✅ | — | Destination for movies |
| `SCAN_INTERVAL` | — | `300` | Seconds between scans |
| `DRY_RUN` | — | unset | Set to any value to log moves without executing them |

## Docker (Unraid)

```bash
docker run -d \
  --name media-sorter \
  -e SOURCE_PATH=/mnt/putio \
  -e TVSHOW_PATH=/mnt/user/TV \
  -e MOVIES_PATH=/mnt/user/Movies \
  -e SCAN_INTERVAL=300 \
  -v /mnt/putio:/mnt/putio \
  -v /mnt/user/TV:/mnt/user/TV \
  -v /mnt/user/Movies:/mnt/user/Movies \
  bharatram1/media-sorter:latest
```

### Test with dry run first

```bash
docker run --rm \
  -e SOURCE_PATH=/mnt/putio \
  -e TVSHOW_PATH=/mnt/user/TV \
  -e MOVIES_PATH=/mnt/user/Movies \
  -e DRY_RUN=true \
  -v /mnt/putio:/mnt/putio \
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
