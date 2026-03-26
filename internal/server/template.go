package server

const HTMLTemplate = `<!DOCTYPE html>
<html lang="it">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🖼️ Galleria v{{.Version}}</title>
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
        .version-badge {
            font-size: 0.55em;
            background: rgba(0, 212, 255, 0.15);
            border: 1px solid rgba(0, 212, 255, 0.4);
            color: #00d4ff;
            padding: 0.2rem 0.5rem;
            border-radius: 12px;
            vertical-align: middle;
            margin-left: 0.3rem;
            font-weight: 600;
            letter-spacing: 0.5px;
            -webkit-text-fill-color: #00d4ff;
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
            width: 100%;
            height: 100%;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .lightbox-media-wrapper {
            position: relative;
            max-width: 95vw;
            max-height: 95vh;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .lightbox img, .lightbox video {
            max-width: 95vw; max-height: 95vh;
            object-fit: contain;
            border-radius: 8px;
            box-shadow: 0 30px 60px rgba(0,0,0,0.5);
        }
        .lightbox video { background: #000; }
        .lightbox audio {
            width: 90vw;
            max-width: 600px;
            min-height: 50px;
            background: linear-gradient(135deg, #2a2a4a 0%, #1a1a2e 100%);
            border-radius: 8px;
            box-shadow: 0 30px 60px rgba(0,0,0,0.5);
            display: block;
        }
        .lightbox-close {
            position: fixed;
            top: 20px;
            right: 20px;
            background: rgba(0,0,0,0.7);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.1);
            color: #fff;
            font-size: 1.5rem;
            cursor: pointer;
            opacity: 0.8;
            transition: all 0.2s;
            z-index: 2100;
            width: 50px;
            height: 50px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .lightbox-close:hover { opacity: 1; background: rgba(255,50,50,0.8); }
        .lightbox-nav {
            position: fixed;
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
            z-index: 2100;
        }
        .lightbox-nav:hover {
            background: rgba(255,255,255,0.2);
            transform: translateY(-50%) scale(1.1);
        }
        .lightbox-prev { left: 20px; }
        .lightbox-next { right: 20px; }
        .lightbox-overlay {
            position: fixed;
            bottom: 20px;
            left: 50%;
            transform: translateX(-50%);
            display: flex;
            flex-direction: column;
            gap: 0.75rem;
            z-index: 2100;
            max-width: 95vw;
            width: auto;
            transition: all 0.3s ease;
        }
        .lightbox-overlay.collapsed .lightbox-overlay-top,
        .lightbox-overlay.collapsed .lightbox-overlay-bottom {
            opacity: 0;
            max-height: 0;
            padding: 0;
            margin: 0;
            overflow: hidden;
            pointer-events: none;
        }
        .lightbox-overlay.expanded .lightbox-overlay-top,
        .lightbox-overlay.expanded .lightbox-overlay-bottom {
            opacity: 1;
            max-height: 500px;
            transition: opacity 0.3s ease, max-height 0.3s ease, padding 0.3s ease;
        }
        .lightbox-overlay-top {
            background: rgba(0, 0, 0, 0.75);
            backdrop-filter: blur(12px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            padding: 0.75rem 1.25rem;
            text-align: center;
            transition: opacity 0.3s ease, max-height 0.3s ease, padding 0.3s ease;
        }
        .lightbox-overlay-bottom {
            background: rgba(0, 0, 0, 0.75);
            backdrop-filter: blur(12px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            padding: 0.75rem 1.25rem;
            display: flex;
            flex-wrap: wrap;
            justify-content: center;
            align-items: center;
            gap: 0.75rem;
            transition: opacity 0.3s ease, max-height 0.3s ease, padding 0.3s ease;
        }
        .lightbox-overlay-toggle {
            background: rgba(0, 0, 0, 0.75);
            backdrop-filter: blur(12px);
            border: 1px solid rgba(255, 255, 255, 0.2);
            border-radius: 24px;
            padding: 0.6rem 1.25rem;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            color: #fff;
            font-size: 0.9rem;
            font-weight: 500;
            transition: all 0.2s ease;
            align-self: center;
            min-height: 44px;
            min-width: 100px;
            user-select: none;
            -webkit-tap-highlight-color: transparent;
        }
        .lightbox-overlay-toggle:hover {
            background: rgba(0, 212, 255, 0.35);
            border-color: rgba(0, 212, 255, 0.6);
            transform: translateY(-2px);
        }
        .lightbox-overlay-toggle:active {
            transform: translateY(0);
        }
        .lightbox-overlay-toggle .chevron {
            transition: transform 0.3s ease;
            font-size: 0.8rem;
        }
        .lightbox-overlay.collapsed .lightbox-overlay-toggle .chevron {
            transform: rotate(0deg);
        }
        .lightbox-overlay.expanded .lightbox-overlay-toggle .chevron {
            transform: rotate(180deg);
        }
        .lightbox-overlay .filename {
            font-size: 1rem;
            color: #fff;
            font-weight: 500;
            margin-bottom: 0.25rem;
            display: flex;
            align-items: center;
            justify-content: center;
            flex-wrap: wrap;
            gap: 0.5rem;
        }
        .lightbox-overlay .filepath {
            font-size: 0.8rem;
            color: #aaa;
            word-break: break-all;
            margin-bottom: 0.25rem;
        }
        .lightbox-overlay .counter {
            font-size: 0.85rem;
            color: #888;
        }
        .lightbox-overlay .video-hint {
            font-size: 0.75rem;
            color: #888;
            margin-top: 0.25rem;
        }
        .rotate-controls {
            display: flex;
            gap: 0.5rem;
        }
        .rotate-btn {
            background: rgba(255,255,255,0.12);
            border: 1px solid rgba(255,255,255,0.2);
            color: #fff;
            padding: 0.5rem 1rem;
            border-radius: 8px;
            cursor: pointer;
            font-size: 0.9rem;
            transition: all 0.2s;
            display: flex;
            align-items: center;
            gap: 0.3rem;
            white-space: nowrap;
        }
        .rotate-btn:hover {
            background: rgba(0,212,255,0.35);
            border-color: rgba(0,212,255,0.6);
            transform: translateY(-1px);
        }
        .rotate-btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none;
        }
        .rotate-btn.spinning {
            animation: pulse 1s infinite;
        }
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5;
        }
        }
        .rename-btn {
            background: rgba(255,255,255,0.12);
            border: 1px solid rgba(255,255,255,0.2);
            color: #fff;
            padding: 0.35rem 0.75rem;
            border-radius: 8px;
            cursor: pointer;
            font-size: 0.8rem;
            transition: all 0.2s;
            display: inline-flex;
            align-items: center;
            gap: 0.3rem;
            white-space: nowrap;
        }
        .rename-btn:hover {
            background: rgba(0,212,255,0.35);
            border-color: rgba(0,212,255,0.6);
            transform: translateY(-1px);
        }
        .rename-modal {
            display: none;
            position: fixed;
            inset: 0;
            background: rgba(0,0,0,0.8);
            z-index: 3000;
            justify-content: center;
            align-items: center;
        }
        .rename-modal.active { display: flex; }
        .rename-modal-content {
            background: linear-gradient(145deg, #1a1a2e, #16213e);
            border: 1px solid rgba(255,255,255,0.1);
            border-radius: 16px;
            padding: 2rem;
            min-width: 400px;
            max-width: 90vw;
        }
        .rename-modal h3 {
            margin-bottom: 1rem;
            color: #fff;
            font-size: 1.2rem;
        }
        .rename-modal .current-name {
            color: #888;
            font-size: 0.85rem;
            margin-bottom: 1rem;
            word-break: break-all;
        }
        .rename-modal input {
            width: 100%;
            background: rgba(255,255,255,0.1);
            border: 1px solid rgba(255,255,255,0.2);
            border-radius: 8px;
            padding: 0.8rem 1rem;
            color: #fff;
            font-size: 1rem;
            margin-bottom: 0.5rem;
        }
        .rename-modal input:focus {
            outline: none;
            border-color: #00d4ff;
            background: rgba(255,255,255,0.15);
        }
        .rename-modal input.error {
            border-color: #ff6b6b;
        }
        .rename-modal .error-msg {
            color: #ff6b6b;
            font-size: 0.85rem;
            margin-bottom: 1rem;
            min-height: 1.2rem;
        }
        .rename-modal .btn-group {
            display: flex;
            gap: 0.5rem;
            justify-content: flex-end;
        }
        .rename-modal button {
            padding: 0.7rem 1.5rem;
            border-radius: 8px;
            border: none;
            cursor: pointer;
            font-size: 0.95rem;
            transition: all 0.2s;
        }
        .rename-modal .btn-cancel {
            background: rgba(255,255,255,0.1);
            color: #fff;
        }
        .rename-modal .btn-cancel:hover {
            background: rgba(255,255,255,0.2);
        }
        .rename-modal .btn-confirm {
            background: linear-gradient(90deg, #00d4ff, #7b2cbf);
            color: #fff;
        }
        .rename-modal .btn-confirm:hover {
            opacity: 0.9;
            transform: translateY(-1px);
        }
        .rename-modal .btn-confirm:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none;
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
            .lightbox-overlay {
                bottom: 10px;
                width: calc(100% - 20px);
                max-width: none;
            }
            .lightbox-overlay-top,
            .lightbox-overlay-bottom {
                padding: 0.6rem 1rem;
            }
            .lightbox-overlay .filename {
                font-size: 0.9rem;
            }
            .lightbox-overlay .filepath {
                font-size: 0.75rem;
            }
            .lightbox-close {
                top: 10px;
                right: 10px;
                width: 44px;
                height: 44px;
                font-size: 1.3rem;
            }
            .rotate-btn {
                padding: 0.5rem 0.75rem;
                font-size: 0.85rem;
                min-height: 44px;
            }
            .rename-btn {
                padding: 0.4rem 0.6rem;
                font-size: 0.75rem;
                min-height: 36px;
            }
            .lightbox-overlay-toggle {
                min-height: 48px;
                min-width: 110px;
                padding: 0.7rem 1.5rem;
                font-size: 0.95rem;
            }
            .lightbox-overlay-toggle .chevron {
                font-size: 0.85rem;
            }
        }
        @media (max-height: 500px) {
            .lightbox-overlay {
                bottom: 5px;
                flex-direction: row;
                flex-wrap: wrap;
                justify-content: center;
                gap: 0.5rem;
            }
            .lightbox-overlay-top,
            .lightbox-overlay-bottom {
                padding: 0.5rem 0.75rem;
            }
            .rotate-btn {
                padding: 0.4rem 0.6rem;
                font-size: 0.8rem;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🖼️ Galleria <span class="version-badge">{{.Version}}</span></h1>
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
            <div class="lightbox-media-wrapper" id="lightboxMedia"></div>
            <button class="lightbox-nav lightbox-next" id="lightboxNext">›</button>
        </div>
        <div class="lightbox-overlay collapsed" id="lightboxOverlay">
            <button class="lightbox-overlay-toggle" id="lightboxOverlayToggle" title="Mostra/Nascondi info (I)">
                <span>ℹ️ Info</span>
                <span class="chevron">▼</span>
            </button>
            <div class="lightbox-overlay-top">
                <div class="filename">
                    <span id="lightboxFilename"></span>
                    <button class="rename-btn" id="renameBtn" title="Rinomina (F2)">✏️ Rinomina</button>
                </div>
                <div class="filepath" id="lightboxFilepath"></div>
                <div class="counter" id="lightboxCounter"></div>
                <div class="video-hint" id="videoHint"></div>
            </div>
            <div class="lightbox-overlay-bottom" id="rotateControls" style="display: none;">
                <div class="rotate-controls">
                    <button class="rotate-btn" id="rotateCCW" title="Ruota -90° (Shift+R)">↺ -90°</button>
                    <button class="rotate-btn" id="rotateCW" title="Ruota +90° (R)">↻ +90°</button>
                </div>
            </div>
        </div>
    </div>
    <div class="rename-modal" id="renameModal">
        <div class="rename-modal-content">
            <h3>✏️ Rinomina file</h3>
            <div class="current-name" id="renameCurrentName"></div>
            <input type="text" id="renameInput" placeholder="Nuovo nome..." autocomplete="off">
            <div class="error-msg" id="renameError"></div>
            <div class="btn-group">
                <button class="btn-cancel" id="renameCancel">Annulla</button>
                <button class="btn-confirm" id="renameConfirm">Conferma</button>
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
            totalFiles: 0,
            overlayExpanded: false
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
            lightboxOverlay: document.getElementById('lightboxOverlay'),
            lightboxMedia: document.getElementById('lightboxMedia'),
            lightboxFilename: document.getElementById('lightboxFilename'),
            lightboxFilepath: document.getElementById('lightboxFilepath'),
            lightboxCounter: document.getElementById('lightboxCounter'),
            videoHint: document.getElementById('videoHint'),
            lightboxClose: document.getElementById('lightboxClose'),
            lightboxPrev: document.getElementById('lightboxPrev'),
            lightboxNext: document.getElementById('lightboxNext'),
            rotateControls: document.getElementById('rotateControls'),
            rotateCW: document.getElementById('rotateCW'),
            rotateCCW: document.getElementById('rotateCCW'),
            renameBtn: document.getElementById('renameBtn'),
            renameModal: document.getElementById('renameModal'),
            renameInput: document.getElementById('renameInput'),
            renameCurrentName: document.getElementById('renameCurrentName'),
            renameError: document.getElementById('renameError'),
            renameCancel: document.getElementById('renameCancel'),
            renameConfirm: document.getElementById('renameConfirm'),
            lightboxOverlayToggle: document.getElementById('lightboxOverlayToggle')
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

        // Disable SSE streaming search to ensure proper fzf relevance sorting
        // The streaming endpoint cannot guarantee global sort order
        const SSE_SUPPORTED = false;

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
                const thumbUrl = '/thumb/' + encodePath(file.path) + '?v=' + file.modTime;
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
            // Reset overlay to collapsed state by default
            setOverlayExpanded(false);
        }

        function showMedia(index) {
            const media = state.media[index];
            // Always use fresh timestamp to bypass cache when opening lightbox
            const cacheBuster = '?t=' + Date.now();
            const mediaUrl = '/raw/' + encodePath(media.path) + cacheBuster;
            elements.lightboxMedia.innerHTML = '';
            elements.rotateControls.style.display = 'none';
            if (media.isVideo) {
                const video = document.createElement('video');
                video.src = mediaUrl;
                video.controls = true;
                video.autoplay = true;
                elements.lightboxMedia.appendChild(video);
                elements.videoHint.textContent = 'Spazio per play/pause • ← → per navigare • ESC per chiudere • F2 per rinominare';
            } else if (media.isAudio) {
                const audio = document.createElement('audio');
                audio.src = mediaUrl;
                audio.controls = true;
                audio.autoplay = true;
                elements.lightboxMedia.appendChild(audio);
                elements.videoHint.textContent = 'Spazio per play/pause • ← → per navigare • ESC per chiudere • F2 per rinominare';
            } else {
                const img = document.createElement('img');
                img.src = mediaUrl;
                img.alt = media.name;
                img.id = 'lightboxImg';
                elements.lightboxMedia.appendChild(img);
                elements.videoHint.textContent = 'R per ruotare +90° • Shift+R per -90° • F2 per rinominare';
                elements.rotateControls.style.display = 'flex';
            }
            elements.lightboxFilename.textContent = media.name;
            elements.lightboxFilepath.textContent = '📂 ' + media.path;
            let mediaIcon = media.isVideo ? ' 🎬' : (media.isAudio ? ' 🎵' : ' 🖼️');
            elements.lightboxCounter.textContent = (index + 1) + ' / ' + state.media.length + mediaIcon;
        }

        async function rotateImage(angle) {
            const media = state.media[state.currentMediaIndex];
            if (!media || !media.isImage) return;

            const btn = angle > 0 ? elements.rotateCW : elements.rotateCCW;
            const originalText = btn.textContent;
            btn.disabled = true;
            btn.classList.add('spinning');
            btn.textContent = '⏳ ...';

            try {
                const response = await fetch('/api/rotate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ path: media.path, angle: angle })
                });

                const result = await response.json();

                if (result.success) {
                    // Add cache-buster to reload image in fullscreen
                    const cacheBuster = '?t=' + Date.now();
                    const img = document.getElementById('lightboxImg');
                    if (img) {
                        img.src = '/raw/' + encodePath(media.path) + cacheBuster;
                    }
                    // Also update thumbnail in gallery
                    // Use iteration instead of CSS selector to handle special chars in path
                    const galleryItems = document.querySelectorAll('.gallery-item');
                    for (const item of galleryItems) {
                        if (item.dataset.path === media.path) {
                            const thumbImg = item.querySelector('img');
                            if (thumbImg) {
                                thumbImg.src = '/thumb/' + encodePath(media.path) + '?t=' + Date.now();
                            }
                            break;
                        }
                    }
                } else {
                    alert('Errore rotazione: ' + (result.error || 'Sconosciuto'));
                }
            } catch (err) {
                console.error('Errore rotazione:', err);
                alert('Errore durante la rotazione');
            } finally {
                btn.disabled = false;
                btn.classList.remove('spinning');
                btn.textContent = originalText;
            }
        }

        // Rename functionality
        const invalidChars = /[\\/:*?"<>|]/;

        function validateNewName(name) {
            if (!name || name.trim() === '') {
                return 'Il nome non può essere vuoto';
            }
            if (name.length > 255) {
                return 'Il nome è troppo lungo (max 255 caratteri)';
            }
            if (invalidChars.test(name)) {
                return 'Il nome contiene caratteri non validi: \\ / : * ? " < > |';
            }
            if (name === '.' || name === '..') {
                return 'Nome non valido';
            }
            return null;
        }

        function openRenameModal() {
            const media = state.media[state.currentMediaIndex];
            if (!media) return;

            elements.renameCurrentName.textContent = 'Attuale: ' + media.name;
            elements.renameInput.value = media.name;
            elements.renameError.textContent = '';
            elements.renameInput.classList.remove('error');
            elements.renameConfirm.disabled = false;
            elements.renameModal.classList.add('active');
            elements.renameInput.focus();
            elements.renameInput.select();
        }

        function closeRenameModal() {
            elements.renameModal.classList.remove('active');
            elements.renameError.textContent = '';
            elements.renameInput.classList.remove('error');
        }

        async function performRename() {
            const media = state.media[state.currentMediaIndex];
            if (!media) return;

            const newName = elements.renameInput.value.trim();
            const error = validateNewName(newName);

            if (error) {
                elements.renameError.textContent = error;
                elements.renameInput.classList.add('error');
                return;
            }

            // Check if name is unchanged
            if (newName === media.name) {
                closeRenameModal();
                return;
            }

            elements.renameConfirm.disabled = true;
            elements.renameError.textContent = '';
            elements.renameInput.classList.remove('error');

            try {
                const response = await fetch('/api/rename', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ path: media.path, newName: newName })
                });

                const result = await response.json();

                if (result.success) {
                    // Update state with new path
                    updateFilePath(result.oldPath, result.newPath, newName);
                    closeRenameModal();
                } else {
                    elements.renameError.textContent = result.error || 'Errore durante la rinomina';
                    elements.renameInput.classList.add('error');
                    elements.renameConfirm.disabled = false;
                }
            } catch (err) {
                console.error('Errore rinomina:', err);
                elements.renameError.textContent = 'Errore di rete durante la rinomina';
                elements.renameInput.classList.add('error');
                elements.renameConfirm.disabled = false;
            }
        }

        function updateFilePath(oldPath, newPath, newName) {
            // Update state.files
            const fileIndex = state.files.findIndex(f => f.path === oldPath);
            if (fileIndex !== -1) {
                state.files[fileIndex].path = newPath;
                state.files[fileIndex].name = newName;
            }

            // Update state.media
            const mediaIndex = state.media.findIndex(m => m.path === oldPath);
            if (mediaIndex !== -1) {
                state.media[mediaIndex].path = newPath;
                state.media[mediaIndex].name = newName;
                // Update current index if this is the current file
                if (mediaIndex === state.currentMediaIndex) {
                    state.currentMediaIndex = mediaIndex;
                }
            }

            // Update renderedFiles Set
            renderedFiles.delete(oldPath);
            renderedFiles.add(newPath);

            // Update lightbox display
            elements.lightboxFilename.textContent = newName;
            elements.lightboxFilepath.textContent = '📂 ' + newPath;

            // Update gallery DOM
            const galleryItem = document.querySelector('.gallery-item[data-path="' + oldPath + '"]');
            if (galleryItem) {
                galleryItem.dataset.path = newPath;
                const nameEl = galleryItem.querySelector('.item-name');
                if (nameEl) {
                    nameEl.textContent = newName;
                    nameEl.title = newName;
                }
                // Update thumbnail src
                const thumbImg = galleryItem.querySelector('img');
                if (thumbImg) {
                    thumbImg.src = '/thumb/' + encodePath(newPath);
                    thumbImg.alt = newName;
                }
            }
        }

        function closeLightbox() {
            const video = elements.lightboxMedia.querySelector('video');
            if (video) video.pause();
            elements.lightbox.classList.remove('active');
            document.body.style.overflow = '';
        }

        function toggleOverlay() {
            state.overlayExpanded = !state.overlayExpanded;
            if (state.overlayExpanded) {
                elements.lightboxOverlay.classList.remove('collapsed');
                elements.lightboxOverlay.classList.add('expanded');
            } else {
                elements.lightboxOverlay.classList.remove('expanded');
                elements.lightboxOverlay.classList.add('collapsed');
            }
        }

        function setOverlayExpanded(expanded) {
            state.overlayExpanded = expanded;
            if (expanded) {
                elements.lightboxOverlay.classList.remove('collapsed');
                elements.lightboxOverlay.classList.add('expanded');
            } else {
                elements.lightboxOverlay.classList.remove('expanded');
                elements.lightboxOverlay.classList.add('collapsed');
            }
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

        // Debounced search input handler
        let searchTimeout;
        elements.searchInput.addEventListener('input', (e) => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                state.searchQuery = e.target.value;
                if (state.searchQuery) {
                    loadFiles(state.currentFolder, state.searchQuery);
                } else {
                    loadFiles(state.currentFolder);
                }
            }, 300);
        });

        elements.lightboxClose.addEventListener('click', closeLightbox);
        elements.lightboxNext.addEventListener('click', (e) => { e.stopPropagation(); nextMedia(); });
        elements.lightboxPrev.addEventListener('click', (e) => { e.stopPropagation(); prevMedia(); });

        // Prevent lightbox content area clicks from closing
        document.querySelector('.lightbox-content').addEventListener('click', (e) => {
            e.stopPropagation();
        });
        elements.rotateCW.addEventListener('click', (e) => { e.stopPropagation(); rotateImage(90); });
        elements.rotateCCW.addEventListener('click', (e) => { e.stopPropagation(); rotateImage(-90); });
        elements.renameBtn.addEventListener('click', (e) => { e.stopPropagation(); openRenameModal(); });
        elements.renameCancel.addEventListener('click', closeRenameModal);
        elements.renameConfirm.addEventListener('click', performRename);
        elements.renameModal.addEventListener('click', (e) => {
            if (e.target === elements.renameModal) closeRenameModal();
        });
        elements.renameInput.addEventListener('input', () => {
            elements.renameError.textContent = '';
            elements.renameInput.classList.remove('error');
        });
        elements.renameInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                performRename();
            } else if (e.key === 'Escape') {
                closeRenameModal();
            }
        });
        elements.lightbox.addEventListener('click', (e) => {
            // Only close if clicking on the lightbox background itself, not on any interactive elements
            if (e.target === elements.lightbox) {
                closeLightbox();
            }
        });

        // Prevent overlay clicks from closing the lightbox
        elements.lightboxOverlay.addEventListener('click', (e) => {
            e.stopPropagation();
        });
        // Toggle overlay on button click
        elements.lightboxOverlayToggle.addEventListener('click', (e) => {
            e.stopPropagation();
            toggleOverlay();
        });
        document.addEventListener('keydown', (e) => {
            // Handle F2 for rename - works both in lightbox and gallery
            if (e.key === 'F2') {
                e.preventDefault();
                if (elements.lightbox.classList.contains('active')) {
                    openRenameModal();
                } else if (state.media.length > 0 && state.currentMediaIndex < state.media.length) {
                    // If no lightbox open, open it first with current media
                    const media = state.media[state.currentMediaIndex];
                    if (media) {
                        openLightbox(media.path);
                        setTimeout(openRenameModal, 100);
                    }
                }
                return;
            }

            if (!elements.lightbox.classList.contains('active')) return;

            // Don't process lightbox shortcuts if rename modal is open
            if (elements.renameModal.classList.contains('active')) return;

            const video = elements.lightboxMedia.querySelector('video');
            const img = elements.lightboxMedia.querySelector('img');
            if (e.key === 'Escape') closeLightbox();
            else if (e.key === 'ArrowRight') nextMedia();
            else if (e.key === 'ArrowLeft') prevMedia();
            else if (e.key === ' ' && video) {
                e.preventDefault();
                video.paused ? video.play() : video.pause();
            } else if (e.key === 'r' || e.key === 'R') {
                if (img) {
                    e.preventDefault();
                    if (e.shiftKey) {
                        rotateImage(-90);
                    } else {
                        rotateImage(90);
                    }
                }
            } else if (e.key === 'i' || e.key === 'I') {
                e.preventDefault();
                toggleOverlay();
            }
        });

        loadFolders();
        loadFiles();
    </script>
</body>
</html>`
