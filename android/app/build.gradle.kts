import java.util.Properties

plugins {
    alias(libs.plugins.ksp)
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.kotlin.serialization)
    alias(libs.plugins.hilt.android)
    alias(libs.plugins.detekt)
    alias(libs.plugins.ktlint)
}

android {
    namespace = "com.velox.app"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.velox.app"
        minSdk = 26
        targetSdk = 35
        versionCode = 108
        versionName = "0.1.8"

        vectorDrawables {
            useSupportLibrary = true
        }

        // Read server URL from local.properties
        val localProperties = Properties()
        val localPropertiesFile = rootProject.file("../local.properties")
        if (localPropertiesFile.exists()) {
            localProperties.load(localPropertiesFile.inputStream())
        }
        val serverUrl = localProperties.getProperty("server.url") ?: "http://192.168.1.1:8098/api/"

        buildConfigField("String", "BASE_URL", "\"$serverUrl\"")
    }

    signingConfigs {
        create("release") {
            storeFile = file("../keystore/velox-release.keystore")
            storePassword = "velox2026"
            keyAlias = "velox"
            keyPassword = "velox2026"
        }
    }

    buildTypes {
        debug {
            isDebuggable = true
            applicationIdSuffix = ".debug"
        }
        release {
            isMinifyEnabled = true
            signingConfig = signingConfigs.getByName("release")
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
        freeCompilerArgs += listOf(
            "-opt-in=androidx.compose.material3.ExperimentalMaterial3Api"
        )
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    packaging {
        resources {
            excludes += "/META-INF/{AL2.0,LGPL2.1}"
        }
    }

    lint {
        // NonNullableMutableLiveDataDetector crashes with IncompatibleClassChangeError
        // on KaCallableMemberCall under newer Kotlin Analysis API. Project doesn't use
        // MutableLiveData (StateFlow + Hilt only), so disabling is safe.
        disable += "NullSafeMutableLiveData"
    }
}

dependencies {
    // Core Android
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.lifecycle.runtime.compose)
    implementation(libs.androidx.lifecycle.viewmodel.compose)
    implementation(libs.androidx.activity.compose)

    // Compose
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.ui)
    implementation(libs.androidx.ui.graphics)
    implementation(libs.androidx.ui.tooling.preview)
    implementation(libs.androidx.material3)
    implementation(libs.androidx.material.icons.extended)
    debugImplementation(libs.androidx.ui.tooling)
    debugImplementation(libs.androidx.ui.test.manifest)

    // Navigation
    implementation(libs.androidx.navigation.compose)

    // TV Compose
    implementation(libs.androidx.tv.foundation)
    implementation(libs.androidx.tv.material)

    // Hilt
    implementation(libs.hilt.android)
    ksp(libs.hilt.android.compiler)
    implementation(libs.hilt.navigation.compose)

    // Network
    implementation(libs.retrofit)
    implementation(libs.okhttp)
    implementation(libs.retrofit.kotlinx.serialization.converter)
    implementation(libs.okhttp.logging)
    implementation(libs.kotlinx.serialization.json)
    implementation(libs.kotlinx.coroutines.core)
    implementation(libs.kotlinx.coroutines.android)

    // Image loading
    implementation(libs.coil.compose)
    implementation(libs.blurhash)

    // Media3 ExoPlayer
    implementation(libs.media3.exoplayer)
    implementation(libs.media3.exoplayer.hls)
    implementation(libs.media3.ui)
    implementation(libs.media3.session)

    // DataStore
    implementation(libs.datastore.preferences)

    // Logging
    implementation(libs.timber)

    // WebView for YouTube
    implementation(libs.webkit)

    // Google Cast
    implementation(libs.cast.framework)

    // Testing
    testImplementation(libs.junit)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
    androidTestImplementation(platform(libs.androidx.compose.bom))
    androidTestImplementation(libs.androidx.ui.test.junit4)
    androidTestImplementation(libs.androidx.runner)
}

// ── Detekt ──────────────────────────────────────────────────────────────
detekt {
    config.setFrom(files("${rootProject.projectDir}/config/detekt/detekt.yml"))
    baseline = file("${rootProject.projectDir}/config/detekt/baseline.xml")
    buildUponDefaultConfig = true
    parallel = true
}

// ── ktlint ──────────────────────────────────────────────────────────────
ktlint {
    android.set(true)
    outputToConsole.set(true)
    ignoreFailures.set(false)
    filter {
        exclude("**/generated/**")
    }
}
