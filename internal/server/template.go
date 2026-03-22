package server

const HTMLTemplate = `<!DOCTYPE html>
<html lang="it">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🖼️ Galleria</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            min-height: 100vh;
            color: #eaeaea;
        }
        .header {
            position: fixed;
            top: 0; left: 0; right: 0;
            background: rgba(26, 26, 46, 0.95);
            backdrop-filter: blur(10px);
            padding: 1rem 2rem;
            z-index: 1000;
            border-bottom: 1px solid rgba(255,255,255,0.1);
            display: flex;
            gap: 1rem;
            align-items: center;
            flex-wrap: wrap;
        }
        .header h1 {
            font-size: 1.5rem;
            background: linear-gradient(90deg, #00d4ff, #7b2cbf);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-right: auto;
        }
        .search-box { position: relative; }
        .search-box input {
            background: rgba(255,255,255,0.1);
            border: 1px solid rgba(255,255,255,0.2);
            border-radius: 25px;
            padding: 0.7rem 1rem 0.7rem 2.5rem;
            color: #fff;
            width: 250px;
            font-size: 0.95rem;
            transition: all 0.3s;
        }
        .search-box input:focus {
            outline: none;
            background: rgba(255,255,255,0.15);
            border-color: #00d4ff;
            width: 300px;
        }
        .search-box::before {
            content: "🔍";
            position: absolute;
            left: 0.8rem;
            top: 50%;
            transform: translateY(-50%);
        }
        .folder-select {
            background: rgba(255,255,255,0.1);
            border: 1px solid rgba(255,255,255,0.2);
            border-radius: 8px;
            padding: 0.7rem 1rem;
            color: #fff;
            font-size: 0.95rem;
            cursor: pointer;
            min-width: 180px;
        }
        .folder-select option { background: #1a1a2e; }
        .stats { font-size: 0.85rem; color: #888; }
        .main-content {
            padding: 6rem 2rem 2rem;
            max-width: 1800px;
            margin: 0 auto;
        }
        .breadcrumb {
            display: flex;
            gap: 0.5rem;
            margin-bottom: 1.5rem;
            flex-wrap: wrap;
            align-items: center;
        }
        .breadcrumb a {
            color: #00d4ff;
            text-decoration: none;
            padding: 0.3rem 0.8rem;
            background: rgba(0, 212, 255, 0.1);
            border-radius: 20px;
            font-size: 0.9rem;
            transition: all 0.2s;
        }
        .breadcrumb a:hover { background: rgba(0, 212, 255, 0.2); }
        .breadcrumb .sep { color: #666; }
        .gallery {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
            gap: 1.5rem;
        }
        .gallery-item {
            background: rgba(255,255,255,0.05);
            border-radius: 16px;
            overflow: hidden;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            cursor: pointer;
            border: 1px solid rgba(255,255,255,0.08);
        }
        .gallery-item:hover {
            transform: translateY(-5px) scale(1.02);
            box-shadow: 0 20px 40px rgba(0,0,0,0.4);
            border-color: rgba(0, 212, 255, 0.3);
        }
        .gallery-item.folder {
            background: linear-gradient(145deg, rgba(0,212,255,0.1), rgba(123,44,191,0.1));
        }
        .gallery-item.video {
            background: linear-gradient(145deg, rgba(255,107,107,0.1), rgba(238,90,111,0.1));
        }
        .item-preview {
            aspect-ratio: 4/3;
            display: flex;
            align-items: center;
            justify-content: center;
            overflow: hidden;
            background: linear-gradient(145deg, rgba(0,0,0,0.2), rgba(0,0,0,0.4));
            position: relative;
        }
        .item-preview img {
            width: 100%; height: 100%;
            object-fit: cover;
            transition: transform 0.3s;
        }
        .gallery-item:hover .item-preview img { transform: scale(1.05); }
        .item-preview video {
            width: 100%; height: 100%;
            object-fit: cover;
        }
        .video-overlay {
            position: absolute;
            inset: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            background: rgba(0,0,0,0.3);
            transition: background 0.3s;
        }
        .gallery-item:hover .video-overlay {
            background: rgba(0,0,0,0.1);
        }
        .play-icon {
            width: 60px;
            height: 60px;
            background: rgba(255,107,107,0.9);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 1.5rem;
            transition: transform 0.3s;
        }
        .gallery-item:hover .play-icon {
            transform: scale(1.1);
            background: rgba(255,107,107,1);
        }
        .video-duration {
            position: absolute;
            bottom: 8px;
            right: 8px;
            background: rgba(0,0,0,0.8);
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 0.75rem;
            font-weight: 600;
        }
        .item-icon { font-size: 4rem; opacity: 0.8; }
        .item-info { padding: 1rem; }
        .item-name {
            font-size: 0.95rem;
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            margin-bottom: 0.3rem;
        }
        .item-meta {
            font-size: 0.75rem;
            color: #888;
            display: flex;
            justify-content: space-between;
        }
        .lightbox {
            display: none;
            position: fixed;
            inset: 0;
            background: rgba(0,0,0,0.98);
            z-index: 2000;
            justify-content: center;
            align-items: center;
        }
        .lightbox.active { display: flex; }
        .lightbox-content { 
            position: relative; 
            max-width: 95vw; 
            max-height: 95vh;
            display: flex;
            flex-direction: column;
            align-items: center;
        }
        .lightbox img, .lightbox video {
            max-width: 100%; max-height: 85vh;
            object-fit: contain;
            border-radius: 8px;
            box-shadow: 0 30px 60px rgba(0,0,0,0.5);
        }
        .lightbox video { background: #000; }
        .lightbox-close {
            position: absolute;
            top: -50px;
            right: 0;
            background: none;
            border: none;
            color: #fff;
            font-size: 2rem;
            cursor: pointer;
            opacity: 0.7;
            transition: opacity 0.2s;
            z-index: 10;
        }
        .lightbox-close:hover { opacity: 1; }
        .lightbox-nav {
            position: absolute;
            top: 50%;
            transform: translateY(-50%);
            background: rgba(255,255,255,0.1);
            border: none;
            color: #fff;
            font-size: 2rem;
            padding: 1rem;
            cursor: pointer;
            border-radius: 50%;
            width: 60px; height: 60px;
            transition: all 0.2s;
            backdrop-filter: blur(10px);
            z-index: 10;
        }
        .lightbox-nav:hover {
            background: rgba(255,255,255,0.2);
            transform: translateY(-50%) scale(1.1);
        }
        .lightbox-prev { left: -80px; }
        .lightbox-next { right: -80px; }
        .lightbox-info {
            margin-top: 1rem;
            text-align: center;
            color: #fff;
        }
        .lightbox-info .filename { font-size: 1rem; margin-bottom: 0.3rem; }
        .lightbox-info .counter { font-size: 0.85rem; color: #888; }
        .video-controls-hint {
            font-size: 0.75rem;
            color: #666;
            margin-top: 0.5rem;
        }
        .empty-state {
            text-align: center;
            padding: 5rem 2rem;
            color: #666;
        }
        .empty-state .icon { font-size: 5rem; margin-bottom: 1rem; opacity: 0.5; }
        .loading {
            display: flex;
            justify-content: center;
            padding: 3rem;
        }
        .spinner {
            width: 50px; height: 50px;
            border: 3px solid rgba(255,255,255,0.1);
            border-top-color: #00d4ff;
            border-radius: 50%;
            animation: spin 1s linear infinite;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
        @media (max-width: 768px) {
            .header { padding: 0.8rem 1rem; }
            .header h1 { font-size: 1.2rem; }
            .search-box input { width: 150px; padding: 0.5rem 0.8rem 0.5rem 2rem; }
            .search-box input:focus { width: 180px; }
            .main-content { padding: 7rem 1rem 1rem; }
            .gallery {
                grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
                gap: 1rem;
            }
            .lightbox-prev { left: 10px; }
            .lightbox-next { right: 10px; }
            .lightbox-nav { width: 45px; height: 45px; font-size: 1.5rem; }
            .play-icon { width: 45px; height: 45px; font-size: 1.2rem; }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🖼️ Galleria</h1>
        <div class="search-box">
            <input type="text" id="searchInput" placeholder="Cerca fuzzy..." autocomplete="off">
        </div>
        <select class="folder-select" id="folderSelect">
            <option value=".">📁 Cartella root</option>
        </select>
        <span class="stats" id="stats"></span>
    </div>
    <div class="main-content">
        <div class="breadcrumb" id="breadcrumb">
            <a href="#" data-folder=".">🏠 Home</a>
        </div>
        <div class="gallery" id="gallery">
            <div class="loading"><div class="spinner"></div></div>
        </div>
    </div>
    <div class="lightbox" id="lightbox">
        <div class="lightbox-content">
            <button class="lightbox-close" id="lightboxClose">✕</button>
            <button class="lightbox-nav lightbox-prev" id="lightboxPrev">‹</button>
            <div id="lightboxMedia"></div>
            <button class="lightbox-nav lightbox-next" id="lightboxNext">›</button>
            <div class="lightbox-info">
                <div class="filename" id="lightboxFilename"></div>
                <div class="counter" id="lightboxCounter"></div>
                <div class="video-controls-hint" id="videoHint"></div>
            </div>
        </div>
    </div>
    <script>
        const state = {
            currentFolder: '.',
            files: [],
            media: [],
            currentMediaIndex: 0,
            searchQuery: ''
        };
        const elements = {
            gallery: document.getElementById('gallery'),
            breadcrumb: document.getElementById('breadcrumb'),
            folderSelect: document.getElementById('folderSelect'),
            searchInput: document.getElementById('searchInput'),
            stats: document.getElementById('stats'),
            lightbox: document.getElementById('lightbox'),
            lightboxMedia: document.getElementById('lightboxMedia'),
            lightboxFilename: document.getElementById('lightboxFilename'),
            lightboxCounter: document.getElementById('lightboxCounter'),
            videoHint: document.getElementById('videoHint'),
            lightboxClose: document.getElementById('lightboxClose'),
            lightboxPrev: document.getElementById('lightboxPrev'),
            lightboxNext: document.getElementById('lightboxNext')
        };
        async function loadFiles(folder = '.', search = '') {
            elements.gallery.innerHTML = '<div class="loading"><div class="spinner"></div></div>';
            try {
                let url = search 
                    ? '/api/search?q=' + encodeURIComponent(search) + '&folder=' + encodeURIComponent(folder)
                    : '/api/files?folder=' + encodeURIComponent(folder);
                const response = await fetch(url);
                state.files = await response.json();
                state.media = state.files.filter(f => f.isImage || f.isVideo);
                renderGallery();
                updateBreadcrumb(folder);
                updateStats();
            } catch (err) {
                elements.gallery.innerHTML = '<div class="empty-state"><div class="icon">⚠️</div><p>Errore caricamento</p></div>';
            }
        }
        async function loadFolders() {
            try {
                const response = await fetch('/api/folders?folder=.');
                const folders = await response.json();
                elements.folderSelect.innerHTML = '<option value=".">📁 Cartella root</option>';
                folders.forEach(f => {
                    const option = document.createElement('option');
                    option.value = f.path;
                    option.textContent = '📁 ' + f.name;
                    elements.folderSelect.appendChild(option);
                });
            } catch (err) { console.error('Errore:', err); }
        }
        function renderGallery() {
            if (state.files.length === 0) {
                elements.gallery.innerHTML = '<div class="empty-state"><div class="icon">📂</div><p>Nessun file trovato</p></div>';
                return;
            }
            elements.gallery.innerHTML = state.files.map(file => {
                const icon = file.isDir ? '📁' : getFileIcon(file.ext);
                const size = formatSize(file.size);
                const imgUrl = file.isImage ? '/raw/' + encodePath(file.path) : '';
                const videoUrl = file.isVideo ? '/raw/' + encodePath(file.path) : '';
                const isMedia = file.isImage || file.isVideo;
                let previewHtml;
                if (file.isImage) {
                    const thumbUrl = '/thumb/' + encodePath(file.path);
                    previewHtml = '<img src="' + thumbUrl + '" loading="lazy" alt="' + file.name + '">';
                } else if (file.isVideo) {
                    previewHtml = '<video src="' + videoUrl + '" preload="metadata" muted></video>' +
                        '<div class="video-overlay"><div class="play-icon">▶</div></div>';
                } else {
                    previewHtml = '<span class="item-icon">' + icon + '</span>';
                }
                return '<div class="gallery-item ' + (file.isDir ? 'folder' : '') + ' ' + (file.isVideo ? 'video' : '') + '" data-path="' + file.path + '" data-isdir="' + file.isDir + '" data-ismedia="' + isMedia + '">' +
                    '<div class="item-preview">' + previewHtml + '</div>' +
                    '<div class="item-info">' +
                    '<div class="item-name" title="' + file.name + '">' + file.name + '</div>' +
                    '<div class="item-meta"><span>' + (file.isDir ? 'Cartella' : (file.isVideo ? '🎬 ' + size : size)) + '</span><span>' + formatDate(file.modTime) + '</span></div>' +
                    '</div></div>';
            }).join('');
            document.querySelectorAll('.gallery-item').forEach(item => {
                item.addEventListener('click', () => {
                    const isDir = item.dataset.isdir === 'true';
                    const isMedia = item.dataset.ismedia === 'true';
                    const path = item.dataset.path;
                    if (isDir) {
                        state.currentFolder = path;
                        elements.folderSelect.value = path;
                        state.searchQuery = '';
                        elements.searchInput.value = '';
                        loadFiles(path);
                    } else if (isMedia) {
                        openLightbox(path);
                    }
                });
            });
        }
        function openLightbox(path) {
            const index = state.media.findIndex(m => m.path === path);
            if (index === -1) return;
            state.currentMediaIndex = index;
            showMedia(index);
            elements.lightbox.classList.add('active');
            document.body.style.overflow = 'hidden';
        }
        function showMedia(index) {
            const media = state.media[index];
            const mediaUrl = '/raw/' + encodePath(media.path);
            elements.lightboxMedia.innerHTML = '';
            if (media.isVideo) {
                const video = document.createElement('video');
                video.src = mediaUrl;
                video.controls = true;
                video.autoplay = true;
                video.style.maxWidth = '100%';
                video.style.maxHeight = '85vh';
                elements.lightboxMedia.appendChild(video);
                elements.videoHint.textContent = 'Spazio per play/pause • ← → per navigare • ESC per chiudere';
            } else {
                const img = document.createElement('img');
                img.src = mediaUrl;
                img.alt = media.name;
                elements.lightboxMedia.appendChild(img);
                elements.videoHint.textContent = '';
            }
            elements.lightboxFilename.textContent = media.name;
            elements.lightboxCounter.textContent = (index + 1) + ' / ' + state.media.length + (media.isVideo ? ' 🎬' : ' 🖼️');
        }
        function closeLightbox() {
            const video = elements.lightboxMedia.querySelector('video');
            if (video) video.pause();
            elements.lightbox.classList.remove('active');
            document.body.style.overflow = '';
        }
        function nextMedia() {
            state.currentMediaIndex = (state.currentMediaIndex + 1) % state.media.length;
            showMedia(state.currentMediaIndex);
        }
        function prevMedia() {
            state.currentMediaIndex = (state.currentMediaIndex - 1 + state.media.length) % state.media.length;
            showMedia(state.currentMediaIndex);
        }
        function updateBreadcrumb(folder) {
            if (folder === '.') {
                elements.breadcrumb.innerHTML = '<a href="#" data-folder=".">🏠 Home</a>';
                return;
            }
            const parts = folder.split('/').filter(p => p);
            let html = '<a href="#" data-folder=".">🏠 Home</a>';
            let currentPath = '';
            parts.forEach(part => {
                currentPath += (currentPath ? '/' : '') + part;
                html += '<span class="sep">/</span><a href="#" data-folder="' + currentPath + '">' + part + '</a>';
            });
            elements.breadcrumb.innerHTML = html;
            document.querySelectorAll('.breadcrumb a').forEach(link => {
                link.addEventListener('click', (e) => {
                    e.preventDefault();
                    const folder = link.dataset.folder;
                    state.currentFolder = folder;
                    elements.folderSelect.value = folder;
                    loadFiles(folder);
                });
            });
        }
        function updateStats() {
            const images = state.files.filter(f => f.isImage).length;
            const videos = state.files.filter(f => f.isVideo).length;
            const folders = state.files.filter(f => f.isDir).length;
            const others = state.files.length - images - videos - folders;
            let text = state.files.length + ' elementi';
            if (images) text += ' • ' + images + ' 🖼️';
            if (videos) text += ' • ' + videos + ' 🎬';
            if (folders) text += ' • ' + folders + ' 📁';
            if (others) text += ' • ' + others + ' 📄';
            elements.stats.textContent = text;
        }
        function encodePath(path) {
            return path.split('/').map(encodeURIComponent).join('/');
        }
        function getFileIcon(ext) {
            const icons = {
                '.jpg': '🖼️', '.jpeg': '🖼️', '.png': '🖼️', '.gif': '🖼️', '.webp': '🖼️',
                '.mp4': '🎬', '.webm': '🎬', '.mov': '🎬', '.avi': '🎬', '.mkv': '🎬', '.flv': '🎬', '.wmv': '🎬',
                '.mp3': '🎵', '.wav': '🎵', '.flac': '🎵',
                '.pdf': '📄', '.doc': '📝', '.docx': '📝', '.txt': '📝',
                '.zip': '📦', '.rar': '📦', '.7z': '📦',
                '.go': '🔵', '.js': '🟡', '.ts': '🔷', '.py': '🐍',
                '.html': '🌐', '.css': '🎨', '.json': '📋'
            };
            return icons[ext] || '📄';
        }
        function formatSize(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        }
        function formatDate(dateStr) {
            return new Date(dateStr).toLocaleDateString('it-IT', { day: '2-digit', month: '2-digit', year: 'numeric' });
        }
        elements.folderSelect.addEventListener('change', () => {
            state.currentFolder = elements.folderSelect.value;
            state.searchQuery = '';
            elements.searchInput.value = '';
            loadFiles(state.currentFolder);
        });
        let searchTimeout;
        elements.searchInput.addEventListener('input', (e) => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                state.searchQuery = e.target.value;
                loadFiles(state.currentFolder, state.searchQuery);
            }, 300);
        });
        elements.lightboxClose.addEventListener('click', closeLightbox);
        elements.lightboxNext.addEventListener('click', (e) => { e.stopPropagation(); nextMedia(); });
        elements.lightboxPrev.addEventListener('click', (e) => { e.stopPropagation(); prevMedia(); });
        elements.lightbox.addEventListener('click', (e) => { 
            if (e.target === elements.lightbox || e.target.closest('.lightbox-content') === elements.lightbox.querySelector('.lightbox-content')) {
                if (e.target.tagName !== 'VIDEO') closeLightbox();
            }
        });
        document.addEventListener('keydown', (e) => {
            if (!elements.lightbox.classList.contains('active')) return;
            const video = elements.lightboxMedia.querySelector('video');
            if (e.key === 'Escape') closeLightbox();
            else if (e.key === 'ArrowRight') nextMedia();
            else if (e.key === 'ArrowLeft') prevMedia();
            else if (e.key === ' ' && video) {
                e.preventDefault();
                video.paused ? video.play() : video.pause();
            }
        });
        loadFolders();
        loadFiles();
    </script>
</body>
</html>`
