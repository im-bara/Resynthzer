import './style.css';
import './app.css';

import { OpenDirectory, OpenImage, ProcessMedia } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

const viewMainMenu = document.getElementById('view-main-menu') as HTMLDivElement | null;
const viewWorkspace = document.getElementById('view-workspace') as HTMLDivElement | null;
const btnEnter = document.getElementById('btn-enter') as HTMLButtonElement | null;
const btnBack = document.getElementById('btn-back') as HTMLButtonElement | null;
const executeButtons = document.querySelectorAll('.btn-execute') as NodeListOf<HTMLButtonElement>;
const btnSelectDir = document.getElementById('btn-select-dir') as HTMLButtonElement | null;
const outputDirText = document.getElementById('output-dir-text') as HTMLParagraphElement | null;
const statusToast = document.getElementById('status-toast') as HTMLDivElement | null;
const statusTitle = document.getElementById('status-title') as HTMLElement | null;
const statusMessage = document.getElementById('status-message') as HTMLElement | null;
const statusClose = document.getElementById('status-close') as HTMLButtonElement | null;
const progressDock = document.getElementById('progress-dock') as HTMLDivElement | null;
const progressFill = document.getElementById('progress-fill') as HTMLDivElement | null;
const menuItems = document.querySelectorAll('.menu-item');
const tabContents = document.querySelectorAll('.tab-content');
const executableTabs = ['tab-video', 'tab-foto', 'tab-audio'];
let currentOutputDir = '';
let toastTimer: number | undefined;
let startupTimer: number | undefined;

EventsOn('media-progress', (progress: { percent: number; label: string }) => {
    if (!progressDock || !progressFill) return;

    const percent = Math.max(0, Math.min(100, progress.percent));
    progressFill.style.width = `${percent}%`;
    progressDock.setAttribute('aria-hidden', 'false');
    progressDock.classList.toggle('complete', percent === 100);
    progressDock.classList.add('visible');

    if (percent === 100 || progress.label === 'Processing failed') {
        window.setTimeout(() => {
            progressDock.classList.remove('visible');
            progressDock.setAttribute('aria-hidden', 'true');
        }, 900);
    }
});

function showStatus(title: string, message: string, type: 'success' | 'error' | 'working' = 'success') {
    if (!statusToast || !statusTitle || !statusMessage) return;

    window.clearTimeout(toastTimer);
    statusTitle.textContent = title;
    statusMessage.textContent = message;
    statusToast.className = `status-toast ${type}`;
    statusToast.setAttribute('aria-hidden', 'false');

    if (type !== 'error') {
        toastTimer = window.setTimeout(() => {
            statusToast.classList.remove('visible');
            statusToast.setAttribute('aria-hidden', 'true');
        }, 3000);
    }
    requestAnimationFrame(() => statusToast.classList.add('visible'));
}

statusClose?.addEventListener('click', () => {
    if (!statusToast) return;
    statusToast.classList.remove('visible');
    statusToast.setAttribute('aria-hidden', 'true');
});

// Hapus baris ini karena tidak digunakan:
// const result = await OpenImage();

function setTab(targetId: string) {
    menuItems.forEach((item) => item.classList.toggle('active', item.getAttribute('data-target') === targetId));
    tabContents.forEach((tab) => tab.classList.toggle('hidden', tab.id !== targetId));

    // Set default active button pas ganti tab
    const activeTab = document.getElementById(targetId);
    if (activeTab) {
        const defaultBtn = activeTab.querySelector('.mode-upscale') as HTMLButtonElement;
        defaultBtn?.click();
    }

    executeButtons.forEach((button) => {
        button.classList.toggle('hidden', !executableTabs.includes(targetId));
    });
}

if (btnEnter && btnBack && viewMainMenu && viewWorkspace) {
    btnEnter.addEventListener('click', () => {
        window.clearTimeout(startupTimer);
        viewMainMenu.classList.add('hidden');
        viewWorkspace.classList.remove('hidden');
    });

    btnBack.addEventListener('click', () => {
        viewWorkspace.classList.add('hidden');
        viewMainMenu.classList.remove('hidden');
    });
}

showStatus('Resynthzer', 'Selamat datang di media processing engine.', 'working');
startupTimer = window.setTimeout(() => {
    viewMainMenu?.classList.add('hidden');
    viewWorkspace?.classList.remove('hidden');
}, 5000);

menuItems.forEach((menu) => {
    menu.addEventListener('click', () => {
        const targetId = menu.getAttribute('data-target');
        if (targetId) {
            setTab(targetId);
        }
    });
});

if (btnSelectDir && outputDirText) {
    btnSelectDir.addEventListener('click', async () => {
        try {
            const dir = await OpenDirectory();
            if (dir) {
                currentOutputDir = dir;
                outputDirText.textContent = `Output folder: ${dir}`;
                outputDirText.style.color = 'var(--accent-green)';
                btnSelectDir.textContent = 'Change Folder...';
            }
        } catch (error) {
            console.error('Failed to open folder:', error);
        }
    });
}

const dropZones = document.querySelectorAll('.drop-zone');
dropZones.forEach((zone) => {
    zone.addEventListener('click', async (event: Event) => {
        event.preventDefault();

        const dropText = zone.querySelector('.drop-text') as HTMLParagraphElement | null;
        if (!dropText) return;

        const originalText = dropText.textContent ?? '';
        dropText.textContent = 'Opening file picker...';
        dropText.style.color = 'var(--text-main)';

        try {
            const fileResult = await OpenImage();
            if (!fileResult || fileResult.toLowerCase().includes('cancel')) {
                dropText.textContent = originalText;
                dropText.style.color = 'var(--text-muted)';
                return;
            }

            dropText.textContent = fileResult;
            dropText.style.color = 'var(--accent-green)';
        } catch (error) {
            dropText.textContent = 'Failed to load file.';
            dropText.style.color = 'var(--accent-red)';
            console.error(error);
        }
    });
});

tabContents.forEach((tab) => {
    const btnUpscale = tab.querySelector('.mode-upscale') as HTMLButtonElement | null;
    const btnCompress = tab.querySelector('.mode-compress') as HTMLButtonElement | null;
    const panelUpscale = tab.querySelector('.panel-upscale') as HTMLDivElement | null;
    const panelCompress = tab.querySelector('.panel-compress') as HTMLDivElement | null;

    if (btnUpscale && btnCompress && panelUpscale && panelCompress) {
        btnUpscale.addEventListener('click', () => {
            btnUpscale.classList.add('active');
            btnCompress.classList.remove('active');
            panelUpscale.classList.remove('hidden');
            panelCompress.classList.add('hidden');
        });

        btnCompress.addEventListener('click', () => {
            btnCompress.classList.add('active');
            btnUpscale.classList.remove('active');
            panelCompress.classList.remove('hidden');
            panelUpscale.classList.add('hidden');
        });
    }

    const slider = tab.querySelector('.slider') as HTMLInputElement | null;
    const sliderText = tab.querySelector('.slider-val') as HTMLSpanElement | null;
    if (slider && sliderText) {
        slider.addEventListener('input', (event) => {
            const value = (event.target as HTMLInputElement).value;
            sliderText.textContent = tab.id === 'tab-video' ? `${value} (CRF)` : `${value}% Quality`;
        });
    }
});

executeButtons.forEach((btnExecute) => {
    btnExecute.addEventListener('click', async () => {
        const activeTab = document.querySelector('.tab-content:not(.hidden)') as HTMLElement | null;
        if (!activeTab) {
            showStatus('Tidak ada tab aktif', 'Pilih jenis media terlebih dahulu.', 'error');
            return;
        }

        const originalText = btnExecute.textContent ?? 'Generate Output';
        btnExecute.textContent = 'Processing...';
        btnExecute.disabled = true;
        showStatus('Sedang diproses', 'Mesin sedang menyiapkan output kamu.', 'working');

        const mediaTypeMap: Record<string, string> = {
            'tab-video': 'video',
            'tab-foto': 'foto',
            'tab-audio': 'audio',
        };
        const activeModeButton = activeTab.querySelector('.toggle-btn.active') as HTMLElement | null;
        const mode = activeModeButton?.classList.contains('mode-upscale') ? 'upscale' : 'compress';
        const mediaType = mediaTypeMap[activeTab.id] ?? 'unknown';

        let val = "0";
        if (mode === 'compress') {
            const slider = activeTab.querySelector('.slider') as HTMLInputElement;
            if (slider) val = slider.value;
        }

        const vramSelect = document.getElementById('setting-vram') as HTMLSelectElement | null;
        const vramMode = vramSelect ? vramSelect.value : 'gaming';
        const engineSelect = activeTab.querySelector('.engine-select') as HTMLSelectElement | null;
        const engine = engineSelect ? engineSelect.value : 'realesrgan';
        const resolutionSelect = activeTab.querySelector('.resolution-select') as HTMLSelectElement | null;
        const resolution = resolutionSelect ? resolutionSelect.value : '2';

        const payload = {
            mediaType: mediaType,
            mode: mode,
            value: val,
            outputDir: currentOutputDir,
            vramMode: vramMode,
            engine: engine,
            resolution: resolution,
        };

        try {
            const result = await ProcessMedia(payload);
            const failed = result.toLowerCase().includes('error');
            showStatus(failed ? 'Proses gagal' : 'Output selesai', result, failed ? 'error' : 'success');
        } catch (error) {
            showStatus('System error', String(error), 'error');
        } finally {
            btnExecute.textContent = originalText;
            btnExecute.disabled = false;
        }
    });
});