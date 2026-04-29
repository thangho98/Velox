// macOS NSOpenGLView subclass that hosts a libmpv RenderContext.
//
// Strategy
// - mpv `vo=libmpv` produces frames that the host (us) draws into our own GL surface.
// - We declare an Obj-C subclass `MpvOpenGLView : NSOpenGLView` at runtime.
// - On `prepareOpenGL` we install the GL swap-interval and lazily create the
//   `mpv_render_context` on the GL thread.
// - On `drawRect:` we call `mpv_render_context_render` against the default FBO (0)
//   using backing-pixel dimensions (Retina).
// - mpv's update callback is dispatched to main thread which calls
//   `setNeedsDisplay:YES` to schedule a redraw.

#![cfg(target_os = "macos")]
#![allow(unexpected_cfgs)]

use cocoa::base::{id, nil, BOOL, NO, YES};
use cocoa::foundation::{NSRect, NSSize};
use objc::declare::ClassDecl;
use objc::runtime::{Class, Object, Sel};
use objc::{class, msg_send, sel, sel_impl};
use std::os::raw::{c_int, c_void};
use std::sync::Once;

use libmpv2::render::{OpenGLInitParams, RenderContext, RenderParam, RenderParamApiType};

const IVAR_RENDER_CTX: &str = "_renderCtx";

// NSOpenGLPixelFormatAttribute values (NSOpenGLPFA*)
const NSOPENGL_PFA_DOUBLE_BUFFER: u32 = 5;
const NSOPENGL_PFA_ACCELERATED: u32 = 73;
const NSOPENGL_PFA_COLOR_SIZE: u32 = 8;
const NSOPENGL_PFA_ALPHA_SIZE: u32 = 11;
const NSOPENGL_PFA_DEPTH_SIZE: u32 = 12;
const NSOPENGL_PFA_OPENGL_PROFILE: u32 = 99;

// NSOpenGLProfileVersion3_2Core
const NSOPENGL_PROFILE_VERSION_3_2_CORE: u32 = 0x3200;

// NSOpenGLContext parameter: NSOpenGLContextParameterSwapInterval
const NSOPENGL_CP_SWAP_INTERVAL: u32 = 222;

// dlsym RTLD_DEFAULT for GL function lookup.
const RTLD_DEFAULT: *mut c_void = -2isize as *mut c_void;
extern "C" {
    fn dlsym(handle: *mut c_void, symbol: *const i8) -> *mut c_void;
}

fn gl_get_proc(_: &(), name: &str) -> *mut c_void {
    let cstr = std::ffi::CString::new(name).unwrap();
    unsafe { dlsym(RTLD_DEFAULT, cstr.as_ptr()) }
}

extern "C" fn prepare_opengl(this: &mut Object, _: Sel) {
    unsafe {
        let _: () = msg_send![super(this, class!(NSOpenGLView)), prepareOpenGL];
        let context: id = msg_send![this, openGLContext];
        let _: () = msg_send![context, makeCurrentContext];
        let swap: c_int = 1;
        let _: () = msg_send![context, setValues: &swap as *const c_int forParameter: NSOPENGL_CP_SWAP_INTERVAL];
    }
}

extern "C" fn reshape(this: &mut Object, _: Sel) {
    unsafe {
        let _: () = msg_send![super(this, class!(NSOpenGLView)), reshape];
        let _: () = msg_send![this, setNeedsDisplay: YES];
    }
}

extern "C" fn draw_rect(this: &mut Object, _: Sel, _dirty: NSRect) {
    unsafe {
        let context: id = msg_send![this, openGLContext];
        let _: () = msg_send![context, makeCurrentContext];

        let ctx_ptr: *mut c_void = *this.get_ivar(IVAR_RENDER_CTX);
        if !ctx_ptr.is_null() {
            let render_ctx = &*(ctx_ptr as *const RenderContext);

            let bounds: NSRect = msg_send![this, bounds];
            let backing: NSSize = msg_send![this, convertSizeToBacking: bounds.size];
            let w = backing.width as i32;
            let h = backing.height as i32;

            let _ = render_ctx.render::<()>(0, w, h, true);
        }

        let _: () = msg_send![context, flushBuffer];
    }
}

extern "C" fn trigger_redraw(this: &mut Object, _: Sel) {
    unsafe {
        let _: () = msg_send![this, setNeedsDisplay: YES];
    }
}

fn ensure_class() -> &'static Class {
    static INIT: Once = Once::new();
    static mut CLASS: *const Class = std::ptr::null();

    unsafe {
        INIT.call_once(|| {
            let superclass = class!(NSOpenGLView);
            let mut decl = ClassDecl::new("MpvOpenGLView", superclass)
                .expect("failed to declare MpvOpenGLView");
            decl.add_ivar::<*mut c_void>(IVAR_RENDER_CTX);
            decl.add_method(
                sel!(prepareOpenGL),
                prepare_opengl as extern "C" fn(&mut Object, Sel),
            );
            decl.add_method(sel!(reshape), reshape as extern "C" fn(&mut Object, Sel));
            decl.add_method(
                sel!(drawRect:),
                draw_rect as extern "C" fn(&mut Object, Sel, NSRect),
            );
            decl.add_method(
                sel!(triggerRedraw),
                trigger_redraw as extern "C" fn(&mut Object, Sel),
            );
            CLASS = decl.register();
        });
        &*CLASS
    }
}

fn make_pixel_format() -> id {
    let attrs: [u32; 11] = [
        NSOPENGL_PFA_ACCELERATED,
        NSOPENGL_PFA_DOUBLE_BUFFER,
        NSOPENGL_PFA_COLOR_SIZE,
        24,
        NSOPENGL_PFA_ALPHA_SIZE,
        8,
        NSOPENGL_PFA_DEPTH_SIZE,
        24,
        NSOPENGL_PFA_OPENGL_PROFILE,
        NSOPENGL_PROFILE_VERSION_3_2_CORE,
        0,
    ];
    unsafe {
        let pf: id = msg_send![class!(NSOpenGLPixelFormat), alloc];
        let pf: id = msg_send![pf, initWithAttributes: attrs.as_ptr()];
        pf
    }
}

/// Create the MpvOpenGLView. ivar starts as nullptr — install render context after.
pub fn create_view(frame: NSRect) -> id {
    let class = ensure_class();
    unsafe {
        let pf = make_pixel_format();
        let view: id = msg_send![class, alloc];
        let view: id = msg_send![view, initWithFrame: frame pixelFormat: pf];
        // Pin to autoresize with parent (full-window).
        // NSViewWidthSizable | NSViewHeightSizable = 2 | 16 = 18
        let _: () = msg_send![view, setAutoresizingMask: 18u64];
        // Wantslayer ensures backing layer composes correctly with WKWebView above.
        let _: () = msg_send![view, setWantsLayer: YES];
        (*view).set_ivar::<*mut c_void>(IVAR_RENDER_CTX, std::ptr::null_mut());
        view
    }
}

/// Install render context on the view. Must be called from main thread after the view is in a window.
/// The returned pointer is owned by the view and freed when the view is detached (caller must call free_render_context).
pub unsafe fn install_render_context(
    view: id,
    mpv: *mut libmpv2_sys::mpv_handle,
) -> Result<(), String> {
    // Force prepareOpenGL by making context current.
    let context: id = msg_send![view, openGLContext];
    if context == nil {
        return Err("openGLContext is nil — view not yet in window?".into());
    }
    let _: () = msg_send![context, makeCurrentContext];

    let params: Vec<RenderParam<()>> = vec![
        RenderParam::ApiType(RenderParamApiType::OpenGl),
        RenderParam::InitParams(OpenGLInitParams {
            get_proc_address: gl_get_proc,
            ctx: (),
        }),
    ];

    let mut render_ctx = RenderContext::new(&mut *mpv, params).map_err(|e| format!("{:?}", e))?;

    // mpv calls update callback from a background thread when a new frame is ready.
    // We schedule a setNeedsDisplay on the main thread so AppKit picks up the redraw.
    let view_addr = view as usize;
    render_ctx.set_update_callback(move || {
        let v = view_addr as id;
        let _: () = msg_send![v, performSelectorOnMainThread: sel!(triggerRedraw)
                                  withObject: nil
                                  waitUntilDone: NO as BOOL];
    });

    let boxed = Box::into_raw(Box::new(render_ctx));
    (*view).set_ivar::<*mut c_void>(IVAR_RENDER_CTX, boxed as *mut c_void);
    Ok(())
}

pub unsafe fn free_render_context(view: id) {
    let ptr: *mut c_void = *(*view).get_ivar(IVAR_RENDER_CTX);
    if !ptr.is_null() {
        drop(Box::from_raw(ptr as *mut RenderContext));
        (*view).set_ivar::<*mut c_void>(IVAR_RENDER_CTX, std::ptr::null_mut());
    }
}
