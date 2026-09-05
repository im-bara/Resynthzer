![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=for-the-badge&logo=go)
![Wails](https://img.shields.io/badge/Wails-v2-red?style=for-the-badge&logo=wails)
![FFmpeg](https://img.shields.io/badge/FFmpeg-Full_GPL-5FB925?style=for-the-badge&logo=ffmpeg)
![Real-ESRGAN](https://img.shields.io/badge/Real--ESRGAN-NCNN_Vulkan-blue?style=for-the-badge)
![License](https://img.shields.io/badge/License-GPL_v3-yellow?style=for-the-badge)

**Resynthzer** is a blazing-fast, lightweight, open-source GUI desktop application for media processing. Built with **Go** and **Wails**, it harnesses the raw power of **FFmpeg** and **Real-ESRGAN AI** to compress and upscale your videos, images, and audio seamlessly without the bloated size of commercial software.

---

## Features

### Video
*   **AI Upscale:** Transform low-res videos into crisp HD using Real-ESRGAN (GPU-accelerated via Vulkan).
*   **Standard Upscale:** Fast, CPU-based upscaling using FFmpeg (Lanczos filter) for low-end machines.
*   **Compression:** Compress videos efficiently using **H.264** or next-gen **AV1** codecs.

### Photo
*   **AI Upscale:** 4x Image resolution enhancement with lossless PNG output.
*   **Compression:** Quickly reduce image file sizes with smart JPG compression.

### Audio
*   **Compression:** Easily downsample and compress audio files (MP3, WAV, FLAC, etc.) by adjusting the bitrate.

---

## Tech Stack
*   **Backend:** Go (Golang)
*   **Frontend UI:** Wails v2 (Vanilla HTML, CSS, JS - Custom Toast & Progress Bar)
*   **Core Engines:** 
    *   `FFmpeg` (Media processing & compression)
    *   `Real-ESRGAN-ncnn-vulkan` (AI Upscaling)

---
