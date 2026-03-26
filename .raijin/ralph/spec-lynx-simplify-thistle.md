# Goal
Fix three UI/UX bugs in the GoGalleryFzf media gallery: MP3 player controls visibility, video/audio streaming stability, and persistent image rotation.

# User Specification

## Bug 1: MP3 Player Missing Controls
The audio player for MP3 files in the lightbox does not display controls, even though music starts playing. Users cannot pause, seek, or adjust volume.

**Expected behavior:** Audio files should display a visible control bar with play/pause, seek, volume controls, and duration display.

**Current behavior:** Audio plays but no controls are visible in the lightbox UI.

**Root cause:** The CSS in `internal/server/template.go` defines styles for `.lightbox img` and `.lightbox video` but has no styles for the `audio` element. The audio element is created with `controls=true` but lacks proper styling to be visible.

## Bug 2: Video/Audio Playback Stops Prematurely
Video and audio files start playing but stop after some time instead of continuing to the end.

**Expected behavior:** Media should play continuously from start to finish without interruption.

**Current behavior:** Playback stops after a period of time, requiring manual restart.

**Root cause hypothesis:** The HTTP range request handling in `internal/server/server.go` function `serveFile` may have issues with:
- Connection keep-alive handling
- Proper content-range header formatting for consecutive range requests
- File handle management during streaming

The server handles range requests but the client may be receiving incomplete responses or connection resets.

## Bug 3: Image Rotation Not Persistent
After rotating an image and closing/reopening the gallery (or refreshing), the rotation is lost or appears incorrect.

**Expected behavior:** Rotated images should retain their rotation permanently on disk.

**Current behavior:** Images appear to lose their rotation or display with incorrect orientation after reopening.

**Root cause:** The `RotateImage` function in `internal/imaging/rotate.go` rotates the pixel data but does not handle EXIF orientation metadata. Many images from cameras/phones contain EXIF orientation tags that tell browsers/viewers how to display the image. When pixels are rotated but the EXIF orientation tag is not reset to "normal" (value 1), browsers apply the EXIF rotation on top of the pixel rotation, causing the image to appear incorrectly oriented.

# Plan

## Fix 1: Audio Player Controls
- [ ] Add CSS styling for `audio` element in the lightbox template
- [ ] Ensure audio controls are visible and properly sized
- [ ] Apply similar styling pattern as video controls (width, max-width, centering)

**Files to modify:**
- `internal/server/template.go` - Add audio element CSS in the `<style>` section

## Fix 2: Video/Audio Streaming Stability
- [ ] Review and fix HTTP range request handling in `serveFile`
- [ ] Ensure proper `Connection: keep-alive` header for media streaming
- [ ] Verify Content-Range header format compliance with HTTP spec
- [ ] Check file seek and read operations for large media files
- [ ] Add proper error handling for partial content requests

**Files to modify:**
- `internal/server/server.go` - Review `serveFile`, `parseRange`, and range request handling

## Fix 3: Persistent Image Rotation
- [ ] Integrate EXIF handling library (e.g., `github.com/rwcarlsen/goexif` or `github.com/barasher/go-exiftool`)
- [ ] After pixel rotation, reset EXIF orientation tag to value 1 (normal)
- [ ] Alternatively, strip all EXIF data after rotation to prevent browser auto-rotation
- [ ] Ensure rotated images display consistently across browsers

**Files to modify:**
- `internal/imaging/rotate.go` - Add EXIF orientation handling after pixel rotation
- `go.mod`/`go.sum` - Add EXIF library dependency if needed

## Constraints
- Maintain backward compatibility with existing image formats (JPEG, PNG, GIF, WebP)
- Do not break existing thumbnail generation
- Keep audio/video player keyboard shortcuts functional (space for play/pause, arrows for navigation)
- Ensure fixes work on both desktop and mobile browsers
