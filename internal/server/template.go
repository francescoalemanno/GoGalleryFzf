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
            gap: 0.3rem;
            margin-bottom: 1.5rem;
            flex-wrap: wrap;
            align-items: center;
            padding: 0.5rem 1rem;
            background: rgba(0, 0, 0, 0.2);
            border-radius: 12px;
            border: 1px solid rgba(255,255,255,0.05);
        }
        .breadcrumb a {
            color: #00d4ff;
            text-decoration: none;
            padding: 0.4rem 0.8rem;
            background: rgba(0, 212, 255, 0.1);
            border-radius: 8px;
            font-size: 0.9rem;
            transition: all 0.2s;
            display: flex;
            align-items: center;
            gap: 0.3rem;
        }
        .breadcrumb a:hover { 
            background: rgba(0, 212, 255, 0.25);
            transform: translateY(-1px);
        }
        .breadcrumb a:first-child {
            background: linear-gradient(135deg, rgba(0, 212, 255, 0.2), rgba(123, 44, 191, 0.2));
            font-weight: 500;
        }
        .breadcrumb a:first-child:hover {
            background: linear-gradient(135deg, rgba(0, 212, 255, 0.3), rgba(123, 44, 191, 0.3));
        }
        .breadcrumb .sep { 
            color: #666; 
            padding: 0 0.2rem;
            font-size: 0.8rem;
        }
        .breadcrumb .parent-nav {
            margin-left: auto;
            background: rgba(255, 255, 255, 0.08) !important;
            color: #aaa !important;
        }
        .breadcrumb .parent-nav:hover {
            background: rgba(255, 255, 255, 0.15) !important;
            color: #fff !important;
        }
        .breadcrumb .sep { color: #666; }
        .gallery {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
            gap: 1.5rem;
            min-height: 200px;
        }
        .gallery-item {
            background: rgba(255,255,255,0.05);
            border-radius: 16px;
            overflow: hidden;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            cursor: pointer;
            border: 1px solid rgba(255,255,255,0.08);
            animation: fadeIn 0.3s ease-out;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
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
        .item-dir {
            font-size: 0.7rem;
            color: #666;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            margin-bottom: 0.2rem;
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
        .lightbox-info .filepath {
            font-size: 0.8rem;
            color: #888;
            margin-bottom: 0.3rem;
            word-break: break-all;
            max-width: 80vw;
        }
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
            grid-column: 1 / -1;
        }
        .empty-state .icon { font-size: 5rem; margin-bottom: 1rem; opacity: 0.5; }
        .loading {
            display: flex;
            justify-content: center;
            padding: 3rem;
            grid-column: 1 / -1;
        }
        .spinner {
            width: 50px; height: 50px;
            border: 3px solid rgba(255,255,255,0.1);
            border-top-color: #00d4ff;
            border-radius: 50%;
            animation: spin 1s linear infinite;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
        .load-more-container {
            grid-column: 1 / -1;
            display: flex;
            justify-content: center;
            padding: 2rem;
        }
        .load-more-btn {
            background: linear-gradient(90deg, #00d4ff, #7b2cbf);
            border: none;
            color: #fff;
            padding: 1rem 2rem;
            border-radius: 30px;
            font-size: 1rem;
            cursor: pointer;
            transition: all 0.3s;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        .load-more-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 30px rgba(0,212,255,0.3);
        }
        .load-more-btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none;
        }
        .pagination-info {
            grid-column: 1 / -1;
            text-align: center;
            color: #888;
            font-size: 0.9rem;
            padding: 1rem;
        }
        .sentinel {
            height: 20px;
            grid-column: 1 / -1;
        }
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
        <div class="sentinel" id="sentinel"></div>
    </div>
    <div class="lightbox" id="lightbox">
        <div class="lightbox-content">
            <button class="lightbox-close" id="lightboxClose">✕</button>
            <button class="lightbox-nav lightbox-prev" id="lightboxPrev">‹</button>
            <div id="lightboxMedia"></div>
            <button class="lightbox-nav lightbox-next" id="lightboxNext">›</button>
            <div class="lightbox-info">
                <div class="filename" id="lightboxFilename"></div>
                <div class="filepath" id="lightboxFilepath"></div>
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
            searchQuery: '',
            currentPage: 1,
            hasMore: false,
            isLoading: false,
            totalFiles: 0
        };
        const mediaTypes = { image: 1, video: 1, audio: 1 };
        const elements = {
            gallery: document.getElementById('gallery'),
            breadcrumb: document.getElementById('breadcrumb'),
            folderSelect: document.getElementById('folderSelect'),
            searchInput: document.getElementById('searchInput'),
            stats: document.getElementById('stats'),
            sentinel: document.getElementById('sentinel'),
            lightbox: document.getElementById('lightbox'),
            lightboxMedia: document.getElementById('lightboxMedia'),
            lightboxFilename: document.getElementById('lightboxFilename'),
            lightboxFilepath: document.getElementById('lightboxFilepath'),
            lightboxCounter: document.getElementById('lightboxCounter'),
            videoHint: document.getElementById('videoHint'),
            lightboxClose: document.getElementById('lightboxClose'),
            lightboxPrev: document.getElementById('lightboxPrev'),
            lightboxNext: document.getElementById('lightboxNext')
        };

        // Intersection Observer for infinite scroll
        const observer = new IntersectionObserver((entries) => {
            if (entries[0].isIntersecting && state.hasMore && !state.isLoading && !loadMoreTimeout) {
                // Debounce: wait a bit to ensure we're really done scrolling
                loadMoreTimeout = setTimeout(() => {
                    loadMoreTimeout = null;
                    if (state.hasMore && !state.isLoading) {
                        loadMore();
                    }
                }, 100);
            }
        }, { rootMargin: '200px' });

        // Track which files are already rendered to prevent duplicates
        const renderedFiles = new Set();
        // Track which page requests are in-flight to prevent duplicate requests
        const loadingPages = new Set();
        // Debounce timer for infinite scroll
        let loadMoreTimeout = null;

        async function loadFiles(folder = '.', search = '', page = 1, append = false) {
            if (!append) {
                state.currentPage = 1;
                state.files = [];
                state.media = [];
                renderedFiles.clear(); // Clear tracking on fresh load
                loadingPages.clear();  // Clear in-flight tracking
                elements.gallery.innerHTML = '<div class="loading"><div class="spinner"></div></div>';
                // Update breadcrumb immediately to show where we are
                updateBreadcrumb(folder);
            }
            
            state.isLoading = true;
            
            try {
                let url = search 
                    ? '/api/search?q=' + encodeURIComponent(search) + '&folder=' + encodeURIComponent(folder) + '&page=' + page + '&limit=100'
                    : '/api/files?folder=' + encodeURIComponent(folder) + '&page=' + page + '&limit=100';
                
                const response = await fetch(url);
                
                if (!response.ok) {
                    throw new Error('HTTP ' + response.status + ': ' + response.statusText);
                }
                
                const data = await response.json();
                
                // Check if data is valid
                if (!data || !Array.isArray(data.files)) {
                    throw new Error('Invalid response data');
                }
                
                if (!append) {
                    state.files = data.files;
                    elements.gallery.innerHTML = '';
                } else {
                    state.files = state.files.concat(data.files);
                }
                
                state.media = state.files.filter(f => f.isImage || f.isVideo || f.isAudio);
                state.hasMore = data.hasMore;
                state.totalFiles = data.total;
                state.currentPage = data.page;
                
                // Check if folder is truly empty (server returned no files)
                if (data.files.length === 0) {
                    appendGallery([], append);
                } else {
                    // Filter out any duplicates before appending
                    const newFiles = append ? data.files.filter(f => !renderedFiles.has(f.path)) : data.files;
                    
                    // Track newly added files
                    newFiles.forEach(f => renderedFiles.add(f.path));
                    
                    appendGallery(newFiles, append);
                }
                
                updateStats(data);
                
                // Observe sentinel for infinite scroll (only once)
                if (elements.sentinel && !append) {
                    observer.observe(elements.sentinel);
                }
            } catch (err) {
                if (!append) {
                    elements.gallery.innerHTML = '<div class="empty-state"><div class="icon">⚠️</div><p>Errore caricamento: ' + escapeHtml(err.message) + '</p></div>';
                }
                console.error('Errore caricamento:', err);
            } finally {
                state.isLoading = false;
            }
        }

        async function loadMore() {
            // Prevent concurrent/duplicate requests with multiple guards
            if (state.isLoading || !state.hasMore) return;
            
            // Calculate next page
            const nextPage = state.currentPage + 1;
            
            // Skip if this page is already being loaded
            if (loadingPages.has(nextPage)) return;
            
            // Mark this page as loading
            loadingPages.add(nextPage);
            state.isLoading = true;
            
            try {
                await loadFiles(state.currentFolder, state.searchQuery, nextPage, true);
            } finally {
                loadingPages.delete(nextPage);
                state.isLoading = false;
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

        function createGalleryItem(file) {
            const icon = file.isDir ? '📁' : getFileIcon(file.ext);
            const size = formatSize(file.size);
            const isMedia = file.isImage || file.isVideo || file.isAudio;
            // Extract directory path for display
            const lastSlash = file.path.lastIndexOf('/');
            const dirPath = lastSlash > 0 ? file.path.substring(0, lastSlash) : '';
            const dirDisplay = dirPath ? '📂 ' + dirPath : '📂 root';
            let previewHtml;
            if (file.isImage) {
                const thumbUrl = '/thumb/' + encodePath(file.path);
                previewHtml = '<img src="' + thumbUrl + '" loading="lazy" alt="' + file.name + '">';
            } else if (file.isVideo) {
                const videoUrl = '/raw/' + encodePath(file.path);
                previewHtml = '<video src="' + videoUrl + '" preload="metadata" muted></video>' +
                    '<div class="video-overlay"><div class="play-icon">▶</div></div>';
            } else if (file.isAudio) {
                previewHtml = '<span class="item-icon">🎵</span>';
            } else {
                previewHtml = '<span class="item-icon">' + icon + '</span>';
            }
            const div = document.createElement('div');
            div.className = 'gallery-item ' + (file.isDir ? 'folder' : '') + ' ' + (file.isVideo ? 'video' : '') + ' ' + (file.isAudio ? 'audio' : '');
            div.dataset.path = file.path;
            div.dataset.isdir = file.isDir;
            div.dataset.ismedia = isMedia;
            div.innerHTML = '<div class="item-preview">' + previewHtml + '</div>' +
                '<div class="item-info">' +
                '<div class="item-name" title="' + file.name + '">' + file.name + '</div>' +
                (file.isDir ? '' : '<div class="item-dir" title="' + dirPath + '">' + dirDisplay + '</div>') +
                '<div class="item-meta"><span>' + (file.isDir ? 'Cartella' : (file.isVideo ? '🎬 ' + size : (file.isAudio ? '🎵 ' + size : size))) + '</span><span>' + formatDate(file.modTime) + '</span></div>' +
                '</div>';
            div.addEventListener('click', () => {
                if (file.isDir) {
                    state.currentFolder = file.path;
                    elements.folderSelect.value = file.path;
                    state.searchQuery = '';
                    elements.searchInput.value = '';
                    loadFiles(file.path);
                } else if (isMedia) {
                    openLightbox(file.path);
                }
            });
            return div;
        }

        function appendGallery(files, append) {
            if (!append && files.length === 0) {
                elements.gallery.innerHTML = '<div class="empty-state"><div class="icon">📂</div><p>Cartella vuota</p></div>';
                return;
            }

            const fragment = document.createDocumentFragment();
            files.forEach(file => {
                fragment.appendChild(createGalleryItem(file));
            });
            
            elements.gallery.appendChild(fragment);
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
            } else if (media.isAudio) {
                const audio = document.createElement('audio');
                audio.src = mediaUrl;
                audio.controls = true;
                audio.autoplay = true;
                audio.style.width = '100%';
                audio.style.maxWidth = '600px';
                elements.lightboxMedia.appendChild(audio);
                elements.videoHint.textContent = 'Spazio per play/pause • ← → per navigare • ESC per chiudere';
            } else {
                const img = document.createElement('img');
                img.src = mediaUrl;
                img.alt = media.name;
                elements.lightboxMedia.appendChild(img);
                elements.videoHint.textContent = '';
            }
            elements.lightboxFilename.textContent = media.name;
            elements.lightboxFilepath.textContent = '📂 ' + media.path;
            let mediaIcon = media.isVideo ? ' 🎬' : (media.isAudio ? ' 🎵' : ' 🖼️');
            elements.lightboxCounter.textContent = (index + 1) + ' / ' + state.media.length + mediaIcon;
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
                addBreadcrumbListeners();
                return;
            }
            const parts = folder.split('/').filter(p => p);
            let html = '<a href="#" data-folder=".">🏠 Home</a>';
            let currentPath = '';
            parts.forEach(part => {
                currentPath += (currentPath ? '/' : '') + part;
                html += '<span class="sep">›</span><a href="#" data-folder="' + encodePath(currentPath) + '">' + escapeHtml(part) + '</a>';
            });
            
            // Add "Parent" button to go up one level
            const lastSlash = folder.lastIndexOf('/');
            const parentFolder = lastSlash > 0 ? folder.substring(0, lastSlash) : '.';
            html += '<span class="sep">|</span><a href="#" data-folder="' + encodePath(parentFolder) + '" class="parent-nav" title="Cartella superiore">⬆️ Parent</a>';
            
            elements.breadcrumb.innerHTML = html;
            addBreadcrumbListeners();
        }

        function addBreadcrumbListeners() {
            document.querySelectorAll('.breadcrumb a').forEach(link => {
                link.addEventListener('click', (e) => {
                    e.preventDefault();
                    const folder = decodePath(link.dataset.folder);
                    state.currentFolder = folder;
                    elements.folderSelect.value = folder;
                    state.searchQuery = '';
                    elements.searchInput.value = '';
                    loadFiles(folder);
                });
            });
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function decodePath(encodedPath) {
            return encodedPath.split('/').map(decodeURIComponent).join('/');
        }

        function updateStats(data) {
            const showing = state.files.length;
            const total = data.total;
            const page = data.page;
            const totalPages = data.totalPages;
            
            let text = showing + ' / ' + total + ' elementi';
            if (totalPages > 1) {
                text += ' (pagina ' + page + ' di ' + totalPages + ')';
            }
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
