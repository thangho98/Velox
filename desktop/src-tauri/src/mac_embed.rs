// Embed an NSOpenGLView underneath the Tauri WKWebView so libmpv can render
// directly into the same window the webview lives in.
//
// View hierarchy after install:
//
//   NSWindow.contentView (container NSView, was WKWebView before)
//     ├── MpvOpenGLView   (added first → drawn at the bottom)
//     └── WKWebView       (added second → drawn on top, transparent background)
//
// Both children get NSViewWidthSizable | NSViewHeightSizable so they fill the
// container as the window resizes.

#![cfg(target_os = "macos")]

use cocoa::base::{id, nil, BOOL, NO, YES};
use cocoa::foundation::{NSRect, NSString};
use objc::{class, msg_send, sel, sel_impl};
use tauri::WebviewWindow;

use crate::gl_view;

const NS_VIEW_AUTORESIZE_ALL: u64 = 2 | 16; // WidthSizable | HeightSizable

pub fn install_video_layer(
    window: &WebviewWindow,
    mpv: *mut libmpv2_sys::mpv_handle,
) -> Result<(), String> {
    eprintln!("[mac_embed] step 1: getting ns_window");
    let ns_window_ptr = window.ns_window().map_err(|e| e.to_string())?;
    if ns_window_ptr.is_null() {
        return Err("ns_window is null".into());
    }
    let ns_window = ns_window_ptr as id;

    unsafe {
        eprintln!("[mac_embed] step 2: get contentView");
        let webview: id = msg_send![ns_window, contentView];
        if webview == nil {
            return Err("ns_window has no contentView".into());
        }
        let frame: NSRect = msg_send![webview, frame];
        eprintln!(
            "[mac_embed] webview frame: {}x{}",
            frame.size.width, frame.size.height
        );

        eprintln!("[mac_embed] step 3: make WKWebView non-opaque");
        // setValue:@NO forKey:@"drawsBackground" is KVC-rejected on modern WKWebView
        // (raises NSUndefinedKeyException). Rely on Tauri's macos-private-api
        // (NSWindow.opaque=NO + clearColor) + transparent body CSS instead, plus
        // setOpaque:NO on the webview itself.
        let _: () = msg_send![webview, setOpaque: NO as BOOL];
        let _: () = msg_send![webview, setWantsLayer: YES];

        eprintln!("[mac_embed] step 4: build container view");
        let container: id = msg_send![class!(NSView), alloc];
        let container: id = msg_send![container, initWithFrame: frame];
        let _: () = msg_send![container, setWantsLayer: YES];
        let _: () = msg_send![container, setAutoresizingMask: NS_VIEW_AUTORESIZE_ALL];

        eprintln!("[mac_embed] step 5: reparent webview into container");
        let _: id = msg_send![webview, retain];
        let _: () = msg_send![webview, removeFromSuperview];
        let _: () = msg_send![ns_window, setContentView: container];

        eprintln!("[mac_embed] step 6: create MpvOpenGLView");
        let gl_v = gl_view::create_view(frame);
        let _: () = msg_send![gl_v, setAutoresizingMask: NS_VIEW_AUTORESIZE_ALL];
        let _: () = msg_send![container, addSubview: gl_v];

        eprintln!("[mac_embed] step 7: re-add webview on top");
        let _: () = msg_send![webview, setFrame: frame];
        let _: () = msg_send![webview, setAutoresizingMask: NS_VIEW_AUTORESIZE_ALL];
        let _: () = msg_send![container, addSubview: webview];
        let _: () = msg_send![webview, release];

        eprintln!("[mac_embed] step 8: realize openGL context");
        let context: id = msg_send![gl_v, openGLContext];
        if context == nil {
            return Err("openGLContext nil after install".into());
        }
        let _: () = msg_send![context, makeCurrentContext];

        eprintln!("[mac_embed] step 9: install mpv render context");
        gl_view::install_render_context(gl_v, mpv)?;
        eprintln!("[mac_embed] DONE");
    }

    Ok(())
}
