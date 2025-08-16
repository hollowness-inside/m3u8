# m3u8

A high-performance Go-based tool for downloading and combining video segments from M3U8 playlists. Features concurrent downloads, custom HTTP headers support, and flexible URL handling.

This is a complete rewrite of the [original m3u8_download](https://github.com/hollowness-inside/m3u8_download) written in Python, delivering significantly faster performance through Go's concurrency model.

## ✨ Key Features

### 🔐 Custom HTTP Headers Support
Specify custom HTTP headers for authentication, user agents, and other requirements:
- **Headers File**: Load headers from a JSON file (`--headers headers.json`)
- **Authentication**: Perfect for protected streams requiring tokens or cookies
- **Flexible**: Support for any HTTP header combination

### 🔗 Force Segment URL Prefix
Override segment URLs with custom prefixes:
- **URL Manipulation**: Use `--force-url-prefix` to prepend custom URLs to segments
- **CDN Switching**: Easily switch between different content delivery networks
- **Local Development**: Point segments to local servers for testing

### ⚡ Performance & Reliability
- **Concurrent Downloads**: Utilizes goroutines for extremely fast segment retrieval
- **Caching**: Optional M3U8 file caching for faster subsequent runs
- **Error Handling**: Graceful failure recovery with force-combine options
- **Cross-Platform**: Single executable for Windows, macOS, and Linux

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or later
- ffmpeg installed and in PATH

### Installation
```bash
# Clone and build
git clone https://github.com/hollowness-inside/m3u8.git
cd m3u8
go mod download
go build
```

### Basic Usage
```bash
# Simple download and combine
./m3u8 "https://example.com/video.m3u8" --combine output.mp4
```

## 📖 Usage Examples

### Authentication with Custom Headers
For protected streams requiring authentication:

```bash
# Create headers file
echo '{
  "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
  "Authorization": "Bearer your-token-here",
  "Referer": "https://example.com"
}' > headers.json

# Download with headers
./m3u8 "https://protected.example.com/video.m3u8" --headers headers.json --combine output.mp4
```

### Force URL Prefix for Segment URLs
When segment URLs need modification or CDN switching:

```bash
# Force all segments to use a specific CDN
./m3u8 "https://example.com/video.m3u8" \
  --force-url-prefix "https://cdn.example.com/segments/" \
  --combine output.mp4

# Use local development server
./m3u8 "https://example.com/video.m3u8" \
  --force-url-prefix "http://localhost:8080/" \
  --combine output.mp4
```

### Advanced Download Options
```bash
# High-performance download with 20 concurrent connections
./m3u8 "https://example.com/video.m3u8" \
  --concurrent 20 \
  --cache playlist.cache \
  --combine output.mp4

# Download specific segment range
./m3u8 "https://example.com/video.m3u8" \
  --skip 10 \
  --limit 100 \
  --combine output.mp4

# Force combine even with failed segments
./m3u8 "https://example.com/video.m3u8" \
  --force-combine output.mp4 \
  --cleanup
```

### Repair and Maintenance
```bash
# Fix missing segments in existing directory
./m3u8 "https://example.com/video.m3u8" --fix ./segments

# Force specific file extension
./m3u8 "https://example.com/video.m3u8" --force-ext .ts
```

## ⚙️ Command Line Reference

```
Usage: m3u8 <url> [options]

Required:
  url                         M3U8 playlist URL

Output Options:
  --segments-dir DIR          Directory to store segments (default: segments)
  --combine OUTPUT            Combine segments into video file
  --force-combine OUTPUT      Combine even with failed downloads
  --cleanup                   Remove segments after successful combination

Authentication & Headers:
  --headers FILE              JSON file with HTTP headers for requests

URL Manipulation:
  --force-url-prefix PREFIX   Override segment URL prefix
  --force-ext EXT            Force file extension for segments

Performance:
  --concurrent N              Concurrent downloads (default: 10)
  --cache FILE               Cache parsed M3U8 for faster reruns

Segment Control:
  --skip N                   Skip first N segments
  --limit N                  Download only N segments
  --fix DIR                  Repair missing segments in directory

System:
  --ffmpeg PATH              Custom ffmpeg executable path
  --filelist FILE            Custom ffmpeg filelist path
  --verbose, -v              Enable detailed output
```

## 🔧 Configuration Files

### Headers File Format
Create a JSON file for HTTP headers:

```json
{
  "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
  "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "Referer": "https://streaming-site.com",
  "Cookie": "session=abc123; auth=xyz789",
  "X-Custom-Header": "custom-value"
}
```

## 💡 Pro Tips

### 🔍 Extracting Headers from Browser
1. Open browser Developer Tools (F12)
2. Go to Network tab
3. Navigate to the streaming page
4. Right-click any request → "Copy as cURL"
5. Use [curlconverter.com/json](https://curlconverter.com/json/) to convert to JSON

### 🎯 Finding Stream URLs
Install [The Stream Detector](https://chromewebstore.google.com/detail/the-stream-detector/iakkmkmhhckcmoiibcfjnooibphlobak) Chrome extension to automatically capture M3U8 URLs from any webpage.

### 🚀 Performance Optimization
- Increase `--concurrent` for faster downloads (try 15-25)
- Use `--cache` to avoid re-parsing large playlists
- Enable `--cleanup` to save disk space

## 📋 Requirements

- **Go**: Version 1.21 or later
- **ffmpeg**: Required for video combination
  - Windows: Download from [ffmpeg.org](https://ffmpeg.org)
  - macOS: `brew install ffmpeg`
  - Linux: `sudo apt install ffmpeg` or equivalent

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

**Need help?** Open an issue on [GitHub](https://github.com/hollowness-inside/m3u8/issues) or check the [Discussions](https://github.com/hollowness-inside/m3u8/discussions) tab.
