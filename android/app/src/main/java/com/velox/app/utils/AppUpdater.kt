package com.velox.app.utils

import android.app.DownloadManager
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.net.Uri
import android.os.Environment
import android.widget.Toast
import androidx.core.content.ContextCompat

object AppUpdater {
    fun downloadAndInstall(context: Context, url: String) {
        val request = DownloadManager.Request(Uri.parse(url)).apply {
            setTitle("Velox Update")
            setDescription("Downloading newest version...")
            setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED)
            // Use PUBLIC download dir, otherwise the Package Installer cannot read it without FileProvider
            setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, "velox-update.apk")
            setMimeType("application/vnd.android.package-archive")
        }

        val downloadManager = context.getSystemService(Context.DOWNLOAD_SERVICE) as DownloadManager
        val downloadId = downloadManager.enqueue(request)

        val receiver = object : BroadcastReceiver() {
            override fun onReceive(c: Context, intent: Intent) {
                val id = intent.getLongExtra(DownloadManager.EXTRA_DOWNLOAD_ID, -1)
                if (id == downloadId) {
                    val uri = downloadManager.getUriForDownloadedFile(downloadId)
                    if (uri != null) {
                        try {
                            val installIntent = Intent(Intent.ACTION_VIEW).apply {
                                setDataAndType(uri, "application/vnd.android.package-archive")
                                addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_GRANT_READ_URI_PERMISSION)
                            }
                            c.startActivity(installIntent)
                        } catch (e: Exception) {
                            Toast.makeText(c, "Error installing update: ${e.message}", Toast.LENGTH_LONG).show()
                        }
                    } else {
                        Toast.makeText(c, "Download failed or file not found", Toast.LENGTH_SHORT).show()
                    }
                    c.unregisterReceiver(this)
                }
            }
        }

        ContextCompat.registerReceiver(
            context,
            receiver,
            IntentFilter(DownloadManager.ACTION_DOWNLOAD_COMPLETE),
            ContextCompat.RECEIVER_EXPORTED
        )
        
        Toast.makeText(context, "Update downloading in background...", Toast.LENGTH_SHORT).show()
    }
}
