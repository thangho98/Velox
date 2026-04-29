fn main() {
    let libdir = std::env::var("MPV_LIB_DIR")
        .unwrap_or_else(|_| "/Users/thawng/Desktop/source/mpv-poc/libmpv/lib".into());
    println!("cargo:rustc-link-search=native={}", libdir);
    println!("cargo:rustc-link-lib=dylib=mpv");
    println!("cargo:rustc-link-arg=-Wl,-rpath,{}", libdir);
    println!("cargo:rerun-if-env-changed=MPV_LIB_DIR");

    tauri_build::build()
}
