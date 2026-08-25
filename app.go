package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	selectedImage string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) OpenVideo() string {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Pilih Media Aset",
		Filters: []runtime.FileFilter{
			{DisplayName: "Video Files (*.mp4, *.mkv, *.mov, etc)",
				Pattern: "*.mp4;*.mkv;*.mov;*.avi;*.webm;*.flv"},
		},
	})
	if err != nil || filePath == "" {
		return "Batal milih file."
	}
	a.selectedImage = filePath
	return fmt.Sprintf("Aset siap: %s", filePath)
}

func (a *App) OpenImage() string {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Pilih Media Aset",
		Filters: []runtime.FileFilter{
			{DisplayName: "Image Files (*.jpg, *.png, *.webp, etc)",
				Pattern: "*.jpg;*.jpeg;*.png;*.webp;*.bmp;*.tiff"},
		},
	})
	if err != nil || filePath == "" {
		return "Batal milih file."
	}
	a.selectedImage = filePath
	return fmt.Sprintf("Aset siap: %s", filePath)
}

func (a *App) OpenAudio() string {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Pilih Media Aset",
		Filters: []runtime.FileFilter{
			{DisplayName: "Audio Files (*.mp3, *.wav, *.aac, etc)",
				Pattern: "*.mp3;*.wav;*.aac;*.flac;*.ogg"},
		},
	})
	if err != nil || filePath == "" {
		return "Batal milih file."
	}
	a.selectedImage = filePath
	return fmt.Sprintf("Aset siap: %s", filePath)
}

func (a *App) OpenDirectory() string {
	dirPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Pilih Folder Output",
	})
	if err != nil || dirPath == "" {
		return ""
	}
	return dirPath
}

type ProcessOptions struct {
	MediaType  string `json:"mediaType"`
	Mode       string `json:"mode"`
	Value      string `json:"value"`
	OutputDir  string `json:"outputDir"`
	VramMode   string `json:"vramMode"`
	Engine     string `json:"engine"`
	Resolution string `json:"resolution"`
}

const maxMediaSize = 4 * 1024 * 1024 * 1024

func (a *App) emitProgress(percent int, label string) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "media-progress", map[string]interface{}{
			"percent": percent,
			"label":   label,
		})
	}
}

func probeDuration(ffmpegPath, inputPath string) float64 {
	output, _ := exec.Command(ffmpegPath, "-hide_banner", "-i", inputPath).CombinedOutput()
	durationPattern := regexp.MustCompile(`Duration: (\d+):(\d+):(\d+(?:\.\d+)?)`)
	matches := durationPattern.FindStringSubmatch(string(output))
	if len(matches) != 4 {
		return 0
	}
	hours, _ := strconv.ParseFloat(matches[1], 64)
	minutes, _ := strconv.ParseFloat(matches[2], 64)
	seconds, _ := strconv.ParseFloat(matches[3], 64)
	return hours*3600 + minutes*60 + seconds
}

func (a *App) runFFmpegWithProgress(ffmpegPath string, args []string, duration float64, startPercent, endPercent int, label string) ([]byte, error) {
	args = append([]string{"-progress", "pipe:1", "-nostats"}, args...)
	command := exec.Command(ffmpegPath, args...)
	progressOutput, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var errorOutput bytes.Buffer
	command.Stderr = &errorOutput
	if err := command.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(progressOutput)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "out_time_us=") || duration <= 0 {
			continue
		}
		microseconds, parseErr := strconv.ParseFloat(strings.TrimPrefix(line, "out_time_us="), 64)
		if parseErr != nil {
			continue
		}
		progress := float64(startPercent) + (microseconds/(duration*1000000))*float64(endPercent-startPercent)
		if progress > float64(endPercent) {
			progress = float64(endPercent)
		}
		a.emitProgress(int(progress), label)
	}

	waitErr := command.Wait()
	if waitErr != nil {
		return errorOutput.Bytes(), waitErr
	}
	a.emitProgress(endPercent, label)
	return errorOutput.Bytes(), nil
}

func (a *App) ProcessMedia(opts ProcessOptions) string {
	a.emitProgress(0, "Checking media")
	if a.selectedImage == "" {
		a.emitProgress(0, "Processing failed")
		return "Error: Lu belum milih file satupun, bre!"
	}

	mediaInfo, err := os.Stat(a.selectedImage)
	if err != nil {
		a.emitProgress(0, "Processing failed")
		return fmt.Sprintf("Error: File tidak bisa dibaca: %v", err)
	}
	if mediaInfo.Size() > maxMediaSize {
		a.emitProgress(0, "Processing failed")
		return "Error: Ukuran video melebihi batas maksimum 4 GB. Pilih file yang lebih kecil."
	}

	outDir := opts.OutputDir
	if outDir == "" {
		outDir = filepath.Join(".", "Output")
	}
	os.MkdirAll(outDir, os.ModePerm)
	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)

	ffmpegPath := filepath.Join(exPath, "build", "tools", "ffmpeg", "ffmpeg.exe")
	esrganPath := filepath.Join(exPath, "build", "tools", "realesrgan", "realesrgan.exe")

	ext := filepath.Ext(a.selectedImage)
	baseName := strings.TrimSuffix(filepath.Base(a.selectedImage), ext)
	outPath := filepath.Join(outDir, fmt.Sprintf("%s_resynth%s", baseName, ext))

	var cmdErr error
	var outLog []byte

	tileSize := "256"
	if opts.VramMode == "potato" {
		tileSize = "64"
	}
	if opts.VramMode == "gaming" {
		tileSize = "128"
	}
	if opts.VramMode == "beast" {
		tileSize = "512"
	}

	if opts.Engine == "" {
		opts.Engine = "realesrgan"
	}
	if opts.Engine != "ffmpeg" && opts.Engine != "realesrgan" {
		opts.Engine = "realesrgan"
	}
	if opts.Resolution != "2" && opts.Resolution != "3" && opts.Resolution != "4" {
		opts.Resolution = "2"
	}
	a.emitProgress(10, "Preparing encoder")
	inputDuration := probeDuration(ffmpegPath, a.selectedImage)

	if opts.MediaType == "foto" && opts.Mode == "upscale" {
		outPath = filepath.Join(outDir, fmt.Sprintf("%s_resynth_HD.png", baseName))
		a.emitProgress(20, "Upscaling image")
		if opts.Engine == "ffmpeg" {
			outLog, cmdErr = exec.Command(ffmpegPath, "-hide_banner", "-y", "-i", a.selectedImage, "-vf", fmt.Sprintf("scale=iw*%s:ih*%s:flags=lanczos", opts.Resolution, opts.Resolution), outPath).CombinedOutput()
		} else {
			outLog, cmdErr = exec.Command(esrganPath, "-i", a.selectedImage, "-o", outPath, "-s", opts.Resolution, "-t", tileSize).CombinedOutput()
		}

	} else if opts.MediaType == "foto" && opts.Mode == "compress" {
		outPath = filepath.Join(outDir, fmt.Sprintf("%s_compressed_resynth.jpg", baseName))
		a.emitProgress(20, "Compressing image")
		outLog, cmdErr = exec.Command(ffmpegPath, "-hide_banner", "-y", "-i", a.selectedImage, "-q:v", opts.Value, outPath).CombinedOutput()

		// --- BLOK KHUSUS VIDEO (Yang lama biarin aja di bawah sini) ---
	} else if opts.MediaType == "video" && opts.Mode == "compress" {
		a.emitProgress(20, "Compressing video")
		outLog, cmdErr = a.runFFmpegWithProgress(ffmpegPath, []string{"-hide_banner", "-y", "-i", a.selectedImage, "-c:v", "libsvtav1", "-preset", "5", "-crf", opts.Value, outPath}, inputDuration, 20, 100, "Compressing video")

	} else if opts.MediaType == "video" && opts.Mode == "upscale" && opts.Engine == "ffmpeg" {
		a.emitProgress(20, "Upscaling with FFmpeg")
		outPath = filepath.Join(outDir, fmt.Sprintf("%s_resynth_HD.mp4", baseName))
		outLog, cmdErr = a.runFFmpegWithProgress(ffmpegPath, []string{"-hide_banner", "-y", "-i", a.selectedImage, "-vf", fmt.Sprintf("scale=iw*%s:ih*%s:flags=lanczos", opts.Resolution, opts.Resolution), "-c:v", "libsvtav1", "-preset", "5", "-crf", "18", "-c:a", "aac", "-pix_fmt", "yuv420p", outPath}, inputDuration, 20, 100, "Upscaling with FFmpeg")

	} else if opts.MediaType == "audio" && opts.Mode == "compress" {
		outPath = filepath.Join(outDir, fmt.Sprintf("%s_compressed.mp3", baseName))

		bitrate := fmt.Sprintf("%vk", opts.Value)

		outLog, cmdErr = exec.Command(ffmpegPath, "-hide_banner", "-y", "-i", a.selectedImage, "-b:a", bitrate, outPath).CombinedOutput()
	} else if opts.MediaType == "video" && opts.Mode == "upscale" && opts.Engine == "realesrgan" {

		tempIn := filepath.Join(outDir, "temp_in")
		tempOut := filepath.Join(outDir, "temp_out")
		os.MkdirAll(tempIn, os.ModePerm)
		os.MkdirAll(tempOut, os.ModePerm)

		outPath = filepath.Join(outDir, fmt.Sprintf("%s_resynth_HD.mp4", baseName))
		audioPath := filepath.Join(outDir, "temp_audio.aac")

		exec.Command(ffmpegPath, "-hide_banner", "-y", "-i", a.selectedImage, "-vn", "-acodec", "copy", audioPath).Run()

		framePattern := filepath.Join(tempIn, "frame_%08d.jpg")
		a.emitProgress(20, "Extracting video frames")
		outLog, cmdErr = exec.Command(ffmpegPath, "-hide_banner", "-y", "-i", a.selectedImage, "-qscale:v", "2", framePattern).CombinedOutput()

		if cmdErr == nil {
			a.emitProgress(45, "Upscaling video frames")
			outLog, cmdErr = exec.Command(esrganPath, "-i", tempIn, "-o", tempOut, "-s", opts.Resolution, "-t", tileSize).CombinedOutput()
		}

		if cmdErr == nil {
			a.emitProgress(75, "Encoding final video")
			upFramePattern := filepath.Join(tempOut, "frame_%08d.jpg")
			if _, statErr := os.Stat(audioPath); statErr == nil {
				outLog, cmdErr = a.runFFmpegWithProgress(ffmpegPath, []string{"-hide_banner", "-y", "-framerate", "24", "-i", upFramePattern, "-i", audioPath, "-c:v", "libsvtav1", "-crf", "18", "-c:a", "aac", "-pix_fmt", "yuv420p", outPath}, inputDuration, 75, 100, "Encoding final video")
			} else {
				outLog, cmdErr = a.runFFmpegWithProgress(ffmpegPath, []string{"-hide_banner", "-y", "-framerate", "24", "-i", upFramePattern, "-c:v", "libsvtav1", "-preset", "5", "-crf", "18", "-pix_fmt", "yuv420p", outPath}, inputDuration, 75, 100, "Encoding final video")
			}
		}

		os.RemoveAll(tempIn)
		os.RemoveAll(tempOut)
		os.Remove(audioPath)

	} else {
		a.emitProgress(0, "Processing failed")
		return fmt.Sprintf("Dev Note: Mode %s buat %s belum ada scriptnya!", strings.ToUpper(opts.Mode), strings.ToUpper(opts.MediaType))
	}

	if cmdErr != nil {
		a.emitProgress(0, "Processing failed")
		lines := strings.Split(string(outLog), "\n")

		startIdx := len(lines) - 10
		if startIdx < 0 {
			startIdx = 0
		}
		shortLog := strings.Join(lines[startIdx:], "\n")

		return fmt.Sprintf("Error :\n%s\n\nKode OS: %v", shortLog, cmdErr)
	}

	a.emitProgress(100, "Complete")
	return fmt.Sprintf("SUKSES! File disave di:\n%s", outPath)
}
